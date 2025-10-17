// internal/middleware/logger.go
package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a custom logging middleware with emojis
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// Get request ID if available
		requestID, _ := c.Get("request_id")
		reqIDStr := ""
		if requestID != nil {
			reqIDStr = fmt.Sprintf(" [%s]", requestID.(string)[:8])
		}

		// Choose emoji based on status code
		var statusEmoji string
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusEmoji = "✅"
		case statusCode >= 300 && statusCode < 400:
			statusEmoji = "↪️"
		case statusCode >= 400 && statusCode < 500:
			statusEmoji = "⚠️"
		case statusCode >= 500:
			statusEmoji = "❌"
		default:
			statusEmoji = "❓"
		}

		// Log with colors and emojis
		fmt.Printf("%s [%s] %d | %12v | %s | %s %s%s\n",
			statusEmoji,
			start.Format("2006-01-02 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			reqIDStr,
		)
	}
}
