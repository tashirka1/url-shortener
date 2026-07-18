package storage

import (
	"context"
	"database/sql"
	"testing"

	"url_shortener"
	"url_shortener/internal/apitoken/model"

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

func TestInsertAndFindByHash(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	token, err := r.Insert(context.Background(), 1, "ci-cd", "abc123hash", "sk_abc12")
	require.NoError(t, err)
	assert.NotZero(t, token.ID)
	assert.Equal(t, "ci-cd", token.Name)

	found, err := r.FindByHash(context.Background(), "abc123hash")
	require.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, 1, found.UserID)
	assert.Equal(t, "ci-cd", found.Name)
	assert.Equal(t, "sk_abc12", found.Prefix)
	assert.Nil(t, found.RevokedAt)
}

func TestFindByHash_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	_, err := r.FindByHash(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, model.ErrTokenNotFound)
}

func TestListByUser(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	_, err := r.Insert(context.Background(), 1, "token-a", "hash1", "sk_a")
	require.NoError(t, err)
	_, err = r.Insert(context.Background(), 1, "token-b", "hash2", "sk_b")
	require.NoError(t, err)

	tokens, err := r.ListByUser(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	names := map[string]bool{}
	for _, tok := range tokens {
		names[tok.Name] = true
	}
	assert.True(t, names["token-a"])
	assert.True(t, names["token-b"])
}

func TestListByUser_Empty(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	tokens, err := r.ListByUser(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestListByUser_OnlyOwnTokens(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	_, err := r.Insert(context.Background(), 1, "my-token", "hash1", "sk_a")
	require.NoError(t, err)
	_, err = r.Insert(context.Background(), 2, "other-token", "hash2", "sk_b")
	require.NoError(t, err)

	tokens, err := r.ListByUser(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "my-token", tokens[0].Name)
}

func TestRevoke_Success(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	token, err := r.Insert(context.Background(), 1, "test", "hash", "sk_test")
	require.NoError(t, err)

	err = r.Revoke(context.Background(), 1, token.ID)
	assert.NoError(t, err)

	found, err := r.FindByHash(context.Background(), "hash")
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
	assert.False(t, found.RevokedAt.IsZero())
}

func TestRevoke_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	err := r.Revoke(context.Background(), 1, 999)
	assert.ErrorIs(t, err, model.ErrTokenNotFound)
}

func TestRevoke_WrongUser(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	token, err := r.Insert(context.Background(), 1, "test", "hash", "sk_test")
	require.NoError(t, err)

	err = r.Revoke(context.Background(), 2, token.ID)
	assert.ErrorIs(t, err, model.ErrTokenNotFound)
}

func TestRevoke_AlreadyRevoked(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	token, err := r.Insert(context.Background(), 1, "test", "hash", "sk_test")
	require.NoError(t, err)

	err = r.Revoke(context.Background(), 1, token.ID)
	require.NoError(t, err)

	err = r.Revoke(context.Background(), 1, token.ID)
	assert.ErrorIs(t, err, model.ErrTokenNotFound)
}

func TestUpdateLastUsed(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	token, err := r.Insert(context.Background(), 1, "test", "hash", "sk_test")
	require.NoError(t, err)
	assert.Nil(t, token.LastUsedAt)

	err = r.UpdateLastUsed(context.Background(), token.ID)
	require.NoError(t, err)

	found, err := r.FindByHash(context.Background(), "hash")
	require.NoError(t, err)
	assert.NotNil(t, found.LastUsedAt)
	assert.False(t, found.LastUsedAt.IsZero())
}

func TestRevokedTokensExcludedFromList(t *testing.T) {
	db := setupDB(t)
	r := NewToken(db)

	t1, _ := r.Insert(context.Background(), 1, "active", "hash1", "sk_a")
	_, _ = r.Insert(context.Background(), 1, "active2", "hash2", "sk_b")
	_ = r.Revoke(context.Background(), 1, t1.ID)

	tokens, err := r.ListByUser(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "active2", tokens[0].Name)
}
