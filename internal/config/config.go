package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// loadEnvFile mencari file .env mulai dari cwd lalu naik ke atas
// sampai menemukan go.mod (root project). Ini memastikan .env selalu
// terbaca meskipun binary dijalankan dari subfolder seperti ./cmd/server.
func loadEnvFile() {
	cwd, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}
	dir := cwd
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		// berhenti jika menemukan go.mod (root module Go)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		dir = parent
	}
	_ = godotenv.Load()
}

// Config menampung semua konfigurasi aplikasi.
type Config struct {
	AppEnv string // "development" atau "production"

	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	AppPort   string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	JWTSecret       string
	JWTAccessTTL    time.Duration
	RefreshTokenTTL time.Duration

	AppName     string
	SMTPFrom    string
	BrevoAPIKey string
	BrevoAPIURL string
	OTPExpire   time.Duration
	OTPMaxTry   int
}

func getenv(keys []string, fallback string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return fallback
}

// Load membaca .env (jika ada) lalu environment variables.
// Mendukung penamaan lama POSTGRES_* dan penamaan baru DB_*.
// Mencari .env dari cwd lalu naik ke atas sampai ketemu (go.mod marker)
// agar tetap terbaca saat binary dijalankan dari subfolder mana pun.
func Load() *Config {
	loadEnvFile()

	return &Config{
		AppEnv:    getenv([]string{"APP_ENV", "GIN_MODE"}, "development"),
		DBHost:    getenv([]string{"DB_HOST", "POSTGRES_HOST"}, "127.0.0.1"),
		DBPort:    getenv([]string{"DB_PORT", "POSTGRES_PORT"}, "5432"),
		DBUser:    getenv([]string{"DB_USER", "POSTGRES_USER"}, "postgres"),
		DBPass:    getenv([]string{"DB_PASSWORD", "DB_PASS", "POSTGRES_PASSWORD"}, ""),
		DBName:    getenv([]string{"DB_NAME", "POSTGRES_DB"}, "db_go_api"),
		DBSSLMode: getenv([]string{"DB_SSLMODE"}, "disable"),
		AppPort:   getenv([]string{"APP_PORT", "PORT"}, "3000"),

		RedisHost:     getenv([]string{"REDIS_HOST"}, "localhost"),
		RedisPort:     getenv([]string{"REDIS_PORT"}, "6379"),
		RedisPassword: getenv([]string{"REDIS_PASSWORD"}, ""),

		JWTSecret:       getenv([]string{"JWT_SECRET"}, "change-me-in-production"),
		JWTAccessTTL:    getenvDuration([]string{"JWT_ACCESS_TTL_MINUTES"}, 15*time.Minute),
		RefreshTokenTTL: getenvDuration([]string{"REFRESH_TOKEN_TTL_DAYS"}, 30*24*time.Hour),

		AppName:     getenv([]string{"APP_NAME"}, "First App"),
		SMTPFrom:    getenv([]string{"SMTP_FROM", "MAIL_FROM"}, "noreply@firstapp.local"),
		BrevoAPIKey: getenv([]string{"BREVO_API_KEY"}, ""),
		BrevoAPIURL: getenv([]string{"BREVO_API_URL"}, "https://api.brevo.com/v3/smtp/email"),
		OTPExpire:   getenvDuration([]string{"OTP_EXPIRE_MINUTES"}, 15*time.Minute),
		OTPMaxTry:   getenvInt([]string{"OTP_MAX_ATTEMPTS"}, 5),
	}
}

// getenvInt membaca env sebagai integer dengan fallback.
func getenvInt(keys []string, fallback int) int {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				return n
			}
		}
	}
	return fallback
}

// getenvDuration membaca env sebagai angka lalu mengalikannya dengan unit.
// Mis. JWT_ACCESS_TTL_MINUTES=15 dengan unit time.Minute → 15 menit.
func getenvDuration(keys []string, fallback time.Duration) time.Duration {
	unit := fallback
	// Turunkan ke satuan menit/jam/hari agar nilai env mudah dibaca manusia.
	switch {
	case fallback%(24*time.Hour) == 0:
		unit = 24 * time.Hour
	case fallback%time.Hour == 0:
		unit = time.Hour
	case fallback%time.Minute == 0:
		unit = time.Minute
	}
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				return time.Duration(n) * unit
			}
		}
	}
	return fallback
}

// DSN (Data Source Name) membangun connection string PostgreSQL.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode,
	)
}

// RedisAddr mengembalikan alamat Redis dalam format host:port.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// Addr mengembalikan alamat listen, mis. ":3000".
func (c *Config) Addr() string {
	if c.AppPort != "" && c.AppPort[0] == ':' {
		return c.AppPort
	}
	return ":" + c.AppPort
}

// IsProduction mengecek apakah aplikasi berjalan di mode production.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// Validate memeriksa konfigurasi kritis dan fail-fast jika tidak aman.
// Dipanggil saat startup sebelum server menerima request.
func (c *Config) Validate() error {
	// JWT secret tidak boleh default/terlalu pendek — attacker bisa forge token.
	const minSecretLen = 32
	weakSecrets := map[string]bool{
		"change-me-in-production": true,
		"secret":                  true,
		"":                        true,
	}
	if weakSecrets[c.JWTSecret] || len(c.JWTSecret) < minSecretLen {
		return fmt.Errorf("JWT_SECRET tidak boleh kosong, default, atau kurang dari %d karakter — set nilai acak yang kuat di .env", minSecretLen)
	}

	// Di production, DB wajib pakai SSL.
	if c.IsProduction() && c.DBSSLMode == "disable" {
		return fmt.Errorf("DB_SSLMODE=disable tidak diizinkan di production — gunakan 'require' atau 'verify-full'")
	}

	return nil
}
