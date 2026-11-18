package model

import (
	"errors"
	"time"
)

// ErrCategoryNotFound adalah error saat category tidak ditemukan.
// Lapisan luar (handler) memetakan error ini ke response 404.
var ErrCategoryNotFound = errors.New("category not found")

// Category adalah entitas untuk tabel categories.
// Soft delete: row tidak benar-benar dihapus, hanya ditandai deleted_at,
// dan semua query default otomatis mengecualikannya.
type Category struct {
	ID        uint       `json:"id"`
	Category  string     `json:"category"`
	Posts     []Post     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// CategoryRepository adalah kontrak akses data untuk Category.
// Didefinisikan di model, bukan di infrastructure, agar usecase
// hanya bergantung pada abstraksi — bukan mengimpor package repositories.
type CategoryRepository interface {
	FindAll() ([]Category, error)
	FindByID(id string) (*Category, error)
	Create(category *Category) error
	Update(category *Category) error
	Delete(id string) error
}
