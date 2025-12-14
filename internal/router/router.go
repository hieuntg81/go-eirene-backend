package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"bamboo-rescue/internal/config"
	"bamboo-rescue/internal/handler"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Handlers holds all handler instances
type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Case         *handler.CaseHandler
	Media        *handler.MediaHandler
	Notification *handler.NotificationHandler
	Geocode      *handler.GeocodeHandler
}

// Setup initializes the router with all routes
func Setup(cfg *config.Config, handlers *Handlers, jwtService *jwt.Service, log *zap.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(log))
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler(log))

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.Requests, cfg.RateLimit.Duration)
	r.Use(middleware.RateLimit(rateLimiter))

	// Endpoint-specific rate limits
	endpointLimiter := middleware.NewEndpointRateLimiter(map[string]middleware.RateLimitEndpointConfig{
		"/api/cases/:id/accept": {Limit: 10, Window: time.Minute},
		"/api/media/upload":     {Limit: 20, Window: time.Minute},
	})
	r.Use(middleware.RateLimitEndpoint(endpointLimiter))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation
	if cfg.IsDevelopment() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/oauth", handlers.Auth.OAuth)
			auth.POST("/refresh", handlers.Auth.RefreshToken)
			auth.POST("/logout", middleware.Auth(jwtService), handlers.Auth.Logout)
		}

		// User routes (authenticated)
		users := api.Group("/users")
		users.Use(middleware.Auth(jwtService))
		{
			users.GET("/me", handlers.User.GetProfile)
			users.PUT("/me", handlers.User.UpdateProfile)
			users.PUT("/me/location", handlers.User.UpdateLocation)
			users.PUT("/me/availability", handlers.User.UpdateAvailability)
			users.GET("/me/preferences", handlers.User.GetPreferences)
			users.PUT("/me/preferences", handlers.User.UpdatePreferences)
			users.GET("/me/stats", handlers.User.GetStats)
			users.POST("/me/push-token", handlers.User.RegisterPushToken)
			users.DELETE("/me/push-token/:token", handlers.User.DeletePushToken)
		}

		// Case routes
		cases := api.Group("/cases")
		{
			// Public/optional auth routes
			cases.GET("", handlers.Case.GetCases)
			cases.POST("", middleware.OptionalAuth(jwtService), handlers.Case.Create)
			cases.GET("/nearby", handlers.Case.GetNearby)
			cases.GET("/:id", handlers.Case.GetByID)
			cases.GET("/:id/updates", handlers.Case.GetUpdates)
			cases.GET("/:id/volunteers", handlers.Case.GetVolunteers)
			cases.GET("/:id/comments", handlers.Case.GetComments)

			// Authenticated routes
			cases.GET("/my-cases", middleware.Auth(jwtService), handlers.Case.GetMyCases)
			cases.GET("/my-volunteer-cases", middleware.Auth(jwtService), handlers.Case.GetMyVolunteerCases)
			cases.PUT("/:id", middleware.Auth(jwtService), handlers.Case.Update)
			cases.DELETE("/:id", middleware.Auth(jwtService), handlers.Case.Delete)
			cases.POST("/:id/accept", middleware.Auth(jwtService), handlers.Case.Accept)
			cases.POST("/:id/withdraw", middleware.Auth(jwtService), handlers.Case.Withdraw)
			cases.PUT("/:id/volunteer-status", middleware.Auth(jwtService), handlers.Case.UpdateVolunteerStatus)
			cases.POST("/:id/updates", middleware.Auth(jwtService), handlers.Case.CreateUpdate)
			cases.POST("/:id/comments", middleware.Auth(jwtService), handlers.Case.CreateComment)
			cases.DELETE("/:id/comments/:commentId", middleware.Auth(jwtService), handlers.Case.DeleteComment)
		}

		// Media routes (authenticated)
		media := api.Group("/media")
		media.Use(middleware.Auth(jwtService))
		{
			media.POST("/upload", handlers.Media.Upload)
			media.POST("/upload-multiple", handlers.Media.UploadMultiple)
			media.DELETE("/:id", handlers.Media.Delete)
		}

		// Notification routes (authenticated)
		notifications := api.Group("/notifications")
		notifications.Use(middleware.Auth(jwtService))
		{
			notifications.GET("", handlers.Notification.GetNotifications)
			notifications.GET("/unread-count", handlers.Notification.GetUnreadCount)
			notifications.PUT("/:id/read", handlers.Notification.MarkAsRead)
			notifications.PUT("/read-all", handlers.Notification.MarkAllAsRead)
		}

		// Geocode routes (public)
		geocode := api.Group("/geocode")
		{
			geocode.GET("/reverse", handlers.Geocode.ReverseGeocode)
			geocode.GET("/search", handlers.Geocode.SearchAddress)
		}

		// Push token routes (authenticated) - separate endpoint for mobile compatibility
		pushTokens := api.Group("/push-tokens")
		pushTokens.Use(middleware.Auth(jwtService))
		{
			pushTokens.POST("", handlers.User.RegisterPushToken)
			pushTokens.DELETE("/:token", handlers.User.DeletePushToken)
		}
	}

	return r
}
