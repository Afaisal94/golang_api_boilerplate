package routes

import (
	"Afaisal94/first-app/internal/handlers"
	"Afaisal94/first-app/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup mendaftarkan semua route API ke engine Gin.
// Handler di-inject dari main agar wiring terpusat di satu tempat.
// jwtSecret dipakai middleware auth untuk memvalidasi bearer access token.
func Setup(router *gin.Engine, health *handlers.HealthHandler, auth *handlers.AuthHandler, jwtSecret string, posts *handlers.PostHandler, categories *handlers.CategoryHandler) {
	// Security headers untuk semua response.
	router.Use(middleware.SecurityHeaders())

	// Body size limit 1MB — cegah payload besar membebani server.
	router.MaxMultipartMemory = 1 << 20

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World!"})
	})

	// Health check — dipakai Docker HEALTHCHECK dan load balancer.
	// Sengaja di luar grup /api dan tanpa auth agar bisa diakses internal.
	router.GET("/health", health.Health)

	// Swagger UI — dokumentasi API interaktif.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		// Auth routes — rate limit ketat (5 req/detik, burst 10) untuk cegah brute force.
		authGroup := api.Group("/auth")
		authGroup.Use(middleware.RateLimit(5, 10))
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/verify-email", auth.VerifyEmail)
			authGroup.POST("/resend-otp", auth.ResendOTP)
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/refresh", auth.Refresh)
			authGroup.POST("/logout", auth.Logout)
			authGroup.POST("/forgot-password", auth.ForgotPassword)
			authGroup.POST("/reset-password", auth.ResetPassword)
		}

		// Post routes — wajib bearer access token.
		postGroup := api.Group("/posts")
		postGroup.Use(middleware.Auth(jwtSecret))
		{
			postGroup.GET("", posts.FindPosts)
			postGroup.POST("", posts.StorePost)
			postGroup.GET("/:id", posts.FindPostById)
			postGroup.PUT("/:id", posts.UpdatePost)
			postGroup.DELETE("/:id", posts.DeletePost)
		}

		// Category routes — wajib bearer access token.
		categoryGroup := api.Group("/categories")
		categoryGroup.Use(middleware.Auth(jwtSecret))
		{
			categoryGroup.GET("", categories.FindCategories)
			categoryGroup.POST("", categories.StoreCategory)
			categoryGroup.GET("/:id", categories.FindCategoryById)
			categoryGroup.PUT("/:id", categories.UpdateCategory)
			categoryGroup.DELETE("/:id", categories.DeleteCategory)
		}
	}
}
