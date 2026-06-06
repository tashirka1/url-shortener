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
db.SetConnMaxIdleTime(15 * time.Minute)  // закрыть idle раньше, чем ConnMaxLifetime
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

## PRAGMA из проекта

Все прагмы применяются к каждому новому соединению через DSN + `db.Exec`:

```go
// DSN: busy_timeout передаётся через URL-параметр
dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)", path)

db.Exec(`
    PRAGMA busy_timeout=10000;        -- ждать до 10s при блокировке вместо SQLITE_BUSY
    PRAGMA foreign_keys=ON;           -- проверка внешних ключей (по умолчанию OFF!)
    PRAGMA journal_mode=WAL;          -- WAL: чтение не блокирует запись и наоборот
    PRAGMA synchronous = NORMAL;      -- безопаснее OFF, быстрее FULL (WAL mode)
    PRAGMA auto_vacuum = INCREMENTAL; -- не отдавать ОС сразу, можно вызвать PRAGMA incremental_vacuum
    PRAGMA journal_size_limit = 67110000;  -- ~64MB — ограничение размера WAL-журнала
    PRAGMA temp_store = MEMORY;       -- временные таблицы/индексы в RAM
    PRAGMA cache_size = -65536;       -- 64MB кеш страниц (отрицательное = килобайты)
    PRAGMA page_size = 4096;          -- размер страницы (до первой таблицы, после — только VACUUM)
`)
```

| PRAGMA | Зачем |
|--------|-------|
| `busy_timeout=10000` | Вместо `SQLITE_BUSY` — ждать 10 с |
| `foreign_keys=ON` | Включить проверку FK (по умолчанию OFF!) |
| `journal_mode=WAL` | Конкурентные чтения без блокировок |
| `synchronous=NORMAL` | Баланс скорости и безопасность (для WAL) |
| `auto_vacuum=INCREMENTAL` | Не фрагментировать БД при каждом DELETE |
| `journal_size_limit=67110000` | ~64 MB лимит на WAL-файл |
| `temp_store=MEMORY` | Временные данные в RAM |
| `cache_size=-65536` | 64 MB кеш страниц |
| `page_size=4096` | 4 KB страница (оптимально для SSD) |

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
DROP TABLE link_link;
DROP TABLE auth_user;
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

## Индексы

```sql
-- простой индекс
CREATE INDEX idx_link_user_id ON link_link(user_id);

-- уникальный индекс
CREATE UNIQUE INDEX idx_link_code ON link_link(code);
```

### Покрывающий индекс (covering index)

Индекс, который содержит ВСЕ поля, нужные запросу. SQLite не ходит в таблицу:

```sql
CREATE INDEX idx_link_covering ON link_link(user_id, code, url);

-- этому SELECT не нужна таблица — всё есть в индексе:
SELECT code, url FROM link_link WHERE user_id = 1;
```

Проверить `USING COVERING INDEX` в `EXPLAIN QUERY PLAN`.

### Кластерный индекс

В SQLite кластерный индекс — только `INTEGER PRIMARY KEY`. Данные хранятся в том же порядке, что и PK. Остальные индексы — некластерные (отдельная B-дерево со ссылками на rowid).

```sql
CREATE TABLE link_link(
    id INTEGER PRIMARY KEY,  -- это кластерный индекс
    ...
);
```

### Частичный индекс (partial index)

Индекс только для подмножества строк:

```sql
CREATE INDEX idx_active_links ON link_link(url) WHERE clicks > 0;
```

Занимает меньше места, ускоряет только запросы с тем же `WHERE`.

### Индекс по выражению

```sql
CREATE INDEX idx_link_url_lower ON link_link(LOWER(url));

SELECT * FROM link_link WHERE LOWER(url) = 'example.com';
```

## EXPLAIN QUERY PLAN

Показать, как SQLite выполняет запрос:

```sql
EXPLAIN QUERY PLAN
SELECT * FROM link_link WHERE user_id = 1 ORDER BY id DESC;
```

Вывод: `SEARCH`, `SCAN`, `USING INDEX`, `USING COVERING INDEX` — видно, используются ли индексы.

## JOIN

```sql
SELECT u.email, l.url, l.clicks
FROM auth_user u
JOIN link_link l ON l.user_id = u.id
WHERE u.email = 'user@example.com';

-- LEFT JOIN — все пользователи, даже без ссылок
SELECT u.email, COUNT(l.id) AS link_count
FROM auth_user u
LEFT JOIN link_link l ON l.user_id = u.id
GROUP BY u.id;
```

## Window functions

