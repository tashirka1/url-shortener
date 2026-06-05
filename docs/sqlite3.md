# SQLite3: основы на примере проекта

## Подключение

`modernc.org/sqlite` — pure-Go драйвер, не требует CGO. DSN может содержать прагмы, которые применяются к каждому новому соединению:

```go
// internal/core/db/db.go
import _ "modernc.org/sqlite"

dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)", path)
db, err := sql.Open("sqlite", dsn)
```

## Connection pool

`sql.DB` — это пул соединений, не одно соединение. Настройка пула для read-heavy сценария:

```go
db.SetMaxOpenConns(8)                    // 8 читателей параллельно
db.SetMaxIdleConns(8)                    // все 8 держать в запасе
db.SetConnMaxLifetime(30 * time.Minute)  // периодическая переустановка
```

**`MaxOpenConns = 8`:** WAL mode позволяет читать параллельно. Писатель только один, но читатели не блокируются. При `busy_timeout=10s` писатель ждёт до 10 секунд, если БД заблокирована — ошибки `SQLITE_BUSY` не будет.

**`MaxIdleConns = MaxOpenConns`:** нет смысла закрывать соединения, если нагрузка постоянная. idle соединения не потребляют ресурсов, только лежат в пуле.

**`ConnMaxLifetime = 30min`:** периодическая переустановка сбрасывает temp-таблицы, prepared statements и накопленное состояние соединения.

### Почему не 1?

| Сценарий | `MaxOpenConns` | Зачем |
|----------|----------------|-------|
| Микро-сервис, 1 rps | 1 | Проще, нет риска `SQLITE_BUSY` |
| **Read-heavy + WAL** (наш) | **4–8** | Конкурентные чтения не блокируют друг друга |
| Write-heavy batch | 1 | Один писатель — нет конкуренции |

Для проекта с WAL + `busy_timeout` безопасно держать 8 соединений.

## WAL mode

```go
if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
    return nil, fmt.Errorf("enable WAL: %w", err)
}
```

WAL (Write-Ahead Logging) позволяет читать данные во время записи без блокировок. Без WAL читатели блокируют писателей и наоборот.

## Миграции через goose

```go
goose.SetBaseFS(url_shortener.EmbeddedMigrations)
goose.SetDialect("sqlite3")
goose.UpContext(ctx, db, "migrations")
```

SQL-миграция:

```sql
-- migrations/20260604192121_create_tables.sql
-- +goose Up
CREATE TABLE auth_user(id INTEGER PRIMARY KEY, email TEXT, password TEXT, UNIQUE(email));

CREATE TABLE link_link(
    id INTEGER PRIMARY KEY,
    code TEXT,
    url TEXT,
    clicks INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES auth_user(id),
    UNIQUE(code)
);

-- +goose Down
DROP TABLE IF EXISTS link_link;
DROP TABLE IF EXISTS auth_user;
```

## Параметризованные запросы

Всегда через `?` — защита от SQL-инъекций:

```go
// internal/link/storage/link.go
r.db.ExecContext(ctx,
    "INSERT INTO link_link(code, url, clicks, user_id) VALUES (?, ?, 0, ?)",
    code, url, userId)

r.db.QueryContext(ctx,
    "SELECT id, code, url, clicks, created_at FROM link_link WHERE user_id = ? AND id < ? ORDER BY id DESC LIMIT 5",
    userId, cursor)
```

## Типы SQLite → Go

| SQLite | Go |
|--------|----|
| `INTEGER` | `int64` (или `int` через драйвер) |
| `TEXT` | `string` |
| `DATETIME` | `time.Time` |
| `DEFAULT CURRENT_TIMESTAMP` | автозаполнение, Go не нужен |

```go
// internal/link/model/link.go
type Link struct {
    Id        int64
    Code      string
    Url       string
    Clicks    int
    CreatedAt time.Time
}
```

## cursor-based пагинация

```sql
WHERE user_id = ? AND id < ? ORDER BY id DESC LIMIT 5
```

Первая страница: `id = MaxInt64`. Каждая следующая: `id = последний полученный id`. Быстрее `OFFSET`, стабилен при вставках.

## ON CONFLICT

```sql
INSERT INTO link_link(code, url, clicks, user_id) VALUES (?, ?, 0, ?)
ON CONFLICT(user_id, url) DO NOTHING
```

Проверка `RowsAffected() == 0` для определения дубликата.
