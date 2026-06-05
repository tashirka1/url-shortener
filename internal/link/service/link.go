package service

import (
	"context"
	"log/slog"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/storage"
)

type LinkService interface {
	CreateLink(ctx context.Context, url string, userId int) (model.Link, error)
	ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error)
	RemoveLink(ctx context.Context, userId int, code string) error
	GetLink(ctx context.Context, code string) (string, error)
	ClickLink(ctx context.Context, code string) error
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
		return model.Link{}, err
	}
	slog.InfoContext(ctx, "link created", "user_id", userId, "code", link.Code, "url", url)
	return link, nil
}

func (s *Link) ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error) {
	links, err := s.r.ListLink(ctx, userId, cursor)
	if err != nil {
		slog.ErrorContext(ctx, "list links failed", "user_id", userId, "cursor", cursor, "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "links listed", "user_id", userId, "cursor", cursor, "count", len(links))
	return links, nil
}

func (s *Link) RemoveLink(ctx context.Context, userId int, code string) error {
	if err := s.r.RemoveLink(ctx, userId, code); err != nil {
		slog.ErrorContext(ctx, "remove link failed", "user_id", userId, "code", code, "error", err)
		return err
	}
	slog.InfoContext(ctx, "link removed", "user_id", userId, "code", code)
	return nil
}

func (s *Link) GetLink(ctx context.Context, code string) (string, error) {
	url, err := s.r.GetLink(ctx, code)
	if err != nil {
		slog.WarnContext(ctx, "get link failed", "code", code, "error", err)
		return "", err
	}
	return url, nil
}

func (s *Link) ClickLink(ctx context.Context, code string) error {
	if err := s.r.ClickLink(ctx, code); err != nil {
		slog.ErrorContext(ctx, "click link failed", "code", code, "error", err)
		return err
	}
	return nil
}
