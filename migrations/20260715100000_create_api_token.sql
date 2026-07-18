-- +goose Up
CREATE TABLE api_token(
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    prefix TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    revoked_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_api_token_hash ON api_token(token_hash);
CREATE INDEX idx_api_token_user_id ON api_token(user_id, created_at DESC);

-- +goose Down
DROP INDEX idx_api_token_user_id;
DROP INDEX idx_api_token_hash;
DROP TABLE api_token;
