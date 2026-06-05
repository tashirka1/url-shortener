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
// internal/auth/handler/user.go
c.Response().Header().Set("HX-Retarget", "#errors")
c.Response().Header().Set("HX-Reswap", "innerHTML")
return coreview.RenderTemplate(c, view.LoginError("email not found"))
```

| Заголовок | Назначение |
|-----------|------------|
| `HX-Redirect` | полная навигация браузера |
| `HX-Retarget` | куда вставить ответ (переопределяет `hx-target`) |
| `HX-Reswap` | как вставить (переопределяет `hx-swap`) |
| `HX-Trigger` | событие на клиенте после ответа |

`HX-Redirect` для логина:

```go
// internal/auth/handler/user.go
c.Response().Header().Set("HX-Redirect", "/link/create-link")
```

## Infinite scroll (cursor-based)

```html
<!-- internal/link/view/link.templ -->
<tr hx-get="/link/list-link?cursor=42"
    hx-trigger="revealed"
    hx-swap="afterend">
```

- `hx-trigger="revealed"` — срабатывает, когда элемент появляется в viewport
- `hx-swap="afterend"` — вставляет ответ **после** текущего элемента
- при скролле подгружается следующая порция (5 записей, `WHERE id < ?`)

## Delete-запрос

```html
<button hx-delete="/link/remove-link/abc123"
        hx-target="closest tr"
        hx-swap="outerHTML">
    Delete
</button>
```

- `hx-delete` — метод DELETE
- `closest tr` — ближайший родительский `<tr>`
- сервер возвращает `200 OK` с пустым телом, HTMX заменяет строку

## hx-boost

```html
<body hx-boost="true">
```

Все обычные ссылки и формы автоматически работают через AJAX.

## Глобальный обработчик 401

```javascript
// static/js/main.js
document.body.addEventListener("htmx:responseError", function (event) {
    if (event.detail.xhr.status === 401) {
        window.location.href = "/auth/login";
    }
});
```

## Переключение таргета при ошибке

```html
<form hx-post="/link/create-link"
      hx-target="#list-links"
      hx-swap="afterbegin"
      hx-on::after-request="this.setAttribute('hx-target', '#list-links'); ..."
      hx-on::error="this.setAttribute('hx-target', '#create-link-errors'); this.setAttribute('hx-swap', 'innerHTML')">
```

Успех: результат добавляется в начало списка. Ошибка: перенаправляется в `#create-link-errors`.
