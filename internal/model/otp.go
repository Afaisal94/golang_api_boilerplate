package model

import (
	"errors"
	"time"
)

// Tipe OTP.
const (
	OTPTypeRegister       = "register"
	OTPTypeForgotPassword = "forgot_password"
)

// Sentinel errors untuk OTP.
var (
	ErrOTPNotFound   = errors.New("otp not found")
	ErrOTPExpired    = errors.New("otp has expired")
	ErrOTPMaxAttempt = errors.New("maximum otp attempts exceeded")
	ErrOTPInvalid    = errors.New("invalid otp code")
)

// OTP adalah entitas untuk tabel otps.
type OTP struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	Type       string     `json:"type"`
	CodeHash   string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	Attempts   int        `json:"attempts"`
	CreatedAt  time.Time  `json:"created_at"`
}

// OTPRepository adalah kontrak akses data untuk OTP.
type OTPRepository interface {
	Create(otp *OTP) error
	FindByUserAndType(userID uint, otpType string) (*OTP, error)
	Update(otp *OTP) error
	Delete(id uint) error
}
