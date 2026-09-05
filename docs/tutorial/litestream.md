# Litestream: непрерывная репликация SQLite в S3

## Что такое Litestream

Litestream — это standalone-демон для непрерывной репликации SQLite в S3-совместимое хранилище. Работает на уровне WAL (Write-Ahead Log): каждый раз, когда SQLite фиксирует транзакцию в WAL, Litestream читает новые страницы и отправляет их в S3.

Особенности:
- **Никаких изменений в приложении** — Litestream читает WAL-файл SQLite, приложение ничего не знает про репликацию
- **Point-in-time recovery** — можно восстановить БД на любой момент времени в пределах окна хранения
- **Минимальная задержка** — `sync-interval` по умолчанию 1 секунда

## Установка

```bash
# Linux (amd64)
wget https://github.com/benbjohnson/litestream/releases/download/v0.5.8/litestream-0.5.8-linux-x86_64.deb
sudo dpkg -i litestream-0.5.8-linux-x86_64.deb
```

## Конфигурация

### Репликация (`litestream/replicate.yml` — фактический шаблон в репо)

```yaml
access-key-id: ${LITESTREAM_ACCESS_KEY_ID}
secret-access-key: ${LITESTREAM_SECRET_ACCESS_KEY}
region: us-east-1
endpoint: ${LITESTREAM_ENDPOINT}
sync-interval: 1s

dbs:
  - path: /db/main.db              # ВАЖНО: в приложении дефолт ./db/url-shortener.db (DB_NAME, internal/core/config/config.go:39); синхронизируй path с DB_NAME или симлинком
    replica:
      type: s3
      bucket: bucket-name          # замени на реальный бакет (в ранней версии доки: 5e8ff288febf-sqlite-backups)
      force-path-style: true       # нужно для MinIO / Yandex Object Storage
      path: path-name              # префикс в бакете (ранее: url-shortener)
```

| Параметр | Описание |
|----------|----------|
| `access-key-id` / `secret-access-key` | credentials для S3 |
| `endpoint` | кастомный endpoint (MinIO, Yandex OBJ, DigitalOcean Spaces) |
| `sync-interval` | как часто проверять WAL (1s = потеря not более 1 секунды) |
| `force-path-style` | true для MinIO и S3-совместимых, кроме AWS |
| `path` | префикс в бакете (папка) |

### Restore (`litestream/restore.yml` — фактический шаблон в репо)

```yaml
access-key-id: ${LITESTREAM_ACCESS_KEY_ID}
secret-access-key: ${LITESTREAM_SECRET_ACCESS_KEY}
region: us-east-1
endpoint: ${LITESTREAM_ENDPOINT}

dbs:
  - path: ./main.db                # куда восстановить (пример)
    replicas:
      - type: s3
        bucket: bucket-name        # замени на реальный (ранее: 5e8ff288febf-sqlite-backups)
        path: path-name            # ранее: url-shortener
```

## Как это работает

```
┌──────────────┐     WAL-запись     ┌─────────────┐
│  SQLite      │ ──────────────────→│  WAL-файл    │
│  (приложение) │                   │  url-shortener.db-wal  │ # имя зависит от DB_NAME
└──────────────┘                    └──────┬──────┘
                                           │ Litestream читает
                                           ▼
                                    ┌──────────────┐
                                    │  S3-бакет     │
                                    │  url-shortener/│
                                    │   └── snapshots/
                                    │   └── wal/
                                    └──────────────┘
```

1. Приложение пишет в SQLite
2. SQLite пишет транзакции в WAL-файл
3. Litestream читает WAL, нарезает на сегменты, отправляет в S3
4. Периодически Litestream делает snapshot (полная копия БД)

### Структура в S3 (пример, bucket/path замени на свои)

```
bucket-name/path-name/
  snapshots/
    20250101T120000Z.db   -- полный слепок БД
    20250102T120000Z.db
  wal/
    0000000000000001.db-wal   -- WAL-сегменты
    0000000000000002.db-wal
    ...
# в ранней версии доки: 5e8ff288febf-sqlite-backups/url-shortener/
```

## Запуск

### Локально (без Docker)

```bash
# 1. Запустить Litestream, который форкает приложение
litestream replicate -config ./litestream/replicate.yml
```

### Восстановление

```bash
litestream restore -config ./litestream/restore.yml ./main.db
```

Восстанавливает latest snapshot + все WAL-сегменты после него. Результат — актуальная на момент восстановления БД.

## Мониторинг

### Логи

Litestream пишет в stderr в структурированном формате:

```
2025-06-06T10:00:00.000Z INF replicating db=/db/url-shortener.db max-age=1s # путь = DB_NAME
2025-06-06T10:00:01.000Z INF snapshot created db=/db/url-shortener.db name=20250606T100001Z.db size=4.2MB
2025-06-06T10:00:05.000Z WAL segment uploaded db=/db/url-shortener.db segment=0000000000000042.db-wal
```

## Аварийное восстановление

Если БД повреждена, а репликация работала:

```bash
# 1. Останавливаем приложение
docker compose down

# 2. Удаляем повреждённую БД (дефолт DB_NAME=./db/url-shortener.db, в шаблоне litestream — /db/main.db)
rm -f db/url-shortener.db db/url-shortener.db-wal db/url-shortener.db-shm
# или: rm -f db/main.db db/main.db-wal db/main.db-shm

# 3. Восстанавливаем из S3
litestream restore -config ./litestream/restore.yml ./db/url-shortener.db

# 4. Запускаем снова
docker compose up -d
```

## Особенности и подводные камни

| Проблема | Решение |
|----------|---------|
| SQLite не в WAL mode | Litestream **требует** WAL. Проверить: `PRAGMA journal_mode=WAL` |
| Размер БД > 10GB | Snapshot раз в сутки, WAL-сегменты накапливаются. Время восстановления растёт |
| Нет S3 | Любое S3-совместимое: MinIO (локально), Yandex OBJ, DigitalOcean Spaces, Backblaze B2 |
| Задержка репликации | `sync-interval: 1s` — не более 1 секунды потери. Для mission-critical можно 100ms |
| Concurrent writers | Litestream работает с одним WAL. Несколько писателей — только через SQLite WAL mode |
| Файл БД на NFS | Litestream не поддерживает NFS. Только локальная ФС |
| Шифрование | Litestream не шифрует данные на клиенте. Использовать S3-server-side encryption (SSE-S3) |

## Сравнение с альтернативами

| Инструмент | Подход | Задержка | Восстановление |
|------------|--------|----------|----------------|
| **Litestream** | WAL-replication → S3 | ~1s | Point-in-time |
| `sqlite3 .backup` | Полный копия | N/A | Только на момент backup |
| rqlite | Raft-кластер | ~50ms | Consensus, HA |
| `litestream` | WAL-replication | ~1s | Point-in-time |

Litestream — самый простой способ получить бекап SQLite с минимальными изменениями в проекте.

## Команды одной строкой

```bash
# запуск репликации
litestream replicate -config ./litestream/replicate.yml -exec "/app/bin/url-shortener"

# восстановление
litestream restore -config ./litestream/restore.yml ./main.db

# список snapshot-ов
litestream snapshots -config ./litestream/restore.yml

# список WAL-сегментов
litestream wal -config ./litestream/restore.yml
```
