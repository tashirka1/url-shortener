package handler

import (
	"log/slog"
	"net/http"
	"strconv"

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

	if name == "" {
		c.Response().Header().Set("HX-Retarget", "#token-errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return c.HTML(http.StatusBadRequest, "Название токена обязательно")
	}

	ctx := c.Request().Context()

	token, rawToken, err := h.s.Generate(ctx, userID, name)
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

	return core_view.RenderTemplate(c, view.TokenListAndDisplay(tokens, rawToken, token.Name))
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
