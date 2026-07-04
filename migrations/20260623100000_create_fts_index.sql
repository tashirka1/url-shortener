-- +goose Up
CREATE VIRTUAL TABLE link_fts USING fts5(
    code, url,
    tokenize='unicode61',
    content='link_link',
    content_rowid='id'
);

-- +goose StatementBegin
CREATE TRIGGER link_ai AFTER INSERT ON link_link BEGIN
    INSERT INTO link_fts(rowid, code, url) VALUES (new.id, new.code, new.url);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER link_ad AFTER DELETE ON link_link BEGIN
    INSERT INTO link_fts(link_fts, rowid, code, url) VALUES('delete', old.id, old.code, old.url);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER link_au AFTER UPDATE ON link_link BEGIN
    INSERT INTO link_fts(link_fts, rowid, code, url) VALUES('delete', old.id, old.code, old.url);
    INSERT INTO link_fts(rowid, code, url) VALUES (new.id, new.code, new.url);
END;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO link_fts(rowid, code, url) SELECT id, code, url FROM link_link;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS link_ai;
DROP TRIGGER IF EXISTS link_ad;
DROP TRIGGER IF EXISTS link_au;
DROP TABLE IF EXISTS link_fts;
