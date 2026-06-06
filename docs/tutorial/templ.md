# templ: основы на примере проекта

## Что такое templ

templ — язык шаблонов с компиляцией в Go. Шаблон пишется в `.templ` файле, компилятор генерирует `*_templ.go` с type-safe функциями.

## Компонент

```templ
// internal/core/view/view.templ
templ Nav(title string, userId int) {
    <nav>
        <ul>
            <li><a href="/"><strong>{ title }</strong></a></li>
        </ul>
        <ul>
            if userId != 0 {
                <li><a href="/auth/logout">Logout</a></li>
            } else {
                <li><a href="/auth/register">Register</a></li>
                <li><a href="/auth/login">Login</a></li>
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
templ Base(title string, userId int) {
    <!DOCTYPE html>
    <html>
        <head>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/css/pico@2.min.css"/>
            <script src="/static/js/htmx.org@2.0.10.min.js"></script>
        </head>
        <body hx-boost="true">
            @Nav(title, userId)
            <main>
                { children... }     <!-- слот для вложенного контента -->
            </main>
        </body>
    </html>
}
```

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
templ ListLink(links []model.Link) {
    for _, link := range links {
        if link.Id == links[len(links)-1].Id {
            @LastLink(link)
        } else {
            @Link(link)
        }
    }
}
```

Работает как обычный Go: `for range`, `if`, `switch`.

## HTMX-атрибуты внутри templ

```templ
templ Link(link model.Link) {
    <tr>
        <td>
            <a href={ templ.URL(fmt.Sprintf("/r/%s", link.Code)) } target="_blank">
                { link.Code }
            </a>
        </td>
        <td>
            <button hx-delete={ string(templ.URL(fmt.Sprintf("/link/remove-link/%s", link.Code))) }
                    hx-target="closest tr"
                    hx-swap="innerHTML">
                Delete
            </button>
        </td>
    </tr>
}
```

URL нужно оборачивать в `templ.URL(...)` — это type-safe, templ проверит валидность.

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
// internal/link/handler/handler.go
return coreview.RenderTemplate(c, view.CreateLink(userId, links))
```

## Команды

```bash
go tool templ generate      # сгенерировать *_templ.go
go tool templ fmt           # отформатировать .templ
```

Важно: `*_templ.go` добавлены в `.gitignore`, при клонировании репозитория нужно запускать `templ generate`.

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
