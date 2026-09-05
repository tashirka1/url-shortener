# Обзор проекта URL Shortener

## Что это вообще такое?

Представь: у тебя есть длинная ссылка, например:
```
https://example.com/очень-длинная-страница/с-кучей-параметров
```

Ты хочешь дать кому-то короткую, красивую ссылку:
```
http://твой-сайт/abc1234
```

Этот проект — **сервис, который это делает**. Ты регистрируешься, заходишь, вставляешь длинную ссылку — получаешь короткую. Когда кто-то переходит по короткой ссылке — его перенаправляет на длинную. И ещё считает, сколько раз кликнули.

Вот 3 простых шага работы сервиса:

```
1. Зарегистрироваться
2. Вставить длинную ссылку → получить короткую
3. Раздать короткую ссылку друзьям
```

## Из чего состоит проект (анатомия)

Проект написан на **Go** (язык программирования). И устроен как **слоёный пирог** из трёх этажей:

```
┌──────────────────────────────────────┐
│  1. Handler (ручка) — принимает       │
│     запросы от браузера               │
├──────────────────────────────────────┤
│  2. Service (сервис) — проверяет,     │
│     что всё правильно, логирует       │
├──────────────────────────────────────┤
│  3. Storage (хранилище) — сохраняет   │
│     в базу данных SQLite              │
└──────────────────────────────────────┘
```

Это называется **чистая архитектура** (clean architecture). Каждый этаж отвечает только за своё, не лезет в чужое.

## Главный вход — `cmd/url-shortener/main.go`

Это как **пульт управления**. Когда запускаешь программу, она:

1. Читает настройки из `.env` файла (`SESSION_KEY` обязателен, `PORT` по умолчанию `8000`, `DB_NAME` по умолчанию `./db/url-shortener.db` — см. `internal/core/config/config.go:18`)
2. Открывает базу данных SQLite (`internal/core/db/db.go:16`)
3. Запускает веб-сервер на порту `:$PORT` (по умолчанию `:8000`)
4. Ждёт, пока кто-то зайдёт на сайт

Представь, что это выключатель света — щёлкнул, и всё заработало.

## База данных — `internal/core/db/db.go`

База — это **толстая тетрадь**, куда всё записывается. Твой проект использует **SQLite** — один файлик `./db/url-shortener.db` (дефолт из `internal/core/config/config.go:39`, переопределяется `DB_NAME`).

Когда программа запускается, она:
- Включает режим WAL (можно читать и писать одновременно)
- Ставит таймаут, если база занята (ждёт до 10 секунд, `busy_timeout=10000`)
- Ставит прагмы через DSN (`internal/core/db/db.go:19`): `foreign_keys=ON`, `journal_mode=WAL`, `synchronous=NORMAL`, `temp_store=MEMORY`, `cache_size=-65536` (64MB), `auto_vacuum=INCREMENTAL`, `journal_size_limit=67110000`, `page_size=4096`
- Настраивает пул: `SetMaxOpenConns(64)` / `SetMaxIdleConns(64)` (`internal/core/db/db.go:38`)
- Применяет миграции через `goose` (`internal/core/db/db.go:51`)

## Миграции — `migrations/`

Миграция — это **инструкция** для базы данных: "создай такие-то таблицы". Всего 7 миграций:

**Таблица 1: `auth_user`** (`migrations/20260604192121_create_tables.sql:2`) — кто зарегистрирован
- `id` — номер пользователя
- `email` — почта
- `password` — пароль (зашифрованный bcrypt)

**Таблица 2: `link_link`** — все ссылки
- `id` — номер ссылки
- `code` — короткий код (типа `abc1234`)
- `url` — длинная ссылка
- `clicks` — сколько раз кликнули
- `created_at` — когда создали
- `user_id` — чья это ссылка
- Индексы: `idx_link_link_user_id_id (user_id, id DESC)` для пагинации, `idx_link_link_user_url (user_id, url)` UNIQUE для детекта дубликатов

