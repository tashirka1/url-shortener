## 1. База данных и слой данных

- [x] 1.1 Создать goose-миграцию для таблицы `api_token` (`id, user_id, name, token_hash, prefix, created_at, last_used_at, revoked_at`) с индексами на `(token_hash)` и `(user_id, created_at DESC)`
- [x] 1.2 Создать `internal/apitoken/model/token.go` — структура `Token` и sentinel-ошибки (`ErrTokenNotFound`, `ErrTokenRevoked`)
- [x] 1.3 Создать `internal/apitoken/storage/token.go` — интерфейс `TokenStorage` и SQLite-реализация: методы `Generate`, `ListByUser`, `Revoke`, `FindByHash`

## 2. Бизнес-логика

- [x] 1.4 Создать `internal/apitoken/service/token.go` — интерфейс `TokenService` и реализация: `Generate` (crypto/rand + SHA-256 + base62), `ListByUser`, `Revoke`, `Authenticate` (поиск по хешу, проверка revoked, обновление last_used_at)
- [x] 1.5 Создать `internal/core/api/middleware.go` — middleware, читающая заголовок `Authorization: Bearer`, вызывающая `TokenService.Authenticate` и устанавливающая ID пользователя в контексте Echo (`c.Set("userId", id)`)
- [x] 1.6 Создать `internal/link/handler/api.go` — JSON API-обработчики: `PostCreateLink`, `ListLink`, `DeleteLink`, `GetStats` (переиспользуя существующие сервисы через `c.Get("userId")`)

## 3. Пользовательский интерфейс

- [x] 1.7 Создать `internal/apitoken/view/token.templ` — PicoCSS-компоненты: таблица токенов с кнопкой отзыва, форма создания токена, отображение созданного токена с кнопкой копирования
- [x] 1.8 Создать `internal/apitoken/handler/token.go` — Echo-обработчики: `Index` (список + форма), `Generate` (POST), `Revoke` (DELETE), с htmx-фрагментами
- [x] 1.9 Сгенерировать templ-код: `templ generate`

## 4. Интеграция

- [x] 1.10 Подключить новый модуль в `cmd/url-shortener/main.go`: инициализация storage/service, регистрация UI-маршрутов на `/link/tokens`, регистрация API-маршрутов на `/api/v1/` с API-middleware
- [x] 1.11 Добавить API-эндпоинт статистики кликов (`GET /api/v1/link/:code/stats`) — переиспользовать существующий `clickService` через тонкую JSON-обёртку

## 5. Тестирование

- [x] 1.12 Написать table-driven тесты для `apitoken/service`: создание токена, список, отзыв, авторизация с валидным/невалидным/отозванным токеном
- [x] 1.13 Написать table-driven тесты для `apitoken/storage`: CRUD-операции с SQLite in-memory
- [x] 1.14 Запустить `make check` и исправить все ошибки
