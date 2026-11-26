# url-shortener

Self-hosted URL shortener. A single binary, SQLite, no external dependencies — runs offline on minimal hardware.

## Features

- Short links (base62), redirects, click counter, QR codes
- FTS5 full-text search, pagination
- Authentication (cookie sessions, CSRF, rate limiting), `Bearer` API tokens for `/api/v1`
- Server-rendered UI: `templ` + `htmx` + `PicoCSS` — no SPA
- Health check at `/health`, ready for Litestream backups

## Stack

Go 1.26 + Echo v4 · templ + htmx + PicoCSS · SQLite3 (`mattn/go-sqlite3`, `fts5`) + goose · Docker

## Quick start

```bash
cp env-example .env
make dev          # development with hot-reload (air), :8000
```

```bash
make up           # prod in Docker, :8001 -> :8000
make down         # stop
```

```bash
# production binary
sudo apt-get install -y build-essential libsqlite3-dev
make build-bin    # -> bin/url-shortener
./bin/url-shortener
```

Environment variables (`.env`):

```
SESSION_KEY=dev_key
DB_NAME=./db/main.db
```

## Commands

```bash
make check      # fmt + vet/lint + tests (run after every change)
make lint       # linters only
```

## Benchmarks (RPS)

Load-testing endpoints — `internal/rps`.

Prefill data for SELECT/UPDATE:

```bash
for i in $(seq 1 1000); do curl -s "http://localhost:8000/rps/templ-page-insert?payload=prefill" > /dev/null; done
```

Run (`wrk`):

```bash
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-text
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-json
wrk -t10 -c100 -d5s http://localhost:8000/rps/simple-templ-page
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-insert?payload=bench'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-join?limit=15'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-update'
wrk -t10 -c100 -d5s 'http://localhost:8000/rps/templ-page-select-simple?limit=15'
```

## Documentation

- [Tutorial](docs/tutorial)
- Architecture and rules: [AGENTS.md](AGENTS.md)

## License

MIT — see [LICENSE](LICENSE).
