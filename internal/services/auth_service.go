package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"Afaisal94/first-app/internal/config"
	"Afaisal94/first-app/internal/model"
	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/mailer"
	"Afaisal94/first-app/pkg/token"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService mendefinisikan kontrak business logic untuk autentikasi.
type AuthService interface {
	Register(input RegisterInput) (*model.User, error)
	VerifyEmail(input VerifyEmailInput) (*model.User, error)
	ResendOTP(email string) error
	Login(input LoginInput) (*LoginResult, error)
	Refresh(input RefreshInput) (*LoginResult, error)
	Logout(refreshToken string) error
	ForgotPassword(email string) error
	ResetPassword(input ResetPasswordInput) error
}

// RegisterInput adalah input untuk registrasi.
type RegisterInput struct {
	Name            string
	Email           string
	Phone           string
	Password        string
	ConfirmPassword string
}

// LoginInput adalah input untuk login.
type LoginInput struct {
	Email    string
	Password string
}

// RefreshInput adalah input untuk refresh token.
type RefreshInput struct {
	RefreshToken string
}

// VerifyEmailInput adalah input untuk verifikasi email dengan OTP.
type VerifyEmailInput struct {
	Email string
	OTP   string
}

// ResetPasswordInput adalah input untuk reset password dengan OTP.
type ResetPasswordInput struct {
	Email           string
	OTP             string
	Password        string
	ConfirmPassword string
}

// LoginResult adalah hasil login: user + token pair.
type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type authService struct {
	cfg         *config.Config
	userRepo    model.UserRepository
	roleRepo    model.RoleRepository
	otpRepo     model.OTPRepository
	refreshRepo model.RefreshTokenRepository
	mailer      mailer.Mailer
}

// NewAuthService membuat implementasi AuthService dengan dependency yang di-inject.
func NewAuthService(
	cfg *config.Config,
	userRepo model.UserRepository,
	roleRepo model.RoleRepository,
	otpRepo model.OTPRepository,
	refreshRepo model.RefreshTokenRepository,
	mailer mailer.Mailer,
) AuthService {
	return &authService{
		cfg:         cfg,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		otpRepo:     otpRepo,
		refreshRepo: refreshRepo,
		mailer:      mailer,
	}
}

