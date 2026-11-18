package repositories

import (
	"database/sql"
	"errors"

	"Afaisal94/first-app/internal/model"
)

// postRepository adalah adapter database/sql untuk kontrak model.PostRepository.
// Package ini boleh tahu database/sql (infrastruktur), tetapi tidak boleh
// membocorkan detail DB ke lapisan luar — error sql.ErrNoRows
// diterjemahkan menjadi model.ErrPostNotFound sebelum keluar dari sini.
type postRepository struct {
	db *sql.DB
}

// NewPostRepository membuat implementasi database/sql dari model.PostRepository.
func NewPostRepository(db *sql.DB) model.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) FindAll() ([]model.Post, error) {
	query := `
		SELECT
			p.id, p.title, p.content, p.category_id, p.created_at, p.updated_at,
			c.id, c.category, c.created_at, c.updated_at
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		ORDER BY p.id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		var cat model.Category
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Content, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt,
			&cat.ID, &cat.Category, &cat.CreatedAt, &cat.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.Category = cat
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) FindByID(id string) (*model.Post, error) {
	query := `
		SELECT
			p.id, p.title, p.content, p.category_id, p.created_at, p.updated_at,
			c.id, c.category, c.created_at, c.updated_at
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE p.id = $1`

	var p model.Post
	var cat model.Category
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Title, &p.Content, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt,
		&cat.ID, &cat.Category, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrPostNotFound
		}
		return nil, err
	}
	p.Category = cat
	return &p, nil
}

func (r *postRepository) Create(post *model.Post) error {
	return r.db.QueryRow(
		"INSERT INTO posts (title, content, category_id, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id, created_at, updated_at",
		post.Title, post.Content, post.CategoryID,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postRepository) Update(post *model.Post) error {
	_, err := r.db.Exec(
		"UPDATE posts SET title = $1, content = $2, category_id = $3, updated_at = NOW() WHERE id = $4",
		post.Title, post.Content, post.CategoryID, post.ID,
	)
	return err
}

func (r *postRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return model.ErrPostNotFound
	}
	return nil
}
