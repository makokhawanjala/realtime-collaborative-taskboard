// internal/middleware/ratelimit.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	requests map[string]*bucket
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*bucket),
		rate:     requestsPerWindow,
		window:   window,
	}

	// Cleanup old entries every minute
	go rl.cleanup()

	return rl
}

// RateLimit middleware enforces rate limiting per IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please slow down",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.requests[key]

	if !exists {
		rl.requests[key] = &bucket{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		return true
	}

	// Reset bucket if window has passed
	if now.Sub(b.lastReset) > rl.window {
		b.tokens = rl.rate - 1
		b.lastReset = now
		return true
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.requests {
			if now.Sub(b.lastReset) > rl.window*2 {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}
