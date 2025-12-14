package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bamboo-rescue/internal/config"
	"bamboo-rescue/internal/handler"
	"bamboo-rescue/internal/repository"
	"bamboo-rescue/internal/router"
	"bamboo-rescue/internal/service"
	"bamboo-rescue/pkg/database"
	"bamboo-rescue/pkg/jwt"
	"bamboo-rescue/pkg/migration"
	"bamboo-rescue/pkg/storage"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// @title Rescue App API
// @version 1.0
// @description API server for Rescue App - Emergency Response Platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.rescue-app.com/support
// @contact.email support@rescue-app.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger
	log, err := initLogger()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	log.Info("Starting Rescue App API",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := database.NewPostgresDB(&cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	// Run database migrations if enabled
	if cfg.Database.AutoMigrate {
		log.Info("Running database migrations...")
		if err := migration.Run(db, "migrations", log); err != nil {
			log.Warn("Migration failed (this may be OK if tables already exist or PostGIS is not installed)", zap.Error(err))
			log.Info("Continuing without migration - please ensure database schema is set up correctly")
		}
	} else {
		log.Info("Auto-migration is disabled, skipping...")
	}

	// Initialize JWT service
	jwtService := jwt.NewService(&cfg.JWT)

	// Initialize S3 storage
	storageClient, err := storage.NewS3Client(cfg, log)
	if err != nil {
		log.Warn("Failed to initialize S3 client, using mock storage", zap.Error(err))
	}

	// Initialize repositories
	repos := initRepositories(db)

	// Initialize services
	services := initServices(repos, jwtService, storageClient, cfg, log)

	// Initialize handlers
	handlers := initHandlers(services)

	// Setup router
	r := router.Setup(cfg, handlers, jwtService, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server is running", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited properly")
}

func initLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build()
}

// Repositories holds all repository instances
type Repositories struct {
	User         repository.UserRepository
	Case         repository.CaseRepository
	Notification repository.NotificationRepository
	Media        repository.MediaRepository
}

func initRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         repository.NewUserRepository(db),
		Case:         repository.NewCaseRepository(db),
		Notification: repository.NewNotificationRepository(db),
		Media:        repository.NewMediaRepository(db),
	}
}

// Services holds all service instances
type Services struct {
	Auth         service.AuthService
	User         service.UserService
	Case         service.CaseService
	Media        service.MediaService
	Notification service.NotificationService
	Geocode      service.GeocodeService
	FCM          service.FCMService
}

func initServices(repos *Repositories, jwtSvc *jwt.Service, storageClient storage.Client, cfg *config.Config, log *zap.Logger) *Services {
	// Initialize FCM service
	fcmSvc, err := service.NewFCMService(cfg, repos.User, log)
	if err != nil {
		log.Warn("Failed to initialize FCM service", zap.Error(err))
	}

	notificationSvc := service.NewNotificationService(repos.Notification, log)

	return &Services{
		Auth:         service.NewAuthService(repos.User, jwtSvc, log),
		User:         service.NewUserService(repos.User, log),
		Case:         service.NewCaseService(repos.Case, repos.User, notificationSvc, fcmSvc, log),
		Media:        service.NewMediaService(repos.Media, storageClient, log),
		Notification: notificationSvc,
		Geocode:      service.NewGeocodeService(cfg, log),
		FCM:          fcmSvc,
	}
}

func initHandlers(services *Services) *router.Handlers {
	return &router.Handlers{
		Auth:         handler.NewAuthHandler(services.Auth),
		User:         handler.NewUserHandler(services.User),
		Case:         handler.NewCaseHandler(services.Case),
		Media:        handler.NewMediaHandler(services.Media),
		Notification: handler.NewNotificationHandler(services.Notification),
		Geocode:      handler.NewGeocodeHandler(services.Geocode),
	}
}
