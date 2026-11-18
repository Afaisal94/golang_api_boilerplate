// Package model berisi entitas untuk resource Post.
// Kontrak akses data (PostRepository) juga didefinisikan di sini agar
// service hanya bergantung pada abstraksi, bukan package infrastructure.
package model

import (
	"errors"
	"time"
)

// ErrPostNotFound adalah error saat post tidak ditemukan.
// Lapisan luar (handler) memetakan error ini ke response 404.
var ErrPostNotFound = errors.New("post not found")

// Post adalah entitas untuk tabel posts.
// Setiap post wajib terelasi ke satu category (belongs to) via CategoryID.
type Post struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CategoryID uint      `json:"category_id"`
	Category   Category  `json:"category,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PostRepository adalah kontrak akses data untuk Post.
// Didefinisikan di model, bukan di infrastructure, agar usecase
// hanya bergantung pada abstraksi — bukan mengimpor package repositories.
type PostRepository interface {
	FindAll() ([]Post, error)
	FindByID(id string) (*Post, error)
	Create(post *Post) error
	Update(post *Post) error
	Delete(id string) error
}
