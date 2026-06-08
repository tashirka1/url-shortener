# BUGS

## 🟡 Логические / Algorithmic

### 1. hx-swap не сбрасывается после ошибки

**Файл:** `internal/link/view/link.templ:19`

После серверной ошибки (echo.NewHTTPError) хендлер `hx-on::error` меняет `hx-swap` на `innerHTML`, но после успешного ответа `hx-swap` не сбрасывается обратно на `afterbegin`. Следующее успешное создание ссылки использует `hx-swap="innerHTML"` на `#list-links` — заменяет все ссылки в таблице одной новой.

```html
hx-on::after-request="this.setAttribute('hx-target', '#list-links'); ..."
hx-on::error="this.setAttribute('hx-target', '#create-link-errors'); this.setAttribute('hx-swap', 'innerHTML')"
<!--           ↑ hx-swap никогда не сбрасывается обратно на afterbegin  -->
```

**Фикс:** добавить сброс `hx-swap` в `hx-on::after-request`.

---

### 2. RemoveLink 404 — HTMX не свапает

**Файл:** `internal/link/handler/handler.go:98`

`c.NoContent(http.StatusNotFound)` — HTMX не выполняет свап контента при статусах 4xx/5xx. <tr> остаётся в DOM, хотя ссылка уже удалена.

```go
if errors.Is(err, sql.ErrNoRows) {
    return c.NoContent(http.StatusNotFound) // ← HTMX проигнорирует
}
```

**Фикс:** вернуть 200 OK с пустым телом.

---

### 3. Валидация email пропускает `a@b`

**Файл:** `internal/auth/handler/user.go:42`

```go
if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
    return errors.New("email must be a valid email address")
}
```

Проверка на наличие `.` без учёта позиции — `a@b` считается валидным.

**Фикс:** добавить проверку, что `.` идёт после `@` (или использовать `net/mail`).

---

### 4. Тест TestLogin_Success не тестирует успешный логин

**Файл:** `internal/auth/handler/user_test.go:40`

Хардкодный bcrypt hash `$2a$12$K6sIYqJZkS8kVq...` не является корректным хешом для пароля `"testpassword"`. `bcrypt.CompareHashAndPassword` всегда возвращает ошибку. Тест проверяет только что нет panic и статус 200, но реально логин не происходит.

```go
return model.User{Id: 1, Email: email, Password: "$2a$12$K6sIYqJZkS8kVq..."}, nil
//       ↑ невалидный hash, bcrypt.CompareHashAndPassword упадёт
```

**Фикс:** генерировать hash через `bcrypt.GenerateFromPassword` внутри теста (как в service тестах).

---

### 5. `class=".container"` — лишняя точка

**Файл:** `internal/core/view/view.templ:38`

```html
<main class=".container">
```

PicoCSS использует класс `container` (без точки). Класс `.container` не применяется.

**Фикс:** `class="container"`.

---

### 6. LinkService — dead code

**Файл:** `internal/link/service/link.go:10-16`

```go
type LinkService interface {
    ...
}
```

Интерфейс `LinkService` объявлен, но нигде не используется (никем не реализуется и не принимается как параметр).

**Фикс:** удалить.

---

## 🔵 Стилистические

### 7. CheckEmail избыточно использует PrepareContext

**Файл:** `internal/auth/storage/user.go:28`

`PrepareContext` + `QueryRowContext` для одного запроса:

```go
stmt, err := r.db.PrepareContext(ctx, query)
...
err = stmt.QueryRowContext(ctx, user.Email).Scan(...)
```

Можно заменить на `r.db.QueryRowContext(ctx, query, email).Scan(...)`.

---

### 8. UNIQUE constraint failed — хрупкая проверка

**Файл:** `internal/auth/storage/user.go:60`

```go
if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
```

Зависит от сообщения об ошибке конкретного драйвера SQLite. `modernc.org/sqlite` возвращает `"UNIQUE constraint failed: auth_user.email"` — работает, но хрупко.

---

### 9. Двойное логирование

**Файлы:** `internal/auth/service/user.go:43,56`, `internal/auth/handler/user.go:83,122`

Сервис и хендлер логируют одни и те же события (user authenticated, user created, user registered).

---

### 10. `math.MaxInt64` как cursor для первой страницы

**Файл:** `internal/link/handler/handler.go:45`

```go
links, err := h.s.ListLink(c.Request().Context(), userId, math.MaxInt64)
```

Работает (запрос `WHERE id < MaxInt64`), но семантически неправильно. Лучше передавать отдельный флаг или 0 как индикатор первой страницы.
