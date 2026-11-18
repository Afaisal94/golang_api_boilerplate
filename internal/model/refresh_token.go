package model

import (
	"errors"
	"time"
)

// Sentinel errors untuk RefreshToken.
var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")
)

// RefreshToken adalah entitas untuk tabel refresh_tokens.
type RefreshToken struct {
	ID        uint       `json:"id"`
	UUID      string     `json:"uuid"`
	UserID    uint       `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// RefreshTokenRepository adalah kontrak akses data untuk RefreshToken.
type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	FindByTokenHash(tokenHash string) (*RefreshToken, error)
	Revoke(id uint) error
	RevokeAllByUserID(userID uint) error
}