**Таблица 3: `link_fts`** (`migrations/20260623100000_create_fts_index.sql:2`) — виртуальная FTS5-таблица для поиска
- `code`, `url`, `content='link_link'`, триггеры `link_ai/ad/au` для синхронизации

**Таблица 4: `link_click`** (`migrations/20260714100000_create_link_click.sql:2`) — детали каждого перехода
- `link_id`, `referrer`, `user_agent`, `clicked_at`

**Таблица 5: `api_token`** (`migrations/20260715100000_create_api_token.sql:2`) — токены для API
- `user_id`, `name`, `token_hash`, `prefix`, `created_at`, `last_used_at`, `revoked_at`

**Таблицы 6-7: `rps_log` / `rps_meta`** (`migrations/20260703100000_create_rps_tables.sql:2`) — логи RPS-теста

## Регистрация и вход — `internal/auth/`

Это как **проходная на заводе**. Чтобы пользоваться сервисом, нужно:

1. **Зарегистрироваться** (`GET /auth/register` → `POST /auth/register`) — ввести email и пароль
2. **Зайти** (`GET /auth/login` → `POST /auth/login`) — ввести email и пароль

Пароль хранится не как обычный текст, а шифруется через **bcrypt** — это надёжный способ, как сейф с 10 замками.

После входа тебе выдаётся **сессия** (как пропуск с фотографией). Пока ты не вышел, сервер знает, кто ты.

```
3 этажа регистрации:

Handler (принимает email и пароль)
    ↓
Service (шифрует пароль bcrypt)
    ↓
Storage (сохраняет в SQLite)
```

Сессии — `internal/core/session/session.go:18` (`AuthMiddleware`, `GetUserId`, cookie `userId` на 7 дней).

## Ссылки — `internal/link/`

Это **сердце проекта**. После входа ты можешь:

1. **Создать ссылку** (`POST /link/` или `POST /link`) — вводишь URL → получаешь короткий код
2. **Посмотреть список** (`GET /link/` — страница с формой + первые 5) и подгрузка `GET /link/list-link?cursor=...`
3. **Поиск** (`GET /link/search-link?q=...`) — FTS5 по `code`/`url` (см. `internal/link/storage/link.go:183`)
4. **Переименовать alias** (`GET /link/:code/edit-form` → `PATCH /link/:code/alias`) — кастомный код 3–32 символа `[a-zA-Z0-9_-]`
5. **QR-код** (`GET /link/:code/qr`)
6. **Статистика кликов** (`GET /link/:code/stats`)
7. **Удалить ссылку** (`DELETE /link/remove-link/:code`)
8. **Перейти** (`GET /:code`) — любой человек переходит → его перенаправляет + инкремент `clicks` и запись в `link_click` (`internal/link/storage/link.go:113`)

Короткий код — это **7 случайных символов** из цифр и букв (a-z, A-Z, 0-9). Всего 62 символа, поэтому называется **base62**. Генерируется через `crypto/rand` — как бросание кубика, только компьютерного (`internal/core/base62/base62.go:11`).

Вероятность совпадения — 62⁷ = 3.5 триллиона вариантов.

Дополнительно:
- **API токены** — `GET /link/tokens` (страница) + `POST /api/v1/...` через `internal/apitoken` (Bearer-токен)
- **RPS** — `internal/rps/handler` (тест нагрузки)

## Страницы (шаблоны) — `*.templ`

Страницы сайта рисуются с помощью **templ** — это как LEGO для HTML.

Есть **базовый шаблон** (`internal/core/view/view.templ:24` `Base`), который повторяется на всех страницах:
- Шапка сайта (`Nav` — Зарегистрироваться / Войти / Токены / Выйти)
- Подключены стили (PicoCSS `pico@2.min.css` + `main.css`) и HTMX 2.0.10
- `body hx-boost="true" hx-target="#main" hx-indicator="#top-loading-bar"` — все ссылки/формы через HTMX + индикатор загрузки
- CSP, favicon

