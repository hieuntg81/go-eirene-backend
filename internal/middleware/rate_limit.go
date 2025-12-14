package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"bamboo-rescue/pkg/response"
)

// RateLimiter represents an in-memory rate limiter
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes stale entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.Sub(info.lastSeen) > rl.window {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[key]

	if !exists {
		rl.requests[key] = &clientInfo{count: 1, lastSeen: now}
		return true
	}

	// Reset if window has passed
	if now.Sub(info.lastSeen) > rl.window {
		info.count = 1
		info.lastSeen = now
		return true
	}

	// Check limit
	if info.count >= rl.limit {
		return false
	}

	info.count++
	info.lastSeen = now
	return true
}

// RateLimit middleware limits requests per client
func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP address as key, or user ID if authenticated
		key := c.ClientIP()
		if userID := GetUserID(c); userID != nil {
			key = userID.String()
		}

		if !limiter.Allow(key) {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimiter provides rate limiting per endpoint
type EndpointRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	configs  map[string]RateLimitEndpointConfig
}

// RateLimitEndpointConfig holds rate limit config for specific endpoints
type RateLimitEndpointConfig struct {
	Limit  int
	Window time.Duration
}

// NewEndpointRateLimiter creates a new endpoint-specific rate limiter
func NewEndpointRateLimiter(configs map[string]RateLimitEndpointConfig) *EndpointRateLimiter {
	erl := &EndpointRateLimiter{
		limiters: make(map[string]*RateLimiter),
		configs:  configs,
	}

	// Initialize limiters for each configured endpoint
	for path, config := range configs {
		erl.limiters[path] = NewRateLimiter(config.Limit, config.Window)
	}

	return erl
}

// RateLimitEndpoint middleware limits requests for specific endpoints
func RateLimitEndpoint(erl *EndpointRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()

		erl.mu.RLock()
		limiter, exists := erl.limiters[path]
		erl.mu.RUnlock()

		if !exists {
			c.Next()
			return
		}

		key := c.ClientIP()
		if userID := GetUserID(c); userID != nil {
			key = userID.String()
		}

		if !limiter.Allow(key) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Rate limit exceeded for this endpoint",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
