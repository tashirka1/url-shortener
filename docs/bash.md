# Bash: команды проекта

## Сборка и запуск

```bash
make run                    # templ generate + go build → ./bin/http
```

Makefile:

```makefile
run: build
    ./bin/http

build:
    go tool templ generate
    go build -o bin/http ./cmd/http

build-bin:
    go build -o bin/http ./cmd/http
```

## Dev-сервер с hot-reload

```bash
go tool air                 # через .air.toml
```

`.air.toml` следит за `.go`, `.templ`, `.html`, `.tpl`, `.tmpl`, игнорирует `*_templ.go` и `*_test.go`. При изменении — `templ generate` + `go build` + перезапуск.

## Docker

```bash
make build                  # docker compose build
make up                     # docker compose up -d  (порт 8001)
make down                   # docker compose down
```

## Работа с модулями

```bash
go mod tidy                 # подчистить зависимости
go mod vendor               # создать vendor/
```

## templ

```bash
go tool templ generate      # сгенерировать *_templ.go из .templ
go tool templ fmt           # отформатировать .templ файлы
```

## Тесты

```bash
go test ./internal/...               # все тесты
go test -v ./internal/auth/handler/  # verbose, конкретный пакет
go test -run TestUserService ./internal/auth/service/  # фильтр по имени
```

## Переменные окружения

Файл `.env`:

```
SESSION_KEY=dev_key
DB_NAME=db/main.db
```

Загружаются через `godotenv` в `main.go`, все обязательны.

## Структура проекта

```
cmd/http/main.go        — точка входа
cmd/worker/main.go      — заглушка
internal/
  core/                 — БД, сессии, base62, общие view
  auth/                 — аутентификация (handler→service→storage)
  link/                 — ссылки (handler→service→storage)
static/                 — PicoCSS, HTMX, main.js
migrations/             — SQL-миграции goose
db/main.db              — SQLite (WAL mode)
```
