package handlers

import (
	"errors"

	"Afaisal94/first-app/internal/dto"
	"Afaisal94/first-app/internal/model"
	"Afaisal94/first-app/internal/services"
	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler menangani HTTP untuk autentikasi.
type AuthHandler struct {
	service services.AuthService
}

// NewAuthHandler membuat handler dengan service yang di-inject.
func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register POST /api/auth/register — registrasi user baru.
// @Summary      Register new user
// @Description  Mendaftarkan user baru dengan data yang diberikan.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration data"
// @Success      201 {object} response.Envelope{data=dto.RegisterResponse}
// @Failure      400 {object} response.Envelope
// @Failure      409 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterRequest
	if !bindAndValidate(c, &input) {
		return
	}

	// Konversi DTO ke service input
	serviceInput := services.RegisterInput{
		Name:            input.Name,
		Email:           input.Email,
		Phone:           input.Phone,
		Password:        input.Password,
		ConfirmPassword: input.ConfirmPassword,
	}

	user, err := h.service.Register(serviceInput)
	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			response.BadRequest(c, "Email already registered", nil)
		case errors.Is(err, model.ErrPhoneAlreadyExists):
			response.BadRequest(c, "Phone number already registered", nil)
		case err.Error() == "password and confirm password do not match":
			response.BadRequest(c, "Password and confirm password do not match", nil)
		default:
			logger.Default.Errorf("Register failed: %v", err)
			response.InternalSafe(c, "Failed to register user")
		}
		return
	}

	response.Created(c, "User registered successfully. Please check your email for OTP verification code.", dto.NewRegisterResponse(user))
}

// VerifyEmail POST /api/auth/verify-email — verifikasi email dengan OTP.
// @Summary      Verify email with OTP
// @Description  Verifikasi email user dengan kode OTP 6 digit yang dikirim ke email saat registrasi. Setelah berhasil, akun aktif dan bisa login.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyEmailRequest true "Verify email data"
// @Success      200 {object} response.Envelope{data=dto.RegisterResponse}
// @Failure      400 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var input dto.VerifyEmailRequest
	if !bindAndValidate(c, &input) {
		return
	}

	user, err := h.service.VerifyEmail(services.VerifyEmailInput{
		Email: input.Email,
		OTP:   input.OTP,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyVerified):
			response.BadRequest(c, "Email already verified", nil)
		case errors.Is(err, model.ErrOTPInvalid):
			response.BadRequest(c, "Invalid OTP code", nil)
		case errors.Is(err, model.ErrOTPExpired):
			response.BadRequest(c, "OTP has expired. Please request a new one", nil)
		case errors.Is(err, model.ErrOTPMaxAttempt):
			response.BadRequest(c, "Maximum OTP attempts exceeded. Please request a new one", nil)
		case errors.Is(err, model.ErrOTPNotFound):
			response.BadRequest(c, "OTP not found. Please request a new one", nil)
		case errors.Is(err, model.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			logger.Default.Errorf("VerifyEmail failed: %v", err)
			response.InternalSafe(c, "Failed to verify email")
		}
		return
	}

	response.Success(c, "Email verified successfully. You can now login.", dto.NewRegisterResponse(user))
}

// ResendOTP POST /api/auth/resend-otp — kirim ulang OTP verifikasi.
// @Summary      Resend OTP verification
// @Description  Mengirim ulang kode OTP 6 digit ke email user yang belum terverifikasi.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResendOTPRequest true "Resend OTP data"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/resend-otp [post]
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var input dto.ResendOTPRequest
	if !bindAndValidate(c, &input) {
		return
	}

	if err := h.service.ResendOTP(input.Email); err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyVerified):
			response.BadRequest(c, "Email already verified", nil)
		case errors.Is(err, model.ErrUserNotFound):
			response.NotFound(c, "User not found")
		default:
			logger.Default.Errorf("ResendOTP failed: %v", err)
			response.InternalSafe(c, "Failed to resend OTP")
		}
		return
	}

	response.OKWithoutData(c, "OTP sent successfully. Please check your email.")
}

// Login POST /api/auth/login — login dengan email & password.
// @Summary      Login user
// @Description  Login dengan email dan password, mengembalikan access token (JWT) dan refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login data"
// @Success      200 {object} response.Envelope{data=dto.LoginResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      403 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginRequest
	if !bindAndValidate(c, &input) {
		return
	}

	result, err := h.service.Login(services.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidCredentials):
			response.Unauthorized(c, "Invalid email or password")
		case errors.Is(err, model.ErrEmailNotVerified):
			response.Forbidden(c, "Email not verified. Please verify your email with OTP first")
		case errors.Is(err, model.ErrUserInactive):
			response.Forbidden(c, "User account is inactive")
		default:
			logger.Default.Errorf("Login failed: %v", err)
			response.InternalSafe(c, "Failed to login")
		}
		return
	}

	response.Success(c, "Login successfully", dto.NewLoginResponse(
		result.User, result.AccessToken, result.RefreshToken, result.ExpiresIn,
	))
}

