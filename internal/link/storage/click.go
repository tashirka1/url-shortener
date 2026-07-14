package storage

import (
	"context"
	"database/sql"
	"log/slog"
	"url_shortener/internal/link/model"
)

type ClickStorage interface {
	GetDailyClicks(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error)
	GetTopReferrers(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error)
}

type Click struct {
	db *sql.DB
}

func NewClick(db *sql.DB) *Click {
	return &Click{db: db}
}

func (r *Click) GetDailyClicks(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT date(clicked_at) as day, COUNT(*) as clicks
		FROM link_click
		WHERE link_id = ? AND clicked_at >= datetime('now', '-' || ? || ' days')
		GROUP BY day
		ORDER BY day
	`, linkID, days)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close rows", "error", err)
		}
	}()
	var result []model.DailyClick
	for rows.Next() {
		var d model.DailyClick
		if err := rows.Scan(&d.Day, &d.Clicks); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Click) GetTopReferrers(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT referrer, COUNT(*) as clicks
		FROM link_click
		WHERE link_id = ?
		GROUP BY referrer
		ORDER BY clicks DESC
		LIMIT ?
	`, linkID, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close rows", "error", err)
		}
	}()
	var result []model.ReferrerStat
	for rows.Next() {
		var r model.ReferrerStat
		if err := rows.Scan(&r.Referrer, &r.Clicks); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
