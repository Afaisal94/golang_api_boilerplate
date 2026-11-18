package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// otpRepository adalah adapter database/sql untuk kontrak model.OTPRepository.
type otpRepository struct {
	db *sql.DB
}

// NewOTPRepository membuat implementasi database/sql dari model.OTPRepository.
func NewOTPRepository(db *sql.DB) model.OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Create(otp *model.OTP) error {
	return r.db.QueryRow(
		`INSERT INTO otps (user_id, type, code_hash, expires_at, attempts, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 RETURNING id, created_at`,
		otp.UserID, otp.Type, otp.CodeHash, otp.ExpiresAt, otp.Attempts,
	).Scan(&otp.ID, &otp.CreatedAt)
}

func (r *otpRepository) FindByUserAndType(userID uint, otpType string) (*model.OTP, error) {
	var otp model.OTP
	err := r.db.QueryRow(
		`SELECT id, user_id, type, code_hash, expires_at, verified_at, attempts, created_at
		 FROM otps WHERE user_id = $1 AND type = $2 ORDER BY id DESC LIMIT 1`,
		userID, otpType,
	).Scan(
		&otp.ID, &otp.UserID, &otp.Type, &otp.CodeHash,
		&otp.ExpiresAt, &otp.VerifiedAt, &otp.Attempts, &otp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrOTPNotFound
		}
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) Update(otp *model.OTP) error {
	_, err := r.db.Exec(
		`UPDATE otps SET code_hash = $1, expires_at = $2, verified_at = $3, attempts = $4
		 WHERE id = $5`,
		otp.CodeHash, otp.ExpiresAt, otp.VerifiedAt, otp.Attempts, otp.ID,
	)
	return err
}

func (r *otpRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM otps WHERE id = $1", id)
	return err
}
