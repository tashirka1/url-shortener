package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"
	"url_shortener/internal/core/base62"
	"url_shortener/internal/link/model"

	"github.com/mattn/go-sqlite3"
)

type LinkStorage interface {
	CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
	ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
	RemoveLink(ctx context.Context, userId int, code string) error
	SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error)
	GetAndClick(ctx context.Context, code string, referrer, userAgent string) (model.Link, error)
	GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error)
	UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error
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
	rows, err := row.RowsAffected()
	if err != nil {
		return model.Link{}, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		var existing model.Link
		if scanErr := r.db.QueryRowContext(ctx, "SELECT id, code, url, clicks, created_at FROM link_link WHERE user_id=? AND url=?", userId, url).
			Scan(&existing.Id, &existing.Code, &existing.Url, &existing.Clicks, &existing.CreatedAt); scanErr != nil {
			return model.Link{}, fmt.Errorf("find existing link: %w", scanErr)
		}
		return existing, model.ErrLinkAlreadyExists
	}
	id, err := row.LastInsertId()
	if err != nil {
		return model.Link{}, fmt.Errorf("last insert id: %w", err)
	}
	return model.Link{Id: id, Code: code, Url: url}, nil
}

func (r *Link) ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error) {
	links := make([]model.Link, 0, 5)

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
	res, err := r.db.ExecContext(ctx, "DELETE FROM link_link WHERE user_id=? AND code=?", userId, code)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Link) GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error) {
	var link model.Link
	err := r.db.QueryRowContext(ctx, "SELECT id, code, url, clicks, created_at FROM link_link WHERE code=? AND user_id=?", code, userId).
		Scan(&link.Id, &link.Code, &link.Url, &link.Clicks, &link.CreatedAt)
	if err != nil {
		return model.Link{}, err
	}
	return link, nil
}

func (r *Link) GetAndClick(ctx context.Context, code string, referrer, userAgent string) (model.Link, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Link{}, err
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, "SELECT id, url FROM link_link WHERE code=?", code)
	var link model.Link
	if scanErr := row.Scan(&link.Id, &link.Url); scanErr != nil {
		return model.Link{}, scanErr
	}

	_, err = tx.ExecContext(ctx, "UPDATE link_link SET clicks=clicks+1 WHERE code=?", code)
	if err != nil {
		return model.Link{}, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO link_click(link_id, referrer, user_agent) VALUES (?, ?, ?)", link.Id, referrer, userAgent)
	if err != nil {
		return model.Link{}, err
	}

	return link, tx.Commit()
}

func ftsQuery(query string) string {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, query)
	tokens := strings.Fields(clean)
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, t := range tokens {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(t)
		b.WriteByte('*')
	}
	return b.String()
}

func (r *Link) UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error {
	if currentCode == newCode {
		return nil
	}
	res, err := r.db.ExecContext(ctx, "UPDATE link_link SET code=? WHERE code=? AND user_id=?", newCode, currentCode, userID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return model.ErrAliasTaken
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Link) SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error) {
	q := ftsQuery(query)
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.code, l.url, l.clicks, l.created_at
		FROM link_link l
		JOIN link_fts f ON l.id = f.rowid
		WHERE l.user_id = ? AND link_fts MATCH ?
		ORDER BY bm25(link_fts)
		LIMIT 20
	`, userId, q)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close rows", "error", err)
		}
	}()
	links := make([]model.Link, 0, 20)
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
