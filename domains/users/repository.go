package users

import (
	"database/sql"
	"fmt"

	"securities-marketplace/domains/shared/events"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	FindByID(userID string) (*UserAggregate, error)
	FindByEmail(email string) (*UserAggregate, error)
	Save(user *UserAggregate) error
}

// EventSourcedUserRepository implements UserRepository using event sourcing
type EventSourcedUserRepository struct {
	eventStore events.EventStore
}

// NewEventSourcedUserRepository creates a new event-sourced user repository
func NewEventSourcedUserRepository(eventStore events.EventStore) *EventSourcedUserRepository {
	return &EventSourcedUserRepository{
		eventStore: eventStore,
	}
}

// FindByID finds a user by ID by replaying events
func (r *EventSourcedUserRepository) FindByID(userID string) (*UserAggregate, error) {
	// Check for snapshot first
	snapshot, err := r.eventStore.GetSnapshot(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	user := NewUserAggregate(userID)
	fromVersion := 0

	// Load from snapshot if available
	if snapshot != nil {
		err = r.loadFromSnapshot(user, snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to load from snapshot: %w", err)
		}
		fromVersion = snapshot.AggregateVersion + 1
	}

	// Load events after snapshot
	eventRecords, err := r.eventStore.GetEvents(userID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(eventRecords) == 0 && snapshot == nil {
		return nil, NewNotFoundError("user", userID)
	}

	// Convert event records to domain events and apply them
	domainEvents, err := r.convertEventRecordsToDomainEvents(eventRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to convert events: %w", err)
	}

	err = user.LoadFromHistory(domainEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to load from history: %w", err)
	}

	return user, nil
}

// FindByEmail finds a user by email using projection
func (r *EventSourcedUserRepository) FindByEmail(email string) (*UserAggregate, error) {
	// This implementation assumes we have a projection or index that maps email to user ID
	// For now, we'll implement a simple approach that scans through user events
	// In a production system, you'd want a proper email-to-ID index
	
	// Get all UserRegistered events and find matching email
	userRegisteredEvents, err := r.eventStore.GetEventsByType("UserRegistered", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get user registration events: %w", err)
	}

	for _, eventRecord := range userRegisteredEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if userRegEvent, ok := domainEvent.(*UserRegistered); ok {
			if userRegEvent.Email == email {
				// Found the user, now load the full aggregate
				return r.FindByID(eventRecord.AggregateID)
			}
		}
	}

	return nil, NewNotFoundError("user", email)
}

// Save saves a user aggregate (this is not typically used in event sourcing)
func (r *EventSourcedUserRepository) Save(user *UserAggregate) error {
	// In event sourcing, saving is typically handled by the service layer
	// through event storage. This method is kept for interface compatibility
	// but could be used for creating snapshots
	return r.createSnapshot(user)
}

// createSnapshot creates a snapshot of the user aggregate
func (r *EventSourcedUserRepository) createSnapshot(user *UserAggregate) error {
	if user.GetVersion() < 10 {
		// Only create snapshots for aggregates with significant event history
		return nil
	}

	snapshotData, err := r.serializeAggregate(user)
	if err != nil {
		return fmt.Errorf("failed to serialize aggregate: %w", err)
	}

	snapshot := &events.Snapshot{
		AggregateID:      user.GetID(),
		AggregateType:    user.GetType(),
		AggregateVersion: user.GetVersion(),
		SnapshotData:     snapshotData,
	}

	return r.eventStore.SaveSnapshot(snapshot)
}

// loadFromSnapshot loads user state from a snapshot
func (r *EventSourcedUserRepository) loadFromSnapshot(user *UserAggregate, snapshot *events.Snapshot) error {
	return r.deserializeAggregate(user, snapshot.SnapshotData)
}

// serializeAggregate serializes the user aggregate to bytes
func (r *EventSourcedUserRepository) serializeAggregate(user *UserAggregate) ([]byte, error) {
	// Implementation would serialize the aggregate state
	// For now, return empty bytes
	return []byte{}, nil
}

// deserializeAggregate deserializes bytes to user aggregate
func (r *EventSourcedUserRepository) deserializeAggregate(user *UserAggregate, data []byte) error {
	// Implementation would deserialize the aggregate state
	// For now, do nothing
	return nil
}

// convertEventRecordsToDomainEvents converts event store records to domain events
func (r *EventSourcedUserRepository) convertEventRecordsToDomainEvents(eventRecords []*events.Event) ([]events.DomainEvent, error) {
	var domainEvents []events.DomainEvent

	for _, eventRecord := range eventRecords {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event %s: %w", eventRecord.EventType, err)
		}
		domainEvents = append(domainEvents, domainEvent)
	}

	return domainEvents, nil
}

