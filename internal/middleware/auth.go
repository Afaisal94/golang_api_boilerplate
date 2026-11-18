package middleware

import (
	"strings"

	"Afaisal94/first-app/pkg/response"
	"Afaisal94/first-app/pkg/token"

	"github.com/gin-gonic/gin"
)

// Context keys untuk claims yang sudah tervalidasi.
const (
	CtxUserID    = "user_id"
	CtxUserUUID  = "user_uuid"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
)

// Auth memvalidasi header `Authorization: Bearer <access_token>`.
// Jika valid, claims disimpan di gin.Context lalu request diteruskan.
// Jika tidak, request dibatalkan dengan 401.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Unauthorized(c, "Invalid authorization header format. Use Bearer <token>")
			c.Abort()
			return
		}

		claims, err := token.ParseAccessToken(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			response.Unauthorized(c, "Invalid or expired access token")
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserUUID, claims.UUID)
		c.Set(CtxUserEmail, claims.Email)
		c.Set(CtxUserRole, claims.Role)
		c.Next()
	}
}
