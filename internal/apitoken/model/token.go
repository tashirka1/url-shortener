package model

import (
	"errors"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenRevoked  = errors.New("token revoked")
)

type Token struct {
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	Name       string
	TokenHash  string
	Prefix     string
	ID         int64
	UserID     int
}

type GenerateResult struct {
	RawToken string
	Token    Token
}
