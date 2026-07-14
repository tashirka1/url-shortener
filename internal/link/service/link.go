package service

import (
	"context"
	"fmt"
	"log/slog"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/storage"
)

type LinkService interface {
	CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
	ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
	RemoveLink(ctx context.Context, userId int, code string) error
	SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error)
	GetAndClick(ctx context.Context, code, referrer, userAgent string) (string, error)
	GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error)
	UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error
}

type Link struct {
	r  storage.LinkStorage
	cr storage.ClickStorage
}

func NewLink(r storage.LinkStorage, cr storage.ClickStorage) *Link {
	return &Link{r: r, cr: cr}
}

func (s *Link) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
	link, err := s.r.CreateLink(ctx, url, userId)
	if err != nil {
		slog.ErrorContext(ctx, "create link failed", "user_id", userId, "url", url, "error", err)
		return model.Link{}, fmt.Errorf("create link: %w", err)
	}
	slog.InfoContext(ctx, "link created", "user_id", userId, "code", link.Code, "url", url)
	return link, nil
}

func (s *Link) ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error) {
	links, err := s.r.ListLink(ctx, userId, cursor)
	if err != nil {
		slog.ErrorContext(ctx, "list links failed", "user_id", userId, "cursor", cursor, "error", err)
		return nil, fmt.Errorf("list links: %w", err)
	}
	slog.InfoContext(ctx, "links listed", "user_id", userId, "cursor", cursor, "count", len(links))
	return links, nil
}

func (s *Link) RemoveLink(ctx context.Context, userId int, code string) error {
	if err := s.r.RemoveLink(ctx, userId, code); err != nil {
		slog.ErrorContext(ctx, "remove link failed", "user_id", userId, "code", code, "error", err)
		return fmt.Errorf("remove link: %w", err)
	}
	slog.InfoContext(ctx, "link removed", "user_id", userId, "code", code)
	return nil
}

func (s *Link) GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error) {
	link, err := s.r.GetLinkByCode(ctx, code, userId)
	if err != nil {
		slog.WarnContext(ctx, "get link by code failed", "code", code, "user_id", userId, "error", err)
		return model.Link{}, fmt.Errorf("get link by code: %w", err)
	}
	return link, nil
}

func (s *Link) UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error {
	if newCode == "" {
		return fmt.Errorf("alias is required")
	}
	if len(newCode) < 3 || len(newCode) > 32 {
		return fmt.Errorf("alias должен быть от 3 до 32 символов")
	}
	for _, r := range newCode {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return fmt.Errorf("alias может содержать только латинские буквы, цифры, дефис и подчёркивание")
		}
	}
	reserved := map[string]bool{"health": true, "auth": true, "link": true, "rps": true, "static": true}
	if reserved[newCode] {
		return fmt.Errorf("этот alias зарезервирован системой")
	}
	if err := s.r.UpdateAlias(ctx, userID, currentCode, newCode); err != nil {
		slog.ErrorContext(ctx, "update alias failed", "user_id", userID, "current_code", currentCode, "new_code", newCode, "error", err)
		return fmt.Errorf("update alias: %w", err)
	}
	slog.InfoContext(ctx, "alias updated", "user_id", userID, "current_code", currentCode, "new_code", newCode)
	return nil
}

func (s *Link) GetAndClick(ctx context.Context, code, referrer, userAgent string) (string, error) {
	link, err := s.r.GetAndClick(ctx, code, referrer, userAgent)
	if err != nil {
		slog.WarnContext(ctx, "get and click failed", "code", code, "error", err)
		return "", fmt.Errorf("get and click: %w", err)
	}
	return link.Url, nil
}

func (s *Link) SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error) {
	links, err := s.r.SearchLink(ctx, userId, query)
	if err != nil {
		slog.ErrorContext(ctx, "search links failed", "user_id", userId, "query", query, "error", err)
		return nil, fmt.Errorf("search links: %w", err)
	}
	slog.InfoContext(ctx, "links searched", "user_id", userId, "query", query, "count", len(links))
	return links, nil
}
