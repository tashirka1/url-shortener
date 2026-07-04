# Layered Architecture: устройство проекта

## Схема

```
handler → service → storage → SQLite
    ↑          ↑         ↑
  HTTP       бизнес     работа с БД
 (Echo)     (логика)   (SQL/queries)
```

Данные передаются через **model**.

## Каждый слой отвечает за своё

На примере `internal/link/`:

```
link/
  model/link.go       — структуры данных и sentinel-ошибки
  storage/link.go     — SQL-запросы, только CRUD
  service/link.go     — бизнес-логика (логгирование, валидация)
  handler/handler.go  — HTTP (Echo), парсинг запроса, рендер
  view/link.templ     — HTML-шаблоны
```

## 1. Model — данные

```go
// internal/link/model/link.go
type Link struct {
    Id        int64
    Code      string
    Url       string
    Clicks    int
    CreatedAt time.Time
}

var ErrLinkAlreadyExists = errors.New("link already exists")
```

## 2. Storage — SQL (самый нижний слой)

```go
// internal/link/storage/link.go
type LinkStorage interface {
    CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
    ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
    RemoveLink(ctx context.Context, userId int, code string) error
    GetLink(ctx context.Context, code string) (string, error)
    ClickLink(ctx context.Context, code string) error
}

type Link struct {
    db *sql.DB
}

func (r *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
    code, _ := base62.NewCode()
    row, _ := r.db.ExecContext(ctx, "INSERT INTO link_link(...) VALUES (...) ON CONFLICT ...", ...)
    // ...
    return model.Link{Id: id, Code: code, Url: url}, nil
}
```

В storage нет логирования и HTTP-логики — только SQL.

## 3. Service — бизнес-логика

```go
// internal/link/service/link.go
type Link struct {
    r storage.LinkStorage   // зависит от интерфейса, не от реализации
}

func (s *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
    link, err := s.r.CreateLink(ctx, url, userId)
    if err != nil {
        slog.ErrorContext(ctx, "create link failed", "error", err)
        return model.Link{}, err
    }
    slog.InfoContext(ctx, "link created", "code", link.Code)
    return link, nil
}
```

- вызывает storage
- логирует
- НЕ знает про HTTP, сессии, templates

## 4. Handler — HTTP

```go
// internal/link/handler/handler.go
type Link struct {
    s *service.Link     // зависит от service
}

func (h *Link) PostCreateLink(c echo.Context) error {
    userId := session.GetUserId(c)     // из сессии
    url := c.FormValue("url")          // из формы

    if err := validateURL(url); err != nil {
        return coreview.RenderTemplate(c, view.CreateLinkError(err.Error()))
    }

    link, err := h.s.CreateLink(c.Request().Context(), url, userId)
    if errors.Is(err, model.ErrLinkAlreadyExists) {
        return coreview.RenderTemplate(c, view.CreateLinkError("this URL already exists"))
    }
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create link")
    }
    return coreview.RenderTemplate(c, view.Link(link))
}
```

- читает запрос (form values, params)
- вызывает service
- рендерит ответ (шаблон или JSON)

Не должен знать SQL или бизнес-логику.

## 5. View — HTML

```templ
templ Link(link model.Link) {
    <tr>
        <td><a href="/r/{ link.Code }">{ link.Code }</a></td>
        <td>{ link.Url }</td>
        <td>{ link.Clicks }</td>
        <td><button hx-delete="/link/remove-link/{ link.Code }">Delete</button></td>
    </tr>
}
```

Только presentation. Не содержит бизнес-логику.

## Dependency Injection

Сборка в `SetupHandlers`:

```go
func SetupHandlers(e *echo.Echo, db *sql.DB, sessionStore *sessions.CookieStore) {
    storage := storage.NewLink(db)                    // SQLite
    service := service.NewLink(storage)                // слой выше
    handler := NewLink(service)                        // HTTP

    group := e.Group("/link")
    group.Use(session.AuthMiddleware)
    group.GET("/create-link", handler.GetCreateLink)
    // ...
}
```
