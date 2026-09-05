# templ: основы на примере проекта

## Что такое templ

templ — язык шаблонов с компиляцией в Go. Шаблон пишется в `.templ` файле, компилятор генерирует `*_templ.go` с type-safe функциями.

## Компонент

```templ
// internal/core/view/view.templ:3
templ Nav(title string, userId int) {
    <nav>
        <ul>
            <li><a href="/"><strong>{ title }</strong></a></li>
        </ul>
        <ul>
            if userId != 0 {
                <li><a href="/link/tokens">Токены</a></li>
                <li><a href="/auth/logout">Выйти</a></li>
            } else {
                <li><a href="/auth/register">Регистрация</a></li>
                <li><a href="/auth/login">Войти</a></li>
            }
        </ul>
    </nav>
}
```

- `templ Nav(...)` — объявление компонента
- `{ title }` — интерполяция Go-переменной
- `if/else` — обычный Go-код

## Layout и children

```templ
// internal/core/view/view.templ:24
templ Base(title string, userId int) {
    <!DOCTYPE html>
    <html lang="ru">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self';"/>
            <link rel="icon" href="/static/icon/favicon.svg" type="image/svg+xml"/>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/css/pico@2.min.css"/>
            <link rel="stylesheet" href="/static/css/main.css?v=2"/>
        </head>
        <body hx-boost="true" hx-target="#main" hx-indicator="#top-loading-bar" style="padding: 0px 30px">
            <div id="top-loading-bar" class="htmx-indicator"></div>
            <div id="main">
                <div class="container">
                    @Nav(title, userId)
                    <main>
                        <section>
                            { children... }     <!-- слот для вложенного контента -->
                        </section>
                    </main>
                </div>
            </div>
            <script src="/static/js/htmx.org@2.0.10.min.js"></script>
            <script src="/static/js/main.js?v=2"></script>
        </body>
    </html>
}
```

Отличия от минимального примера: `lang="ru"`, CSP, favicon, `hx-target="#main"` + `hx-indicator`, контейнер `#main` для HTMX-буста, `main.css` и `main.js` с `?v=2` (кеш-бастинг).

Дочерний компонент встраивается через `@`:

```templ
templ Login(userId int) {
    @Base("URL Shortener", userId) {
        <div class="grid">
            <article>
                <h1>Login</h1>
                @LoginForm()
            </article>
        </div>
    }
}
```

## Итерация и условия

```templ
// internal/link/view/link.templ:159
templ ListLink(links []model.Link) {
    for _, link := range links {
        if link.Id == links[len(links)-1].Id {
            @LastLink(link)
        } else {
            @Link(link)
        }
    }
}
templ SearchResults(links []model.Link) {
    for _, link := range links {
        @Link(link) // без пагинации
    }
}
```

Работает как обычный Go: `for range`, `if`, `switch`. `LastLink` добавляет `hx-trigger="revealed"` для infinite scroll.

## HTMX-атрибуты внутри templ

```templ
// internal/link/view/link.templ:113
templ Link(link model.Link) {
    <tr>
        @CodeDisplay(link.Code)
        <td><a href={ templ.URL(link.Url) } target="_blank">{ link.Url }</a></td>
        <td>{ fmt.Sprintf("%d", link.Clicks) }</td>
        <td><a href={ templ.URL(fmt.Sprintf("/link/%s/stats", link.Code)) }>Статистика</a></td>
        <td><button hx-get={ string(templ.URL(fmt.Sprintf("/link/%s/qr", link.Code))) } hx-target="#qr-modal-container" hx-swap="innerHTML">QR</button></td>
        <td><button hx-delete={ string(templ.URL(fmt.Sprintf("/link/remove-link/%s", link.Code))) } hx-target="closest tr" hx-swap="outerHTML">Удалить</button></td>
    </tr>
}
```

URL нужно оборачивать в `templ.URL(...)` — это type-safe, templ проверит валидность. Для HTMX-атрибутов (`hx-get`, `hx-delete`, `hx-patch`) — `string(templ.URL(...))`.

### CodeDisplay / CodeEditField (alias)

```templ
// internal/link/view/link.templ:9
templ CodeDisplay(code string) {
    <td><a href={ templ.URL(fmt.Sprintf("/%s", code)) } class="shortLink">{ code }</a>
        <button hx-get={ string(templ.URL(fmt.Sprintf("/link/%s/edit-form", code))) } hx-target="closest td" hx-swap="outerHTML">✏️</button>
    </td>
}
templ CodeEditField(code string, errMsg ...string) {
    <td>
        <input name="code" value={ code } hx-patch={ string(templ.URL(fmt.Sprintf("/link/%s/alias", code))) } hx-target="closest tr" hx-swap="outerHTML" hx-trigger="keyup[key=='Enter']"/>
        <button hx-get={ string(templ.URL(fmt.Sprintf("/link/%s/display", code))) } hx-target="closest td">✕</button>
        @CodeEditErrors(errMsg)
    </td>
}
```

### OOB-swap при создании

```templ
// internal/link/view/link.templ:60
templ CreateLinkSuccess(link model.Link) {
    @Link(link)
    <div id="create-link-errors" style="color:red;" hx-swap-oob="true"></div>
}
```

`hx-swap-oob="true"` чистит блок ошибок после успеха.

## Рендеринг в Go-хендлере

```go
// internal/core/view/render.go
func RenderTemplate(c echo.Context, cmp templ.Component) error {
    c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
    return cmp.Render(c.Request().Context(), c.Response().Writer)
}
```

Использование в хендлере:

```go
// internal/link/handler/handler.go:63
return core_view.RenderTemplate(c, view.CreateLink(userId, links))
// ошибки: view.CreateLinkError, view.CodeEditField
// фрагменты: view.Link, view.ListLink, view.SearchResults, view.QRCodeModal
```

## Команды

```bash
go tool templ generate      # сгенерировать *_templ.go (вызывается в make build-bin / make check)
go tool templ fmt           # отформатировать .templ
```

Важно: `*_templ.go` добавлены в `.gitignore:4`, при клонировании репозитория нужно запускать `templ generate` (делает `make build-bin` и `make check`).

## Импорт

```templ
package view

import (
    "fmt"
    "url_shortener/internal/core/view"
    "url_shortener/internal/link/model"
)
```

Только `package view` — каждый `.templ` принадлежит Go-пакету, никаких package-level const/var.
