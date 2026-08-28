## Context

Система одного бинаря на SQLite WAL. Текущие дефекты не требуют миграций, но затрагивают 3 модуля: `rps`, `link`, `core/session+auth`. Пользователь запретил трогать `core/db` pool/mmap и `ContextTimeout` middleware — дизайн учитывает это.

## Goals / Non-Goals

**Goals:**
- Починить RPS бенчмарк (`UPDATE` синтаксис) и привести handler к конвенциям проекта (логирование, `HTTPError`).
- Устранить TOCTOU в `UpdateAlias` одним атомарным `UPDATE` + обработкой `UNIQUE`.
- Сохранять `User-Agent` при редиректе (обрезка 512).
- Унифицировать cookie `Secure:false` и вернуть ошибки `ClearSession` наружу.

**Non-Goals:**
- Не менять `internal/core/db/db.go` (pool, pragmas).
- Не менять `cmd/url-shortener/main.go` middleware `ContextTimeout`.
- Не добавлять новые таблицы/индексы, не вводить кэш/очередь.

## Decisions

1. **RPS SQL фикс — прямой `ExecContext` без транзакции.** Альтернатива `Prepare` избыточна для одноразового `WHERE id=1`. Выбран самый простой `UPDATE SET WHERE`.

2. **UpdateAlias — полагаться на `UNIQUE(code)` (`migrations/20260604192121`).** Альтернатива `SELECT ... FOR UPDATE` недоступна в SQLite без `BEGIN IMMEDIATE`; `INSERT ... ON CONFLICT` не подходит (нужен `UPDATE`). Обработка `sqlite3.Error{ExtendedCode: ErrConstraintUnique}` → `model.ErrAliasTaken` — идиоматично, уже используется в `auth/storage/user.go:51`.

3. **User-Agent обрезка 512 байт в handler, не в storage.** Handler — граница валидации (аналогично `validateURL`), storage остается чистым SQL. 512 — баланс между аналитикой и bloat (средний UA ~150, max 4KB).

4. **`Secure:false` всегда.** Альтернатива `isSecureRequest(c)` отклонена по требованию пользователя — проще для офлайн-бинаря и `make dev` на http. Продакшн за TLS-терминатором все равно требует `Secure:true`, но пользователь явно выбрал `false` — фиксируем в спеках.

5. **`ClearSession() error` с едиными `Options`.** `SetUserId` и `ClearSession` должны использовать одинаковый `Path="/"` и `SameSite`, иначе удаление cookie не сработает в браузере (RFC 6265). Возврат `error` позволяет `Logout` вернуть 500 вместо молчаливого 303.

## Risks / Trade-offs

- **Secure:false → cookie по http.** Риск перехвата в незащищенной сети → Mitigation: задокументировать в `design.md` и рекомендовать reverse-proxy с TLS; в коде оставить комментарий.
- **ErrAliasTaken из sqlite ошибки может замаскировать другой constraint.** Mitigation: проверять только `ErrConstraintUnique`, логировать исходную ошибку с `code`.
- **UA truncation теряет данные ботов с длинным UA.** Trade-off: защита БД от bloat важнее; 512 покрывает 99% кейсов.
- **Сигнатура `ClearSession` breaking для тестов.** Mitigation: обновить единственный call site `auth/handler/user.go:88` и `core/session.AuthMiddleware:22`.

## Migration Plan

1. Применить код-изменения, `make check` (требует `go test -tags fts5`).
2. Smoke `curl /rps/templ-page-update`, `curl -L /<code>` с `User-Agent`, `PATCH /link/:code/alias` конкурентно (`ab -n 20 -c 10`).
3. Rollback — revert коммита, миграций нет.

## Open Questions

- Нужен ли `X-Forwarded-Proto` учет для `Secure` в будущем, если пользователь передумает? Сейчас — нет.
