package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisEventBus implements EventBus using Redis pub/sub
type RedisEventBus struct {
	client     *redis.Client
	handlers   map[string][]EventHandler
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	subscribed map[string]*redis.PubSub
}

// NewEventBus creates a new Redis-based event bus
func NewEventBus(client *redis.Client) *RedisEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RedisEventBus{
		client:     client,
		handlers:   make(map[string][]EventHandler),
		ctx:        ctx,
		cancel:     cancel,
		subscribed: make(map[string]*redis.PubSub),
	}
}

// Publish publishes a domain event to the event bus
func (eb *RedisEventBus) Publish(event DomainEvent) error {
	eventType := event.GetEventType()
	
	// Serialize event data
	eventData, err := event.GetEventData()
	if err != nil {
		return fmt.Errorf("failed to get event data: %w", err)
	}

	// Create message payload
	message := EventMessage{
		EventType:     eventType,
		AggregateID:   event.GetAggregateID(),
		AggregateType: event.GetAggregateType(),
		EventData:     eventData,
		Metadata:      event.GetMetadata(),
		PublishedAt:   time.Now(),
	}

	// Serialize message
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize event message: %w", err)
	}

	// Publish to Redis
	channelName := fmt.Sprintf("events:%s", eventType)
	err = eb.client.Publish(eb.ctx, channelName, messageJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event to Redis: %w", err)
	}

	// Also publish to a general events channel
	err = eb.client.Publish(eb.ctx, "events:all", messageJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event to all channel: %w", err)
	}

	return nil
}

// Subscribe subscribes to events of a specific type
func (eb *RedisEventBus) Subscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Add handler to the list
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// If this is the first handler for this event type, start subscription
	if len(eb.handlers[eventType]) == 1 {
		channelName := fmt.Sprintf("events:%s", eventType)
		pubsub := eb.client.Subscribe(eb.ctx, channelName)
		eb.subscribed[eventType] = pubsub

		eb.wg.Add(1)
		go eb.handleSubscription(eventType, pubsub)
	}

	return nil
}

// Unsubscribe removes a handler from event subscriptions
func (eb *RedisEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	if handlers == nil {
		return nil // No handlers for this event type
	}

	// Remove the handler (this is a simple implementation - in practice you'd want better handler identification)
	newHandlers := make([]EventHandler, 0, len(handlers))
	for _, h := range handlers {
		// Simple pointer comparison - in practice you'd want better handler identification
		if fmt.Sprintf("%p", h) != fmt.Sprintf("%p", handler) {
			newHandlers = append(newHandlers, h)
		}
	}

	eb.handlers[eventType] = newHandlers

	// If no more handlers, close the subscription
	if len(newHandlers) == 0 {
		if pubsub, exists := eb.subscribed[eventType]; exists {
			pubsub.Close()
			delete(eb.subscribed, eventType)
		}
		delete(eb.handlers, eventType)
	}

	return nil
}

// SubscribeToAll subscribes to all events
func (eb *RedisEventBus) SubscribeToAll(handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Add handler to the "all" list
	eb.handlers["all"] = append(eb.handlers["all"], handler)

	// If this is the first handler for all events, start subscription
	if len(eb.handlers["all"]) == 1 {
		pubsub := eb.client.Subscribe(eb.ctx, "events:all")
		eb.subscribed["all"] = pubsub

		eb.wg.Add(1)
		go eb.handleSubscription("all", pubsub)
	}

	return nil
}

// Close closes the event bus and all subscriptions
func (eb *RedisEventBus) Close() error {
	eb.cancel()

	eb.mu.Lock()
	for _, pubsub := range eb.subscribed {
		pubsub.Close()
	}
	eb.mu.Unlock()

	eb.wg.Wait()
	return nil
}

// handleSubscription handles incoming messages for a subscription
func (eb *RedisEventBus) handleSubscription(eventType string, pubsub *redis.PubSub) {
	defer eb.wg.Done()

	ch := pubsub.Channel()
	for {
		select {
		case <-eb.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			// Deserialize the message
			var eventMessage EventMessage
			if err := json.Unmarshal([]byte(msg.Payload), &eventMessage); err != nil {
				log.Printf("Failed to deserialize event message: %v", err)
				continue
			}

			// Create a domain event from the message
			domainEvent := &GenericDomainEvent{
				EventType:     eventMessage.EventType,
				AggregateID:   eventMessage.AggregateID,
				AggregateType: eventMessage.AggregateType,
				EventData:     eventMessage.EventData,
				Metadata:      eventMessage.Metadata,
			}

			// Execute handlers
			eb.mu.RLock()
			handlers := eb.handlers[eventType]
			eb.mu.RUnlock()

			for _, handler := range handlers {
				go func(h EventHandler) {
					if err := h(domainEvent); err != nil {
						log.Printf("Event handler error for %s: %v", eventType, err)
					}
				}(handler)
			}
		}
	}
}

// EventMessage represents a serialized event message
type EventMessage struct {
	EventType     string    `json:"event_type"`
	AggregateID   string    `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	EventData     []byte    `json:"event_data"`
	Metadata      Metadata  `json:"metadata"`
	PublishedAt   time.Time `json:"published_at"`
}

// GenericDomainEvent implements DomainEvent for deserialized events
type GenericDomainEvent struct {
	EventType     string
	AggregateID   string
	AggregateType string
	EventData     []byte
	Metadata      Metadata
}

// GetEventType returns the event type
func (e *GenericDomainEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateID returns the aggregate ID
func (e *GenericDomainEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetAggregateType returns the aggregate type
func (e *GenericDomainEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetEventData returns the event data
func (e *GenericDomainEvent) GetEventData() ([]byte, error) {
	return e.EventData, nil
}

// GetMetadata returns the event metadata
func (e *GenericDomainEvent) GetMetadata() Metadata {
	return e.Metadata
}

// InMemoryEventBus provides a simple in-memory event bus for testing
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event to all registered handlers
func (eb *InMemoryEventBus) Publish(event DomainEvent) error {
	eventType := event.GetEventType()

	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	allHandlers := eb.handlers["all"]
	eb.mu.RUnlock()

	// Execute specific handlers
	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(event); err != nil {
				log.Printf("Event handler error for %s: %v", eventType, err)
			}
		}(handler)
	}

	// Execute "all" handlers
	for _, handler := range allHandlers {
		go func(h EventHandler) {
			if err := h(event); err != nil {
				log.Printf("Event handler error for all events: %v", err)
			}
		}(handler)
	}

	return nil
}

// Subscribe subscribes to events of a specific type
func (eb *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes a handler from event subscriptions
func (eb *InMemoryEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	if handlers == nil {
		return nil
	}

	// Remove the handler
	newHandlers := make([]EventHandler, 0, len(handlers))
	for _, h := range handlers {
		if fmt.Sprintf("%p", h) != fmt.Sprintf("%p", handler) {
			newHandlers = append(newHandlers, h)
		}
	}

	eb.handlers[eventType] = newHandlers
	return nil
}