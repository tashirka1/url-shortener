package handler

import (
	"net/http"
	"strconv"
	"time"
	core_view "url_shortener/internal/core/view"
	"url_shortener/internal/rps/storage"
	"url_shortener/internal/rps/view"

	"github.com/labstack/echo/v4"
)

type RPS struct {
	storage storage.RPSStorage
}

func NewRPS(storage storage.RPSStorage) *RPS {
	return &RPS{storage: storage}
}

func (h *RPS) SetupRoutes(e *echo.Echo) {
	e.GET("/rps/simple-text", h.SimpleText)
	e.GET("/rps/simple-json", h.SimpleJSON)
	e.GET("/rps/simple-templ-page", h.SimpleTemplPage)
	e.GET("/rps/templ-page-insert", h.TemplPageInsert)
	e.GET("/rps/templ-page-select-join", h.TemplPageSelectJoin)
	e.GET("/rps/templ-page-select-join-update", h.TemplPageSelectJoinUpdate)
}

func (h *RPS) SimpleText(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func (h *RPS) SimpleJSON(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RPS) SimpleTemplPage(c echo.Context) error {
	return core_view.RenderTemplate(c, view.SimplePage())
}

func (h *RPS) TemplPageInsert(c echo.Context) error {
	payload := c.QueryParam("payload")
	if payload == "" {
		payload = "rps_bench"
	}
	ts := time.Now().Unix()

	ctx := c.Request().Context()
	id, err := h.storage.Insert(ctx, payload, ts, 0)
	if err != nil {
		return err
	}

	return core_view.RenderTemplate(c, view.InsertPage(id, payload))
}

func (h *RPS) TemplPageSelectJoin(c echo.Context) error {
	limit := parseLimit(c.QueryParam("limit"))

	ctx := c.Request().Context()
	rows, err := h.storage.SelectJoin(ctx, limit)
	if err != nil {
		return err
	}

	return core_view.RenderTemplate(c, view.SelectJoinPage(rows))
}

func (h *RPS) TemplPageSelectJoinUpdate(c echo.Context) error {
	limit := parseLimit(c.QueryParam("limit"))

	ctx := c.Request().Context()
	rows, err := h.storage.SelectJoin(ctx, limit)
	if err != nil {
		return err
	}

	ids := make([]int64, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
		rows[i].Duration++ // отражает UPDATE, чтобы не делать второй SELECT
	}
	if bulkErr := h.storage.BulkUpdateDuration(ctx, ids); bulkErr != nil {
		return bulkErr
	}

	return core_view.RenderTemplate(c, view.SelectJoinPage(rows))
}

func parseLimit(s string) int {
	if s == "" {
		return 10
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 10
	}
	return n
}
