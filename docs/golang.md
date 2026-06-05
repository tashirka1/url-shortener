# Go: основы на примере проекта

## Пакеты и импорты

```go
package handler  // один package на директорию

import (
    "errors"
    "log/slog"
    "url_shortener/internal/auth/model"     // свой пакет
    "github.com/labstack/echo/v4"           // внешний
)
```

## Структуры и методы

```go
// internal/auth/storage/user.go
type UserStorage interface {
    CheckEmail(ctx context.Context, email string) (model.User, error)
    CreateUser(ctx context.Context, email string, hashedPassword []byte) error
}

type User struct {
    db *sql.DB
}

func NewUser(db *sql.DB) *User {
    return &User{db: db}
}

func (r *User) CheckEmail(ctx context.Context, email string) (model.User, error) {
    // метод на структуре User
}
```

**Правило:** интерфейс определяет потребитель, не производитель. `service.User` зависит от `storage.UserStorage` (интерфейса), а не от конкретной реализации.

## Обработка ошибок

```go
// internal/auth/handler/user.go
err := h.s.CheckUser(ctx, email, password)
if errors.Is(err, model.ErrInvalidPassword) {
    return coreview.RenderTemplate(c, view.LoginError("password isn't correct"))
} else if errors.Is(err, model.ErrUserNotFound) {
    return coreview.RenderTemplate(c, view.LoginError("email not found"))
}
```

`sentinel error` — переменная уровня пакета, сравнение через `errors.Is`.

## slog — структурированное логирование

```go
slog.Info("user logged in", "user_id", user.Id, "email", email)
slog.Warn("validation error", "user_id", userId, "error", err.Error())
slog.Error("failed to open database", "error", err)
```

Ключ-значение, без форматирования строк.

## defer

```go
defer func() {
    if err := stmt.Close(); err != nil {
        slog.WarnContext(ctx, "failed to close statement", "error", err)
    }
}()
```

Выполняется при выходе из функции. С ресурсами (stmt, rows) обязателен.

## Context

Контекст передаётся через всё приложение — от HTTP-запроса до SQL-запроса:

```go
// cmd/http/main.go — контекст с таймаутом для graceful shutdown
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
e.Shutdown(shutdownCtx)

// internal/link/handler/handler.go — контекст из echo-запроса
func (h *Link) ListLink(c echo.Context) error {
    links, err := h.s.ListLink(c.Request().Context(), userId, cursor)
    // ...
}

// internal/link/storage/link.go — контекст в SQL-запросе
rows, err := r.db.QueryContext(ctx, "SELECT ...", userId, cursor)
```

Правила:
- **Всегда** передавай `ctx` первым аргументом в функцию
- **Никогда** не храни `ctx` в структуре — только через аргументы
- Контекст отменяется при: timeout, отмене клиента, сигнале ОС

## HTTP-клиент с контекстом

```go
// GET с таймаутом через контекст
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/data", nil)
if err != nil {
    return err
}

resp, err := http.DefaultClient.Do(req)
if err != nil {
    // context.DeadlineExceeded при таймауте
    return err
}
defer resp.Body.Close()

// POST с телом
body := bytes.NewReader([]byte(`{"url":"https://example.com"}`))
req, _ = http.NewRequestWithContext(ctx, http.MethodPost, "https://api.example.com/check", body)
req.Header.Set("Content-Type", "application/json")
resp, _ = http.DefaultClient.Do(req)
```

**Важно:** всегда создавай новый `http.Request` с `NewRequestWithContext`. Стандартный клиент без контекста будет висеть вечно при недоступном сервере.

## if-else

```go
// internal/auth/handler/user.go
if errors.Is(err, model.ErrInvalidPassword) {
    return coreview.RenderTemplate(c, view.LoginError("password isn't correct"))
} else if errors.Is(err, model.ErrUserNotFound) {
    return coreview.RenderTemplate(c, view.LoginError("email not found"))
}

// internal/core/session/session.go — проверка с объявлением переменной
if userId, ok := sessions.Values[UserIdKey].(int); !ok || userId == 0 {
    return view.RenderTemplate(c, view.Unathorized(0))
}
```

Идиома: `if err := doSomething(); err != nil {`.

## switch

```go
// internal/auth/handler/user.go — switch по типу ошибки
switch {
case errors.Is(err, model.ErrInvalidPassword):
    return coreview.RenderTemplate(c, view.LoginError("password isn't correct"))
case errors.Is(err, model.ErrUserNotFound):
    return coreview.RenderTemplate(c, view.LoginError("email not found"))
default:
    return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
}

// switch по значению
switch method {
case "GET":
    // GET logic
case "POST":
    // POST logic
default:
    // fallback
}

// type switch
switch v := anyVal.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("unknown")
}
```

В Go `switch` без выражения — это `switch true`. Каждый `case` — любое булево выражение. `break` не нужен, проваливание только через `fallthrough`.

## for range

