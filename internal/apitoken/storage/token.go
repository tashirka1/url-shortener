package storage

import (
	"context"
	"database/sql"
	"log/slog"
	"url_shortener/internal/apitoken/model"
)

type TokenStorage interface {
	Insert(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error)
	ListByUser(ctx context.Context, userID int) ([]model.Token, error)
	Revoke(ctx context.Context, userID int, tokenID int64) error
	FindByHash(ctx context.Context, hash string) (model.Token, error)
	UpdateLastUsed(ctx context.Context, tokenID int64) error
}

type Token struct {
	db *sql.DB
}

func NewToken(db *sql.DB) *Token {
	return &Token{db: db}
}

func (r *Token) Insert(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO api_token(user_id, name, token_hash, prefix) VALUES (?, ?, ?, ?)",
		userID, name, tokenHash, prefix)
	if err != nil {
		return model.Token{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Token{}, err
	}
	return model.Token{ID: id, UserID: userID, Name: name, TokenHash: tokenHash, Prefix: prefix}, nil
}

func (r *Token) ListByUser(ctx context.Context, userID int) ([]model.Token, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, name, token_hash, prefix, created_at, last_used_at, revoked_at FROM api_token WHERE user_id = ? AND revoked_at IS NULL ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close rows", "error", err)
		}
	}()
	var tokens []model.Token
	for rows.Next() {
		var t model.Token
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.TokenHash, &t.Prefix, &t.CreatedAt, &t.LastUsedAt, &t.RevokedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *Token) Revoke(ctx context.Context, userID int, tokenID int64) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE api_token SET revoked_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ? AND revoked_at IS NULL",
		tokenID, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrTokenNotFound
	}
	return nil
}

func (r *Token) FindByHash(ctx context.Context, hash string) (model.Token, error) {
	var t model.Token
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, name, token_hash, prefix, created_at, last_used_at, revoked_at FROM api_token WHERE token_hash = ?",
		hash).Scan(&t.ID, &t.UserID, &t.Name, &t.TokenHash, &t.Prefix, &t.CreatedAt, &t.LastUsedAt, &t.RevokedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Token{}, model.ErrTokenNotFound
		}
		return model.Token{}, err
	}
	return t, nil
}

func (r *Token) UpdateLastUsed(ctx context.Context, tokenID int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE api_token SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?", tokenID)
	return err
}
