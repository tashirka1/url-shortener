package handler

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"url_shortener/internal/core/session"
	core_view "url_shortener/internal/core/view"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/service"
	"url_shortener/internal/link/view"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"
)

type Link struct {
	s  service.LinkService
	cs service.ClickService
}

func NewLink(s service.LinkService, cs service.ClickService) *Link {
	return &Link{s: s, cs: cs}
}

func validateURL(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("ссылка обязательна")
	}
	if len(raw) > model.MaxURLLength {
		return "", errors.New("ссылка слишком длинная")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("только http и https")
	}
	if u.Host == "" || !strings.Contains(u.Host, ".") {
		return "", errors.New("invalid url")
	}
	return raw, nil
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

	url, err := validateURL(url)
	if err != nil {
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
		return core_view.RenderTemplate(c, view.CreateLinkError("эта ссылка уже существует"))
	}
	if err != nil {
		slog.Error("failed to create link", "user_id", userId, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка создания ссылки")
	}
	c.Response().Header().Set("HX-Trigger", "reset-create-form")
	return core_view.RenderTemplate(c, view.CreateLinkSuccess(link))
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
	ref := c.Request().Referer()
	url, err := h.s.GetAndClick(c.Request().Context(), code, ref, "")
	if err != nil {
		slog.Warn("link not found", "code", code, "error", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Ссылка не найдена")
	}
	return c.Redirect(http.StatusSeeOther, url)
}

func (h *Link) SearchLink(c echo.Context) error {
	userId := session.GetUserId(c)
	query := c.QueryParam("q")

	if query == "" {
		links, err := h.s.ListLink(c.Request().Context(), userId, math.MaxInt64)
		if err != nil {
			slog.Error("failed to list links on empty search", "user_id", userId, "error", err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Внутренняя ошибка")
		}
		return core_view.RenderTemplate(c, view.ListLink(links))
	}

	links, err := h.s.SearchLink(c.Request().Context(), userId, query)
	if err != nil {
		slog.Error("search failed", "user_id", userId, "query", query, "error", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
	}

	if len(links) == 0 {
		return c.HTML(http.StatusOK, `<tr><td colspan="4" style="text-align:center;">Ничего не найдено</td></tr>`)
	}
	return core_view.RenderTemplate(c, view.SearchResults(links))
}

func (h *Link) GetCodeDisplay(c echo.Context) error {
	userId := session.GetUserId(c)
	code := c.Param("code")

	_, err := h.s.GetLinkByCode(c.Request().Context(), code, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Ссылка не найдена")
	}

	return core_view.RenderTemplate(c, view.CodeDisplay(code))
}

func (h *Link) GetEditForm(c echo.Context) error {
	userId := session.GetUserId(c)
	currentCode := c.Param("code")

	link, err := h.s.GetLinkByCode(c.Request().Context(), currentCode, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Ссылка не найдена")
	}

	return core_view.RenderTemplate(c, view.CodeEditField(link.Code, ""))
}

func (h *Link) UpdateAlias(c echo.Context) error {
	userId := session.GetUserId(c)
	currentCode := c.Param("code")
	newCode := c.FormValue("code")

	if err := h.s.UpdateAlias(c.Request().Context(), userId, currentCode, newCode); err != nil {
		return core_view.RenderTemplate(c, view.CodeEditField(currentCode, err.Error()))
	}

	link, err := h.s.GetLinkByCode(c.Request().Context(), newCode, userId)
	if err != nil {
		slog.Error("update alias: failed to fetch updated link", "code", newCode, "user_id", userId, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка обновления")
	}
	return core_view.RenderTemplate(c, view.Link(link))
}

func (h *Link) GetQRCode(c echo.Context) error {
	userId := session.GetUserId(c)
	code := c.Param("code")

	link, err := h.s.GetLinkByCode(c.Request().Context(), code, userId)
	if err != nil {
		slog.Warn("qr code: link not found", "code", code, "user_id", userId, "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "Ссылка не найдена")
	}

	png, err := qrcode.Encode(link.Url, qrcode.Medium, 256)
	if err != nil {
		slog.Error("qr encode failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Ошибка генерации QR")
	}

	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return core_view.RenderTemplate(c, view.QRCodeModal(link.Code, dataURI))
}

func (h *Link) Main(c echo.Context) error {
	userId := session.GetUserId(c)
	return core_view.RenderTemplate(c, view.Main(userId))
}

func SetupHandlers(e *echo.Echo, s service.LinkService, cs service.ClickService) {
	h := NewLink(s, cs)

	group := e.Group("/link")
	group.Use(session.AuthMiddleware)
	group.GET("/create-link", h.GetCreateLink)
	group.POST("/create-link", h.PostCreateLink)
	group.GET("/list-link", h.ListLink)
	group.GET("/search-link", h.SearchLink)
	group.GET("/:code/display", h.GetCodeDisplay)
	group.GET("/:code/edit-form", h.GetEditForm)
	group.PATCH("/:code/alias", h.UpdateAlias)
	group.GET("/:code/qr", h.GetQRCode)
	group.GET("/:code/stats", h.Stats)
	group.DELETE("/remove-link/:code", h.RemoveLink)
	e.GET("/:code", h.RedirectLink)
	e.GET("/", h.Main)
}
