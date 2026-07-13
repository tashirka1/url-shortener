# url-shortener — Project Overview

## Domain
Self-hosted URL shortener. Users register, create short links, share them, and track click counts. Designed for maximum autonomy — single binary, no external network dependencies at runtime.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26.3 (idiomatic, `database/sql`, no ORM) |
| HTTP framework | Echo v4 (`github.com/labstack/echo/v4`) |
| Frontend | htmx 2.0.10 (AJAX/partials) + templ 0.3 (`github.com/a-h/templ`, Go HTML components) |
| CSS | PicoCSS v2 (semantic HTML, minimal classes) |
| Database | SQLite3 via `github.com/mattn/go-sqlite3` (CGO, `-tags fts5`) |
| Migrations | goose v3 (`github.com/pressly/goose/v3`, embedded via `//go:embed`) |
| Sessions | Gorilla sessions (`github.com/gorilla/sessions`, cookie store) |
| Auth | bcrypt via `golang.org/x/crypto` |
| Testing | testify (`github.com/stretchr/testify`, table-driven tests, struct mocks) |
| Linter | golangci-lint v2 with strict rules (no panic, no reflect, no init, no `any`) |
| Build | templ generate + go build → single binary in `bin/` |
| Docker | Multi-stage: `ghcr.io/a-h/templ:latest` → `golang:1.26-alpine` → `alpine:3.23` |
| Live reload | air (`github.com/air-verse/air`) |

## Module Path
`url_shortener` — `cmd/url-shortener/main.go` (not `cmd/api/` as noted in AGENTS.md)

## Build Commands
- `make build-bin` — `go tool templ generate && go build -tags fts5 -ldflags="-s -w" -o bin/url-shortener cmd/url-shortener/main.go`
- `make check` — `golangci-lint run ./... && go test -tags fts5 ./...`
- `make air` — live-reload with air
- `make up/down` — Docker Compose orchestration

## Directory Structure
```
.
├── cmd/url-shortener/main.go    # Entry point, DI wiring, middleware, routes
├── migrations/                   # goose SQL migrations (YYYYMMDDHHMMSS_name.sql)
│   ├── 20260604192121_create_tables.sql
│   ├── 20260623100000_create_fts_index.sql
│   ├── 20260703100000_create_rps_tables.sql
│   └── 20260709120726_seed_data.sql
├── internal/
│   ├── core/                    # Infrastructure (no business logic, no imports from other modules)
│   │   ├── base62/              # crypto/rand 7-char code generation
│   │   ├── db/                  # SQLite init, pragmas, goose runner
│   │   ├── health/              # GET /health endpoint
│   │   ├── session/             # AuthMiddleware, GetUserId, SetUserId, ClearSession
│   │   └── view/                # Shared templ layout (Base shell, Nav, Unauthorized)
│   ├── auth/                    # User authentication module
│   │   ├── model/               # User struct, sentinel errors
│   │   ├── storage/             # UserStorage interface + SQL implementation
│   │   ├── service/             # UserService interface + business logic (bcrypt)
│   │   ├── handler/             # Echo handlers (form validation, render)
│   │   └── view/                # Login/Register templ components
│   ├── link/                    # URL shortening module
│   │   ├── model/               # Link struct, sentinel errors, MaxURLLength
│   │   ├── storage/             # LinkStorage interface + SQL (FTS5 search, cursor pagination, click increment)
│   │   ├── service/             # LinkService interface + business logic
│   │   ├── handler/             # Echo handlers (CRUD, search, redirect, main page)
│   │   └── view/                # CreateLink/ListLink/SearchResults/Main templ components
│   └── rps/                     # RPS benchmarking module (separate, no interface/service split)
│       ├── model/               # JoinRow struct
│       ├── storage/             # SQL operations (insert, join, update)
│       ├── handler/             # Echo handlers for benchmark endpoints
│       └── view/                # Minimal templ pages
├── static/                      # Embedded static assets
│   ├── css/ (pico@2.min.css, main.css)
│   ├── js/  (htmx.org@2.0.10.min.js, main.js)
│   └── icon/ (favicon.ico/.svg/.png)
├── static.go                    # //go:embed static
├── migrations.go                # //go:embed migrations
├── openspec/                    # OpenSpec framework (config.yaml, specs/, changes/)
├── .golangci.yml                # Linter config (forbidigo: panic, reflect, init)
└── .env                         # SESSION_KEY, DB_NAME
```

## Database Schema

### `auth_user`
| Column | Type | Constraints |
|--------|------|------------|
| id | INTEGER | PRIMARY KEY |
| email | TEXT | UNIQUE |
| password | TEXT | bcrypt hash |

### `link_link`
| Column | Type | Constraints |
|--------|------|------------|
| id | INTEGER | PRIMARY KEY |
| code | TEXT | UNIQUE, 7-char base62 |
| url | TEXT | max 2048 |
| clicks | INTEGER | default 0 |
| created_at | DATETIME | default CURRENT_TIMESTAMP |
| user_id | INTEGER | FK → auth_user(id) |

Indexes: `(user_id, id DESC)`, UNIQUE `(user_id, url)`

### `link_fts` (FTS5 virtual table)
Columns: `code, url` — content-synced via triggers (`link_ai`, `link_ad`, `link_au`). Tokenizer: `unicode61`. Ranked by `bm25`.

### `rps_log`
| Column | Type |
|--------|------|
| id | INTEGER PRIMARY KEY AUTOINCREMENT |
| payload | TEXT NOT NULL |
| ts | INTEGER NOT NULL |
| duration | INTEGER NOT NULL DEFAULT 0 |

