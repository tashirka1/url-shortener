## Why

Проект содержит критичные баги и узкие места, блокирующие корректность и наблюдаемость: сломанный бенчмарк RPS (`UPDATE ... WHERE ... SET` синтаксис), TOCTOU race в `UpdateAlias`, потеря `User-Agent` в редиректе, небезопасная/нерабочая сессионная cookie (`Secure:true` на http) и игнорирование ошибок `ClearSession`. Фикс нужен сейчас — до расширения функционала, чтобы не размножать дефекты.

## What Changes

- Исправить SQL в `rps/storage.Update`: `UPDATE rps_log SET duration=duration+1 WHERE id=1` и вернуть корректный HTTP статус в handler.
- Устранить race в `link/storage.UpdateAlias`: убрать `SELECT COUNT(*)`, полагаться на `UNIQUE(code)` и парсить `sqlite3.ErrConstraintUnique` → `ErrAliasTaken`; оставить один `UPDATE`.
- Исправить редирект `link/handler.RedirectLink`: прокидывать `c.Request().UserAgent()` (с обрезкой до 512) в `GetAndClick` вместо `""`, чтобы `link_click.user_agent` заполнялся.
- Сделать `Secure:false` всегда в `core/session` (`SetUserId`, `ClearSession`) и унифицировать `Options{Path, HttpOnly, SameSite}` для set/clear.
- Поменять `ClearSession` на `(c echo.Context) error`, обрабатывать ошибку в `auth/handler.Logout` и `core/session.AuthMiddleware`.
- Не трогать: `core/db` pool/mmap (п.5) и `ContextTimeout` middleware (п.10) — по требованию.

## Capabilities

### New Capabilities
- `session-cookie-handling`: требования к установке/удалению сессионной cookie (Secure, path, ошибки).

### Modified Capabilities
- `rps-benchmark`: исправление контракта `GET /rps/templ-page-update` (должен возвращать 200 с HTML).
- `custom-aliases`: уточнение поведения `UpdateAlias` при конкурентной записи (один запрос успешен, остальные `409`/ошибка "alias already taken").
- `link-click-stats`: требование сохранять `User-Agent` при редиректе (ранее терялся).

## Impact

- Затронуты: `internal/rps/storage`, `internal/rps/handler`, `internal/link/storage`, `internal/link/handler`, `internal/core/session`, `internal/auth/handler`.
- Миграций нет, breaking changes нет (поведение только исправляется к уже задокументированному в спеках).
- Риск: изменение сигнатуры `ClearSession` — требует обновления всех call sites (только `Logout` и `AuthMiddleware`).
