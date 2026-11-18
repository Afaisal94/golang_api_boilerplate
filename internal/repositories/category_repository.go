package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// categoryRepository adalah adapter database/sql untuk kontrak model.CategoryRepository.
// Package ini boleh tahu database/sql (infrastruktur), tetapi tidak boleh
// membocorkan detail DB ke lapisan luar — error sql.ErrNoRows
// diterjemahkan menjadi model.ErrCategoryNotFound sebelum keluar dari sini.
// Delete memakai soft delete (mengisi deleted_at, row tetap ada).
type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository membuat implementasi database/sql dari model.CategoryRepository.
func NewCategoryRepository(db *sql.DB) model.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll() ([]model.Category, error) {
	rows, err := r.db.Query("SELECT id, category, created_at, updated_at FROM categories WHERE deleted_at IS NULL ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Category, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *categoryRepository) FindByID(id string) (*model.Category, error) {
	var c model.Category
	err := r.db.QueryRow("SELECT id, category, created_at, updated_at FROM categories WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&c.ID, &c.Category, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.QueryRow(
		"INSERT INTO categories (category, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id, created_at, updated_at",
		category.Category,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *categoryRepository) Update(category *model.Category) error {
	_, err := r.db.Exec(
		"UPDATE categories SET category = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL",
		category.Category, category.ID,
	)
	return err
}

func (r *categoryRepository) Delete(id string) error {
	// Soft delete: mengisi deleted_at dan mengecualikan row dari query berikutnya.
	result, err := r.db.Exec("UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return model.ErrCategoryNotFound
	}
	return nil
}
