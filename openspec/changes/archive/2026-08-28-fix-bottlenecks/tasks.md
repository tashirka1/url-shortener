## 1. RPS benchmark fix

- [x] 1.1 Исправить SQL в `internal/rps/storage/storage.go:90` на `UPDATE rps_log SET duration = duration + 1 WHERE id = 1`
- [x] 1.2 Обновить `internal/rps/handler/handler.go:72` — логировать ошибку и возвращать `echo.NewHTTPError(500)` вместо `return err`
- [x] 1.3 Проверить `make check` и ручной `GET /rps/templ-page-update` → 200

## 2. UpdateAlias race

- [x] 2.1 Переписать `internal/link/storage/link.go:158` — убрать `SELECT COUNT(*)`, один `UPDATE`, маппинг `sqlite3.ErrConstraintUnique` → `model.ErrAliasTaken`
- [x] 2.2 Удалить неиспользуемый импорт если появится, `go fmt`
- [x] 2.3 Добавить конкурентный тест в `internal/link/storage/link_test.go` (10 горутин на один alias → 1 успех)

## 3. Redirect User-Agent

- [x] 3.1 В `internal/link/handler/handler.go:122` добавить `ua := c.Request().UserAgent()` с обрезкой 512 и передать в `GetAndClick`
- [x] 3.2 Обновить `internal/link/service/link_test.go` мок для `GetAndClick` с проверкой `userAgent` аргумента

## 4. Session cookie Secure:false и ClearSession error

- [x] 4.1 В `internal/core/session/session.go` поменять `SetUserId` — `Secure:false` всегда, вынести единые `Options`
- [x] 4.2 Поменять сигнатуру `ClearSession(c echo.Context) error`, унифицировать `Options{Path:"/", HttpOnly:true, SameSite:Lax, Secure:false, MaxAge:-1}`, возвращать `sess.Save` error
- [x] 4.3 Обновить `internal/auth/handler/user.go:88` `Logout` — обработать error от `ClearSession`, вернуть 500 при ошибке
- [x] 4.4 Обновить `internal/core/session/session.go:18` `AuthMiddleware` — обработать `Save` error при очистке инвалидной сессии
- [x] 4.5 Поправить `internal/auth/handler/user_test.go` и `internal/core/session` тесты под новую сигнатуру

## 5. Верификация

- [x] 5.1 `make check` (fmt + vet + `go test -tags fts5 ./...`) — зелёный
- [x] 5.2 Smoke: `POST /auth/login` → Set-Cookie без `Secure`, `GET /auth/logout` → 303, `GET /<code>` с кастомным UA → запись в `link_click`