HTMX — это штука, которая позволяет обновлять часть страницы без перезагрузки. Например, когда ты создаёшь ссылку, она просто добавляется в список, а не вся страница перерисовывается.

## Разделение по папкам

```
cmd/url-shortener/main.go   ← пульт (запуск сервера)
internal/
  core/
    config/config.go        ← ENV: SESSION_KEY, PORT, DB_NAME
    db/db.go                ← база данных + goose
    session/session.go      ← сессии (кто ты)
    base62/base62.go        ← генератор коротких кодов
    view/view.templ         ← базовый шаблон HTML
  auth/                     ← регистрация и вход
    handler/user.go         ← ручки (принимают запросы)
    service/user.go         ← сервис (шифрует пароль)
    storage/user.go         ← хранилище (SQLite)
    view/                   ← страницы (шаблоны)
  link/                     ← ссылки + FTS + клики + alias + QR
    handler/handler.go      ← ручки (+ API)
    service/link.go         ← сервис
    storage/link.go         ← хранилище
    storage/click.go        ← клики
    model/link.go           ← модель (структура данных)
    view/                   ← страницы (link.templ, stats.templ)
  apitoken/                 ← API-токены
  rps/                      ← RPS-тест
migrations/                 ← инструкции для базы (goose)
static/                     ← CSS, JS, картинки
litestream/                 ← репликация SQLite → S3
```

## Как запустить

1. Скопируй файл настроек:
```bash
cp env-example .env
# SESSION_KEY обязателен, DB_NAME по умолчанию ./db/url-shortener.db
```

2. Запусти через `air` (с автоперезагрузкой):
```bash
make dev
```

Или собери вручную:
```bash
make build-bin   # templ generate + go build -tags fts5 -o bin/url-shortener
./bin/url-shortener
```

3. Открой в браузере: `http://localhost:8000` (в Docker — `http://localhost:8001`)

## Полный цикл работы

```
Твой браузер:
┌───────────────────┐
│ 1. Заходишь на    │
│    localhost:8000  │
│    → видишь       │
│    "Register"     │
└────────┬──────────┘
         ↓
┌───────────────────┐
│ 2. Регистрация    │
│    email + пароль │
│    → жмёшь        │
│    "Register"     │
└────────┬──────────┘
         ↓
┌───────────────────┐
│ 3. Вход           │
│    email + пароль │
│    → жмёшь "Login"│
└────────┬──────────┘
         ↓
┌───────────────────┐
│ 4. Вставляешь     │
│    длинную ссылку │
│    → жмёшь "Create"│
└────────┬──────────┘
         ↓
┌───────────────────┐
│ 5. Получаешь      │
│    короткую ссылку│
│    (типа /Xk9m2aB)│
└────────┬──────────┘
         ↓
┌───────────────────┐
│ 6. Даёшь ссылку   │
│    друзьям        │
│    → они переходят│
│    → счётчик +++  │
└───────────────────┘
```

---

# Как работает шифрование паролей

Пароль хранится **не как текст**. Его превращают в "кашу" через **bcrypt** — это как мясорубка: обратно собрать нельзя.

**Как это выглядит в коде (`internal/auth/service/user.go`):**

### Когда регистрируешься (CreateUser):

```
Ты вводишь пароль "hello123"
          ↓
bcrypt берёт пароль + генерирует случайную соль (128 бит)
          ↓
Прокручивает 2^10 = 1024 раза (cost = 10)
          ↓
Получается строка типа:
$2a$10$СОЛЬ_22СИМВОЛА_БАЙТЫ_КАК_СТРОКА
          ↓
Сохраняется в базу
```

**Соль** — это случайные байты, встроенные прямо в строку хеша. Каждый раз разные. Даже если у двух человек одинаковый пароль — в базе будут разные строки.

### Когда заходишь (CheckUser):

```
Ты вводишь пароль "hello123"
          ↓
Находят твою строку в базе
          ↓
bcrypt сам достаёт соль из строки
          ↓
Смешивает соль с твоим паролем и прокручивает 1024 раза
          ↓
Если получился тот же хеш — пароль верный!
```

