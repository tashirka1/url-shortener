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
  model/link.go       — структуры данных и sentinel-ошибки (ErrLinkAlreadyExists, ErrAliasTaken)
  storage/link.go     — SQL-запросы, только CRUD + FTS + транзакция GetAndClick
  storage/click.go    — агрегация кликов (ClickStorage)
  service/link.go     — бизнес-логика (логгирование, валидация alias)
  handler/handler.go  — HTTP (Echo), парсинг запроса, рендер
  view/link.templ     — HTML-шаблоны
```

Другие модули по той же схеме: `internal/auth/`, `internal/apitoken/`, `internal/rps/` — изолированы, не импортируют друг друга, связываются через интерфейсы в `cmd/url-shortener/main.go:88`.

## 1. Model — данные

```go
// internal/link/model/link.go
const MaxURLLength = 2048

var ErrLinkAlreadyExists = errors.New("link already exists")
var ErrAliasTaken = errors.New("alias already taken")

type Link struct {
    Id        int64
    Code      string
    Url       string
    Clicks    int
    CreatedAt time.Time
}
```

Чистые структуры, без внешних импортов (кроме `time`, `errors`).

## 2. Storage — SQL (самый нижний слой)

```go
// internal/link/storage/link.go:17
type LinkStorage interface {
    CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
    ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
    RemoveLink(ctx context.Context, userId int, code string) error
    SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error)
    GetAndClick(ctx context.Context, code string, referrer, userAgent string) (model.Link, error)
    GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error)
    UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error
}

type Link struct {
    db *sql.DB
}

func (r *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
    code, _ := base62.NewCode()
    row, _ := r.db.ExecContext(ctx, "INSERT INTO link_link(code, url, clicks, user_id) VALUES (?, ?, 0, ?) ON CONFLICT(user_id, url) DO NOTHING", code, url, userId)
    // RowsAffected==0 → SELECT существующей + ErrLinkAlreadyExists
    // ...
    return model.Link{Id: id, Code: code, Url: url}, nil
}
```

Второй storage для кликов:

```go
// internal/link/storage/click.go
type ClickStorage interface {
    ListClicks(ctx context.Context, linkId int64) ([]model.Click, error)
}
```

В storage нет логирования (кроме `slog.Warn` при `rows.Close`) и HTTP-логики — только SQL, транзакции (`GetAndClick` делает `BEGIN; SELECT; UPDATE clicks+1; INSERT link_click; COMMIT` в `storage/link.go:113`), FTS5 (`SearchLink` → `JOIN link_fts ... MATCH ?`).

## 3. Service — бизнес-логика

```go
// internal/link/service/link.go:11
type LinkService interface {
    CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
    ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
    RemoveLink(ctx context.Context, userId int, code string) error
    SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error)
    GetAndClick(ctx context.Context, code, referrer, userAgent string) (string, error)
    GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error)
    UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error
}

type Link struct {
    r  storage.LinkStorage   // зависит от интерфейса, не от реализации
    cr storage.ClickStorage
}

func (s *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
    link, err := s.r.CreateLink(ctx, url, userId)
    if err != nil {
        slog.ErrorContext(ctx, "create link failed", "error", err)
        return model.Link{}, fmt.Errorf("create link: %w", err)
    }
    slog.InfoContext(ctx, "link created", "code", link.Code)
    return link, nil
}

func (s *Link) UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error {
    if len(newCode) < 3 || len(newCode) > 32 { return fmt.Errorf("alias должен быть от 3 до 32 символов") }
    // проверка [a-zA-Z0-9_-] + reserved: health, auth, link, rps, static
    return s.r.UpdateAlias(ctx, userID, currentCode, newCode)
}
```

- вызывает storage
- логирует (`slog.*Context`)
- валидирует alias (`service/link.go:68`)
- НЕ знает про HTTP, сессии, templates

## 4. Handler — HTTP

```go
// internal/link/handler/handler.go:24
type Link struct {
    s  service.LinkService
    cs service.ClickService
}

