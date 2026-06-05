# AGENTS.md — URL Shortener

## Быстрый старт

```bash
make run                    # сборка (templ codegen + go build)
./bin/http                  # запуск на :8000
```

Dev-сервер с hot-reload (через `.air.toml`, polling):
```bash
go tool air
```

Docker:
```bash
make build   # сборка образов
make up      # docker compose up -d
make down    # docker compose down
```

## Архитектура

```
cmd/http/main.go   — точка входа (Echo + сессии + хендлеры)
cmd/worker/main.go — заглушка (печатает "Hello"), игнорировать
internal/
  core/       — общее: БД, миграции, middleware сессий, BaseView
  auth/       — handler → service → storage → SQLite (bcrypt cost 12)
  link/       — handler → service → repository → SQLite
static/       — PicoCSS, HTMX 2.0, main.js, пустой main.css
```

- Аутентификация: только сессии (Gorilla Sessions), без JWT. `core.AuthMiddleware` защищает `/link/*`.
- `link.Service` — pass-through к репозиторию, бизнес-логики пока нет.
- Короткие коды: 7-символьный base62 (`[0-9a-zA-Z]`) через `crypto/rand`.
- Пагинация: cursor-based, 5 на страницу (`WHERE id < ? ORDER BY id DESC LIMIT 5`), infinite scroll через `hx-trigger="revealed"`.
- SQLite в режиме WAL. БД в `db/main.db`.

## Ключевые команды

| Команда | Назначение |
|---|---|
| `make run` | Сборка + генерация templ |
| `make build` | Сборка Docker-образов |
| `make up` | Запуск Docker-контейнеров |
| `make down` | Остановка Docker-контейнеров |
| `go tool templ generate` | Генерация `*_templ.go` из `.templ` |
| `go test ./internal/...` | Запуск всех тестов |
| `air` (через `.air.toml`) | Dev-сервер с автоперезагрузкой |

## Переменные окружения

```
SESSION_KEY=dev_key
DB_NAME=db/main.db
```

Загружаются через `godotenv` в `main.go`. Обе обязательны.
