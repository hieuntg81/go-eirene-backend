package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey = "requestID"
)

// Logger middleware logs HTTP requests
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Get status code
		statusCode := c.Writer.Status()

		// Get user ID if available
		var userID string
		if uid := GetUserID(c); uid != nil {
			userID = uid.String()
		}

		// Log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log based on status code
		if statusCode >= 500 {
			log.Error("Server error", fields...)
		} else if statusCode >= 400 {
			log.Warn("Client error", fields...)
		} else {
			log.Info("Request completed", fields...)
		}
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}
