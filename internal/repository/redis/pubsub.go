// internal/repository/redis/pubsub.go
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	// Channel names for different event types
	TaskEventsChannel = "task:events"
)

// PubSubClient handles Redis pub/sub operations
type PubSubClient struct {
	client *redis.Client
}

// NewPubSubClient creates a new Redis pub/sub client
func NewPubSubClient(client *redis.Client) *PubSubClient {
	return &PubSubClient{client: client}
}

// PublishTaskEvent publishes a task event to Redis
func (p *PubSubClient) PublishTaskEvent(ctx context.Context, event *domain.WebSocketEvent) error {
	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis channel
	if err := p.client.Publish(ctx, TaskEventsChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("📤 Published event: %s (board: %s)", event.Type, event.BoardID)
	return nil
}

// SubscribeToTaskEvents subscribes to task events from Redis
func (p *PubSubClient) SubscribeToTaskEvents(ctx context.Context) (<-chan *domain.WebSocketEvent, error) {
	// Subscribe to the task events channel
	pubsub := p.client.Subscribe(ctx, TaskEventsChannel)

	// Wait for subscription confirmation
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Printf("📥 Subscribed to channel: %s", TaskEventsChannel)

	// Create channel for events
	eventChan := make(chan *domain.WebSocketEvent, 100)

	// Start goroutine to receive messages
	go func() {
		defer close(eventChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				log.Println("🔌 Unsubscribing from task events")
				return
			case msg, ok := <-ch:
				if !ok {
					log.Println("⚠️ Redis channel closed")
					return
				}

				// Unmarshal the event
				var event domain.WebSocketEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Printf("❌ Failed to unmarshal event: %v", err)
					continue
				}

				// Send to event channel
				select {
				case eventChan <- &event:
					log.Printf("📨 Received event: %s (board: %s)", event.Type, event.BoardID)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventChan, nil
}

// Ping checks if Redis connection is alive
func (p *PubSubClient) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
