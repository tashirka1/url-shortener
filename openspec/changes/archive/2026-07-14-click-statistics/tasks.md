## 1. Data Layer (goose + model + storage)

- [x] 1.1 Создать goose-миграцию `link_click` с индексом по `(link_id, clicked_at)`
- [x] 1.2 Добавить `LinkClick` модель в `internal/link/model/click.go`
- [x] 1.3 Обновить `LinkStorage` — `GetAndClick` возвращает `link_id` (или всю модель `Link`)
- [x] 1.4 Добавить `ClickStorage` в `internal/link/storage/click.go` с методом `RecordClick`
- [x] 1.5 Добавить `ClickStorage` методы для статистики: `GetDailyClicks`, `GetTopReferrers`, `GetLinkOwner`

## 2. Business Logic (service)

- [x] 2.1 Обновить `link/service.GetAndClick` — добавить параметры `referrer`, `userAgent`; вызывать `clickStorage.RecordClick` в той же транзакции
- [x] 2.2 Добавить `ClickService` в `internal/link/service/click.go` — методы для получения статистики
- [x] 2.3 Написать table-driven тесты для обновлённого `GetAndClick` и `ClickService`

## 3. Интерфейс (views + handler)

- [x] 3.1 Создать `internal/link/view/stats.templ` — SVG bar chart кликов по дням + таблица реферреров
- [x] 3.2 Добавить ссылку «Статистика» в `internal/link/view/link.templ` в строку таблицы
- [x] 3.3 Создать `internal/link/handler/stats.go` — `GET /link/:code/stats` с проверкой владельца
- [x] 3.4 Зарегистрировать новый handler в `SetupHandlers`
- [x] 3.5 Обновить `cmd/url-shortener/main.go` — передать `ClickStorage` в `link_service.NewLink` и `link_service.NewClick`

## 4. Проверка

- [x] 4.1 Запустить `make check` и исправить ошибки
- [x] 4.2 Проверить руками: создать ссылку → перейти → открыть статистику
