# BUGS

## 🟥 Critical

### 1. Session.Get — nil pointer panic

**Файл:** `internal/core/session/session.go:18,28,37`

`sessions, _ := session.Get(UserSessionsKey, c)` — ошибка игнорируется. Если `Get()` возвращает ошибку, `sessions == nil`, и доступ к `sessions.Values` даёт panic. Три места: `AuthMiddleware`, `GetUserId`, `SetUserId`.

```go
sessions, _ := session.Get(UserSessionsKey, c)
// если sessions == nil, следующая строка паникует:
if userId, ok := sessions.Values[UserIdKey].(int); !ok || userId == 0 {
```

### 2. CustomHTTPErrorHandler пишет JSON в HTMX-приложение

**Файл:** `cmd/http/main.go:26-41`

Все HTMX-запросы ожидают HTML, но при ошибке 500+ приходит JSON. Форма создания ссылки через `hx-on::error` пытается ретаргетнуть ответ, но тело — JSON, а не HTML.

```go
if he, ok := err.(*echo.HTTPError); ok {
    c.JSON(he.Code, map[string]any{ // ← JSON, хотя клиент ждёт HTML
```

### 3. ListLink — panic при пустом списке

**Файл:** `internal/link/view/link.templ:81-89`

При `links == []model.Link{}` (новый пользователь, ни одной ссылки) `links[len(links)-1]` даёт индекс -1 — runtime panic. Guard `if len(links) == 0 { return }` отсутствует.

```templ
templ ListLink(links []model.Link) {
    for _, link := range links {
        if link.Id == links[len(links)-1].Id {
```

---

## 🟡 Логические

### 4. RowsAffected и LastInsertId ошибки заглушены

**Файл:** `internal/link/storage/link.go:36,40`

`rows, _ := row.RowsAffected()` — если ошибка, `rows = 0`, что ложно возвращает `ErrLinkAlreadyExists`.
`id, _ := row.LastInsertId()` — если ошибка, `id = 0`, ссылка создаётся с `Id = 0`.
Комментарий "SQLite doesn't support LastInsertId" неактуален для `modernc.org/sqlite`.

```go
rows, _ := row.RowsAffected()
if rows == 0 {
    return model.Link{}, model.ErrLinkAlreadyExists
}
id, _ := row.LastInsertId()
```

### 5. RemoveLink не проверяет, что строка удалена

**Файл:** `internal/link/storage/link.go:70`

DELETE на несуществующем или не принадлежащем пользователю коде возвращает успех (0 rows affected). HTMX удалит `<tr>` из DOM без реального удаления.

```go
_, err := r.db.ExecContext(ctx, "DELETE FROM link_link WHERE user_id=? AND code=?", userId, code)
```

### 6. ClickLink не проверяет, что строка обновлена

**Файл:** `internal/link/storage/link.go:88`

UPDATE на несуществующем коде возвращает успех (0 rows affected). Ошибка не логируется.

```go
_, err := r.db.ExecContext(ctx, "UPDATE link_link SET clicks=clicks + 1 WHERE code=?", code)
```

### 7. Миграции без таймаута

**Файл:** `internal/core/db/db.go:63`

`goose.UpContext(context.Background(), ...)` использует `context.Background()` — миграции могут висеть бесконечно при проблемах с БД.

```go
if err := goose.UpContext(context.Background(), db, "migrations"); err != nil {
```

### 8. SetUserId не проверяет ошибку Save

**Файл:** `internal/core/session/session.go:44`

`sess.Save(c.Request(), c.Response())` — ошибка игнорируется. Пользователь не узнает, что сессия не сохранилась.

---

## 🔵 Стилистические

### 9. Опечатка: Unathorized → Unauthorized

**Файлы:** `internal/core/view/view.templ:49`, `internal/core/session/session.go:21`

Название шаблона и текст в `<h1>`:

```templ
templ Unathorized(userId int) {
    @Base("URL Shortener", userId) {
        <div>
            <h1>Unathorized</h1>
```

### 10. Shadow import переменной sessions

**Файл:** `internal/core/session/session.go:18`

Переменная `sessions` shadow-ит импорт пакета `github.com/gorilla/sessions`:

```go
sessions, _ := session.Get(UserSessionsKey, c)
```

### 11. Избыточная функция sentinelErr в тестах

**Файл:** `internal/auth/service/user_test.go:16-18`

Обёртка не нужна — `return model.User{}, dbErr` работает без предупреждений go vet:

```go
func sentinelErr(err error) error {
    return err
}
```
