package session

import (
	"log/slog"
	"url_shortener/internal/core/view"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	UserSessionsKey string = "user"
	UserIdKey       string = "userId"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(UserSessionsKey, c)
		if err != nil || sess == nil {
			slog.Warn("session get failed", "error", err)
			c.Response().Header().Set("HX-Redirect", "/auth/login")
			return view.RenderTemplate(c, view.Unathorized(0))
		}
		if userId, ok := sess.Values[UserIdKey].(int); !ok || userId == 0 {
			c.Response().Header().Set("HX-Redirect", "/auth/login")
			return view.RenderTemplate(c, view.Unathorized(0))
		}
		return next(c)
	}
}

func GetUserId(c echo.Context) int {
	sess, err := session.Get(UserSessionsKey, c)
	if err != nil || sess == nil {
		slog.Warn("session get failed", "error", err)
		return 0
	}
	userId, ok := sess.Values[UserIdKey].(int)
	if !ok {
		return 0
	}
	return userId
}

func SetUserId(c echo.Context, value int) {
	sess, err := session.Get(UserSessionsKey, c)
	if err != nil || sess == nil {
		slog.Error("session get failed", "error", err)
		return
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
		HttpOnly: true,
	}
	sess.Values[UserIdKey] = value
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		slog.Error("session save failed", "error", err)
	}
}