func (s *authService) Register(input RegisterInput) (*model.User, error) {
	// Validasi password confirmation
	if input.Password != input.ConfirmPassword {
		return nil, errors.New("password and confirm password do not match")
	}

	// Normalisasi input
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)

	// Cek apakah email sudah terdaftar
	existingUser, _ := s.userRepo.FindByEmail(input.Email)
	if existingUser != nil {
		return nil, model.ErrEmailAlreadyExists
	}

	// Cek apakah phone sudah terdaftar
	existingUser, _ = s.userRepo.FindByPhone(input.Phone)
	if existingUser != nil {
		return nil, model.ErrPhoneAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Default.Errorf("failed to hash password: %v", err)
		return nil, errors.New("failed to process password")
	}

	// Dapatkan role "user" (default untuk registrasi)
	userRole, err := s.roleRepo.FindByName("user")
	if err != nil {
		logger.Default.Errorf("failed to find user role: %v", err)
		return nil, errors.New("system configuration error")
	}

	// Generate UUID untuk user
	userUUID := uuid.New().String()

	// Buat user baru — nonaktif sampai email diverifikasi via OTP.
	user := &model.User{
		UUID:         userUUID,
		Name:         input.Name,
		Email:        input.Email,
		Phone:        input.Phone,
		PasswordHash: string(hashedPassword),
		RoleID:       userRole.ID,
		IsActive:     false,
	}

	if err := s.userRepo.Create(user); err != nil {
		logger.Default.Errorf("failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}

	// Generate OTP numerik + kirim ke email via Brevo.
	// PENTING: jangan swallow error — log detail agar terlihat di console server.
	if err := s.issueRegisterOTP(user); err != nil {
		logger.Default.Errorf("failed to send register OTP to %s from=%s: %v", user.Email, s.cfg.SMTPFrom, err)
		// User tetap terdaftar — client bisa pakai /auth/resend-otp.
		// Cek log di atas untuk penyebab (auth/sender/koneksi).
	}

	return user, nil
}

// VerifyEmail mencocokkan OTP register, lalu mengaktifkan user.
func (s *authService) VerifyEmail(input VerifyEmailInput) (*model.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	code := strings.TrimSpace(input.OTP)
	if email == "" || code == "" {
		return nil, model.ErrOTPNotFound
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	if user.EmailVerifiedAt != nil {
		return nil, model.ErrEmailAlreadyVerified
	}

	stored, err := s.otpRepo.FindByUserAndType(user.ID, model.OTPTypeRegister)
	if err != nil {
		return nil, model.ErrOTPNotFound
	}

	if stored.VerifiedAt != nil {
		return nil, model.ErrOTPNotFound
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, model.ErrOTPExpired
	}

	if stored.Attempts >= s.cfg.OTPMaxTry {
		return nil, model.ErrOTPMaxAttempt
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.CodeHash), []byte(code)); err != nil {
		stored.Attempts++
		_ = s.otpRepo.Update(stored)
		return nil, model.ErrOTPInvalid
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.IsActive = true
	if err := s.userRepo.Update(user); err != nil {
		logger.Default.Errorf("verify update user failed: %v", err)
		return nil, errors.New("failed to verify email")
	}

	// OTP sekali pakai — hapus setelah berhasil.
	_ = s.otpRepo.Delete(stored.ID)

	return user, nil
}

// ResendOTP membuat OTP baru dan mengirim ulang ke email.
func (s *authService) ResendOTP(email string) error {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return model.ErrUserNotFound
	}

	user, err := s.userRepo.FindByEmail(normalized)
	if err != nil {
		return model.ErrUserNotFound
	}

	if user.EmailVerifiedAt != nil {
		return model.ErrEmailAlreadyVerified
	}

	if err := s.issueRegisterOTP(user); err != nil {
		logger.Default.Errorf("resend OTP to %s failed: %v", normalized, err)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// Login memverifikasi email & password, lalu menerbitkan access token (JWT)
// dan refresh token (opaque, hash-nya disimpan di database).
// Pesan error sengaja generik agar tidak membocorkan user mana yang terdaftar.
func (s *authService) Login(input LoginInput) (*LoginResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrInvalidCredentials
		}
		logger.Default.Errorf("login FindByEmail failed: %v", err)
		return nil, errors.New("failed to process login")
	}

	if !user.IsActive {
		if user.EmailVerifiedAt == nil {
			return nil, model.ErrEmailNotVerified
		}
		return nil, model.ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	roleName := ""
	if role, err := s.roleRepo.FindByID(user.RoleID); err == nil {
		roleName = role.Name
	}

	accessToken, err := token.GenerateAccessToken(
		s.cfg.JWTSecret, user.ID, user.UUID, user.Email, roleName, s.cfg.JWTAccessTTL,
	)
	if err != nil {
		logger.Default.Errorf("login GenerateAccessToken failed: %v", err)
		return nil, errors.New("failed to issue access token")
	}

	pair, err := token.GenerateRefreshToken()
	if err != nil {
		logger.Default.Errorf("login GenerateRefreshToken failed: %v", err)
		return nil, errors.New("failed to issue refresh token")
	}

	if err := s.refreshRepo.Create(&model.RefreshToken{
		UUID:      pair.UUID,
		UserID:    user.ID,
		TokenHash: pair.Hash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}); err != nil {
		logger.Default.Errorf("login persist refresh token failed: %v", err)
		return nil, errors.New("failed to persist refresh token")
	}

	// Catat waktu login terakhir — kegagalan di sini tidak menggagalkan login.
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(user); err != nil {
		logger.Default.Errorf("login update last_login_at failed: %v", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: pair.Token,
		ExpiresIn:    int64(s.cfg.JWTAccessTTL.Seconds()),
	}, nil
}

// Refresh menukar refresh token yang valid dengan token pair baru (rotasi).
// Token lama langsung di-revoke agar tidak bisa dipakai ulang (replay protection).
func (s *authService) Refresh(input RefreshInput) (*LoginResult, error) {
	raw := strings.TrimSpace(input.RefreshToken)
	if raw == "" {
		return nil, model.ErrRefreshTokenNotFound
	}

	stored, err := s.refreshRepo.FindByTokenHash(token.HashRefreshToken(raw))
	if err != nil {
		return nil, model.ErrRefreshTokenNotFound
	}

	if stored.RevokedAt != nil {
		return nil, model.ErrRefreshTokenRevoked
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, model.ErrRefreshTokenExpired
	}

	user, err := s.userRepo.FindByID(strconv.FormatUint(uint64(stored.UserID), 10))
	if err != nil {
		return nil, model.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, model.ErrUserInactive
	}

	roleName := ""
	if role, err := s.roleRepo.FindByID(user.RoleID); err == nil {
		roleName = role.Name
	}

	accessToken, err := token.GenerateAccessToken(
		s.cfg.JWTSecret, user.ID, user.UUID, user.Email, roleName, s.cfg.JWTAccessTTL,
	)
	if err != nil {
		logger.Default.Errorf("refresh GenerateAccessToken failed: %v", err)
		return nil, errors.New("failed to issue access token")
	}

	pair, err := token.GenerateRefreshToken()
	if err != nil {
		logger.Default.Errorf("refresh GenerateRefreshToken failed: %v", err)
		return nil, errors.New("failed to issue refresh token")
	}

	if err := s.refreshRepo.Create(&model.RefreshToken{
		UUID:      pair.UUID,
		UserID:    user.ID,
		TokenHash: pair.Hash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}); err != nil {
		logger.Default.Errorf("refresh persist refresh token failed: %v", err)
		return nil, errors.New("failed to persist refresh token")
	}

	// Revoke token lama setelah token baru berhasil disimpan.
	if err := s.refreshRepo.Revoke(stored.ID); err != nil {
		logger.Default.Errorf("refresh revoke old token failed: %v", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: pair.Token,
		ExpiresIn:    int64(s.cfg.JWTAccessTTL.Seconds()),
	}, nil
}

// Logout me-revoke refresh token sehingga tidak bisa dipakai untuk refresh lagi.
// Bersifat idempoten: token yang sudah revoked/expired tetap dianggap sukses logout.
func (s *authService) Logout(refreshToken string) error {
	raw := strings.TrimSpace(refreshToken)
	if raw == "" {
		return model.ErrRefreshTokenNotFound
	}

	stored, err := s.refreshRepo.FindByTokenHash(token.HashRefreshToken(raw))
	if err != nil {
		return model.ErrRefreshTokenNotFound
	}

	if stored.RevokedAt != nil {
		return nil
	}

	if err := s.refreshRepo.Revoke(stored.ID); err != nil {
		logger.Default.Errorf("logout revoke failed: %v", err)
		return errors.New("failed to logout")
	}

	return nil
}

// ForgotPassword mengirim OTP reset password ke email user.
// Selalu mengembalikan sukses (meski email tidak terdaftar) untuk mencegah
// enumeration attack — namun OTP hanya dikirim jika user ditemukan & aktif.
func (s *authService) ForgotPassword(email string) error {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return nil
	}

	user, err := s.userRepo.FindByEmail(normalized)
	if err != nil {
		// Jangan bocorkan apakah email terdaftar — tetap sukses.
		return nil
	}

	// Hanya user aktif yang bisa reset password.
	if !user.IsActive {
		return nil
	}

	if err := s.issueForgotPasswordOTP(user); err != nil {
		logger.Default.Errorf("forgot password OTP to %s failed: %v", normalized, err)
		// Tetap return nil agar tidak membocorkan kegagalan ke client.
	}

	return nil
}

// ResetPassword memverifikasi OTP reset password, lalu mengganti password user.
// Setelah berhasil, semua refresh token user yang ada di-revoke agar dipaksa login ulang.
func (s *authService) ResetPassword(input ResetPasswordInput) error {
	if input.Password != input.ConfirmPassword {
		return errors.New("password and confirm password do not match")
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	code := strings.TrimSpace(input.OTP)
	if email == "" || code == "" {
		return model.ErrOTPNotFound
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return model.ErrUserNotFound
	}

	stored, err := s.otpRepo.FindByUserAndType(user.ID, model.OTPTypeForgotPassword)
	if err != nil {
		return model.ErrOTPNotFound
	}

	if stored.VerifiedAt != nil {
		return model.ErrOTPNotFound
	}

	if time.Now().After(stored.ExpiresAt) {
		return model.ErrOTPExpired
	}

	if stored.Attempts >= s.cfg.OTPMaxTry {
		return model.ErrOTPMaxAttempt
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.CodeHash), []byte(code)); err != nil {
		stored.Attempts++
		_ = s.otpRepo.Update(stored)
		return model.ErrOTPInvalid
	}

	// Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Default.Errorf("reset password hash failed: %v", err)
		return errors.New("failed to process password")
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		logger.Default.Errorf("reset password update failed: %v", err)
		return errors.New("failed to reset password")
	}

	// Revoke semua refresh token user agar dipaksa login ulang dengan password baru.
	if err := s.refreshRepo.RevokeAllByUserID(user.ID); err != nil {
		logger.Default.Errorf("reset password revoke tokens failed: %v", err)
	}

	// OTP sekali pakai — hapus setelah berhasil.
	_ = s.otpRepo.Delete(stored.ID)

	return nil
}

// issueRegisterOTP membuat OTP register baru, menyimpan hash-nya,
// lalu mengirim kode mentah ke email user via Brevo SMTP relay.
func (s *authService) issueRegisterOTP(user *model.User) error {
	otpCode, err := generateNumericOTP(6)
	if err != nil {
		return err
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	otp := &model.OTP{
		UserID:    user.ID,
		Type:      model.OTPTypeRegister,
		CodeHash:  string(hashedOTP),
		ExpiresAt: time.Now().Add(s.cfg.OTPExpire),
		Attempts:  0,
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return err
	}

	if s.mailer == nil {
		return nil
	}

	if err := s.mailer.SendOTP(user.Email, user.Name, otpCode); err != nil {
		logger.Default.Errorf("mailer.SendOTP to %s gagal: %v", user.Email, err)
		return err
	}

	return nil
}

// issueForgotPasswordOTP membuat OTP reset password baru, menyimpan hash-nya,
// lalu mengirim kode mentah ke email user via Brevo HTTP API.
func (s *authService) issueForgotPasswordOTP(user *model.User) error {
	otpCode, err := generateNumericOTP(6)
	if err != nil {
		return err
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	otp := &model.OTP{
		UserID:    user.ID,
		Type:      model.OTPTypeForgotPassword,
		CodeHash:  string(hashedOTP),
		ExpiresAt: time.Now().Add(s.cfg.OTPExpire),
		Attempts:  0,
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return err
	}

	if s.mailer == nil {
		return nil
	}

	if err := s.mailer.SendPasswordResetOTP(user.Email, user.Name, otpCode); err != nil {
		logger.Default.Errorf("mailer.SendPasswordResetOTP to %s gagal: %v", user.Email, err)
		return err
	}

	return nil
}

// generateNumericOTP menghasilkan OTP digit 0-9 sepanjang length.
func generateNumericOTP(length int) (string, error) {
	digits := make([]byte, length)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}
	return string(digits), nil
}
