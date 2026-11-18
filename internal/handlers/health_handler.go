package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler menangani health check untuk orchestrator (Docker, K8s, load balancer).
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler membuat handler dengan dependency yang di-inject.
func NewHealthHandler(db *sql.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient}
}

// HealthResponse merinci status tiap dependency.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}

// Health GET /health — health check dependency (DB & Redis).
// @Summary      Health check
// @Description  Memeriksa status server beserta koneksi database dan Redis.
// @Tags         Health
// @Produce      json
// @Success      200 {object} response.Envelope{data=handlers.HealthResponse}
// @Failure      503 {object} response.Envelope
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// Batasi waktu probe agar health check tidak menggantung saat dependency lambat.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	result := HealthResponse{
		Status:   "ok",
		Database: "ok",
		Redis:    "ok",
	}
	healthy := true

	if err := h.db.PingContext(ctx); err != nil {
		logger.Default.Errorf("health: database ping failed: %v", err)
		result.Database = "down"
		healthy = false
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		logger.Default.Errorf("health: redis ping failed: %v", err)
		result.Redis = "down"
		healthy = false
	}

	if !healthy {
		result.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, response.Envelope{
			Success: false,
			Message: "Service unhealthy",
			Data:    result,
		})
		return
	}

	response.Success(c, "Service healthy", result)
}
