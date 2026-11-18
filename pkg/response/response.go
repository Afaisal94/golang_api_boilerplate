package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope adalah format standar semua response API.
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// JSON menulis response dengan envelope standar.
func JSON(c *gin.Context, status int, success bool, message string, data any) {
	c.JSON(status, Envelope{
		Success: success,
		Message: message,
		Data:    data,
	})
}

// Success 200 OK dengan data.
func Success(c *gin.Context, message string, data any) {
	JSON(c, http.StatusOK, true, message, data)
}

// Created 201 Created dengan data.
func Created(c *gin.Context, message string, data any) {
	JSON(c, http.StatusCreated, true, message, data)
}

// OKWithoutData 200 OK tanpa data (mis. delete).
func OKWithoutData(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message})
}

// NotFound 404 — record tidak ditemukan.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Envelope{Success: false, Message: message})
}

// BadRequest 400 — request tidak valid + detail error validasi.
func BadRequest(c *gin.Context, message string, errors any) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// BadRequestWithError 400 — body tidak valid + pesan error mentah.
func BadRequestWithError(c *gin.Context, message, err string) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// Unauthorized 401 — kredensial tidak valid / belum login.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Envelope{Success: false, Message: message})
}

// Forbidden 403 — tidak punya akses (mis. akun nonaktif).
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Envelope{Success: false, Message: message})
}

// Internal 500 — kesalahan server.
// Catatan: err dikirim ke client — HANYA untuk development/debug.
// Di production, gunakan InternalSafe agar tidak membocorkan detail internal.
func Internal(c *gin.Context, message, err string) {
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// InternalSafe 500 — kesalahan server tanpa membocorkan detail internal ke client.
// err hanya dicatat di log server, tidak dikirim ke response.
func InternalSafe(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Message: message,
	})
}
