package dto

import (
	"time"

	"Afaisal94/first-app/internal/model"
)

// CreateCategoryRequest adalah DTO untuk POST /api/categories.
type CreateCategoryRequest struct {
	Category string `json:"category" binding:"required,min=2,max=100" example:"Teknologi"`
}

// UpdateCategoryRequest adalah DTO untuk PUT /api/categories/:id.
type UpdateCategoryRequest struct {
	Category string `json:"category" binding:"required,min=2,max=100" example:"Teknologi Terbaru"`
}

// CategoryResponse adalah DTO output untuk resource Category.
// Memisahkan bentuk response API dari entitas model.
type CategoryResponse struct {
	ID        uint      `json:"id" example:"1"`
	Category  string    `json:"category" example:"Teknologi"`
	CreatedAt time.Time `json:"created_at" example:"2026-09-24T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-09-24T10:00:00Z"`
}

// NewCategoryResponse memetakan satu entitas model Category ke CategoryResponse.
func NewCategoryResponse(category *model.Category) CategoryResponse {
	return CategoryResponse{
		ID:        category.ID,
		Category:  category.Category,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// NewCategoryResponseList memetakan slice entitas model ke slice CategoryResponse.
func NewCategoryResponseList(categories []model.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(categories))
	for i := range categories {
		out[i] = NewCategoryResponse(&categories[i])
	}
	return out
}
