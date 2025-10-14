// internal/middleware/recovery.go
package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("💥 PANIC RECOVERED: %v", err)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
