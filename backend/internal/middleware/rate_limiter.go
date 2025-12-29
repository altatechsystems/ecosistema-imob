package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds the rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Requests allowed per second
	Burst             int           // Maximum burst size
	CleanupInterval   time.Duration // Interval to cleanup old limiters
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10.0,            // 10 requests per second
		Burst:             20,               // Burst of up to 20 requests
		CleanupInterval:   5 * time.Minute, // Cleanup every 5 minutes
	}
}

// StrictRateLimiterConfig returns strict rate limiter for public endpoints
func StrictRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 2.0,             // 2 requests per second
		Burst:             5,                // Burst of up to 5 requests
		CleanupInterval:   5 * time.Minute, // Cleanup every 5 minutes
	}
}

// visitor holds rate limiter and last seen time
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a middleware that limits requests per IP
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		config:   config,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// getVisitor retrieves or creates a rate limiter for the given IP
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.Burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes old visitors periodically
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			// Remove visitors not seen in the last 3 minutes
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit returns a Gin middleware that limits requests per IP
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Get limiter for this IP
		limiter := rl.getVisitor(ip)

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit is a convenience function to create a rate limiter with default config
func RateLimit() gin.HandlerFunc {
	return NewRateLimiter(DefaultRateLimiterConfig()).Limit()
}

// StrictRateLimit is a convenience function for strict rate limiting (public endpoints)
func StrictRateLimit() gin.HandlerFunc {
	return NewRateLimiter(StrictRateLimiterConfig()).Limit()
}
