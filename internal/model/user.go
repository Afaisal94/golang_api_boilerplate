package model

import (
	"errors"
	"time"
)

// Sentinel errors untuk User.
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailAlreadyExists   = errors.New("email already registered")
	ErrPhoneAlreadyExists   = errors.New("phone already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

// User adalah entitas untuk tabel users.
type User struct {
	ID              uint       `json:"id"`
	UUID            string     `json:"uuid"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	PasswordHash    string     `json:"-"`
	RoleID          uint       `json:"role_id"`
	Role            *Role      `json:"role,omitempty"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"-"`
}

// UserRepository adalah kontrak akses data untuk User.
type UserRepository interface {
	FindByID(id string) (*User, error)
	FindByUUID(uuid string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByPhone(phone string) (*User, error)
	Create(user *User) error
	Update(user *User) error
}
