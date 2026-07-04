package storage

import (
	"context"
	"database/sql"
	"testing"
	"url_shortener"
	"url_shortener/internal/auth/model"

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

func TestCheckEmail_Found(t *testing.T) {
	db := setupDB(t)
	r := NewUser(db)

	_, err := db.Exec("INSERT INTO auth_user(email, password) VALUES (?, ?)", "test@example.com", "hash1")
	require.NoError(t, err)

	user, err := r.CheckEmail(context.Background(), "test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "hash1", user.Password)
	assert.NotZero(t, user.Id)
}

func TestCheckEmail_NotFound(t *testing.T) {
	db := setupDB(t)
	r := NewUser(db)

	_, err := r.CheckEmail(context.Background(), "missing@example.com")

	assert.ErrorIs(t, err, model.ErrUserNotFound)
}

func TestCreateUser_Success(t *testing.T) {
	db := setupDB(t)
	r := NewUser(db)

	err := r.CreateUser(context.Background(), "new@example.com", []byte("hash2"))

	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM auth_user WHERE email=?", "new@example.com").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateUser_Duplicate(t *testing.T) {
	db := setupDB(t)
	r := NewUser(db)

	err := r.CreateUser(context.Background(), "same@example.com", []byte("hash3"))
	require.NoError(t, err)

	err = r.CreateUser(context.Background(), "same@example.com", []byte("hash4"))

	assert.ErrorIs(t, err, model.ErrUserAlreadyExists)
}