// Refresh POST /api/auth/refresh — tukar refresh token dengan token pair baru.
// @Summary      Refresh access token
// @Description  Menukar refresh token yang valid dengan access token + refresh token baru (rotasi). Token lama langsung di-revoke.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshRequest true "Refresh token"
// @Success      200 {object} response.Envelope{data=dto.LoginResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      403 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var input dto.RefreshRequest
	if !bindAndValidate(c, &input) {
		return
	}

	result, err := h.service.Refresh(services.RefreshInput{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrRefreshTokenNotFound),
			errors.Is(err, model.ErrRefreshTokenRevoked),
			errors.Is(err, model.ErrRefreshTokenExpired):
			response.Unauthorized(c, "Invalid or expired refresh token")
		case errors.Is(err, model.ErrUserInactive):
			response.Forbidden(c, "User account is inactive")
		case errors.Is(err, model.ErrUserNotFound):
			response.Unauthorized(c, "Invalid or expired refresh token")
		default:
			logger.Default.Errorf("Refresh failed: %v", err)
			response.InternalSafe(c, "Failed to refresh token")
		}
		return
	}

	response.Success(c, "Token refreshed successfully", dto.NewLoginResponse(
		result.User, result.AccessToken, result.RefreshToken, result.ExpiresIn,
	))
}

// Logout POST /api/auth/logout — revoke refresh token.
// @Summary      Logout user
// @Description  Me-revoke refresh token sehingga tidak bisa dipakai untuk refresh lagi. Bersifat idempoten.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LogoutRequest true "Refresh token"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var input dto.LogoutRequest
	if !bindAndValidate(c, &input) {
		return
	}

	if err := h.service.Logout(input.RefreshToken); err != nil {
		switch {
		case errors.Is(err, model.ErrRefreshTokenNotFound):
			response.BadRequest(c, "Invalid refresh token", nil)
		default:
			logger.Default.Errorf("Logout failed: %v", err)
			response.InternalSafe(c, "Failed to logout")
		}
		return
	}

	response.OKWithoutData(c, "Logged out successfully")
}

// ForgotPassword POST /api/auth/forgot-password — kirim OTP reset password ke email.
// @Summary      Forgot password
// @Description  Mengirim kode OTP reset password ke email user. Selalu mengembalikan sukses meskipun email tidak terdaftar (anti-enumeration).
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Forgot password data"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input dto.ForgotPasswordRequest
	if !bindAndValidate(c, &input) {
		return
	}

	_ = h.service.ForgotPassword(input.Email)

	// Selalu sukses untuk mencegah enumeration attack.
	response.OKWithoutData(c, "If the email is registered, a password reset OTP has been sent to your email.")
}

// ResetPassword POST /api/auth/reset-password — reset password dengan OTP.
// @Summary      Reset password
// @Description  Memverifikasi OTP reset password dan mengganti password user. Setelah berhasil, semua refresh token di-revoke sehingga user harus login ulang.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Reset password data"
// @Success      200 {object} response.Envelope
// @Failure      400 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input dto.ResetPasswordRequest
	if !bindAndValidate(c, &input) {
		return
	}

	err := h.service.ResetPassword(services.ResetPasswordInput{
		Email:           input.Email,
		OTP:             input.OTP,
		Password:        input.Password,
		ConfirmPassword: input.ConfirmPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, model.ErrOTPInvalid):
			response.BadRequest(c, "Invalid OTP code", nil)
		case errors.Is(err, model.ErrOTPExpired):
			response.BadRequest(c, "OTP has expired. Please request a new one", nil)
		case errors.Is(err, model.ErrOTPMaxAttempt):
			response.BadRequest(c, "Maximum OTP attempts exceeded. Please request a new one", nil)
		case errors.Is(err, model.ErrOTPNotFound):
			response.BadRequest(c, "OTP not found. Please request a new one", nil)
		case errors.Is(err, model.ErrUserNotFound):
			response.NotFound(c, "User not found")
		case err.Error() == "password and confirm password do not match":
			response.BadRequest(c, "Password and confirm password do not match", nil)
		default:
			logger.Default.Errorf("ResetPassword failed: %v", err)
			response.InternalSafe(c, "Failed to reset password")
		}
		return
	}

	response.OKWithoutData(c, "Password reset successfully. Please login with your new password.")
}
