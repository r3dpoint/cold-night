package events

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event in the event store
type Event struct {
	// Core event fields
	EventNumber     int64     `json:"event_number" db:"event_number"`
	EventID         string    `json:"event_id" db:"event_id"`
	EventType       string    `json:"event_type" db:"event_type"`
	AggregateID     string    `json:"aggregate_id" db:"aggregate_id"`
	AggregateType   string    `json:"aggregate_type" db:"aggregate_type"`
	AggregateVersion int      `json:"aggregate_version" db:"aggregate_version"`
	EventVersion    int       `json:"event_version" db:"event_version"`
	EventData       []byte    `json:"event_data" db:"event_data"`
	Metadata        Metadata  `json:"metadata" db:"metadata"`
	OccurredAt      time.Time `json:"occurred_at" db:"occurred_at"`
	
	// Audit fields
	UserID        string    `json:"user_id" db:"user_id"`
	CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	CausationID   *string   `json:"causation_id" db:"causation_id"`
	IPAddress     *string   `json:"ip_address" db:"ip_address"`
	UserAgent     *string   `json:"user_agent" db:"user_agent"`
	SessionID     *string   `json:"session_id" db:"session_id"`
	Checksum      string    `json:"checksum" db:"checksum"`
}

// Metadata contains additional event metadata
type Metadata struct {
	EventID       string    `json:"eventId"`
	EventType     string    `json:"eventType"`
	AggregateID   string    `json:"aggregateId"`
	AggregateType string    `json:"aggregateType"`
	EventVersion  int       `json:"eventVersion"`
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"userId"`
	CorrelationID string    `json:"correlationId"`
	CausationID   string    `json:"causationId"`
	IPAddress     string    `json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	SessionID     string    `json:"sessionId"`
	Checksum      string    `json:"checksum"`
}

// DomainEvent interface that all domain events must implement
type DomainEvent interface {
	GetEventType() string
	GetAggregateID() string
	GetAggregateType() string
	GetEventData() ([]byte, error)
	GetMetadata() Metadata
}

// BaseEvent provides common functionality for all domain events
type BaseEvent struct {
	EventID       string    `json:"eventId"`
	AggregateID   string    `json:"aggregateId"`
	AggregateType string    `json:"aggregateType"`
	Timestamp     time.Time `json:"timestamp"`
	Version       int       `json:"version"`
}

// NewBaseEvent creates a new base event
func NewBaseEvent(aggregateID, aggregateType string) BaseEvent {
	return BaseEvent{
		EventID:       uuid.New().String(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Timestamp:     time.Now(),
		Version:       1,
	}
}

// EventStore interface for storing and retrieving events
type EventStore interface {
	SaveEvent(event *Event) error
	SaveEvents(events []*Event) error
	GetEvents(aggregateID string, fromVersion int) ([]*Event, error)
	GetEventsByType(eventType string, limit int) ([]*Event, error)
	GetAllEvents(fromEventNumber int64, limit int) ([]*Event, error)
	GetSnapshot(aggregateID string) (*Snapshot, error)
	SaveSnapshot(snapshot *Snapshot) error
}

// EventBus interface for publishing and subscribing to events
type EventBus interface {
	Publish(event DomainEvent) error
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
}

// EventHandler function type for handling events
type EventHandler func(event DomainEvent) error

// Snapshot represents an aggregate snapshot for performance optimization
type Snapshot struct {
	AggregateID      string    `json:"aggregate_id" db:"aggregate_id"`
	AggregateType    string    `json:"aggregate_type" db:"aggregate_type"`
	AggregateVersion int       `json:"aggregate_version" db:"aggregate_version"`
	SnapshotData     []byte    `json:"snapshot_data" db:"snapshot_data"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ProjectionCheckpoint tracks the progress of event projections
type ProjectionCheckpoint struct {
	ProjectionName            string    `json:"projection_name" db:"projection_name"`
	LastProcessedEventNumber  int64     `json:"last_processed_event_number" db:"last_processed_event_number"`
	LastProcessedAt           time.Time `json:"last_processed_at" db:"last_processed_at"`
	Status                    string    `json:"status" db:"status"` // active, rebuilding, failed
}

// Aggregate interface that all aggregates must implement
type Aggregate interface {
	GetID() string
	GetType() string
	GetVersion() int
	GetUncommittedEvents() []DomainEvent
	ApplyEvent(event DomainEvent) error
	LoadFromHistory(events []DomainEvent) error
	MarkEventsAsCommitted()
}

// AggregateRoot provides common functionality for all aggregates
type AggregateRoot struct {
	ID                string
	Type              string
	Version           int
	UncommittedEvents []DomainEvent
}

// NewAggregateRoot creates a new aggregate root
func NewAggregateRoot(id, aggregateType string) AggregateRoot {
	return AggregateRoot{
		ID:                id,
		Type:              aggregateType,
		Version:           0,
		UncommittedEvents: make([]DomainEvent, 0),
	}
}

// GetID returns the aggregate ID
func (a *AggregateRoot) GetID() string {
	return a.ID
}

// GetType returns the aggregate type
func (a *AggregateRoot) GetType() string {
	return a.Type
}

// GetVersion returns the current version
func (a *AggregateRoot) GetVersion() int {
	return a.Version
}

// GetUncommittedEvents returns uncommitted events
func (a *AggregateRoot) GetUncommittedEvents() []DomainEvent {
	return a.UncommittedEvents
}

// AddEvent adds a new event to the uncommitted events
func (a *AggregateRoot) AddEvent(event DomainEvent) {
	a.UncommittedEvents = append(a.UncommittedEvents, event)
}

// MarkEventsAsCommitted clears the uncommitted events
func (a *AggregateRoot) MarkEventsAsCommitted() {
	a.UncommittedEvents = make([]DomainEvent, 0)
}

// IncrementVersion increments the aggregate version
func (a *AggregateRoot) IncrementVersion() {
	a.Version++
}