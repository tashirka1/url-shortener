-- +goose Up
CREATE TABLE rps_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    payload TEXT NOT NULL,
    ts INTEGER NOT NULL,
    duration INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE rps_meta (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_id INTEGER NOT NULL REFERENCES rps_log(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE INDEX idx_rps_meta_log_id ON rps_meta(log_id);

-- +goose Down
DROP INDEX idx_rps_meta_log_id;
DROP TABLE rps_meta;
DROP TABLE rps_log;
