package api

import (
	"errors"
	"net/http"
	"strings"

	"url_shortener/internal/apitoken/model"
	"url_shortener/internal/apitoken/service"

	"github.com/labstack/echo/v4"
)

const UserIDKey = "userId"

type AuthMiddleware struct {
	s service.TokenService
}

func NewAuthMiddleware(s service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{s: s}
}

func (m *AuthMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		if header == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		rawToken, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			rawToken, ok = strings.CutPrefix(header, "bearer ")
		}
		if !ok || rawToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
		}

		token, err := m.s.Authenticate(c.Request().Context(), rawToken)
		if err != nil {
			if errors.Is(err, model.ErrTokenNotFound) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}
			if errors.Is(err, model.ErrTokenRevoked) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token revoked"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		c.Set(UserIDKey, token.UserID)
		return next(c)
	}
}

func GetUserID(c echo.Context) int {
	id, ok := c.Get(UserIDKey).(int)
	if !ok {
		return 0
	}
	return id
}
