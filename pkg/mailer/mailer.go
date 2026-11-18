package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Afaisal94/first-app/pkg/logger"
)

// Config menampung konfigurasi pengiriman email via Brevo HTTP API.
type Config struct {
	From    string // alamat sender yang sudah diverifikasi di Brevo
	AppName string
	APIKey  string // Brevo API key (format xkeysib-...)
	APIURL  string // endpoint Brevo, default https://api.brevo.com/v3/smtp/email
}

// Mailer adalah kontrak pengirim email.
type Mailer interface {
	SendOTP(toEmail, toName, otpCode string) error
	SendPasswordResetOTP(toEmail, toName, otpCode string) error
}

type brevoMailer struct {
	cfg Config
}

// New membuat Mailer Brevo yang mengirim email via HTTP API (HTTPS :443).
func New(cfg Config) Mailer {
	return &brevoMailer{cfg: cfg}
}

// SendOTP mengirim email berisi kode OTP verifikasi registrasi via Brevo HTTP API.
func (m *brevoMailer) SendOTP(toEmail, toName, otpCode string) error {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("mailer: tujuan email kosong")
	}
	if m.cfg.From == "" {
		return fmt.Errorf("mailer: from belum dikonfigurasi (cek SMTP_FROM)")
	}
	apiKey := strings.TrimSpace(m.cfg.APIKey)
	if apiKey == "" {
		return fmt.Errorf("mailer: BREVO_API_KEY kosong — isi BREVO_API_KEY di .env")
	}

	subject := fmt.Sprintf("[%s] Kode Verifikasi Email", m.cfg.AppName)
	body := fmt.Sprintf(`Halo %s,

Terima kasih telah mendaftar di %s.
Kode OTP verifikasi email kamu adalah:

    %s

Kode berlaku 15 menit. Jangan bagikan kode ini ke siapa pun.

Jika kamu tidak merasa mendaftar, abaikan email ini.
`, toName, m.cfg.AppName, otpCode)

	logger.Default.Infof("mailer: mencoba Brevo API (key=%d char) ke %s ...", len(apiKey), toEmail)
	if err := m.sendViaBrevoAPI(toEmail, toName, subject, body); err != nil {
		logger.Default.Errorf("mailer: Brevo API gagal: %v", err)
		return err
	}
	logger.Default.Infof("mailer: OTP terkirim ke %s via Brevo API (from=%s)", toEmail, m.cfg.From)
	return nil
}

// SendPasswordResetOTP mengirim email berisi kode OTP reset password via Brevo HTTP API.
func (m *brevoMailer) SendPasswordResetOTP(toEmail, toName, otpCode string) error {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("mailer: tujuan email kosong")
	}
	if m.cfg.From == "" {
		return fmt.Errorf("mailer: from belum dikonfigurasi (cek SMTP_FROM)")
	}
	apiKey := strings.TrimSpace(m.cfg.APIKey)
	if apiKey == "" {
		return fmt.Errorf("mailer: BREVO_API_KEY kosong — isi BREVO_API_KEY di .env")
	}

	subject := fmt.Sprintf("[%s] Kode Reset Password", m.cfg.AppName)
	body := fmt.Sprintf(`Halo %s,

Kami menerima permintaan untuk mereset password akun %s kamu.
Kode OTP reset password kamu adalah:

    %s

Kode berlaku 15 menit. Jangan bagikan kode ini ke siapa pun.

Jika kamu tidak merasa meminta reset password, abaikan email ini dan password kamu tidak akan diubah.
`, toName, m.cfg.AppName, otpCode)

	logger.Default.Infof("mailer: mencoba Brevo API (key=%d char) ke %s ...", len(apiKey), toEmail)
	if err := m.sendViaBrevoAPI(toEmail, toName, subject, body); err != nil {
		logger.Default.Errorf("mailer: Brevo API gagal: %v", err)
		return err
	}
	logger.Default.Infof("mailer: OTP reset password terkirim ke %s via Brevo API (from=%s)", toEmail, m.cfg.From)
	return nil
}

// sendViaBrevoAPI mengirim via POST https://api.brevo.com/v3/smtp/email (HTTPS :443).
// Butuh API key dari dashboard Brevo (SMTP & API > API Keys), BUKAN smtp key xsmtpsib-*.
func (m *brevoMailer) sendViaBrevoAPI(toEmail, toName, subject, body string) error {
	apiURL := strings.TrimSpace(m.cfg.APIURL)
	if apiURL == "" {
		apiURL = "https://api.brevo.com/v3/smtp/email"
	}
	payload := map[string]any{
		"sender":      map[string]string{"name": m.cfg.AppName, "email": m.cfg.From},
		"to":          []map[string]string{{"email": toEmail, "name": toName}},
		"subject":     subject,
		"textContent": body,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mailer: marshal Brevo API payload gagal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("mailer: buat request Brevo API gagal: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", strings.TrimSpace(m.cfg.APIKey))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: Brevo API request gagal: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mailer: Brevo API ditolak (status %d): %s — pastikan BREVO_API_KEY benar dan sender %s terverifikasi", resp.StatusCode, strings.TrimSpace(string(respBody)), m.cfg.From)
	}
	return nil
}
