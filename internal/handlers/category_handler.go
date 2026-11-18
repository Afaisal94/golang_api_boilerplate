package handlers

import (
	"errors"

	"Afaisal94/first-app/internal/dto"
	"Afaisal94/first-app/internal/model"
	"Afaisal94/first-app/internal/services"
	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/response"

	"github.com/gin-gonic/gin"
)

// CategoryHandler menangani HTTP untuk resource Category.
// Semua business logic dilakukan lewat service, bukan repository/GORM langsung.
type CategoryHandler struct {
	service services.CategoryService
}

// NewCategoryHandler membuat handler dengan service yang di-inject.
func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// FindCategories GET /api/categories — list semua category.
// @Summary      Get all categories
// @Description  Mengambil semua data category.
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Envelope{data=[]dto.CategoryResponse}
// @Failure      401 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /categories [get]
func (h *CategoryHandler) FindCategories(c *gin.Context) {
	categories, err := h.service.FindAll()
	if err != nil {
		logger.Default.Errorf("FindCategories failed: %v", err)
		response.InternalSafe(c, "Failed to fetch categories")
		return
	}

	response.Success(c, "Lists Data Categories", dto.NewCategoryResponseList(categories))
}

// StoreCategory POST /api/categories — buat category baru.
// @Summary      Create a category
// @Description  Membuat category baru dengan nama category.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCategoryRequest true "Category data"
// @Success      201 {object} response.Envelope{data=dto.CategoryResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /categories [post]
func (h *CategoryHandler) StoreCategory(c *gin.Context) {
	var input dto.CreateCategoryRequest
	if !bindAndValidate(c, &input) {
		return
	}

	category, err := h.service.Create(input.Category)
	if err != nil {
		logger.Default.Errorf("StoreCategory failed: %v", err)
		response.InternalSafe(c, "Failed to create category")
		return
	}

	response.Created(c, "Category Created Successfully", dto.NewCategoryResponse(category))
}

// FindCategoryById GET /api/categories/:id — detail category.
// @Summary      Get category by ID
// @Description  Mengambil detail satu category berdasarkan ID.
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} response.Envelope{data=dto.CategoryResponse}
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /categories/{id} [get]
func (h *CategoryHandler) FindCategoryById(c *gin.Context) {
	category, err := h.service.FindByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, model.ErrCategoryNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		logger.Default.Errorf("FindCategoryById failed: %v", err)
		response.InternalSafe(c, "Failed to fetch category")
		return
	}

	response.Success(c, "Detail Data Category By ID : "+c.Param("id"), dto.NewCategoryResponse(category))
}

// UpdateCategory PUT /api/categories/:id — update category.
// @Summary      Update a category
// @Description  Mengubah nama category berdasarkan ID.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Param        request body dto.UpdateCategoryRequest true "Category data"
// @Success      200 {object} response.Envelope{data=dto.CategoryResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var input dto.UpdateCategoryRequest
	if !bindAndValidate(c, &input) {
		return
	}

	category, err := h.service.Update(c.Param("id"), input.Category)
	if err != nil {
		if errors.Is(err, model.ErrCategoryNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		logger.Default.Errorf("UpdateCategory failed: %v", err)
		response.InternalSafe(c, "Failed to update category")
		return
	}

	response.Success(c, "Category Updated Successfully", dto.NewCategoryResponse(category))
}

// DeleteCategory DELETE /api/categories/:id — hapus category (soft delete).
// @Summary      Delete a category
// @Description  Menghapus category berdasarkan ID (soft delete — row ditandai deleted_at, tidak benar-benar dihapus).
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		if errors.Is(err, model.ErrCategoryNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		logger.Default.Errorf("DeleteCategory failed: %v", err)
		response.InternalSafe(c, "Failed to delete category")
		return
	}

	response.OKWithoutData(c, "Category Deleted Successfully")
}
