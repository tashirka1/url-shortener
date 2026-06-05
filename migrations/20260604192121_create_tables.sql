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
CREATE INDEX idx_link_link_user_id ON link_link(user_id);
CREATE UNIQUE INDEX idx_link_link_user_url ON link_link(user_id, url);

-- +goose Down
DROP INDEX idx_link_link_user_id;
DROP INDEX idx_link_link_user_url;
DROP TABLE link_link;
DROP TABLE auth_user;