```go
// internal/core/base62/base62.go — генерация 7 символов
code := make([]byte, codeLength)  // codeLength = 7
for i := range code {
    n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
    code[i] = charset[n.Int64()]
}

// internal/link/storage/link.go — итерация строк результата
for rows.Next() {
    var link model.Link
    rows.Scan(&link.Id, &link.Code, &link.Url, &link.Clicks, &link.CreatedAt)
    links = append(links, link)
}

// internal/link/view/link.templ — итерация слайса в шаблоне
for _, link := range links {
    if link.Id == links[len(links)-1].Id {
        @LastLink(link)
    } else {
        @Link(link)
    }
}
```

| Форма | Когда использовать |
|-------|-------------------|
| `for i := range slice` | Нужен только индекс |
| `for i, v := range slice` | Нужны индекс и значение |
| `for _, v := range slice` | Нужно только значение |
| `for k, v := range map` | Итерация мапы |
| `for range ch` | Канал |

## Slice

```go
// internal/link/storage/link.go
var links []model.Link                         // nil-слайс, 0 элементов
links = append(links, link)                    // добавление
// ...
return links, nil

// internal/link/handler/handler.go
cursor, err := strconv.Atoi(c.QueryParam("cursor"))  // string → int
```

Слайс — это ссылка на underlying array. `append` аллоцирует новый массив при превышении capacity.

## Map

```go
// cmd/http/main.go — литерал map
c.JSON(he.Code, map[string]any{
    "error":   http.StatusText(he.Code),
    "message": he.Message,
})

// internal/core/session/session.go — чтение с проверкой
userId := sess.Values[UserIdKey]                // Values — map[interface{}]interface{}
if userId != nil {
    return userId.(int)
}
```

Паттерн `ok` для мапы:

```go
if val, ok := m["key"]; ok {
    // ключ существует
}
```

## Pointer

```go
// internal/auth/storage/user.go — возврат указателя из конструктора
func NewUser(db *sql.DB, sessionStore *sessions.CookieStore) *User {
    return &User{db: db, sessionStore: sessionStore}
}

// internal/auth/service/user.go — указатель на структуру в методе
func (s *User) CheckUser(ctx context.Context, email, password string) (model.User, error) {
    // метод на указателе (s *User), а не на значении (s User)
}
```

Когда ставить `*`:
- **Конструкторы** всегда возвращают `*Struct`
- **Методы с изменением состояния** — на `*T`
- **БД, HTTP-клиенты, большие структуры** — передавать как `*T`
- **int, string, bool, маленькие struct** — передавать значением

## Goroutine Worker

```go
// cmd/http/main.go — goroutine для старта сервера
go func() {
    slog.Info("server starting on :8000")
    if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
        slog.Error("server error", "error", err)
        stop()  // дёргаем cancel, основной поток завершается
    }
}()

<-ctx.Done()  // main() блокируется до сигнала SIGINT/SIGTERM
```

Паттерн graceful shutdown:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

go func() { /* запуск сервера */ }()

<-ctx.Done()
// здесь — очистка, закрытие соединений
dbs.Close()
```

## sync.Pool

Пул для переиспользования временных объектов. Снижает нагрузку на GC:

```go
import "sync"

var bufPool = sync.Pool{
    New: func() any {
        return make([]byte, 4096)
    },
}

func handle() {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // используем buf...
}
```

В проекте не используется (SQLite-драйвер управляет памятью сам), но полезен для **буферов, сериализации, temporary объектов**.

## Map и Filter (имитация)

Go не имеет встроенных `map`/`filter`. Пишется руками через цикл:

```go
// Filter: оставить только ссылки с кликами > 0
var popular []model.Link
for _, link := range links {
    if link.Clicks > 0 {
        popular = append(popular, link)
    }
}

// Map: достать только URL
urls := make([]string, len(links))
for i, link := range links {
    urls[i] = link.Url
}
```

Для частых трансформаций — `slices` и `maps` из `golang.org/x/exp` (Go 1.21+).

## Sorting

```go
import "sort"

// сортировка по возрастанию ID (аналог ORDER BY id)
sort.Slice(links, func(i, j int) bool {
    return links[i].Id < links[j].Id
})

// сортировка по убыванию кликов
sort.Slice(links, func(i, j int) bool {
    return links[i].Clicks > links[j].Clicks
})
```

В проекте сортировка не нужна — `ORDER BY id DESC` в SQL делает то же самое. Но для in-memory данных — `sort.Slice`.

## Работа с SQLite

Полный разбор в `docs/sqlite3.md`. Кратко:

```go
import _ "modernc.org/sqlite"       // pure-Go, без CGO

db, err := sql.Open("sqlite", "file:db/main.db?_pragma=busy_timeout(10000)")
db.SetMaxOpenConns(8)               // WAL + busy_timeout = concurrent reads
db.SetMaxIdleConns(8)

// параметризованные запросы через ?
rows, _ := db.QueryContext(ctx, "SELECT * FROM link_link WHERE user_id = ?", id)

// одна строка
row := db.QueryRowContext(ctx, "SELECT url FROM link_link WHERE code = ?", code)
row.Scan(&url)
```
