package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"

	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/service"
	"url_shortener/internal/auth/view"
	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"

	"github.com/labstack/echo/v4"
)

type User struct {
	s service.UserService
}

func NewUser(s service.UserService) *User {
	return &User{s: s}
}

func validateLogin(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return nil
}

func validateRegister(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email must be a valid email address")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}

func (h *User) GetLogin(c echo.Context) error {
	userId := session.GetUserId(c)
	return core_view.RenderTemplate(c, view.Login(userId))
}

func (h *User) PostLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if err := validateLogin(email, password); err != nil {
		slog.Warn("login validation error", "email", email, "error", err.Error())
		c.Response().Header().Set("HX-Retarget", "#errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return core_view.RenderTemplate(c, view.LoginError(err.Error()))
	}

	user, err := h.s.CheckUser(c.Request().Context(), email, password)

	if err != nil {
		c.Response().Header().Set("HX-Retarget", "#errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		if errors.Is(err, model.ErrInvalidPassword) || errors.Is(err, model.ErrUserNotFound) {
			slog.Warn("login failed", "email", email, "error", err)
			return core_view.RenderTemplate(c, view.LoginError("invalid email or password"))
		}
		slog.Error("login failed", "email", email, "error", err)
		return core_view.RenderTemplate(c, view.LoginError("internal error"))
	}

	slog.Info("user logged in", "user_id", user.Id, "email", email)
	session.SetUserId(c, user.Id)
	c.Response().Header().Set("HX-Redirect", "/link/create-link")

	return nil
}

func (h *User) Logout(c echo.Context) error {
	session.ClearSession(c)
	return c.Redirect(http.StatusSeeOther, "/auth/login")
}

func (h *User) GetRegister(c echo.Context) error {
	userId := session.GetUserId(c)
	return core_view.RenderTemplate(c, view.Register(userId))
}

func (h *User) PostRegister(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if err := validateRegister(email, password); err != nil {
		slog.Warn("register validation error", "email", email, "error", err.Error())
		c.Response().Header().Set("HX-Retarget", "#errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return core_view.RenderTemplate(c, view.RegisterError(err.Error()))
	}

	err := h.s.CreateUser(c.Request().Context(), email, password)

	if err != nil {
		if errors.Is(err, model.ErrUserAlreadyExists) {
			slog.Warn("register failed", "email", email, "error", "email already in use")
			c.Response().Header().Set("HX-Retarget", "#errors")
			c.Response().Header().Set("HX-Reswap", "innerHTML")
			return core_view.RenderTemplate(c, view.RegisterError("the email is already in use"))
		}
		slog.Error("register failed", "email", email, "error", err)
		c.Response().Header().Set("HX-Retarget", "#errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return core_view.RenderTemplate(c, view.RegisterError("internal error"))
	}

	slog.Info("user registered", "email", email)
	c.Response().Header().Set("HX-Redirect", "/auth/login")
	return nil
}

func SetupHandlers(e *echo.Echo, s service.UserService) {
	h := NewUser(s)

	group := e.Group("/auth")
	group.GET("/login", h.GetLogin)
	group.POST("/login", h.PostLogin)
	group.GET("/logout", h.Logout)
	group.GET("/register", h.GetRegister)
	group.POST("/register", h.PostRegister)
}
