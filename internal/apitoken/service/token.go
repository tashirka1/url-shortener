package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"url_shortener/internal/apitoken/model"
	"url_shortener/internal/apitoken/storage"
)

const (
	tokenPrefix    = "sk_"
	tokenBytes     = 32
	tokenBase62Len = 44
	tokenCharset   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	prefixShowLen  = 8
)

type TokenService interface {
	Generate(ctx context.Context, userID int, name string) (model.GenerateResult, error)
	ListByUser(ctx context.Context, userID int) ([]model.Token, error)
	Revoke(ctx context.Context, userID int, tokenID int64) error
	Authenticate(ctx context.Context, rawToken string) (model.Token, error)
}

type Token struct {
	r storage.TokenStorage
}

func NewToken(r storage.TokenStorage) *Token {
	return &Token{r: r}
}

func (s *Token) Generate(ctx context.Context, userID int, name string) (model.GenerateResult, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return model.GenerateResult{}, fmt.Errorf("generate token: %w", err)
	}

	var n big.Int
	n.SetBytes(b)
	var base big.Int
	base.SetInt64(int64(len(tokenCharset)))
	raw := make([]byte, tokenBase62Len)
	var rem big.Int
	for i := tokenBase62Len - 1; i >= 0; i-- {
		n.DivMod(&n, &base, &rem)
		raw[i] = tokenCharset[rem.Int64()]
	}

	fullToken := tokenPrefix + string(raw)
	hash := sha256.Sum256([]byte(fullToken))
	tokenHash := hex.EncodeToString(hash[:])
	prefix := fullToken[:prefixShowLen]

	t, err := s.r.Insert(ctx, userID, name, tokenHash, prefix)
	if err != nil {
		return model.GenerateResult{}, fmt.Errorf("save token: %w", err)
	}
	slog.InfoContext(ctx, "api token created", "user_id", userID, "token_id", t.ID)
	return model.GenerateResult{Token: t, RawToken: fullToken}, nil
}

func (s *Token) ListByUser(ctx context.Context, userID int) ([]model.Token, error) {
	tokens, err := s.r.ListByUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "list tokens failed", "user_id", userID, "error", err)
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	return tokens, nil
}

func (s *Token) Revoke(ctx context.Context, userID int, tokenID int64) error {
	if err := s.r.Revoke(ctx, userID, tokenID); err != nil {
		slog.ErrorContext(ctx, "revoke token failed", "user_id", userID, "token_id", tokenID, "error", err)
		return fmt.Errorf("revoke token: %w", err)
	}
	slog.InfoContext(ctx, "api token revoked", "user_id", userID, "token_id", tokenID)
	return nil
}

func (s *Token) Authenticate(ctx context.Context, rawToken string) (model.Token, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	t, err := s.r.FindByHash(ctx, tokenHash)
	if err != nil {
		return model.Token{}, err
	}

	if t.RevokedAt != nil && !t.RevokedAt.IsZero() {
		return model.Token{}, model.ErrTokenRevoked
	}

	if err := s.r.UpdateLastUsed(ctx, t.ID); err != nil {
		slog.WarnContext(ctx, "failed to update last_used_at", "token_id", t.ID, "error", err)
	}

	now := time.Now()
	t.LastUsedAt = &now
	return t, nil
}
