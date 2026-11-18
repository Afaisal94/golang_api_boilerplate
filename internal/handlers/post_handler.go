package handlers

import (
	"errors"

	"Afaisal94/first-app/internal/dto"
	"Afaisal94/first-app/internal/model"
	"Afaisal94/first-app/internal/services"
	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/response"
	appvalidator "Afaisal94/first-app/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// bindAndValidate mengikat JSON body lalu memformat error validasi.
// Menulis response 400 sendiri dan mengembalikan false jika gagal.
// Binding gin tetap di handler — pkg/validator murni tanpa gin.
func bindAndValidate(c *gin.Context, input any) bool {
	if err := c.ShouldBindJSON(input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			response.BadRequest(c, "Validation error", appvalidator.Format(ve))
			return false
		}
		response.BadRequestWithError(c, "Invalid request body", err.Error())
		return false
	}
	return true
}

// PostHandler menangani HTTP untuk resource Post.
// Semua business logic dilakukan lewat service, bukan repository/GORM langsung.
type PostHandler struct {
	service services.PostService
}

// NewPostHandler membuat handler dengan service yang di-inject.
func NewPostHandler(service services.PostService) *PostHandler {
	return &PostHandler{service: service}
}

// FindPosts GET /api/posts — list semua post.
// @Summary      Get all posts
// @Description  Mengambil semua data post.
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Envelope{data=[]dto.PostResponse}
// @Failure      401 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /posts [get]
func (h *PostHandler) FindPosts(c *gin.Context) {
	posts, err := h.service.FindAll()
	if err != nil {
		logger.Default.Errorf("FindPosts failed: %v", err)
		response.InternalSafe(c, "Failed to fetch posts")
		return
	}

	response.Success(c, "Lists Data Posts", dto.NewPostResponseList(posts))
}

// StorePost POST /api/posts — buat post baru.
// @Summary      Create a post
// @Description  Membuat post baru dengan title, content, dan category_id yang wajib merujuk ke category yang ada.
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreatePostRequest true "Post data"
// @Success      201 {object} response.Envelope{data=dto.PostResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /posts [post]
func (h *PostHandler) StorePost(c *gin.Context) {
	var input dto.CreatePostRequest
	if !bindAndValidate(c, &input) {
		return
	}

	post, err := h.service.Create(input.Title, input.Content, input.CategoryID)
	if err != nil {
		if errors.Is(err, model.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found!")
			return
		}
		logger.Default.Errorf("StorePost failed: %v", err)
		response.InternalSafe(c, "Failed to create post")
		return
	}

	response.Created(c, "Post Created Successfully", dto.NewPostResponse(post))
}

// FindPostById GET /api/posts/:id — detail post.
// @Summary      Get post by ID
// @Description  Mengambil detail satu post berdasarkan ID.
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Success      200 {object} response.Envelope{data=dto.PostResponse}
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /posts/{id} [get]
func (h *PostHandler) FindPostById(c *gin.Context) {
	post, err := h.service.FindByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, model.ErrPostNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		logger.Default.Errorf("FindPostById failed: %v", err)
		response.InternalSafe(c, "Failed to fetch post")
		return
	}

	response.Success(c, "Detail Data Post By ID : "+c.Param("id"), dto.NewPostResponse(post))
}

// UpdatePost PUT /api/posts/:id — update post.
// @Summary      Update a post
// @Description  Mengubah title, content, dan category_id post berdasarkan ID. category_id wajib merujuk ke category yang ada.
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Param        request body dto.UpdatePostRequest true "Post data"
// @Success      200 {object} response.Envelope{data=dto.PostResponse}
// @Failure      400 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	var input dto.UpdatePostRequest
	if !bindAndValidate(c, &input) {
		return
	}

	post, err := h.service.Update(c.Param("id"), input.Title, input.Content, input.CategoryID)
	if err != nil {
		if errors.Is(err, model.ErrPostNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		if errors.Is(err, model.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found!")
			return
		}
		logger.Default.Errorf("UpdatePost failed: %v", err)
		response.InternalSafe(c, "Failed to update post")
		return
	}

	response.Success(c, "Post Updated Successfully", dto.NewPostResponse(post))
}

// DeletePost DELETE /api/posts/:id — hapus post.
// @Summary      Delete a post
// @Description  Menghapus post berdasarkan ID.
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Success      200 {object} response.Envelope
// @Failure      401 {object} response.Envelope
// @Failure      404 {object} response.Envelope
// @Failure      500 {object} response.Envelope
// @Router       /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		if errors.Is(err, model.ErrPostNotFound) {
			response.NotFound(c, "Record not found!")
			return
		}
		logger.Default.Errorf("DeletePost failed: %v", err)
		response.InternalSafe(c, "Failed to delete post")
		return
	}

	response.OKWithoutData(c, "Post Deleted Successfully")
}
