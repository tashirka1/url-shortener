-- +goose Up
DROP INDEX IF EXISTS idx_rps_meta_log_id;
CREATE INDEX idx_rps_meta_log_id ON rps_meta(log_id, key, value);

-- +goose Down
DROP INDEX IF EXISTS idx_rps_meta_log_id;
CREATE INDEX idx_rps_meta_log_id ON rps_meta(log_id);