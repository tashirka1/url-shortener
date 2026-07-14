package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"url_shortener"
	"url_shortener/internal/link/model"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	goose.SetBaseFS(url_shortener.EmbeddedMigrations)
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.Up(db, "migrations"))

	return db
}

var userSeq int

func addUser(t *testing.T, db *sql.DB) int {
	t.Helper()
	userSeq++
	email := fmt.Sprintf("user%d@test.com", userSeq)
	res, err := db.Exec("INSERT INTO auth_user(email, password) VALUES (?, ?)", email, "hash")
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

func insertLink(t *testing.T, db *sql.DB, code, url string, userId int) {
	t.Helper()
	_, err := db.Exec("INSERT INTO link_link(code, url, clicks, user_id) VALUES (?, ?, 0, ?)", code, url, userId)
	require.NoError(t, err)
}

func TestCreateLink_Success(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	link, err := r.CreateLink(context.Background(), "https://example.com", userId)

	assert.NoError(t, err)
	assert.NotZero(t, link.Id)
	assert.NotEmpty(t, link.Code)
	assert.Equal(t, "https://example.com", link.Url)
}

func TestCreateLink_Duplicate(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	_, err := r.CreateLink(context.Background(), "https://example.com", userId)
	require.NoError(t, err)

	_, err = r.CreateLink(context.Background(), "https://example.com", userId)

	assert.ErrorIs(t, err, model.ErrLinkAlreadyExists)
}

func TestListLink_Empty(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	links, err := r.ListLink(context.Background(), userId, 999)

	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestListLink_Pagination(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	insertLink(t, db, "c1", "https://a.com", userId)
	insertLink(t, db, "c2", "https://b.com", userId)
	insertLink(t, db, "c3", "https://c.com", userId)

	links, err := r.ListLink(context.Background(), userId, 100)
	assert.NoError(t, err)
	assert.Len(t, links, 3)
	assert.Equal(t, "https://c.com", links[0].Url)

	links2, err := r.ListLink(context.Background(), userId, int(links[len(links)-1].Id))
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(links2), 2)
}

func TestListLink_OtherUser(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId1 := addUser(t, db)
	userId2 := addUser(t, db)

	insertLink(t, db, "c1", "https://a.com", userId1)

	links, err := r.ListLink(context.Background(), userId2, 999)
	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestRemoveLink_Success(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	insertLink(t, db, "c1", "https://a.com", userId)

	err := r.RemoveLink(context.Background(), userId, "c1")
	assert.NoError(t, err)

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM link_link WHERE code=?", "c1").Scan(&count))
	assert.Equal(t, 0, count)
}

func TestRemoveLink_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	err := r.RemoveLink(context.Background(), userId, "missing")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestRemoveLink_WrongUser(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId1 := addUser(t, db)
	userId2 := addUser(t, db)

	insertLink(t, db, "c1", "https://a.com", userId1)

	err := r.RemoveLink(context.Background(), userId2, "c1")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetAndClick_Success(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	insertLink(t, db, "c1", "https://example.com", userId)

	link, err := r.GetAndClick(context.Background(), "c1", "", "")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", link.Url)

	var clicks int
	require.NoError(t, db.QueryRow("SELECT clicks FROM link_link WHERE code=?", "c1").Scan(&clicks))
	assert.Equal(t, 1, clicks)

	var clickCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM link_click WHERE link_id=?", link.Id).Scan(&clickCount))
	assert.Equal(t, 1, clickCount)
}

func TestGetAndClick_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)

	_, err := r.GetAndClick(context.Background(), "missing", "", "")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSearchLink_Success(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	insertLink(t, db, "c1", "https://example.com", userId)

	links, err := r.SearchLink(context.Background(), userId, "example")
	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, "https://example.com", links[0].Url)
}

func TestSearchLink_NoResults(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	links, err := r.SearchLink(context.Background(), userId, "nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestUpdateAlias_Success(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)
	insertLink(t, db, "abc123", "https://example.com", userId)

	err := r.UpdateAlias(context.Background(), userId, "abc123", "my-link")

	assert.NoError(t, err)
	var code string
	require.NoError(t, db.QueryRow("SELECT code FROM link_link WHERE user_id=?", userId).Scan(&code))
	assert.Equal(t, "my-link", code)
}

func TestUpdateAlias_AlreadyTaken(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	user1 := addUser(t, db)
	user2 := addUser(t, db)
	insertLink(t, db, "abc123", "https://a.com", user1)
	insertLink(t, db, "def456", "https://b.com", user2)

	err := r.UpdateAlias(context.Background(), user1, "abc123", "def456")

	assert.ErrorIs(t, err, model.ErrAliasTaken)
}

func TestUpdateAlias_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)

	err := r.UpdateAlias(context.Background(), userId, "missing", "new-code")

	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestUpdateAlias_WrongUser(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	user1 := addUser(t, db)
	user2 := addUser(t, db)
	insertLink(t, db, "abc123", "https://a.com", user1)

	err := r.UpdateAlias(context.Background(), user2, "abc123", "my-link")

	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestUpdateAlias_SameCode(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId := addUser(t, db)
	insertLink(t, db, "abc123", "https://example.com", userId)

	err := r.UpdateAlias(context.Background(), userId, "abc123", "abc123")

	assert.NoError(t, err)
}

func TestSearchLink_OtherUser(t *testing.T) {
	db := setupDB(t)
	r := NewLink(db)
	userId1 := addUser(t, db)
	userId2 := addUser(t, db)

	insertLink(t, db, "c1", "https://example.com", userId1)

	links, err := r.SearchLink(context.Background(), userId2, "example")
	assert.NoError(t, err)
	assert.Empty(t, links)
}