**Три хитрости bcrypt:**
1. **Встроенная соль** — соль хранится прямо в хеше, не нужно хранить её отдельно
2. **Cost-фактор** (10) — можно увеличить со временем, чтобы компенсировать рост мощности компьютеров
3. **Постоянное время сравнения** — bcrypt сам проверяет так, чтобы хакер не мог догадаться по времени ответа

---

# Как HTMX делает страницу быстрой

HTMX — это маленькая JavaScript-библиотека. Она позволяет **обновлять только часть страницы**, без перезагрузки всей.

В проекте используется **HTMX 2.0**. Он подключён в `internal/core/view/view.templ:50`:

```html
<script src="/static/js/htmx.org@2.0.10.min.js"></script>
```

И включён **режим глобальных анимаций** в `static/js/main.js:2`:

```javascript
htmx.config.globalViewTransitions = true;
```

### Пример 1: создание ссылки без перезагрузки

Смотри шаблон `internal/link/view/link.templ:71`:

```html
<form hx-post="/link/"
      hx-target="#list-links"
      hx-swap="afterbegin"
      hx-on:reset-create-form="this.reset()">
```

Перевод на человеческий:
- `hx-post="/link/"` — отправь форму НЕ через обычный POST, а через HTMX
- `hx-target="#list-links"` — результат вставь в элемент с id="list-links"
- `hx-swap="afterbegin"` — вставь новой строкой **в начало** списка
- `hx-on:reset-create-form="this.reset()"` — после успеха (`HX-Trigger: reset-create-form` из `internal/link/handler/handler.go:89`) очисти форму

Инлайн-ошибки валидации рендерятся через `HX-Retarget: #create-link-errors` + `HX-Reswap: innerHTML` (`internal/link/handler/handler.go:73`).

**Результат:** ты жмёшь "Create" — новая ссылка появляется в таблице без blinking-эффекта. Всё плавно, без перезагрузки.

### Пример 2: удаление строки

```html
<button hx-delete="/link/remove-link/Xk9m2aB"
        hx-target="closest tr"
        hx-swap="outerHTML">
  Удалить
</button>
```

- `hx-delete` — отправь DELETE-запрос
- `hx-target="closest tr"` — найди ближайшую строку таблицы
- `hx-swap="outerHTML"` — замени её на результат (пустоту, `c.NoContent` в `internal/link/handler/handler.go:107`)

Строка исчезает. Магия!

### Пример 3: плавные переходы между страницами

В `internal/core/view/view.templ:38` на теге `<body>` написано `hx-boost="true" hx-target="#main" hx-indicator="#top-loading-bar"`. Это значит — все обычные ссылки на сайте тоже перехватываются HTMX. Страница не перегружается целиком, а обновляется только `#main` между шапкой и подвалом. + стоят анимации перехода (globalViewTransitions) и индикатор загрузки.

### Главная фишка

Всё, что тебе нужно — это **атрибуты в HTML** (hx-post, hx-target, hx-swap). Не надо писать сложный JavaScript. HTMX всё делает сам.

---

# Как работает cursor-based пагинация (по 5 ссылок)

**Проблема:** если у тебя 1000 ссылок, нельзя загрузить все сразу — будет долго.

**Решение:** загружаем по 5 штук. Когда пользователь доскроллил до последней — подгружаем следующие 5.

### Как это устроено в коде

**Хранилище (`internal/link/storage/link.go:66`):**

```go
SELECT id, code, url, clicks, created_at
FROM link_link
WHERE user_id = ? AND id < ?
ORDER BY id DESC
LIMIT 5
```

Перевод: "Дай 5 ссылок этого пользователя, у которых id меньше, чем курсор. Отсортируй от новых к старым."

**Курсор** — это id последней загруженной ссылки. Если загрузили ссылки с id 10, 9, 8, 7, 6 — следующий курсор будет **6**. Получаем WHERE id < 6 → LIMIT 5 → получим 5, 4, 3, 2, 1.

