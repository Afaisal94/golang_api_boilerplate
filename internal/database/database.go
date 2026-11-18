package database

import (
	"database/sql"
	"fmt"

	"Afaisal94/first-app/internal/config"
	"Afaisal94/first-app/pkg/logger"

	_ "github.com/lib/pq"
)

// Connect membuka koneksi PostgreSQL dari config, menjalankan migrasi manual,
// lalu mengembalikan *sql.DB untuk di-inject ke repository.
// Tidak memakai global var agar dependency eksplisit dan mudah diuji.
func Connect(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		logger.Default.Fatalf("failed to connect database: %v", err)
	}

	if err := db.Ping(); err != nil {
		logger.Default.Fatalf("failed to ping database: %v", err)
	}

	// Jalankan migrasi manual — buat tabel jika belum ada.
	if err := migrate(db); err != nil {
		logger.Default.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

// migrate membuat skema tabel categories dan posts jika belum ada.
// Urutan penting: categories dulu karena posts.category_id ber-FK ke categories.
func migrate(db *sql.DB) error {
	queries := []string{
		// Tabel roles (harus dibuat dulu karena users.role_id ber-FK ke roles)
		`CREATE TABLE IF NOT EXISTS roles (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		// Insert default roles jika belum ada
		`INSERT INTO roles (name) VALUES ('admin'), ('user') ON CONFLICT (name) DO NOTHING`,
		// Tabel users
		`CREATE TABLE IF NOT EXISTS users (
			id                SERIAL PRIMARY KEY,
			uuid              VARCHAR(36) NOT NULL UNIQUE,
			name              VARCHAR(100) NOT NULL,
			email             VARCHAR(255) NOT NULL UNIQUE,
			phone             VARCHAR(15) NOT NULL UNIQUE,
			password_hash     VARCHAR(255) NOT NULL,
			role_id           INTEGER NOT NULL REFERENCES roles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			is_active         BOOLEAN NOT NULL DEFAULT true,
			email_verified_at TIMESTAMP,
			last_login_at     TIMESTAMP,
			created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at        TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid)`,
		`CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at)`,
		// Tabel otps
		`CREATE TABLE IF NOT EXISTS otps (
			id          SERIAL PRIMARY KEY,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
			type        VARCHAR(20) NOT NULL,
			code_hash   VARCHAR(255) NOT NULL,
			expires_at  TIMESTAMP NOT NULL,
			verified_at TIMESTAMP,
			attempts    INTEGER NOT NULL DEFAULT 0,
			created_at  TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_otps_user_id_type ON otps(user_id, type)`,
		// Tabel refresh_tokens
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         SERIAL PRIMARY KEY,
			uuid       VARCHAR(36) NOT NULL UNIQUE,
			user_id    INTEGER NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			revoked_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`,
		// Tabel categories
		`CREATE TABLE IF NOT EXISTS categories (
			id         SERIAL PRIMARY KEY,
			category   VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at)`,
		// Tabel posts
		`CREATE TABLE IF NOT EXISTS posts (
			id          SERIAL PRIMARY KEY,
			title       VARCHAR(255) NOT NULL,
			content     TEXT NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts(category_id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	logger.Default.Info("database migration completed")
	return nil
}