func (h *Link) PostCreateLink(c echo.Context) error {
    userId := session.GetUserId(c)     // из сессии
    url := c.FormValue("url")          // из формы

    url, err := validateURL(url) // добавляет https://, проверяет схему/host, MaxURLLength
    if err != nil {
        c.Response().Header().Set("HX-Retarget", "#create-link-errors")
        c.Response().Header().Set("HX-Reswap", "innerHTML")
        return coreview.RenderTemplate(c, view.CreateLinkError(err.Error()))
    }

    link, err := h.s.CreateLink(c.Request().Context(), url, userId)
    if errors.Is(err, model.ErrLinkAlreadyExists) {
        c.Response().Header().Set("HX-Retarget", "#create-link-errors")
        c.Response().Header().Set("HX-Reswap", "innerHTML")
        return coreview.RenderTemplate(c, view.CreateLinkError("эта ссылка уже существует"))
    }
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка создания ссылки")
    }
    c.Response().Header().Set("HX-Trigger", "reset-create-form")
    return coreview.RenderTemplate(c, view.CreateLinkSuccess(link))
}
```

- читает запрос (form values, params, `c.Param("code")`, `c.QueryParam("cursor")`)
- вызывает service
- рендерит HTML-фрагмент (templ) или `echo.NewHTTPError`, управляет HTMX через `HX-Redirect`/`HX-Retarget`/`HX-Reswap`/`HX-Trigger`
- не знает SQL

Дополнительно: `RedirectLink` (`GET /:code`), `SearchLink` (`GET /link/search-link`), `UpdateAlias` (`PATCH /link/:code/alias`), `GetQRCode`, `Stats`.

## 5. View — HTML

```templ
// internal/link/view/link.templ:113
templ Link(link model.Link) {
    <tr>
        @CodeDisplay(link.Code)
        <td><a href={ templ.URL(link.Url) } target="_blank">{ link.Url }</a></td>
        <td>{ link.Clicks }</td>
        <td><a href={ templ.URL(fmt.Sprintf("/link/%s/stats", link.Code)) }>Статистика</a></td>
        <td><button hx-get={ string(templ.URL(fmt.Sprintf("/link/%s/qr", link.Code))) } hx-target="#qr-modal-container">QR</button></td>
        <td><button hx-delete={ string(templ.URL(fmt.Sprintf("/link/remove-link/%s", link.Code))) } hx-target="closest tr" hx-swap="outerHTML">Удалить</button></td>
    </tr>
}
```

Только presentation. Не содержит бизнес-логику. Код рендерится через `CodeDisplay` (ссылка `/{code}` + кнопка редактирования alias).

## Dependency Injection

Сборка в `cmd/url-shortener/main.go:88`:

```go
linkStrg := link_storage.NewLink(database)
clickStrg := link_storage.NewClick(database)
linkSvc := link_service.NewLink(linkStrg, clickStrg)
clickSvc := link_service.NewClick(clickStrg)
link_handler.SetupHandlers(e, linkSvc, clickSvc) // внутри: e.Group("/link") + session.AuthMiddleware

tokenStrg := apitoken_storage.NewToken(database)
tokenSvc := apitoken_service.NewToken(tokenStrg)
apitoken_handler.SetupHandlers(e, tokenSvc)
apiAuth := apitoken_middleware.NewAuthMiddleware(tokenSvc)
apiGroup := e.Group("/api/v1"); apiGroup.Use(apiAuth.Middleware)
link_handler.NewLinkAPI(linkSvc, clickSvc).SetupAPIRoutes(apiGroup)

rpsStrg := rps_storage.NewRPS(database)
rps_handler.NewRPS(rpsStrg).SetupRoutes(e)
```

Модули не импортируют друг друга — только интерфейсы, внедрение в `main.go`.
