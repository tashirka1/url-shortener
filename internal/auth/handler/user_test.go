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
)

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
	// setup mock repository
	mockStorage := &MockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: "$2a$12$K6sIYqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kV"}, nil
		},
	}

	// create service with mock
	service := service.NewUser(mockStorage)
	handler := NewUser(service)

	// setup echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Form = map[string][]string{
		"email":    {"test@example.com"},
		"password": {"testpassword"},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	err := handler.PostLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLogin_InvalidPassword(t *testing.T) {
	// setup mock repository — return real bcrypt hash, service will fail Compare
	realHash := "$2a$12$K6sIYqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kVqJZkS8kV"
	mockStorage := &MockUserStorage{
		checkEmailFunc: func(ctx context.Context, email string) (model.User, error) {
			return model.User{Id: 1, Email: email, Password: realHash}, nil
		},
	}

	service := service.NewUser(mockStorage)
	handler := NewUser(service)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Form = map[string][]string{
		"email":    {"test@example.com"},
		"password": {"wrongpassword"},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.PostLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "password isn&#39;t correct")
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
	assert.Contains(t, rec.Body.String(), "email not found")
}
