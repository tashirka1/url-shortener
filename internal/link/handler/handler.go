package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/service"
	"url_shortener/internal/link/storage"
	"url_shortener/internal/link/view"

	"github.com/labstack/echo/v4"
)

type Link struct {
	s service.LinkService
}

func NewLink(s service.LinkService) *Link {
	return &Link{s: s}
}

func validateURL(url string) error {
	if url == "" {
		return errors.New("url is required")
	}
	if len(url) > model.MaxURLLength {
		return errors.New("url is too long")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("url must start with http:// or https://")
	}
	return nil
}

func (h *Link) GetCreateLink(c echo.Context) error {
	userId := session.GetUserId(c)
	links, err := h.s.ListLink(c.Request().Context(), userId, math.MaxInt64)
	if err != nil {
		slog.Error("failed to list links", "user_id", userId, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
	}
	return core_view.RenderTemplate(c, view.CreateLink(userId, links))
}

func (h *Link) PostCreateLink(c echo.Context) error {
	userId := session.GetUserId(c)
	url := c.FormValue("url")

	if err := validateURL(url); err != nil {
		slog.Warn("validation error", "user_id", userId, "error", err.Error())
		c.Response().Header().Set("HX-Retarget", "#create-link-errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return core_view.RenderTemplate(c, view.CreateLinkError(err.Error()))
	}

	link, err := h.s.CreateLink(c.Request().Context(), url, userId)
	if errors.Is(err, model.ErrLinkAlreadyExists) {
		slog.Warn("duplicate link", "user_id", userId, "url", url)
		c.Response().Header().Set("HX-Retarget", "#create-link-errors")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return core_view.RenderTemplate(c, view.CreateLinkError("this URL already exists"))
	}
	if err != nil {
		slog.Error("failed to create link", "user_id", userId, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create link")
	}
	return core_view.RenderTemplate(c, view.Link(link))
}

func (h *Link) ListLink(c echo.Context) error {
	userId := session.GetUserId(c)
	cursor, err := strconv.Atoi(c.QueryParam("cursor"))
	if err != nil {
		cursor = 0
	}
	links, err := h.s.ListLink(c.Request().Context(), userId, cursor)
	if err != nil {
		slog.Error("failed to list links", "user_id", userId, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
	}
	return core_view.RenderTemplate(c, view.ListLink(links))
}

func (h *Link) RemoveLink(c echo.Context) error {
	userId := session.GetUserId(c)
	code := c.Param("code")
	err := h.s.RemoveLink(c.Request().Context(), userId, code)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Warn("link not found for removal", "user_id", userId, "code", code)
		return c.NoContent(http.StatusOK)
	}
	if err != nil {
		slog.Error("failed to remove link", "user_id", userId, "code", code, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
	}
	return nil
}

func (h *Link) RedirectLink(c echo.Context) error {
	code := c.Param("code")
	url, err := h.s.GetLink(c.Request().Context(), code)
	if err != nil {
		slog.Warn("link not found", "code", code, "error", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Link not found")
	}
	if url != "" {
		if err := h.s.ClickLink(c.Request().Context(), code); err != nil {
			slog.Error("failed to increment click", "code", code, "error", err.Error())
		}
		return c.Redirect(http.StatusSeeOther, url)
	}
	return echo.NewHTTPError(http.StatusNotFound, "Link not found")
}

func (h *Link) Main(c echo.Context) error {
	return core_view.RenderTemplate(c, view.Main())
}

func SetupHandlers(e *echo.Echo, db *sql.DB) {
	storage := storage.NewLink(db)
	service := service.NewLink(storage)
	handler := NewLink(service)

	group := e.Group("/link")
	group.Use(session.AuthMiddleware)
	group.GET("/create-link", handler.GetCreateLink)
	group.POST("/create-link", handler.PostCreateLink)
	group.GET("/list-link", handler.ListLink)
	group.DELETE("/remove-link/:code", handler.RemoveLink)
	e.GET("/:code", handler.RedirectLink)
	e.GET("/", handler.Main)
}
