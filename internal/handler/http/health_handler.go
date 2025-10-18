// internal/handler/http/health_handler.go
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *sqlx.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sqlx.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime,omitempty"`
}

var startTime = time.Now()

// Check performs a detailed health check
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// Check PostgreSQL
	if err := h.db.PingContext(ctx); err != nil {
		services["postgres"] = "❌ unhealthy"
	} else {
		services["postgres"] = "✅ healthy"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		services["redis"] = "❌ unhealthy"
	} else {
		services["redis"] = "✅ healthy"
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, status := range services {
		if status[:1] == "❌" {
			overallStatus = "degraded"
			break
		}
	}

	uptime := time.Since(startTime).Round(time.Second).String()

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Unix(),
		Services:  services,
		Uptime:    uptime,
	}

	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
