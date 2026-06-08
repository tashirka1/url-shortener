package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"

	"url_shortener/internal/auth/model"
	"url_shortener/internal/auth/service"
	"url_shortener/internal/auth/storage"
	"url_shortener/internal/auth/view"
	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"

	"github.com/labstack/echo/v4"
)

type User struct {
	s *service.User
}

func NewUser(s *service.User) *User {
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
		if errors.Is(err, model.ErrInvalidPassword) {
			slog.Warn("login failed", "email", email, "error", "invalid password")
			return core_view.RenderTemplate(c, view.LoginError("password isn't correct"))
		} else if errors.Is(err, model.ErrUserNotFound) {
			slog.Warn("login failed", "email", email, "error", "email not found")
			return core_view.RenderTemplate(c, view.LoginError("email not found"))
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
	session.SetUserId(c, 0)
	return c.Redirect(http.StatusSeeOther, "/auth/login")
}

func (h *User) Register(c echo.Context) error {
	if c.Request().Method == "POST" {
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

	userId := session.GetUserId(c)
	return core_view.RenderTemplate(c, view.Register(userId))
}

func SetupHandlers(e *echo.Echo, db *sql.DB) {
	storage := storage.NewUser(db)
	service := service.NewUser(storage)
	handler := NewUser(service)

	group := e.Group("/auth")
	group.GET("/login", handler.GetLogin)
	group.POST("/login", handler.PostLogin)
	group.GET("/logout", handler.Logout)
	group.GET("/register", handler.Register)
	group.POST("/register", handler.Register)
}
