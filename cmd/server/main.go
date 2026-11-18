package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "Afaisal94/first-app/docs" // docs swagger yang digenerate swag init
	"Afaisal94/first-app/internal/config"
	"Afaisal94/first-app/internal/database"
	"Afaisal94/first-app/internal/handlers"
	"Afaisal94/first-app/internal/repositories"
	"Afaisal94/first-app/internal/routes"
	"Afaisal94/first-app/internal/services"
	"Afaisal94/first-app/pkg/logger"
	"Afaisal94/first-app/pkg/mailer"
	appvalidator "Afaisal94/first-app/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// @title           First App REST API
// @version         1.0
// @description     REST API — contoh clean architecture dengan Gin, GORM, dan PostgreSQL.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3000
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer" diikuti spasi dan access token JWT.
func main() {
	// load konfigurasi (.env + environment)
	cfg := config.Load()

	// fail-fast: validasi konfigurasi kritis sebelum server mulai.
	if err := cfg.Validate(); err != nil {
		logger.Default.Fatalf("config validation failed: %v", err)
	}

	// Set Gin mode berdasarkan APP_ENV.
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// koneksi database + auto migrate
	db := database.Connect(cfg)

	// koneksi redis (cache/session) — fail-fast jika tidak bisa connect.
	redisClient := database.ConnectRedis(cfg)

	// wiring dependency langsung di main (repo -> service -> handler per fitur).
	// Category di-wire lebih dulu karena PostService butuh categoryRepo
	// untuk memvalidasi relasi category_id saat create/update.
	categoryRepo := repositories.NewCategoryRepository(db)
	postRepo := repositories.NewPostRepository(db)
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	otpRepo := repositories.NewOTPRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	logger.Default.Infof("brevo config: from=%s (api_key_set=%t api_key_len=%d)", cfg.SMTPFrom, cfg.BrevoAPIKey != "", len(cfg.BrevoAPIKey))
	mailSender := mailer.New(mailer.Config{
		From:    cfg.SMTPFrom,
		AppName: cfg.AppName,
		APIKey:  cfg.BrevoAPIKey,
		APIURL:  cfg.BrevoAPIURL,
	})

	categoryService := services.NewCategoryService(categoryRepo)
	postService := services.NewPostService(postRepo, categoryRepo)
	authService := services.NewAuthService(cfg, userRepo, roleRepo, otpRepo, refreshTokenRepo, mailSender)

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	postHandler := handlers.NewPostHandler(postService)
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(db, redisClient)

	// inisialisasi Gin
	router := gin.Default()

	// daftarkan validator khusus auth (mis. eqfield untuk confirm_password)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		appvalidator.RegisterAuthValidators(v)
	}

	// daftarkan semua route
	routes.Setup(router, healthHandler, authHandler, cfg.JWTSecret, postHandler, categoryHandler)

	// HTTP server dengan timeout untuk mencegah slowloris dan request yang hang.
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Jalankan server di goroutine agar main goroutine bisa menangkap sinyal shutdown.
	go func() {
		logger.Default.Infof("server running on %s (env=%s)", cfg.Addr(), cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Default.Fatalf("failed to run server: %v", err)
		}
	}()

	// Tunggu sinyal interrupt (Ctrl+C) atau SIGTERM untuk graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Default.Info("shutting down server...")

	// Beri waktu 30 detik untuk request in-flight selesai.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Default.Errorf("server forced to shutdown: %v", err)
	}

	// Tutup koneksi database.
	if err := db.Close(); err != nil {
		logger.Default.Errorf("database close failed: %v", err)
	}

	// Tutup koneksi redis.
	if err := redisClient.Close(); err != nil {
		logger.Default.Errorf("redis close failed: %v", err)
	}

	logger.Default.Info("server exited")
}
