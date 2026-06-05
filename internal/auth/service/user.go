package service

import (
	"context"
	"errors"
	"log/slog"
	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/storage"

	"golang.org/x/crypto/bcrypt"
)

type UserUser interface {
	CheckUser(ctx context.Context, email, password string) (model.User, error)
	CreateUser(ctx context.Context, email, password string) error
}

type User struct {
	r storage.UserStorage
}

func NewUser(r storage.UserStorage) *User {
	return &User{r: r}
}

func (s *User) CheckUser(ctx context.Context, email, password string) (model.User, error) {
	user, err := s.r.CheckEmail(ctx, email)
	if err != nil {
		slog.WarnContext(ctx, "check email failed", "email", email, "error", err)
		return model.User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		slog.WarnContext(ctx, "password mismatch", "email", email, "error", err)
		return model.User{}, errors.Join(model.ErrInvalidPassword, err)
	}

	slog.InfoContext(ctx, "user authenticated", "user_id", user.Id, "email", email)
	return user, nil
}

func (s *User) CreateUser(ctx context.Context, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		slog.ErrorContext(ctx, "bcrypt hash failed", "error", err)
		return err
	}
	if err := s.r.CreateUser(ctx, email, hashedPassword); err != nil {
		slog.WarnContext(ctx, "create user failed", "email", email, "error", err)
		return err
	}
	slog.InfoContext(ctx, "user created", "email", email)
	return nil
}
