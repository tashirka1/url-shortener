package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func TestGenerate(t *testing.T) {
	tests := []struct {
		mock    *mockTokenStorage
		name    string
		tokenNm string
		wantNm  string
		wantID  int64
		userID  int
		wantErr bool
	}{
		{
			name:    "success",
			userID:  1,
			tokenNm: "ci-cd",
			mock: &mockTokenStorage{
				insertFunc: func(ctx context.Context, userID int, name, tokenHash, prefix string) (model.Token, error) {
					assert.Equal(t, 1, userID)
					assert.Equal(t, "ci-cd", name)
					assert.Len(t, tokenHash, 64)
					assert.Equal(t, "sk_", prefix[:3])
					return model.Token{ID: 1, UserID: userID, Name: name, TokenHash: tokenHash, Prefix: prefix}, nil
				},
			},
			wantID: 1,
			wantNm: "ci-cd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewToken(tt.mock)
			token, raw, err := s.Generate(context.Background(), tt.userID, tt.tokenNm)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, token.ID)
			assert.Equal(t, tt.wantNm, token.Name)
			assert.Len(t, raw, 3+tokenBase62Len)
			assert.Equal(t, "sk_", raw[:3])
		})
	}
}

func TestListByUser(t *testing.T) {
	expected := []model.Token{
		{ID: 1, Name: "token-1", Prefix: "sk_abc12"},
		{ID: 2, Name: "token-2", Prefix: "sk_def34"},
	}
	tests := []struct {
		mock    *mockTokenStorage
		name    string
		wantLen int
		wantErr bool
	}{
		{
			name: "with tokens",
			mock: &mockTokenStorage{
				listByUserFunc: func(ctx context.Context, userID int) ([]model.Token, error) {
					assert.Equal(t, 1, userID)
					return expected, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "empty list",
			mock: &mockTokenStorage{
				listByUserFunc: func(ctx context.Context, userID int) ([]model.Token, error) {
					return []model.Token{}, nil
				},
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewToken(tt.mock)
			tokens, err := s.ListByUser(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, tokens, tt.wantLen)
		})
	}
}

func TestRevoke(t *testing.T) {
	tests := []struct {
		wantErr error
		mock    *mockTokenStorage
		name    string
	}{
		{
			name: "success",
			mock: &mockTokenStorage{
				revokeFunc: func(ctx context.Context, userID int, tokenID int64) error {
					assert.Equal(t, 1, userID)
					assert.Equal(t, int64(1), tokenID)
					return nil
				},
			},
		},
		{
			name: "not found",
			mock: &mockTokenStorage{
				revokeFunc: func(ctx context.Context, userID int, tokenID int64) error {
					return model.ErrTokenNotFound
				},
			},
			wantErr: model.ErrTokenNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewToken(tt.mock)
			err := s.Revoke(context.Background(), 1, 1)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	hash := sha256.Sum256([]byte("sk_validtoken123"))
	validHash := hex.EncodeToString(hash[:])
	revokedAt := time.Now()

	tests := []struct {
		wantErr  error
		mock     *mockTokenStorage
		name     string
		rawToken string
		wantID   int64
	}{
		{
			name:     "success",
			rawToken: "sk_validtoken123",
			mock: &mockTokenStorage{
				findByHashFunc: func(ctx context.Context, h string) (model.Token, error) {
					assert.Equal(t, validHash, h)
					return model.Token{ID: 1, UserID: 1, Name: "test", TokenHash: h}, nil
				},
				updateLastUsedFunc: func(ctx context.Context, tokenID int64) error {
					return nil
				},
			},
			wantID: 1,
		},
		{
			name:     "invalid token",
			rawToken: "sk_invalidtoken123",
			mock: &mockTokenStorage{
				findByHashFunc: func(ctx context.Context, hash string) (model.Token, error) {
					return model.Token{}, model.ErrTokenNotFound
				},
			},
			wantErr: model.ErrTokenNotFound,
		},
		{
			name:     "revoked token",
			rawToken: "sk_revokedtoken",
			mock: &mockTokenStorage{
				findByHashFunc: func(ctx context.Context, hash string) (model.Token, error) {
					return model.Token{ID: 1, RevokedAt: &revokedAt}, nil
				},
			},
			wantErr: model.ErrTokenRevoked,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewToken(tt.mock)
			result, err := s.Authenticate(context.Background(), tt.rawToken)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, result.ID)
		})
	}
}
