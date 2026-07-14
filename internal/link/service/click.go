package service

import (
	"context"
	"fmt"
	"log/slog"
	"url_shortener/internal/link/model"
	"url_shortener/internal/link/storage"
)

type ClickService interface {
	GetDailyClicks(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error)
	GetTopReferrers(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error)
}

type Click struct {
	r storage.ClickStorage
}

func NewClick(r storage.ClickStorage) *Click {
	return &Click{r: r}
}

func (s *Click) GetDailyClicks(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
	result, err := s.r.GetDailyClicks(ctx, linkID, days)
	if err != nil {
		slog.ErrorContext(ctx, "get daily clicks failed", "link_id", linkID, "days", days, "error", err)
		return nil, fmt.Errorf("get daily clicks: %w", err)
	}
	return result, nil
}

func (s *Click) GetTopReferrers(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
	result, err := s.r.GetTopReferrers(ctx, linkID, limit)
	if err != nil {
		slog.ErrorContext(ctx, "get top referrers failed", "link_id", linkID, "limit", limit, "error", err)
		return nil, fmt.Errorf("get top referrers: %w", err)
	}
	return result, nil
}
