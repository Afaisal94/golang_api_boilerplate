package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// roleRepository adalah adapter database/sql untuk kontrak model.RoleRepository.
type roleRepository struct {
	db *sql.DB
}

// NewRoleRepository membuat implementasi database/sql dari model.RoleRepository.
func NewRoleRepository(db *sql.DB) model.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(
		"SELECT id, name, created_at, updated_at FROM roles WHERE name = $1", name,
	).Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.QueryRow(
		"SELECT id, name, created_at, updated_at FROM roles WHERE id = $1", id,
	).Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}
