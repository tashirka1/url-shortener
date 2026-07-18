package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"url_shortener/internal/apitoken/model"
	"url_shortener/internal/apitoken/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenStorage struct {
	insertFunc         func(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error)
	listByUserFunc     func(ctx context.Context, userID int) ([]model.Token, error)
	revokeFunc         func(ctx context.Context, userID int, tokenID int64) error
	findByHashFunc     func(ctx context.Context, hash string) (model.Token, error)
	updateLastUsedFunc func(ctx context.Context, tokenID int64) error
}

func (m *mockTokenStorage) Insert(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
	if m.insertFunc != nil {
		return m.insertFunc(ctx, userID, name, tokenHash, prefix)
	}
	return model.Token{}, errors.New("Insert not implemented")
}

func (m *mockTokenStorage) ListByUser(ctx context.Context, userID int) ([]model.Token, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(ctx, userID)
	}
	return nil, errors.New("ListByUser not implemented")
}

func (m *mockTokenStorage) Revoke(ctx context.Context, userID int, tokenID int64) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, userID, tokenID)
	}
	return errors.New("Revoke not implemented")
}

func (m *mockTokenStorage) FindByHash(ctx context.Context, hash string) (model.Token, error) {
	if m.findByHashFunc != nil {
		return m.findByHashFunc(ctx, hash)
	}
	return model.Token{}, errors.New("FindByHash not implemented")
}

func (m *mockTokenStorage) UpdateLastUsed(ctx context.Context, tokenID int64) error {
	if m.updateLastUsedFunc != nil {
		return m.updateLastUsedFunc(ctx, tokenID)
	}
	return errors.New("UpdateLastUsed not implemented")
}

var _ storage.TokenStorage = (*mockTokenStorage)(nil)

func TestGenerate_Success(t *testing.T) {
	mock := &mockTokenStorage{
		insertFunc: func(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, "ci-cd", name)
			assert.Len(t, tokenHash, 64) // SHA-256 hex = 64 chars
			assert.Equal(t, "sk_", prefix[:3])
			return model.Token{ID: 1, UserID: userID, Name: name, TokenHash: tokenHash, Prefix: prefix}, nil
		},
	}
	s := NewToken(mock)

	token, raw, err := s.Generate(context.Background(), 1, "ci-cd")

	require.NoError(t, err)
	assert.Equal(t, int64(1), token.ID)
	assert.Equal(t, "ci-cd", token.Name)
	assert.Len(t, raw, 3+32) // "sk_" + 32 base62 chars
	assert.Equal(t, "sk_", raw[:3])
}

func TestGenerate_EmptyName(t *testing.T) {
	mock := &mockTokenStorage{
		insertFunc: func(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
			return model.Token{ID: 1}, nil
		},
	}
	s := NewToken(mock)

	token, raw, err := s.Generate(context.Background(), 1, "")

	require.NoError(t, err)
	assert.Equal(t, int64(1), token.ID)
	assert.NotEmpty(t, raw)
}

func TestListByUser_Success(t *testing.T) {
	expected := []model.Token{
		{ID: 1, Name: "token-1", Prefix: "sk_abc12"},
		{ID: 2, Name: "token-2", Prefix: "sk_def34"},
	}
	mock := &mockTokenStorage{
		listByUserFunc: func(ctx context.Context, userID int) ([]model.Token, error) {
			assert.Equal(t, 1, userID)
			return expected, nil
		},
	}
	s := NewToken(mock)

	tokens, err := s.ListByUser(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "token-1", tokens[0].Name)
}

func TestListByUser_Empty(t *testing.T) {
	mock := &mockTokenStorage{
		listByUserFunc: func(ctx context.Context, userID int) ([]model.Token, error) {
			return []model.Token{}, nil
		},
	}
	s := NewToken(mock)

	tokens, err := s.ListByUser(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, tokens)
}

func TestRevoke_Success(t *testing.T) {
	mock := &mockTokenStorage{
		revokeFunc: func(ctx context.Context, userID int, tokenID int64) error {
			assert.Equal(t, 1, userID)
			assert.Equal(t, int64(1), tokenID)
			return nil
		},
	}
	s := NewToken(mock)

	err := s.Revoke(context.Background(), 1, 1)

	assert.NoError(t, err)
}

func TestRevoke_NotFound(t *testing.T) {
	mock := &mockTokenStorage{
		revokeFunc: func(ctx context.Context, userID int, tokenID int64) error {
			return model.ErrTokenNotFound
		},
	}
	s := NewToken(mock)

	err := s.Revoke(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrTokenNotFound))
}

func TestAuthenticate_Success(t *testing.T) {
	var savedHash string
	insertMock := &mockTokenStorage{
		insertFunc: func(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
			savedHash = tokenHash
			return model.Token{ID: 1, UserID: userID, Name: name, TokenHash: tokenHash, Prefix: prefix}, nil
		},
	}
	s := NewToken(insertMock)

	_, raw, err := s.Generate(context.Background(), 1, "test-token")
	require.NoError(t, err)
	require.NotEmpty(t, savedHash)

	mock := &mockTokenStorage{
		findByHashFunc: func(ctx context.Context, hash string) (model.Token, error) {
			assert.Equal(t, savedHash, hash)
			return model.Token{ID: 1, UserID: 1, Name: "test-token", TokenHash: hash}, nil
		},
		updateLastUsedFunc: func(ctx context.Context, tokenID int64) error {
			return nil
		},
	}
	s2 := NewToken(mock)

	result, err := s2.Authenticate(context.Background(), raw)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, 1, result.UserID)
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	mock := &mockTokenStorage{
		findByHashFunc: func(ctx context.Context, hash string) (model.Token, error) {
			return model.Token{}, model.ErrTokenNotFound
		},
	}
	s := NewToken(mock)

	_, err := s.Authenticate(context.Background(), "sk_invalidtoken123")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrTokenNotFound))
}

func TestAuthenticate_RevokedToken(t *testing.T) {
	revokedAt := time.Now()
	mock := &mockTokenStorage{
		findByHashFunc: func(ctx context.Context, hash string) (model.Token, error) {
			return model.Token{ID: 1, RevokedAt: &revokedAt}, nil
		},
	}
	s := NewToken(mock)

	_, err := s.Authenticate(context.Background(), "sk_revokedtoken")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrTokenRevoked))
}
