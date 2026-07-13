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
	GetAndClick(ctx context.Context, code string) (string, error)
	GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error)
}

type Link struct {
	r storage.LinkStorage
}

func NewLink(r storage.LinkStorage) *Link {
	return &Link{r: r}
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

func (s *Link) GetAndClick(ctx context.Context, code string) (string, error) {
	url, err := s.r.GetAndClick(ctx, code)
	if err != nil {
		slog.WarnContext(ctx, "get and click failed", "code", code, "error", err)
		return "", fmt.Errorf("get and click: %w", err)
	}
	return url, nil
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
