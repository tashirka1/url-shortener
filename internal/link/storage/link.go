package storage

import (
	"context"
	"database/sql"
	"log/slog"
	"url_shortener/internal/core/base62"
	"url_shortener/internal/link/model"
)

type LinkStorage interface {
	CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
	ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
	RemoveLink(ctx context.Context, userId int, code string) error
	GetLink(ctx context.Context, code string) (string, error)
	ClickLink(ctx context.Context, code string) error
}

type Link struct {
	db *sql.DB
}

func NewLink(db *sql.DB) *Link {
	return &Link{db: db}
}

func (r *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
	code, err := base62.NewCode()
	if err != nil {
		return model.Link{}, err
	}
	row, err := r.db.ExecContext(ctx, "INSERT INTO link_link(code, url, clicks, user_id) VALUES (?, ?, 0, ?) ON CONFLICT(user_id, url) DO NOTHING", code, url, userId)
	if err != nil {
		return model.Link{}, err
	}
	rows, _ := row.RowsAffected()
	if rows == 0 {
		return model.Link{}, model.ErrLinkAlreadyExists
	}
	id, _ := row.LastInsertId() // SQLite doesn't support LastInsertId, ignore error
	return model.Link{Id: id, Code: code, Url: url}, nil
}

func (r *Link) ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error) {
	var links []model.Link

	rows, err := r.db.QueryContext(ctx, "SELECT id, code, url, clicks, created_at FROM link_link WHERE user_id = ? AND id < ? ORDER BY id DESC LIMIT 5", userId, cursor)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close rows", "error", err)
		}
	}()
	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.Id, &link.Code, &link.Url, &link.Clicks, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *Link) RemoveLink(ctx context.Context, userId int, code string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM link_link WHERE user_id=? AND code=?", userId, code)
	if err != nil {
		return err
	}
	return nil
}

func (r *Link) GetLink(ctx context.Context, code string) (string, error) {
	row := r.db.QueryRowContext(ctx, "SELECT url FROM link_link WHERE code=?", code)
	var link model.Link
	err := row.Scan(&link.Url)
	if err != nil {
		return "", err
	}
	return link.Url, nil
}

func (r *Link) ClickLink(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE link_link SET clicks=clicks + 1 WHERE code=?", code)
	if err != nil {
		return err
	}
	return nil
}
