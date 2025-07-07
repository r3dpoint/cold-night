package testutil

import (
	"time"

	"github.com/google/uuid"

	"securities-marketplace/domains/shared/events"
)

// TestEventStore provides a simple in-memory event store for testing
type TestEventStore struct {
	events []events.Event
}

func NewTestEventStore() *TestEventStore {
	return &TestEventStore{
		events: make([]events.Event, 0),
	}
}

func (s *TestEventStore) SaveEvent(evt *events.Event) error {
	s.events = append(s.events, *evt)
	return nil
}

func (s *TestEventStore) SaveEvents(evts []*events.Event) error {
	for _, evt := range evts {
		s.events = append(s.events, *evt)
	}
	return nil
}

func (s *TestEventStore) GetEvents(aggregateID string, fromVersion int) ([]*events.Event, error) {
	var result []*events.Event
	for _, evt := range s.events {
		if evt.AggregateID == aggregateID && evt.AggregateVersion >= fromVersion {
			result = append(result, &evt)
		}
	}
	return result, nil
}

func (s *TestEventStore) GetEventsByType(eventType string, limit int) ([]*events.Event, error) {
	var result []*events.Event
	count := 0
	for _, evt := range s.events {
		if evt.EventType == eventType {
			result = append(result, &evt)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *TestEventStore) GetEventsFromVersion(aggregateID string, fromVersion int) ([]*events.Event, error) {
	var result []*events.Event
	for _, evt := range s.events {
		if evt.AggregateID == aggregateID && evt.AggregateVersion >= fromVersion {
			result = append(result, &evt)
		}
	}
	return result, nil
}

func (s *TestEventStore) CreateEventFromDomain(domainEvent events.DomainEvent, userID, correlationID string, causationID *string) (*events.Event, error) {
	eventData, err := domainEvent.GetEventData()
	if err != nil {
		return nil, err
	}

	event := &events.Event{
		EventID:          uuid.New().String(),
		EventType:        domainEvent.GetEventType(),
		AggregateID:      domainEvent.GetAggregateID(),
		AggregateType:    domainEvent.GetAggregateType(),
		EventVersion:     1,
		EventData:        eventData,
		Metadata:         domainEvent.GetMetadata(),
		OccurredAt:       time.Now(),
		UserID:           userID,
		CorrelationID:    correlationID,
		CausationID:      causationID,
		IPAddress:        stringPtr("127.0.0.1"),
		UserAgent:        stringPtr("test-agent"),
		SessionID:        stringPtr("test-session"),
		Checksum:         "test-checksum",
	}
	return event, nil
}

func (s *TestEventStore) GetAllEvents(fromEventNumber int64, limit int) ([]*events.Event, error) {
	var result []*events.Event
	count := 0
	for i, evt := range s.events {
		if int64(i+1) >= fromEventNumber {
			result = append(result, &evt)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *TestEventStore) GetSnapshot(aggregateID string) (*events.Snapshot, error) {
	// No snapshots in test store
	return nil, nil
}

func (s *TestEventStore) SaveSnapshot(snapshot *events.Snapshot) error {
	// No-op for testing
	return nil
}

// TestEventBus provides a simple in-memory event bus for testing
type TestEventBus struct {
	publishedEvents []events.DomainEvent
}

func NewTestEventBus() *TestEventBus {
	return &TestEventBus{
		publishedEvents: make([]events.DomainEvent, 0),
	}
}

func (b *TestEventBus) Publish(event events.DomainEvent) error {
	b.publishedEvents = append(b.publishedEvents, event)
	return nil
}

func (b *TestEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	// No-op for testing
	return nil
}

func (b *TestEventBus) GetPublishedEvents() []events.DomainEvent {
	return b.publishedEvents
}

func (b *TestEventBus) GetEventsByType(eventType string) []events.DomainEvent {
	var result []events.DomainEvent
	for _, evt := range b.publishedEvents {
		if evt.GetEventType() == eventType {
			result = append(result, evt)
		}
	}
	return result
}

func (b *TestEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	// No-op for testing
	return nil
}

// Test Data Builders

// Generic test builders will be defined in domain-specific test files
// to avoid import cycles

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Test Constants
const (
	TestUserID      = "test-user-123"
	TestSecurityID  = "test-security-456"
	TestListingID   = "test-listing-789"
	TestBidID       = "test-bid-101"
	TestTradeID     = "test-trade-112"
	TestAdminUserID = "admin-user-999"
)

// Test Date/Time Helpers
var (
	TestTime     = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	TestTimeNext = time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	FutureTime   = time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	PastTime     = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
)

// AssertEventType checks if an event of the specified type was published
func AssertEventType(events []events.DomainEvent, eventType string) bool {
	for _, evt := range events {
		if evt.GetEventType() == eventType {
			return true
		}
	}
	return false
}

// AssertEventCount checks if the expected number of events were published
func AssertEventCount(events []events.DomainEvent, expected int) bool {
	return len(events) == expected
}

// AssertNoEvents checks if no events were published
func AssertNoEvents(events []events.DomainEvent) bool {
	return len(events) == 0
}