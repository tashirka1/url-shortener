package storage

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"url_shortener/internal/auth/model"
)

type UserStorage interface {
	CheckEmail(ctx context.Context, email string) (model.User, error)
	CreateUser(ctx context.Context, email string, hashedPassword []byte) error
}

type User struct {
	db *sql.DB
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

func (r *User) CheckEmail(ctx context.Context, email string) (model.User, error) {
	query := "SELECT id, email, password FROM auth_user WHERE email = ?"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return model.User{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close statement", "error", err)
		}
	}()

	user := model.User{}
	user.Email = email
	err = stmt.QueryRowContext(
		ctx,
		user.Email,
	).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *User) CreateUser(ctx context.Context, email string, hashedPassword []byte) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO auth_user(email, password) VALUES (?, ?)", email, hashedPassword)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return model.ErrUserAlreadyExists
	}
	return err
}
