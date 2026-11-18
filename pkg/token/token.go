package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims adalah payload JWT access token.
type Claims struct {
	UserID uint   `json:"user_id"`
	UUID   string `json:"uuid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken membuat JWT access token berumur pendek.
func GenerateAccessToken(secret string, userID uint, userUUID, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		UUID:   userUUID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessToken memvalidasi JWT dan mengembalikan claims-nya.
func ParseAccessToken(secret, tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// RefreshTokenPair adalah refresh token mentah (dikirim ke client)
// beserta hash-nya (disimpan di database).
type RefreshTokenPair struct {
	UUID  string
	Token string
	Hash  string
}

// GenerateRefreshToken membuat refresh token opaque (random 32 byte hex)
// beserta hash SHA-256-nya untuk disimpan di database.
func GenerateRefreshToken() (RefreshTokenPair, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return RefreshTokenPair{}, err
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return RefreshTokenPair{
		UUID:  uuid.NewString(),
		Token: token,
		Hash:  hex.EncodeToString(sum[:]),
	}, nil
}

// HashRefreshToken menghitung hash SHA-256 dari refresh token mentah,
// dipakai saat lookup token di database.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
