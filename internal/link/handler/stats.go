package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"
	"url_shortener/internal/link/view"

	"github.com/labstack/echo/v4"
)

func (h *Link) Stats(c echo.Context) error {
	userId := session.GetUserId(c)
	code := c.Param("code")
	ctx := c.Request().Context()

	link, err := h.s.GetLinkByCode(ctx, code, userId)
	if err != nil {
		slog.Warn("stats: link not found or access denied", "code", code, "user_id", userId, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "Ссылка не найдена")
		}
		return echo.NewHTTPError(http.StatusForbidden, "Доступ запрещён")
	}

	daily, err := h.cs.GetDailyClicks(ctx, link.Id, 30)
	if err != nil {
		slog.Error("stats: failed to get daily clicks", "link_id", link.Id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка загрузки статистики")
	}

	referrers, err := h.cs.GetTopReferrers(ctx, link.Id, 10)
	if err != nil {
		slog.Error("stats: failed to get referrers", "link_id", link.Id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка загрузки статистики")
	}

	return core_view.RenderTemplate(c, view.StatsPage(userId, link, daily, referrers))
}
