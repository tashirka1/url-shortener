package session

import (
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
		sessions, _ := session.Get(UserSessionsKey, c)
		if userId, ok := sessions.Values[UserIdKey].(int); !ok || userId == 0 {
			c.Response().Header().Set("HX-Redirect", "/auth/login")
			return view.RenderTemplate(c, view.Unathorized(0))
		}
		return next(c)
	}
}

func GetUserId(c echo.Context) int {
	sess, _ := session.Get(UserSessionsKey, c)
	userId, ok := sess.Values[UserIdKey].(int)
	if !ok {
		return 0
	}
	return userId
}

func SetUserId(c echo.Context, value int) {
	sess, _ := session.Get(UserSessionsKey, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
		HttpOnly: true,
	}
	sess.Values[UserIdKey] = value
	sess.Save(c.Request(), c.Response())
}
