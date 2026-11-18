package dto

import "Afaisal94/first-app/internal/model"

// ──────────────────────────────────────────────
// Request DTOs
// ──────────────────────────────────────────────

// RegisterRequest adalah DTO untuk POST /api/auth/register.
type RegisterRequest struct {
	Name            string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email           string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
	Phone           string `json:"phone" binding:"required,min=10,max=15" example:"08123456789"`
	Password        string `json:"password" binding:"required,min=8,max=100" example:"secret123"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password" example:"secret123"`
}

// LoginRequest adalah DTO untuk POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"secret123"`
}

// RefreshRequest adalah DTO untuk POST /api/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"a1b2c3d4e5f6..."`
}

// LogoutRequest adalah DTO untuk POST /api/auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"a1b2c3d4e5f6..."`
}

// VerifyEmailRequest adalah DTO untuk POST /api/auth/verify-email.
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

// ResendOTPRequest adalah DTO untuk POST /api/auth/resend-otp.
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
}

// ForgotPasswordRequest adalah DTO untuk POST /api/auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
}

// ResetPasswordRequest adalah DTO untuk POST /api/auth/reset-password.
type ResetPasswordRequest struct {
	Email           string `json:"email" binding:"required,email,max=255" example:"john@example.com"`
	OTP             string `json:"otp" binding:"required,len=6" example:"123456"`
	Password        string `json:"password" binding:"required,min=8,max=100" example:"newsecret123"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password" example:"newsecret123"`
}

// ──────────────────────────────────────────────
// Response DTOs
// ──────────────────────────────────────────────

// RegisterResponse adalah DTO output untuk register.
type RegisterResponse struct {
	ID    uint   `json:"id" example:"1"`
	UUID  string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
	Phone string `json:"phone" example:"08123456789"`
}

// NewRegisterResponse memetakan entitas User ke RegisterResponse.
func NewRegisterResponse(user *model.User) RegisterResponse {
	return RegisterResponse{
		ID:    user.ID,
		UUID:  user.UUID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
	}
}

// LoginResponse adalah DTO output untuk login.
type LoginResponse struct {
	AccessToken  string           `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string           `json:"refresh_token" example:"a1b2c3d4..."`
	ExpiresIn    int64            `json:"expires_in" example:"900"`
	TokenType    string           `json:"token_type" example:"Bearer"`
	User         RegisterResponse `json:"user"`
}

// NewLoginResponse memetakan hasil login ke LoginResponse.
func NewLoginResponse(user *model.User, accessToken, refreshToken string, expiresIn int64) LoginResponse {
	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User:         NewRegisterResponse(user),
	}
}