### `rps_meta`
| Column | Type | Constraints |
|--------|------|------------|
| id | INTEGER PRIMARY KEY AUTOINCREMENT |
| log_id | INTEGER NOT NULL | FK → rps_log(id) ON DELETE CASCADE |
| key | TEXT NOT NULL |
| value | TEXT NOT NULL |

## Architecture Rules

### Module isolation
- Modules in `internal/` never import each other directly.
- Communication is strictly via interfaces wired in `cmd/url-shortener/main.go`.
- `internal/core/` imports no other internal packages.

### Per-module 5-directory layout
```
module/
├── model/       # DB entities + DTOs. Zero external imports. Concrete types.
├── storage/     # Interface XStorage + SQL implementation. Pure database/sql.
├── service/     # Interface XService + business logic. Depends only on storage interfaces.
├── handler/     # Echo handlers. Calls service via interface. Renders templ components.
└── view/        # .templ files with PicoCSS + htmx attributes.
```
Exception: `rps/` has no `service/` layer (handler → storage directly).

### Data flow
htmx click/form → handler (parse + validate) → service (business logic via interface) → storage (SQL via interface) → service → handler (render templ fragment) → htmx updates DOM

### Hard constraints
- No ORM (`database/sql` only)
- No global variables (DI via structs)
- No `reflect`, no `any`/`interface{}`, no `init()`, no `panic`
- No inheritance (composition only)
- Max 3 levels of nesting (early return)
- Max 2 return values: `(result, error)`
- `ctx context.Context` is first arg in all service and storage functions
- `(*T, error)` with `nil` = "not found" (not an error)
- SQL: column-specific SELECT (no `SELECT *`), parameterized queries only (`?` placeholders)

### Database/sql safety
- `defer rows.Close()` immediately after error check on `QueryContext`/`QueryRowContext`
- Always check `rows.Err()` after `rows.Next()` loop
- Transactions: `BeginTx` + `defer tx.Rollback()` + `tx.Commit()` on success
- No string concatenation in SQL — always `?` placeholders

## Routing

| Method | Path | Handler | Auth | Module |
|--------|------|---------|------|--------|
| GET | `/` | link.Main | No | link |
| GET | `/health` | health.Handler | No | core |
| GET | `/auth/login` | auth.GetLogin | No | auth |
| POST | `/auth/login` | auth.PostLogin | No | auth |
| GET | `/auth/logout` | auth.Logout | No | auth |
| GET | `/auth/register` | auth.GetRegister | No | auth |
| POST | `/auth/register` | auth.PostRegister | No | auth |
| GET | `/link/create-link` | link.GetCreateLink | Yes | link |
| POST | `/link/create-link` | link.PostCreateLink | Yes | link |
| GET | `/link/list-link` | link.ListLink | Yes | link |
| GET | `/link/search-link` | link.SearchLink | Yes | link |
| DELETE | `/link/remove-link/:code` | link.RemoveLink | Yes | link |
| GET | `/:code` | link.RedirectLink | No | link |
| GET | `/rps/simple-text` | rps.SimpleText | No | rps |
| GET | `/rps/simple-json` | rps.SimpleJSON | No | rps |
| GET | `/rps/simple-templ-page` | rps.SimpleTemplPage | No | rps |
| GET | `/rps/templ-page-insert` | rps.TemplPageInsert | No | rps |
| GET | `/rps/templ-page-select-join` | rps.TemplPageSelectJoin | No | rps |
| GET | `/rps/templ-page-select-join-update` | rps.TemplPageSelectJoinUpdate | No | rps |

Auth middleware (`session.AuthMiddleware`) is applied via `group.Use()` on `/link/*` routes.

## Middleware Stack (in order)
1. CrossOriginProtection (custom CSRF)
2. Request logger
3. Session middleware (Gorilla via echo-contrib)
4. Context timeout (10s)
5. Rate limiter (3 req/s on `POST /link/create-link` only)

## Configuration (`.env`)
- `SESSION_KEY` — required, string for cookie signing
- `DB_NAME` — required, path to SQLite file (e.g. `./db/main.db`)

## Seed Data
- User: `test@test.ru` / `Test10293847` (bcrypt hash)
- 5 sample links: x.com, vk.com, youtube.com, instagram.com, facebook.com

## Testing Convention
- `_test.go` files next to tested file
- Table-driven tests (testify)
- Lightweight struct mocks for storage interfaces (no generators)
- Tags: `-tags fts5` (required for FTS5 support)
- Command: `go test -tags fts5 ./...`

## Existing Modules Summary

### auth — User authentication
- Register with email/password validation (≥8 chars, ≤72 chars, valid email format, duplicate check)
- Login with bcrypt verification
- Logout (clear session)
- Session: 7-day cookie, HttpOnly, Secure, SameSite=Lax

### link — URL shortening CRUD
- Create: generates 7-char crypto-random code, ON CONFLICT DO NOTHING for duplicate URLs per user
- List: cursor-based pagination (5 per page, infinite scroll via `hx-trigger="revealed"`)
- Search: FTS5 full-text via `bm25` ranking, trigger-synced content table
- Delete: scoped by `user_id`, returns 200 even if not found
- Redirect: lookup + transaction-based click counter increment → 303 redirect

### rps — Requests Per Second benchmarking
- 6 endpoints for load-testing: plain text, JSON, templ render, insert, select+join, select+join+update
- No `service/` layer (storage → handler directly)
- Not for production use

## Not Yet Implemented
- Payments/billing module (`internal/yookassa/` is mentioned in AGENTS.md but does not exist)
- Custom aliases, link expiry, QR codes, click analytics, API tokens, password-protected links
