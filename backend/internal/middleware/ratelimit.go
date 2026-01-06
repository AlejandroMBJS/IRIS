/*
Package middleware - Rate Limiting Middleware

==============================================================================
FILE: internal/middleware/ratelimit.go
==============================================================================

DESCRIPTION:
    Implements rate limiting to protect against brute force attacks and abuse.
    Uses a token bucket algorithm with in-memory storage (consider Redis for
    distributed deployments).

    Features:
    - Per-IP rate limiting for unauthenticated endpoints
    - Configurable limits for different endpoint types
    - Automatic cleanup of expired entries
    - Informative rate limit headers

USER PERSPECTIVE:
    - Protects accounts from brute force password attacks
    - Returns 429 Too Many Requests when limit exceeded
    - Includes Retry-After header for client guidance

DEVELOPER GUIDELINES:
    OK to modify: Rate limits, window sizes, key generation
    CAUTION: Memory usage grows with unique IPs - implement cleanup
    DO NOT modify: Core rate limiting logic without testing

PRODUCTION NOTES:
    - For multi-instance deployments, replace in-memory store with Redis
    - Consider implementing exponential backoff for repeated violations
    - Monitor rate limit hits in production logs

==============================================================================
*/
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per minute
	RequestsPerMinute int
	// WindowDuration is the time window for rate limiting
	WindowDuration time.Duration
	// CleanupInterval is how often to clean up expired entries
	CleanupInterval time.Duration
}

// rateLimitEntry tracks request count for a client
type rateLimitEntry struct {
	count      int
	windowStart time.Time
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	appConfig *config.AppConfig
	config    RateLimitConfig
	entries   map[string]*rateLimitEntry
	mu        sync.RWMutex
	stopClean chan struct{}
}

// DefaultAuthRateLimitConfig returns rate limit config for auth endpoints
// More restrictive: 10 requests per minute to prevent brute force
func DefaultAuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 10,
		WindowDuration:    time.Minute,
		CleanupInterval:   time.Minute * 5,
	}
}

// DefaultAPIRateLimitConfig returns rate limit config for general API endpoints
// Less restrictive: 60 requests per minute for normal usage
func DefaultAPIRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		WindowDuration:    time.Minute,
		CleanupInterval:   time.Minute * 5,
	}
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(appConfig *config.AppConfig, rlConfig RateLimitConfig) *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		appConfig: appConfig,
		config:    rlConfig,
		entries:   make(map[string]*rateLimitEntry),
		stopClean: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go rl.cleanupExpiredEntries()

	return rl
}

// Limit returns a Gin middleware that enforces rate limits
func (rl *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientIP := c.ClientIP()

		// Check rate limit
		allowed, remaining, resetTime := rl.checkAndIncrement(clientIP)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(rl.config.RequestsPerMinute)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		if !allowed {
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", string(rune(retryAfter)))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkAndIncrement checks if request is allowed and increments counter
// Returns: (allowed, remaining, resetTime)
func (rl *RateLimitMiddleware) checkAndIncrement(key string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	// Create new entry or reset if window expired
	if !exists || now.Sub(entry.windowStart) >= rl.config.WindowDuration {
		rl.entries[key] = &rateLimitEntry{
			count:       1,
			windowStart: now,
		}
		return true, rl.config.RequestsPerMinute - 1, now.Add(rl.config.WindowDuration)
	}

	// Check if limit exceeded
	if entry.count >= rl.config.RequestsPerMinute {
		resetTime := entry.windowStart.Add(rl.config.WindowDuration)
		return false, 0, resetTime
	}

	// Increment counter
	entry.count++
	remaining := rl.config.RequestsPerMinute - entry.count
	resetTime := entry.windowStart.Add(rl.config.WindowDuration)

	return true, remaining, resetTime
}

// cleanupExpiredEntries periodically removes expired rate limit entries
func (rl *RateLimitMiddleware) cleanupExpiredEntries() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, entry := range rl.entries {
				if now.Sub(entry.windowStart) >= rl.config.WindowDuration*2 {
					delete(rl.entries, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopClean:
			return
		}
	}
}

// Stop stops the cleanup goroutine (for graceful shutdown)
func (rl *RateLimitMiddleware) Stop() {
	close(rl.stopClean)
}

// AuthRateLimiter creates a rate limiter specifically for auth endpoints
func AuthRateLimiter(appConfig *config.AppConfig) *RateLimitMiddleware {
	return NewRateLimitMiddleware(appConfig, DefaultAuthRateLimitConfig())
}

// APIRateLimiter creates a rate limiter for general API endpoints
func APIRateLimiter(appConfig *config.AppConfig) *RateLimitMiddleware {
	return NewRateLimitMiddleware(appConfig, DefaultAPIRateLimitConfig())
}
