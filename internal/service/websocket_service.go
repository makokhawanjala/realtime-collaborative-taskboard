// internal/service/websocket_service.go
package service

import (
	"context"
	"log"
	"sync"

	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
	"github.com/makokhawanjala/realtime-taskboard/internal/repository/redis"
)

// WebSocketService manages WebSocket broadcast operations
type WebSocketService struct {
	pubsubClient *redis.PubSubClient
	subscribers  map[string][]chan *domain.WebSocketEvent
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(pubsubClient *redis.PubSubClient) *WebSocketService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &WebSocketService{
		pubsubClient: pubsubClient,
		subscribers:  make(map[string][]chan *domain.WebSocketEvent),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start listening to Redis events
	go service.startRedisListener()

	return service
}

// startRedisListener listens to Redis pub/sub and broadcasts to subscribers
func (s *WebSocketService) startRedisListener() {
	log.Println("🎧 Starting Redis event listener...")

	eventChan, err := s.pubsubClient.SubscribeToTaskEvents(s.ctx)
	if err != nil {
		log.Printf("❌ Failed to subscribe to task events: %v", err)
		return
	}

	for {
		select {
		case <-s.ctx.Done():
			log.Println("🔌 Stopping Redis event listener")
			return
		case event, ok := <-eventChan:
			if !ok {
				log.Println("⚠️ Event channel closed")
				return
			}

			// Broadcast to all subscribers for this board
			s.broadcastToBoard(event.BoardID.String(), event)
		}
	}
}

// Subscribe subscribes a client to a specific board's events
func (s *WebSocketService) Subscribe(boardID string) <-chan *domain.WebSocketEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a buffered channel for this subscriber
	eventChan := make(chan *domain.WebSocketEvent, 100)

	// Add to subscribers map
	s.subscribers[boardID] = append(s.subscribers[boardID], eventChan)

	log.Printf("➕ New subscriber for board: %s (total: %d)", boardID, len(s.subscribers[boardID]))

	return eventChan
}

// Unsubscribe removes a subscriber from a board
func (s *WebSocketService) Unsubscribe(boardID string, eventChan <-chan *domain.WebSocketEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers := s.subscribers[boardID]
	for i, ch := range subscribers {
		if ch == eventChan {
			// Remove from slice
			s.subscribers[boardID] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)

			log.Printf("➖ Removed subscriber from board: %s (remaining: %d)", boardID, len(s.subscribers[boardID]))

			// Clean up empty board entries
			if len(s.subscribers[boardID]) == 0 {
				delete(s.subscribers, boardID)
			}
			break
		}
	}
}

// broadcastToBoard sends an event to all subscribers of a board
func (s *WebSocketService) broadcastToBoard(boardID string, event *domain.WebSocketEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subscribers, exists := s.subscribers[boardID]
	if !exists || len(subscribers) == 0 {
		return
	}

	log.Printf("📡 Broadcasting %s to %d subscribers (board: %s)",
		event.Type, len(subscribers), boardID)

	// Send to all subscribers
	for _, ch := range subscribers {
		select {
		case ch <- event:
			// Successfully sent
		default:
			// Channel is full, skip this subscriber
			log.Printf("⚠️ Subscriber channel full, skipping event")
		}
	}
}

// BroadcastEvent directly broadcasts an event (alternative to Redis pub/sub)
func (s *WebSocketService) BroadcastEvent(event *domain.WebSocketEvent) {
	s.broadcastToBoard(event.BoardID.String(), event)
}

// GetSubscriberCount returns the number of subscribers for a board
func (s *WebSocketService) GetSubscriberCount(boardID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers[boardID])
}

// GetTotalSubscribers returns the total number of subscribers across all boards
func (s *WebSocketService) GetTotalSubscribers() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, subs := range s.subscribers {
		total += len(subs)
	}
	return total
}

// Close gracefully shuts down the WebSocket service
func (s *WebSocketService) Close() {
	log.Println("🔌 Closing WebSocket service...")

	// Cancel context to stop Redis listener
	s.cancel()

	// Close all subscriber channels
	s.mu.Lock()
	defer s.mu.Unlock()

	for boardID, subscribers := range s.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
		delete(s.subscribers, boardID)
	}

	log.Println("✅ WebSocket service closed")
}
