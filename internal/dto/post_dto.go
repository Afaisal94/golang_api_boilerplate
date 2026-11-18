package dto

import (
	"time"

	"Afaisal94/first-app/internal/model"
)

// CreatePostRequest adalah DTO untuk POST /api/posts.
type CreatePostRequest struct {
	Title      string `json:"title" binding:"required,min=3,max=255" example:"Judul Post Pertama"`
	Content    string `json:"content" binding:"required,min=3" example:"Ini adalah isi dari post pertama."`
	CategoryID uint   `json:"category_id" binding:"required,gt=0" example:"1"`
}

// UpdatePostRequest adalah DTO untuk PUT /api/posts/:id.
type UpdatePostRequest struct {
	Title      string `json:"title" binding:"required,min=3,max=255" example:"Judul Post Diperbarui"`
	Content    string `json:"content" binding:"required,min=3" example:"Ini adalah isi post yang sudah diperbarui."`
	CategoryID uint   `json:"category_id" binding:"required,gt=0" example:"1"`
}

// PostResponse adalah DTO output untuk resource Post.
// Memisahkan bentuk response API dari entitas model.
// Category disematkan agar konsumen API tahu relasi tanpa request tambahan.
type PostResponse struct {
	ID         uint             `json:"id" example:"1"`
	Title      string           `json:"title" example:"Judul Post Pertama"`
	Content    string           `json:"content" example:"Ini adalah isi dari post pertama."`
	CategoryID uint             `json:"category_id" example:"1"`
	Category   CategoryResponse `json:"category"`
	CreatedAt  time.Time        `json:"created_at" example:"2026-09-21T10:00:00Z"`
	UpdatedAt  time.Time        `json:"updated_at" example:"2026-09-21T10:00:00Z"`
}

// NewPostResponse memetakan satu entitas model Post ke PostResponse.
func NewPostResponse(post *model.Post) PostResponse {
	return PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		Content:    post.Content,
		CategoryID: post.CategoryID,
		Category:   NewCategoryResponse(&post.Category),
		CreatedAt:  post.CreatedAt,
		UpdatedAt:  post.UpdatedAt,
	}
}

// NewPostResponseList memetakan slice entitas model ke slice PostResponse.
func NewPostResponseList(posts []model.Post) []PostResponse {
	out := make([]PostResponse, len(posts))
	for i := range posts {
		out[i] = NewPostResponse(&posts[i])
	}
	return out
}