**Первая загрузка** — когда пользователь впервые заходит на `GET /link/`, вызывается `GetCreateLink()` с курсором = `math.MaxInt64` (огромное число 9223372036854775807, `internal/link/handler/handler.go:58`). Все id в базе гарантированно меньше этого числа. Поэтому запрос `WHERE id < 9223372036854775807` возвращает последние 5 ссылок.

А для подгрузки через `ListLink` — курсор приходит из query-параметра `?cursor=...` (`internal/link/handler/handler.go:95`).

### Как это работает в шаблоне

Смотри `internal/link/view/link.templ:137` — последняя ссылка в списке обёрнута в `LastLink`:

```html
<tr hx-get="/link/list-link?cursor=6"
    hx-trigger="revealed"
    hx-swap="beforeend"
    hx-target="#list-links">
```

- `hx-trigger="revealed"` — когда этот элемент появился на экране (пользователь доскроллил)
- `hx-get="/link/list-link?cursor=6"` — запроси следующие 5 (id < 6)
- `hx-swap="beforeend"` + `hx-target="#list-links"` — вставь их в конец `tbody#list-links`

**Это бесконечный скролл**. Доехал до конца → подгрузились ещё 5 → среди них новая последняя → опять сработает `revealed` → подгрузятся следующие 5. И так до конца. Поиск (`/link/search-link`) возвращает без пагинации, лимит 20 (`internal/link/storage/link.go:191`).

---

# Как собрать Docker-образ

Docker — как **контейнер для перевозки**. Всё нужное (программа, файлы, настройки) пакуется в один образ. Можно запустить на любом компьютере.

### Dockerfile (3 этапа)

**Этап 1: templ** (`FROM ghcr.io/a-h/templ:latest AS templ`)

Берёт специальный образ с установленным `templ` и генерирует Go-код из `.templ`-файлов. Результат: файлы `*_templ.go`.

**Этап 2: сборка** (`FROM golang:1.26-alpine AS builder`)

Берёт образ Go, ставит `build-base sqlite-dev` (нужен CGO для `mattn/go-sqlite3`), копирует сгенерированные файлы, скачивает зависимости (`go mod download`) и компилирует программу в один бинарник (`CGO_ENABLED=1 go build -tags fts5 -ldflags="-s -w" -o bin/url-shortener cmd/url-shortener/main.go`). Флаги `-s -w` делают файл меньше, `-tags fts5` включает FTS5.

**Этап 3: запуск** (`FROM alpine:3.23 AS run`)

Берёт минимальный Linux (alpine — ~5 МБ). Копирует:
- Собранную программу `bin/url-shortener`
- Статические файлы `static/`
- Миграции `migrations/`

Создаёт пользователя `noroot` (безопасность — программа не имеет root-прав). Запускает на порту 8000 (`EXPOSE 8000`, `CMD /app/bin/url-shortener`).

**Итоговый образ весит ~15-20 МБ.**

### docker compose (compose.yml)

Описывает, как запускать сервис:

```yaml
services:
  app:
    build:
      dockerfile: ./Dockerfile       # собрать из этого файла
    restart: always                  # если упал — перезапусти
    env_file: .env                   # прочитать настройки
    volumes:
      - ".env:/app/.env"             # прокинуть .env
      - "./db:/app/db"               # база данных сохраняется на диске
    ports:
      - "8001:8000"                  # наружу порт 8001, внутри 8000
    healthcheck: ...                 # curl http://localhost:8000/health
```

### Как собрать и запустить

```bash
# Собрать образ (выполнить Dockerfile)
make build

# Запустить (прочитать compose.yml)
make up

# Остановить
make down
```

Или руками:

```bash
docker compose -p url-shortener build
docker compose -p url-shortener up -d
```

После этого открываешь `http://localhost:8001` — работает!

Сборка без Docker:
```bash
make build-bin  # templ generate + go build -tags fts5 -o bin/url-shortener
```
