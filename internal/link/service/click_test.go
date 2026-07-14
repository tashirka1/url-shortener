package service

import (
	"context"
	"errors"
	"testing"

	"url_shortener/internal/link/model"

	"github.com/stretchr/testify/assert"
)

func TestGetDailyClicks_Success(t *testing.T) {
	expected := []model.DailyClick{
		{Day: "2026-07-01", Clicks: 10},
		{Day: "2026-07-02", Clicks: 5},
	}
	mock := &mockClickStorage{
		getDailyClicksFunc: func(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
			assert.Equal(t, int64(1), linkID)
			assert.Equal(t, 30, days)
			return expected, nil
		},
	}
	s := NewClick(mock)

	result, err := s.GetDailyClicks(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetDailyClicks_Empty(t *testing.T) {
	mock := &mockClickStorage{
		getDailyClicksFunc: func(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
			return []model.DailyClick{}, nil
		},
	}
	s := NewClick(mock)

	result, err := s.GetDailyClicks(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetDailyClicks_StorageError(t *testing.T) {
	mock := &mockClickStorage{
		getDailyClicksFunc: func(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
			return nil, errors.New("db error")
		},
	}
	s := NewClick(mock)

	_, err := s.GetDailyClicks(context.Background(), 1, 30)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetTopReferrers_Success(t *testing.T) {
	expected := []model.ReferrerStat{
		{Referrer: "https://google.com", Clicks: 20},
		{Referrer: "https://twitter.com", Clicks: 5},
	}
	mock := &mockClickStorage{
		getTopReferrersFunc: func(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
			assert.Equal(t, int64(1), linkID)
			assert.Equal(t, 10, limit)
			return expected, nil
		},
	}
	s := NewClick(mock)

	result, err := s.GetTopReferrers(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetTopReferrers_Empty(t *testing.T) {
	mock := &mockClickStorage{
		getTopReferrersFunc: func(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
			return []model.ReferrerStat{}, nil
		},
	}
	s := NewClick(mock)

	result, err := s.GetTopReferrers(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetTopReferrers_StorageError(t *testing.T) {
	mock := &mockClickStorage{
		getTopReferrersFunc: func(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
			return nil, errors.New("db error")
		},
	}
	s := NewClick(mock)

	_, err := s.GetTopReferrers(context.Background(), 1, 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
