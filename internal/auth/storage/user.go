package storage

import (
	"context"
	"database/sql"
	"errors"
	"url_shortener/internal/auth/model"

	"modernc.org/sqlite"
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
	user := model.User{}
	user.Email = email
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, password FROM auth_user WHERE email = ?",
		email,
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
	if err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == 2067 {
			return model.ErrUserAlreadyExists
		}
	}
	return err
}
