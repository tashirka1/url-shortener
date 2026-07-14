package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDailyClicks_Empty(t *testing.T) {
	db := setupDB(t)
	db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES ('c1', 'https://a.com', 0, 1)")
	r := NewClick(db)

	result, err := r.GetDailyClicks(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetDailyClicks_Success(t *testing.T) {
	db := setupDB(t)
	db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES ('c1', 'https://a.com', 0, 1)")
	db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://google.com')")
	db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://twitter.com')")

	r := NewClick(db)
	result, err := r.GetDailyClicks(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 2, result[0].Clicks)
}

func TestGetTopReferrers_Empty(t *testing.T) {
	db := setupDB(t)
	db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES ('c1', 'https://a.com', 0, 1)")
	r := NewClick(db)

	result, err := r.GetTopReferrers(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetTopReferrers_Success(t *testing.T) {
	db := setupDB(t)
	db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES ('c1', 'https://a.com', 0, 1)")
	for range 5 {
		db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://google.com')")
	}
	for range 3 {
		db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://twitter.com')")
	}

	r := NewClick(db)
	result, err := r.GetTopReferrers(context.Background(), 1, 10)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "https://google.com", result[0].Referrer)
	assert.Equal(t, 5, result[0].Clicks)
	assert.Equal(t, "https://twitter.com", result[1].Referrer)
	assert.Equal(t, 3, result[1].Clicks)
}

func TestGetTopReferrers_Limit(t *testing.T) {
	db := setupDB(t)
	db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES ('c1', 'https://a.com', 0, 1)")
	for range 5 {
		db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://google.com')")
	}
	for range 3 {
		db.Exec("INSERT INTO link_click(link_id, referrer) VALUES (1, 'https://twitter.com')")
	}

	r := NewClick(db)
	result, err := r.GetTopReferrers(context.Background(), 1, 1)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}
