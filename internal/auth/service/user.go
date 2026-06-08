package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/storage"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

func argonHash(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

func argonVerify(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid argon2id hash format")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("not argon2id hash")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}
	var memory, time, threads int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	computed := argon2.IDKey([]byte(password), salt, uint32(time), uint32(memory), uint8(threads), uint32(len(hash)))
	if subtle.ConstantTimeCompare(hash, computed) == 1 {
		return true, nil
	}
	return false, nil
}

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

	ok, err := argonVerify(password, user.Password)
	if err != nil {
		slog.WarnContext(ctx, "argon2 verify failed", "email", email, "error", err)
		return model.User{}, errors.Join(model.ErrInvalidPassword, err)
	}
	if !ok {
		slog.WarnContext(ctx, "password mismatch", "email", email)
		return model.User{}, model.ErrInvalidPassword
	}

	return user, nil
}

func (s *User) CreateUser(ctx context.Context, email, password string) error {
	hashedPassword, err := argonHash(password)
	if err != nil {
		slog.ErrorContext(ctx, "argon2 hash failed", "error", err)
		return err
	}
	if err := s.r.CreateUser(ctx, email, []byte(hashedPassword)); err != nil {
		slog.WarnContext(ctx, "create user failed", "email", email, "error", err)
		return err
	}
	slog.InfoContext(ctx, "user created", "email", email)
	return nil
}
