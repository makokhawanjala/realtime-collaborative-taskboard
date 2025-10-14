// internal/handler/websocket/handler.go
package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (configure this properly in production)
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get board ID from query parameter
	boardID := c.Query("board_id")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "board_id is required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := NewClient(h.hub, conn, boardID)

	// Register client with hub
	h.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	log.Printf("🔌 WebSocket connection established (board: %s)", boardID)
}

// HandleStats returns WebSocket hub statistics
func (h *Handler) HandleStats(c *gin.Context) {
	stats := h.hub.GetStats()
	c.JSON(http.StatusOK, stats)
}
