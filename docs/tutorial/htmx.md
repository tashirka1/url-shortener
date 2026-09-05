# HTMX: основы на примере проекта

## Базовая атрибутика

```html
<!-- internal/auth/view/login.templ -->
<form hx-post="/auth/login">
    <input type="email" name="email" required/>
    <input type="password" name="password" required/>
    <button>Login</button>
    <div id="errors"></div>
</form>
```

- `hx-post="/auth/login"` — отправляет форму POST-запросом (AJAX)
- ответ целиком заменяет контент формы (или target)

## HX-заголовки из Go

Сервер управляет поведением через заголовки ответа:

```go
// internal/auth/handler/user.go:63
c.Response().Header().Set("HX-Retarget", "#errors")
c.Response().Header().Set("HX-Reswap", "innerHTML")
return coreview.RenderTemplate(c, view.LoginError("email not found"))
```

| Заголовок | Назначение |
|-----------|------------|
| `HX-Redirect` | полная навигация браузера |
| `HX-Retarget` | куда вставить ответ (переопределяет `hx-target`) |
| `HX-Reswap` | как вставить (переопределяет `hx-swap`) |
| `HX-Trigger` | событие на клиенте после ответа (`reset-create-form`) |

`HX-Redirect` для логина:

```go
// internal/auth/handler/user.go:83
c.Response().Header().Set("HX-Redirect", "/link/")
// после регистрации → HX-Redirect: /auth/login
```

`HX-Trigger` для сброса формы создания ссылки:

```go
// internal/link/handler/handler.go:89
c.Response().Header().Set("HX-Trigger", "reset-create-form")
// + в шаблоне hx-on:reset-create-form="this.reset()"
```

## Infinite scroll (cursor-based)

```html
<!-- internal/link/view/link.templ:137 LastLink -->
<tr hx-get="/link/list-link?cursor=42"
    hx-trigger="revealed"
    hx-swap="beforeend"
    hx-target="#list-links">
```

- `hx-trigger="revealed"` — срабатывает, когда элемент появляется в viewport
- `hx-target="#list-links"` — куда вставлять (tbody)
- `hx-swap="beforeend"` — вставляет ответ в конец `tbody` (а не `afterend` текущего `tr`)
- при скролле подгружается следующая порция (5 записей, `WHERE user_id=? AND id < ? ORDER BY id DESC LIMIT 5` в `internal/link/storage/link.go:66`)

Первая страница грузится с `cursor=math.MaxInt64` (`internal/link/handler/handler.go:58`).

## Delete-запрос

```html
<button hx-delete="/link/remove-link/abc123"
        hx-target="closest tr"
        hx-swap="outerHTML">
    Удалить
</button>
```

- `hx-delete` — метод DELETE
- `closest tr` — ближайший родительский `<tr>`
- сервер возвращает `200 OK` с пустым телом (`c.NoContent` в `internal/link/handler/handler.go:113`), HTMX заменяет строку на пустоту

## hx-boost + indicator

```html
<!-- internal/core/view/view.templ:38 -->
<body hx-boost="true" hx-target="#main" hx-indicator="#top-loading-bar">
    <div id="top-loading-bar" class="htmx-indicator"></div>
    <div id="main"> ... @Nav ... { children... } </div>
</body>
```

Все обычные ссылки и формы автоматически работают через AJAX и обновляют только `#main`. Индикатор `#top-loading-bar` показывает загрузку (стили в `static/css/main.css`). Включены `globalViewTransitions` (`static/js/main.js:2`).

## Поиск (FTS5)

```html
<!-- internal/link/view/link.templ:80 -->
<input type="search" name="q"
       hx-get="/link/search-link"
       hx-trigger="keyup delay:300ms, search"
       hx-target="#list-links"
       placeholder="Поиск ссылок..."/>
```

Дебаунс 300ms, сервер делает `JOIN link_fts ... MATCH ? ORDER BY bm25` (`internal/link/storage/link.go:185`), лимит 20. При `q=""` возвращает пагинацию (`handler.go:141`).

## Alias-редактирование

```html
<!-- internal/link/view/link.templ:13 CodeDisplay -->
<button hx-get="/link/abc123/edit-form" hx-target="closest td" hx-swap="outerHTML">✏️</button>

<!-- CodeEditField -->
<input name="code" hx-patch="/link/abc123/alias" hx-target="closest tr" hx-swap="outerHTML" hx-trigger="keyup[key=='Enter']"/>
<button hx-get="/link/abc123/display" hx-target="closest td" hx-swap="outerHTML">✕</button>
```

Патч `PATCH /link/:code/alias` валидирует 3–32 символа `[a-zA-Z0-9_-]` + reserved (`internal/link/service/link.go:68`). Успех → перерисовка `Link`, ошибка → `CodeEditField` с сообщением.

## QR-код

```html
<button hx-get="/link/abc123/qr" hx-target="#qr-modal-container" hx-swap="innerHTML">QR</button>
<div id="qr-modal-container"></div>
```

Хендлер `GetQRCode` (`internal/link/handler/handler.go:203`) генерит PNG через `go-qrcode` и отдает `QRCodeModal` (`link.templ:43`).

## Переключение таргета при ошибке (создание ссылки)

```html
<!-- internal/link/view/link.templ:71 актуально -->
<form hx-post="/link/" hx-target="#list-links" hx-swap="afterbegin"
      hx-on:reset-create-form="this.reset()">
```

Успех: `CreateLinkSuccess` (`link.templ:60`) вставляет `Link` в начало + `hx-swap-oob` чистит `#create-link-errors`. Ошибка: хендлер переопределяет `HX-Retarget: #create-link-errors` + `HX-Reswap: innerHTML` (`handler.go:73`) и рендерит `CreateLinkError`.

## Копирование ссылки (без HTMX)

```js
// static/js/main.js:5 — делегирование на .shortLink / .copy-token
document.body.addEventListener("click", async (e) => {
  const shortLink = e.target.closest(".shortLink");
  if (shortLink) { await navigator.clipboard.writeText(shortLink.href); /* "Скопировано!" 2s */ }
});
```
