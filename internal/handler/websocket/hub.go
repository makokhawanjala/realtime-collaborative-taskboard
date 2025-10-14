// internal/handler/websocket/hub.go
package websocket

import (
	"log"
	"sync"

	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
	"github.com/makokhawanjala/realtime-taskboard/internal/service"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients organized by board ID
	clients map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// WebSocket service for event subscription
	wsService *service.WebSocketService

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(wsService *service.WebSocketService) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		wsService:  wsService,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	log.Println("🚀 Starting WebSocket Hub...")

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize board clients map if it doesn't exist
	if h.clients[client.boardID] == nil {
		h.clients[client.boardID] = make(map[*Client]bool)

		// Subscribe to events for this board
		go h.subscribeToBoard(client.boardID)
	}

	// Add client to the board
	h.clients[client.boardID][client] = true

	log.Printf("✅ Client registered (board: %s, total: %d)",
		client.boardID, len(h.clients[client.boardID]))
}

// unregisterClient removes a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.boardID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			log.Printf("❌ Client unregistered (board: %s, remaining: %d)",
				client.boardID, len(h.clients[client.boardID]))

			// Clean up empty board entries
			if len(clients) == 0 {
				delete(h.clients, client.boardID)
				log.Printf("🗑️ No more clients for board: %s", client.boardID)
			}
		}
	}
}

// subscribeToBoard subscribes to events for a specific board
func (h *Hub) subscribeToBoard(boardID string) {
	log.Printf("🎧 Subscribing to events for board: %s", boardID)

	eventChan := h.wsService.Subscribe(boardID)
	defer h.wsService.Unsubscribe(boardID, eventChan)

	for event := range eventChan {
		h.broadcastToBoard(boardID, event)
	}

	log.Printf("🔌 Stopped listening to board: %s", boardID)
}

// broadcastToBoard broadcasts an event to all clients of a board
func (h *Hub) broadcastToBoard(boardID string, event *domain.WebSocketEvent) {
	h.mu.RLock()
	clients := h.clients[boardID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	log.Printf("📡 Broadcasting to %d clients (board: %s, type: %s)",
		len(clients), boardID, event.Type)

	for client := range clients {
		if err := client.SendEvent(event); err != nil {
			log.Printf("❌ Failed to send event to client: %v", err)
		}
	}
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalClients := 0
	boardStats := make(map[string]int)

	for boardID, clients := range h.clients {
		count := len(clients)
		boardStats[boardID] = count
		totalClients += count
	}

	return map[string]interface{}{
		"total_clients": totalClients,
		"boards":        len(h.clients),
		"board_stats":   boardStats,
	}
}
