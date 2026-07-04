package service

import (
	"context"
	"fmt"
	"log/slog"
	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/storage"

	"golang.org/x/crypto/bcrypt"
)

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

type UserService interface {
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
		return model.User{}, fmt.Errorf("check email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		slog.WarnContext(ctx, "password mismatch", "email", email)
		return model.User{}, model.ErrInvalidPassword
	}

	return user, nil
}

func (s *User) CreateUser(ctx context.Context, email, password string) error {
	hashedPassword, err := bcryptHash(password)
	if err != nil {
		slog.ErrorContext(ctx, "bcrypt hash failed", "error", err)
		return fmt.Errorf("bcrypt hash: %w", err)
	}
	if err := s.r.CreateUser(ctx, email, []byte(hashedPassword)); err != nil {
		slog.WarnContext(ctx, "create user failed", "email", email, "error", err)
		return fmt.Errorf("create user: %w", err)
	}
	slog.InfoContext(ctx, "user created", "email", email)
	return nil
}
