-- +goose Up
CREATE TABLE link_click (
    id         INTEGER PRIMARY KEY,
    link_id    INTEGER NOT NULL REFERENCES link_link(id) ON DELETE CASCADE,
    referrer   TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    clicked_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_link_click_link_id_clicked_at ON link_click(link_id, clicked_at);

-- +goose Down
DROP INDEX idx_link_click_link_id_clicked_at;
DROP TABLE link_click;
