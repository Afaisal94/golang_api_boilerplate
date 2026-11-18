package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// refreshTokenRepository adalah adapter database/sql untuk kontrak model.RefreshTokenRepository.
type refreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository membuat implementasi database/sql dari model.RefreshTokenRepository.
func NewRefreshTokenRepository(db *sql.DB) model.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.QueryRow(
		`INSERT INTO refresh_tokens (uuid, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, created_at`,
		token.UUID, token.UserID, token.TokenHash, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.QueryRow(
		`SELECT id, uuid, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	).Scan(
		&t.ID, &t.UUID, &t.UserID, &t.TokenHash,
		&t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *refreshTokenRepository) Revoke(id uint) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL", id,
	)
	return err
}

func (r *refreshTokenRepository) RevokeAllByUserID(userID uint) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", userID,
	)
	return err
}
