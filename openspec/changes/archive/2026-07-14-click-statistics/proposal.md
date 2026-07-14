## Why

Сейчас ссылки хранят только общий счётчик кликов (`link_link.clicks`) без какой-либо истории. Невозможно увидеть динамику переходов по дням, топ реферреров или понять, какие каналы трафика работают. Нужна постраничная статистика кликов для каждой ссылки.

## What Changes

- Новая таблица `link_click` для записи каждого отдельного перехода (link_id, referrer, user_agent, clicked_at)
- `link_link.clicks` остаётся как денормализованный быстрый счётчик, обновляется в той же транзакции
- `GetAndClick` в сервисе ссылок принимает referrer и user-agent, записывает клик в `link_click`
- Новая страница статистики для ссылки: график кликов по дням (inline SVG) + топ реферреров
- В таблице списка ссылок добавляется кнопка/ссылка «Статистика» для каждой ссылки
- Модуль `internal/link/` расширяется: новые model, storage, handler, view для кликов и статистики

## Capabilities

### New Capabilities

- `link-click-stats`: учёт каждого перехода по ссылке (link_click), отображение дневной статистики и реферреров

### Modified Capabilities

_Нет. Модуль ссылок не имеет существующих specs._

## Impact

- Новая goose-миграция: `link_click` + индекс
- Расширение `internal/link/model/click.go`
- Новый `internal/link/storage/click.go` — хранение и чтение кликов
- `link/service.GetAndClick` меняет сигнатуру: принимает referrer + userAgent
- Новый `internal/link/handler/stats.go` — эндпоинты статистики
- Новый `internal/link/view/stats.templ` — HTML-фрагменты с графиком и таблицей
- `internal/link/view/link.templ` — добавляется ссылка «Статистика» в строку таблицы
- `cmd/url-shortener/main.go` — без изменений (всё внутри модуля link)
