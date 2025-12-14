package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"bamboo-rescue/pkg/jwt"
	"bamboo-rescue/pkg/response"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer token
	BearerPrefix = "Bearer "
	// UserIDKey is the context key for user ID
	UserIDKey = "userID"
	// UserEmailKey is the context key for user email
	UserEmailKey = "userEmail"
)

// Auth middleware validates JWT token and sets user context
func Auth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)

		c.Next()
	}
}

// OptionalAuth middleware validates JWT token if present, but doesn't require it
func OptionalAuth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) *uuid.UUID {
	if id, exists := c.Get(UserIDKey); exists {
		if userID, ok := id.(uuid.UUID); ok {
			return &userID
		}
	}
	return nil
}

// GetUserIDRequired retrieves the user ID from context, returns error if not found
func GetUserIDRequired(c *gin.Context) (uuid.UUID, bool) {
	userID := GetUserID(c)
	if userID == nil {
		return uuid.UUID{}, false
	}
	return *userID, true
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(c *gin.Context) string {
	if email, exists := c.Get(UserEmailKey); exists {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}
