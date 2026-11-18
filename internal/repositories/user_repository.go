package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// userRepository adalah adapter database/sql untuk kontrak model.UserRepository.
type userRepository struct {
	db *sql.DB
}

// NewUserRepository membuat implementasi database/sql dari model.UserRepository.
func NewUserRepository(db *sql.DB) model.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, uuid, name, email, phone, password_hash, role_id, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash,
		&u.RoleID, &u.IsActive, &u.EmailVerifiedAt, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByUUID(uuid string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, uuid, name, email, phone, password_hash, role_id, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE uuid = $1 AND deleted_at IS NULL`, uuid,
	).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash,
		&u.RoleID, &u.IsActive, &u.EmailVerifiedAt, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, uuid, name, email, phone, password_hash, role_id, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash,
		&u.RoleID, &u.IsActive, &u.EmailVerifiedAt, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByPhone(phone string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		`SELECT id, uuid, name, email, phone, password_hash, role_id, is_active,
		        email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE phone = $1 AND deleted_at IS NULL`, phone,
	).Scan(
		&u.ID, &u.UUID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash,
		&u.RoleID, &u.IsActive, &u.EmailVerifiedAt, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO users (uuid, name, email, phone, password_hash, role_id, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		user.UUID, user.Name, user.Email, user.Phone,
		user.PasswordHash, user.RoleID, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) Update(user *model.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET name = $1, email = $2, phone = $3, password_hash = $4,
		        role_id = $5, is_active = $6, email_verified_at = $7,
		        last_login_at = $8, updated_at = NOW()
		 WHERE id = $9 AND deleted_at IS NULL`,
		user.Name, user.Email, user.Phone, user.PasswordHash,
		user.RoleID, user.IsActive, user.EmailVerifiedAt,
		user.LastLoginAt, user.ID,
	)
	return err
}
