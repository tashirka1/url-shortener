package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func testBcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// MockUserStorage implements RepositoryInterface for testing
type MockUserStorage struct {
	checkEmailFunc func(ctx context.Context, email string) (model.User, error)
	createUserFunc func(ctx context.Context, email string, hashedPassword []byte) error
}

func (m *MockUserStorage) CheckEmail(ctx context.Context, email string) (model.User, error) {
	if m.checkEmailFunc != nil {
		return m.checkEmailFunc(ctx, email)
	}
	return model.User{}, errors.New("CheckEmail not implemented")
}

func (m *MockUserStorage) CreateUser(ctx context.Context, email string, hashedPassword []byte) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, email, hashedPassword)
	}
	return errors.New("CreateUser not implemented")
}

func TestLogin_Success(t *testing.T) {
	password := "testpassword"
	hash, err := testBcryptHash(password)
	require.NoError(t, err)

	mockStorage := &MockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: hash}, nil
		},
	}

	service := service.NewUser(mockStorage)
	handler := NewUser(service)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Form = map[string][]string{
		"email":    {"test@example.com"},
		"password": {password},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.PostLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("HX-Redirect"), "/link/create-link")
}

func TestLogin_InvalidPassword(t *testing.T) {
	password := "correct-password"
	hash, err := testBcryptHash(password)
	require.NoError(t, err)

	mockStorage := &MockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: hash}, nil
		},
	}

	service := service.NewUser(mockStorage)
	handler := NewUser(service)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Form = map[string][]string{
		"email":    {"test@example.com"},
		"password": {"wrong-password"},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.PostLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid email or password")
}

func TestLogin_EmailNotFound(t *testing.T) {
	// setup mock repository
	mockStorage := &MockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{}, model.ErrUserNotFound
		},
	}

	service := service.NewUser(mockStorage)
	handler := NewUser(service)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Form = map[string][]string{
		"email":    {"notfound@example.com"},
		"password": {"password123"},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.PostLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid email or password")
}