// convertEventRecordToDomainEvent converts a single event record to domain event
func (r *EventSourcedUserRepository) convertEventRecordToDomainEvent(eventRecord *events.Event) (events.DomainEvent, error) {
	switch eventRecord.EventType {
	case "UserRegistered":
		return r.deserializeUserRegistered(eventRecord.EventData)
	case "AccreditationSubmitted":
		return r.deserializeAccreditationSubmitted(eventRecord.EventData)
	case "AccreditationVerified":
		return r.deserializeAccreditationVerified(eventRecord.EventData)
	case "AccreditationRevoked":
		return r.deserializeAccreditationRevoked(eventRecord.EventData)
	case "ComplianceCheckPerformed":
		return r.deserializeComplianceCheckPerformed(eventRecord.EventData)
	case "UserSuspended":
		return r.deserializeUserSuspended(eventRecord.EventData)
	case "UserReinstated":
		return r.deserializeUserReinstated(eventRecord.EventData)
	case "UserProfileUpdated":
		return r.deserializeUserProfileUpdated(eventRecord.EventData)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventRecord.EventType)
	}
}

// Event deserialization methods (simplified implementations)
func (r *EventSourcedUserRepository) deserializeUserRegistered(data []byte) (*UserRegistered, error) {
	// Implementation would deserialize JSON to UserRegistered
	// For now, return a basic event
	return &UserRegistered{}, nil
}

func (r *EventSourcedUserRepository) deserializeAccreditationSubmitted(data []byte) (*AccreditationSubmitted, error) {
	return &AccreditationSubmitted{}, nil
}

func (r *EventSourcedUserRepository) deserializeAccreditationVerified(data []byte) (*AccreditationVerified, error) {
	return &AccreditationVerified{}, nil
}

func (r *EventSourcedUserRepository) deserializeAccreditationRevoked(data []byte) (*AccreditationRevoked, error) {
	return &AccreditationRevoked{}, nil
}

func (r *EventSourcedUserRepository) deserializeComplianceCheckPerformed(data []byte) (*ComplianceCheckPerformed, error) {
	return &ComplianceCheckPerformed{}, nil
}

func (r *EventSourcedUserRepository) deserializeUserSuspended(data []byte) (*UserSuspended, error) {
	return &UserSuspended{}, nil
}

func (r *EventSourcedUserRepository) deserializeUserReinstated(data []byte) (*UserReinstated, error) {
	return &UserReinstated{}, nil
}

func (r *EventSourcedUserRepository) deserializeUserProfileUpdated(data []byte) (*UserProfileUpdated, error) {
	return &UserProfileUpdated{}, nil
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// ProjectionUserRepository implements UserRepository using read model projections
type ProjectionUserRepository struct {
	db *sql.DB
}

// NewProjectionUserRepository creates a new projection-based user repository
func NewProjectionUserRepository(db *sql.DB) *ProjectionUserRepository {
	return &ProjectionUserRepository{db: db}
}

// FindByID finds a user by ID from the projection
func (r *ProjectionUserRepository) FindByID(userID string) (*UserAggregate, error) {
	// This would query the user_profiles projection table
	// For now, return not implemented
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByEmail finds a user by email from the projection
func (r *ProjectionUserRepository) FindByEmail(email string) (*UserAggregate, error) {
	// This would query the user_profiles projection table
	// For now, return not implemented
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// Save saves a user aggregate (not applicable for read-only projections)
func (r *ProjectionUserRepository) Save(user *UserAggregate) error {
	return fmt.Errorf("save operation not supported for projection repository")
}