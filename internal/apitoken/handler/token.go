package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"url_shortener/internal/apitoken/model"
	"url_shortener/internal/apitoken/service"
	"url_shortener/internal/apitoken/view"
	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"

	"github.com/labstack/echo/v4"
)

type Token struct {
	s service.TokenService
}

func NewToken(s service.TokenService) *Token {
	return &Token{s: s}
}

func (h *Token) Index(c echo.Context) error {
	userID := session.GetUserId(c)
	ctx := c.Request().Context()

	tokens, err := h.s.ListByUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "token index: list failed", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка")
	}

	return core_view.RenderTemplate(c, view.TokenPage(userID, tokens))
}

func (h *Token) Generate(c echo.Context) error {
	userID := session.GetUserId(c)
	name := c.FormValue("name")

	if name == "" || len(name) > 128 {
		c.Response().Header().Set("HX-Retarget", "#token-errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		msg := "Название токена обязательно"
		if len(name) > 128 {
			msg = "Название токена не может быть длиннее 128 символов"
		}
		return c.HTML(http.StatusBadRequest, "<span>"+msg+"</span>")
	}

	ctx := c.Request().Context()

	res, err := h.s.Generate(ctx, userID, name)
	if err != nil {
		slog.ErrorContext(ctx, "token generate failed", "user_id", userID, "error", err)
		c.Response().Header().Set("HX-Retarget", "#token-errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return c.HTML(http.StatusInternalServerError, "Ошибка создания токена")
	}

	tokens, err := h.s.ListByUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "token generate: list after create failed", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка")
	}

	c.Response().Header().Set("HX-Trigger", "reset-token-form")

	return core_view.RenderTemplate(c, view.TokenListAndDisplay(tokens, res.RawToken, res.Token.Name))
}

func (h *Token) Revoke(c echo.Context) error {
	userID := session.GetUserId(c)
	tokenIDStr := c.Param("id")

	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID токена")
	}

	ctx := c.Request().Context()

	if err := h.s.Revoke(ctx, userID, tokenID); err != nil {
		slog.WarnContext(ctx, "token revoke failed", "user_id", userID, "token_id", tokenID, "error", err)
		if errors.Is(err, model.ErrTokenNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Токен не найден")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка отзыва токена")
	}
	return c.NoContent(http.StatusOK)
}

func SetupHandlers(e *echo.Echo, s service.TokenService) {
	h := NewToken(s)

	group := e.Group("/link")
	group.Use(session.AuthMiddleware)
	group.GET("/tokens", h.Index)
	group.POST("/tokens", h.Generate)
	group.DELETE("/tokens/:id", h.Revoke)
}
