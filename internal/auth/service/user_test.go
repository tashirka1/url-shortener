package service

import (
	"context"
	"errors"
	"testing"

	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func sentinelErr(err error) error {
	return err
}

type mockUserStorage struct {
	checkEmailFunc func(ctx context.Context, email string) (model.User, error)
	createUserFunc func(ctx context.Context, email string, hashedPassword []byte) error
}

func (m *mockUserStorage) CheckEmail(ctx context.Context, email string) (model.User, error) {
	if m.checkEmailFunc != nil {
		return m.checkEmailFunc(ctx, email)
	}
	return model.User{}, errors.New("CheckEmail not implemented")
}

func (m *mockUserStorage) CreateUser(ctx context.Context, email string, hashedPassword []byte) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, email, hashedPassword)
	}
	return errors.New("CreateUser not implemented")
}

var _ storage.UserStorage = (*mockUserStorage)(nil)

func TestCheckUser_Success(t *testing.T) {
	password := "correct-password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	mock := &mockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: string(hash)}, nil
		},
	}
	s := NewUser(mock)

	user, err := s.CheckUser(context.Background(), "test@example.com", password)

	assert.NoError(t, err)
	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestCheckUser_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)

	mock := &mockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: string(hash)}, nil
		},
	}
	s := NewUser(mock)

	_, err = s.CheckUser(context.Background(), "test@example.com", "wrong-password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidPassword))
}

func TestCheckUser_EmailNotFound(t *testing.T) {
	mock := &mockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{}, model.ErrUserNotFound
		},
	}
	s := NewUser(mock)

	_, err := s.CheckUser(context.Background(), "missing@example.com", "password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrUserNotFound))
}

func TestCheckUser_StorageError(t *testing.T) {
	dbErr := errors.New("db connection failed")
	mock := &mockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{}, sentinelErr(dbErr)
		},
	}
	s := NewUser(mock)

	_, err := s.CheckUser(context.Background(), "test@example.com", "password")

	assert.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}

func TestCreateUser_Success(t *testing.T) {
	mock := &mockUserStorage{
		createUserFunc: func(ctx context.Context, email string, hashedPassword []byte) error {
			assert.Equal(t, "new@example.com", email)
			assert.NotEmpty(t, hashedPassword)
			// verify the password was hashed with bcrypt cost 12
			err := bcrypt.CompareHashAndPassword(hashedPassword, []byte("user-password"))
			assert.NoError(t, err)
			return nil
		},
	}
	s := NewUser(mock)

	err := s.CreateUser(context.Background(), "new@example.com", "user-password")

	assert.NoError(t, err)
}

func TestCreateUser_StorageError(t *testing.T) {
	mock := &mockUserStorage{
		createUserFunc: func(ctx context.Context, email string, hashedPassword []byte) error {
			return sentinelErr(model.ErrUserAlreadyExists)
		},
	}
	s := NewUser(mock)

	err := s.CreateUser(context.Background(), "existing@example.com", "password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrUserAlreadyExists))
}