```sql
-- номер строки в группе
SELECT url, clicks,
       ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY clicks DESC) AS rank
FROM link_link;

-- скользящая сумма кликов по времени
SELECT created_at, clicks,
       SUM(clicks) OVER (ORDER BY created_at) AS running_total
FROM link_link
WHERE user_id = 1;
```

## CTE (Common Table Expression)

```sql
-- рекурсивный CTE: генерация чисел
WITH RECURSIVE cnt(x) AS (
    SELECT 1
    UNION ALL
    SELECT x + 1 FROM cnt WHERE x < 10
)
SELECT x FROM cnt;

-- именованный подзапрос
WITH user_links AS (
    SELECT user_id, COUNT(*) AS cnt
    FROM link_link
    GROUP BY user_id
)
SELECT u.email, ul.cnt
FROM auth_user u
JOIN user_links ul ON ul.user_id = u.id;
```

## ANALYZE

Собирает статистику для оптимизатора запросов:

```sql
ANALYZE;
```

После `ANALYZE` оптимизатор лучше выбирает индексы. Запускать после загрузки данных или больших изменений.

## VACUUM

Перестраивает БД — возвращает неиспользуемое место ОС:

```sql
VACUUM;
```

Блокирует БД на время выполнения. Для ежедневного использования — `PRAGMA auto_vacuum=1` (инкрементальный).

## PRAGMA optimize

Оптимизация без блокировок (анализ, обновление статистики):

```sql
PRAGMA optimize;
```

Безопасно запускать периодически (например, в cron или при завершении приложения).

## FTS5 — полнотекстовый поиск

```sql
-- создаём виртуальную FTS5-таблицу
CREATE VIRTUAL TABLE link_fts USING fts5(code, url, content='link_link', content_rowid='id');

-- синхронизируем с основной таблицей
INSERT INTO link_fts(rowid, code, url) SELECT id, code, url FROM link_link;

-- поиск
SELECT * FROM link_fts WHERE link_fts MATCH 'example';

-- обновление через триггеры (чтобы FTS не расходилась с данными)
CREATE TRIGGER link_fts_insert AFTER INSERT ON link_link BEGIN
    INSERT INTO link_fts(rowid, code, url) VALUES (new.id, new.code, new.url);
END;
```

## Transactions

```sql
-- явная транзакция: всё или ничего
BEGIN TRANSACTION;
UPDATE link_link SET clicks = clicks + 1 WHERE id = 1;
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('abc', 'https://x.com', 0, 1);
COMMIT;  -- или ROLLBACK при ошибке

-- SAVEPOINT — вложенные транзакции
SAVEPOINT sp;
UPDATE link_link SET clicks = clicks + 1 WHERE id = 1;
ROLLBACK TO sp;  -- откатить только часть
```

В Go:

```go
tx, _ := db.BeginTx(ctx, nil)
tx.ExecContext(ctx, "UPDATE link_link SET clicks = clicks + 1 WHERE id = ?", id)
tx.Commit()  // или tx.Rollback()
```

## Atomic UPDATE

Безопасное конкурентное обновление (не теряет клики):

```sql
-- атомарно: читает старое значение, прибавляет 1, записывает
UPDATE link_link SET clicks = clicks + 1 WHERE id = 1;

-- условие в UPDATE (CAS — compare-and-swap)
UPDATE link_link SET url = ? WHERE id = ? AND url = ?;
```

`RETURNING` — получить данные после обновления:

```sql
UPDATE link_link SET clicks = clicks + 1 WHERE id = 1 RETURNING clicks;
```

## sqlite-vec — векторный поиск

Расширение для хранения и поиска embedding-векторов (например, из LLM).

```bash
# загрузить .so / .dll со страницы релизов
# https://github.com/asg017/sqlite-vec/releases
```

```sql
-- загрузить расширение
.load ./vec0;

-- создать виртуальную таблицу для векторов размерностью 384
CREATE VIRTUAL TABLE link_embeddings USING vec0(
    id INTEGER PRIMARY KEY,
    embedding float[384]
);

-- вставка вектора (шестнадцатеричный blob из float32)
INSERT INTO link_embeddings(id, embedding)
VALUES (1, :vec_blob);

-- KNN-поиск: 10 ближайших соседей
SELECT id, distance
FROM link_embeddings
WHERE embedding MATCH :query_vec
    AND k = 10;

-- в Go: передаём []byte как BLOB
-- embedding_blob := Float32SliceToBytes(embedding)
-- db.Exec("INSERT INTO link_embeddings(id, embedding) VALUES (?, ?)", id, embedding_blob)
```

```go
// конвертация []float32 в blob для sqlite-vec
func Float32SliceToBytes(vec []float32) []byte {
    b := make([]byte, len(vec)*4)
    for i, v := range vec {
        binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
    }
    return b
}
```
