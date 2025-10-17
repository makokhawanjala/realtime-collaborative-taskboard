// internal/middleware/requestid.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in context and response header
		c.Set("request_id", requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}
