## Context

Сейчас каждый переход по короткой ссылке инкрементит `link_link.clicks` в транзакции (`link/storage.GetAndClick`). Никакой информации о времени перехода, источнике (referrer) или браузере не сохраняется.

Добавляется таблица `link_click`, куда будет писаться каждый переход. Это позволит строить график кликов по дням и собирать топ реферреров.

Всё реализуется внутри модуля `internal/link/` — нового модуля не требуется.

## Goals / Non-Goals

**Goals:**
- Запись каждого перехода с referrer и user-agent в `link_click`
- График кликов по дням (inline SVG, без JS-библиотек)
- Таблица топ реферреров для каждой ссылки
- Кнопка/ссылка «Статистика» в списке ссылок
- `link_link.clicks` остаётся как денормализованный счётчик

**Non-Goals:**
- Геолокация по IP, анализ устройств, UTM-метки
- Экспорт статистики
- WebSocket / live-обновления
- Общая статистика по всем ссылкам пользователя (дашборд)

## Decisions

### Таблица link_click

```sql
CREATE TABLE link_click (
    id         INTEGER PRIMARY KEY,
    link_id    INTEGER NOT NULL REFERENCES link_link(id) ON DELETE CASCADE,
    referrer   TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    clicked_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX idx_link_click_link_id_clicked_at ON link_click(link_id, clicked_at);
```

- `clicked_at` в ISO8601 — удобно для GROUP BY date()
- `referrer` и `user_agent` как TEXT — достаточно для статистики
- Индекс по `(link_id, clicked_at)` — покрывает оба запроса (ежедневная статистика и реферреры по ссылке)

### link_link.clicks остаётся

Денормализованный счётчик быстрее для списка ссылок, чем `COUNT(*)` с join. Обновляется в той же транзакции, что и INSERT в link_click — атомарность гарантирована.

### GetAndClick расширяется

Сигнатура меняется с `(code string) (string, error)` на `(code, referrer, userAgent string) (string, error)`. В транзакции:
1. `SELECT url FROM link_link WHERE code=?`
2. `UPDATE link_link SET clicks=clicks+1 WHERE code=?`
3. `INSERT INTO link_click(link_id, referrer, user_agent) VALUES (?, ?, ?)`

link_id получаем через вложенный SELECT или в том же SELECT, что url.

### Статистика: inline SVG

График кликов по дням рисуется как SVG-элемент с колонками (bar chart). Каждая колонка — `<rect>` с высотой пропорциональной числу кликов. Всё рендерится в templ-компоненте, ноль JS.

Реферреры — `<table>` с колонками «Источник» и «Переходы».

### User-agent не пишется в редиректе (пока)

В `GET /:code` у нас нет доступа к User-Agent через echo request. Referrer доступен через `c.Request().Referer()`. UA можно добавить позже — он не критичен для первых двух фич (график + реферреры).

_Решение: пока пишем только referrer. UA опционален._

## Risks / Trade-offs

- **Риск**: Увеличение времени редиректа из-за INSERT в link_click
  → **Митигация**: SQLite в WAL-режиме, вставка в одной транзакции с апдейтом, индекс помогает избежать table scan. Для MVP допустимо.
- **Риск**: Рост таблицы link_click при большом числе переходов
  → **Митигация**: Нет автоочистки в MVP. Можно добавить в будущем (например, retention 90 дней через джобу).
- **Риск**: Изменение сигнатуры GetAndClick ломает существующий код
  → **Митигация**: единственный caller — handler.RedirectLink. Меняем оба места синхронно.
