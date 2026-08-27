# url-shortener

Self-hosted сервис сокращения ссылок. Один бинарник, SQLite, без внешних зависимостей — работает офлайн на минимальном железе.

## Возможности

- Короткие ссылки (base62), редирект, счётчик кликов, QR-код
- Поиск по FTS5, пагинация
- Аутентификация (cookie-сессии, CSRF, rate-limit), API-токены `Bearer` для `/api/v1`
- Server-rendered UI: `templ` + `htmx` + `PicoCSS` — без SPA
- Health-check `/health`, готов к бэкапу через Litestream

## Стек

Go 1.26 + Echo v4 · templ + htmx + PicoCSS · SQLite3 (`mattn/go-sqlite3`, `fts5`) + goose · Docker

## Быстрый старт

```bash
cp env-example .env
make dev          # разработка с hot-reload (air), :8000
```

```bash
make up           # prod в Docker, :8001 -> :8000
make down         # остановить
```

```bash
# prod-бинарник
sudo apt-get install -y build-essential libsqlite3-dev
make build-bin    # -> bin/url-shortener
./bin/url-shortener
```

Переменные окружения (`.env`):

```
SESSION_KEY=dev_key
DB_NAME=./db/main.db
```

## Команды

```bash
make check      # fmt + vet/lint + тесты (запускай после каждого изменения)
make lint       # только линтеры
```

## Бенчмарки (RPS)

Эндпоинты для нагрузочного тестирования — `internal/rps`.

Подготовка данных для SELECT/UPDATE:

```bash
for i in $(seq 1 1000); do curl -s "http://localhost:8000/rps/templ-page-insert?payload=prefill" > /dev/null; done
```

Запуск (`wrk`):

```bash
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-text
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-json
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-templ-page
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-insert?payload=bench'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-join?limit=15'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-update'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-simple?limit=15'
```

## Документация

- [Туториал](docs/tutorial)
- Архитектура и правила: [AGENTS.md](AGENTS.md)

## Лицензия

MIT — см. [LICENSE](LICENSE).
