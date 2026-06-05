# Найденные ошибки

## 🔴 Критические

### 1. Login bypass при неизвестной ошибке

**Файл:** `internal/auth/handler/user.go:70-80`

Если `CheckUser` возвращает ошибку, не равную `ErrInvalidPassword` или `ErrUserNotFound` (например, DB timeout), код не делает `return` и продолжает выполнение — пользователь считается аутентифицированным.

```go
if err != nil {
    c.Response().Header().Set("HX-Retarget", "#errors")
    c.Response().Header().Set("HX-Reswap", "innerHTML")
    if errors.Is(err, model.ErrInvalidPassword) {
        return coreview.RenderTemplate(c, view.LoginError("password isn't correct"))
    } else if errors.Is(err, model.ErrUserNotFound) {
        return coreview.RenderTemplate(c, view.LoginError("email not found"))
    }
}
// fallthrough — user.Id = 0, но SetUserId всё равно вызывается
slog.Info("user logged in", "user_id", user.Id, "email", email)
session.SetUserId(c, user.Id)
```

### 2. Регистрация «успешна» при неизвестной ошибке

**Файл:** `internal/auth/handler/user.go:106-119`

Аналогичная проблема: если `CreateUser` возвращает ошибку, отличную от `ErrUserAlreadyExists`, код падает на `slog.Info("user registered")` и редиректит на `/auth/login`.

```go
if err != nil {
    if errors.Is(err, model.ErrUserAlreadyExists) {
        return coreview.RenderTemplate(c, view.RegisterError("the email is already in use"))
    }
}
// fallthrough — пользователь думает, что зарегистрирован
slog.Info("user registered", "email", email)
c.Response().Header().Set("HX-Redirect", "/auth/login")
return nil
```

### 3. Panic при пустом списке ссылок

**Файл:** `internal/link/view/link.templ:79-86`

При `links == []model.Link{}` (новый пользователь, ни одной ссылки) `links[len(links)-1]` даёт индекс -1 — runtime panic.

```templ
templ ListLink(links []model.Link) {
    for _, link := range links {
        if link.Id == links[len(links)-1].Id {
```

---

## 🟡 Логические / UI

### 4. Удаление не удаляет `<tr>` из DOM

**Файл:** `internal/link/view/link.templ:55`

`hx-swap="innerHTML"` с пустым телом ответа очищает содержимое строки, но `<tr>` остаётся в таблице. Пользователь видит пустую строку.

### 5. RedirectLink редиректит на логин при любой ошибке

**Файл:** `internal/link/handler/handler.go:109`

Любая ошибка `GetLink` (даже DB timeout) редиректит на `/auth/login` без сообщения.

### 6. ClickLink error silently ignored

**Файл:** `internal/link/handler/handler.go:112`

Возврат `h.s.ClickLink(...)` не проверяется — ошибки инкремента клика не логируются и не обрабатываются.

---

## 🔵 Мелкие

### 7. PRAGMA journal_mode — ошибка игнорируется

**Файл:** `internal/core/db/db.go:21`

`db.Exec("PRAGMA journal_mode=WAL")` — если WAL не включился, база работает в режиме по умолчанию без предупреждения.

### 8. Unsafe type assertion в GetUserId

**Файл:** `internal/core/session/session.go:31`

`userId.(int)` без comma-ok — panic, если в сессии лежит значение не типа `int`.

### 9. Неиспользуемое поле sessionStore

**Файл:** `internal/auth/storage/user.go:22`

Поле `sessionStore *sessions.CookieStore` в структуре `User` никогда не используется.

### 10. Избыточное создание model.Link в GetLink

**Файл:** `internal/link/storage/link.go:79`

Для сканирования одного поля `url` создаётся `var link model.Link` вместо `var url string`.
