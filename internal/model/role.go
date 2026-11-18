package model

import (
	"errors"
	"time"
)

// Sentinel errors untuk Role.
var (
	ErrRoleNotFound = errors.New("role not found")
)

// Role adalah entitas untuk tabel roles.
type Role struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RoleRepository adalah kontrak akses data untuk Role.
type RoleRepository interface {
	FindByName(name string) (*Role, error)
	FindByID(id uint) (*Role, error)
}
