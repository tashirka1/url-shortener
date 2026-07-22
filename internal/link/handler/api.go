package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"url_shortener/internal/apitoken/middleware"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/service"

	"github.com/labstack/echo/v4"
)

type LinkAPI struct {
	s  service.LinkService
	cs service.ClickService
}

func NewLinkAPI(s service.LinkService, cs service.ClickService) *LinkAPI {
	return &LinkAPI{s: s, cs: cs}
}

func (h *LinkAPI) SetupAPIRoutes(g *echo.Group) {
	g.POST("/link", h.PostCreateLink)
	g.GET("/link", h.ListLink)
	g.DELETE("/link/:code", h.DeleteLink)
	g.GET("/link/:code/stats", h.GetStats)
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	Code     string `json:"code"`
	URL      string `json:"url"`
	ShortURL string `json:"short_url"`
}

type listLinkResponse struct {
	Links      []linkItem `json:"links"`
	NextCursor int64      `json:"next_cursor"`
}

type linkItem struct {
	Code      string `json:"code"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	Clicks    int    `json:"clicks"`
}

type deleteLinkResponse struct {
	Deleted string `json:"deleted"`
}

type errorResponse struct {
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
}

type statsResponse struct {
	Code         string         `json:"code"`
	DailyClicks  []dailyClick   `json:"daily_clicks"`
	TopReferrers []referrerItem `json:"top_referrers"`
	TotalClicks  int            `json:"total_clicks"`
}

type dailyClick struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}

type referrerItem struct {
	Referrer string `json:"referrer"`
	Clicks   int    `json:"clicks"`
}

func (h *LinkAPI) PostCreateLink(c echo.Context) error {
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	var req createLinkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body"})
	}

	url, err := validateURL(req.URL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}

	link, err := h.s.CreateLink(ctx, url, userID)
	if err != nil {
		if errors.Is(err, model.ErrLinkAlreadyExists) {
			return c.JSON(http.StatusConflict, errorResponse{Error: "link already exists", Code: link.Code})
		}
		slog.ErrorContext(ctx, "api create link failed", "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	scheme := c.Scheme()
	host := c.Request().Host
	shortURL := scheme + "://" + host + "/" + link.Code

	return c.JSON(http.StatusCreated, createLinkResponse{
		Code:     link.Code,
		URL:      link.Url,
		ShortURL: shortURL,
	})
}

func (h *LinkAPI) ListLink(c echo.Context) error {
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	cursorStr := c.QueryParam("cursor")
	cursor := math.MaxInt64
	if cursorStr != "" {
		if v, err := strconv.Atoi(cursorStr); err == nil {
			cursor = v
		}
	}

	links, err := h.s.ListLink(ctx, userID, cursor)
	if err != nil {
		slog.ErrorContext(ctx, "api list links failed", "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	items := make([]linkItem, 0, len(links))
	var nextCursor int64
	for _, l := range links {
		items = append(items, linkItem{
			Code:      l.Code,
			URL:       l.Url,
			Clicks:    l.Clicks,
			CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
		nextCursor = l.Id
	}

	return c.JSON(http.StatusOK, listLinkResponse{Links: items, NextCursor: nextCursor})
}

func (h *LinkAPI) DeleteLink(c echo.Context) error {
	userID := middleware.GetUserID(c)
	code := c.Param("code")

	err := h.s.RemoveLink(c.Request().Context(), userID, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, errorResponse{Error: "link not found"})
		}
		slog.Error("api delete link failed", "user_id", userID, "code", code, "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	return c.JSON(http.StatusOK, deleteLinkResponse{Deleted: code})
}

func (h *LinkAPI) GetStats(c echo.Context) error {
	userID := middleware.GetUserID(c)
	code := c.Param("code")
	ctx := c.Request().Context()

	link, err := h.s.GetLinkByCode(ctx, code, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, errorResponse{Error: "link not found"})
		}
		slog.ErrorContext(ctx, "api stats: get link failed", "code", code, "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	daily, err := h.cs.GetDailyClicks(ctx, link.Id, 30)
	if err != nil {
		slog.ErrorContext(ctx, "api stats: daily clicks failed", "link_id", link.Id, "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	refs, err := h.cs.GetTopReferrers(ctx, link.Id, 10)
	if err != nil {
		slog.ErrorContext(ctx, "api stats: referrers failed", "link_id", link.Id, "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
	}

	dailyItems := make([]dailyClick, 0, len(daily))
	for _, d := range daily {
		dailyItems = append(dailyItems, dailyClick{Date: d.Day, Clicks: d.Clicks})
	}

	refItems := make([]referrerItem, 0, len(refs))
	for _, r := range refs {
		refItems = append(refItems, referrerItem{Referrer: r.Referrer, Clicks: r.Clicks})
	}

	return c.JSON(http.StatusOK, statsResponse{
		Code:         link.Code,
		TotalClicks:  link.Clicks,
		DailyClicks:  dailyItems,
		TopReferrers: refItems,
	})
}
