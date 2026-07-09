-- +goose Up
INSERT INTO auth_user(email, password) VALUES ('test@test.ru', '$2a$10$dqJiFSXWUfCwb/h6yYbGMuBBpwXuzh5aJtDjIQk86wx9Nftl9VrjS');
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('S9NEfOF', 'https://x.com', 0, 1);
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('ThZ7rgY', 'https://vk.com', 0, 1);
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('nVJVrym', 'https://youtube.com', 0, 1);
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('itUDVt5', 'https://instagram.com', 0, 1);
INSERT INTO link_link(code, url, clicks, user_id) VALUES ('CbpaPIl', 'https://facebook.com', 0, 1);

-- +goose Down
DELETE FROM link_link WHERE code IN ('S9NEfOF', 'ThZ7rgY', 'nVJVrym', 'itUDVt5', 'CbpaPIl');
DELETE FROM auth_user WHERE email = 'test@test.ru';
