package securities

import (
	"database/sql"
	"fmt"

	"securities-marketplace/domains/shared/events"
)

// SecurityRepository defines the interface for security persistence
type SecurityRepository interface {
	FindByID(securityID string) (*SecurityAggregate, error)
	FindBySymbol(symbol string) (*SecurityAggregate, error)
	FindByIssuer(issuerID string) ([]*SecurityAggregate, error)
	FindByType(securityType SecurityType) ([]*SecurityAggregate, error)
	FindByStatus(status SecurityStatus) ([]*SecurityAggregate, error)
	FindByOwner(ownerID string) ([]*SecurityAggregate, error)
	Save(security *SecurityAggregate) error
}

// EventSourcedSecurityRepository implements SecurityRepository using event sourcing
type EventSourcedSecurityRepository struct {
	eventStore events.EventStore
}

// NewEventSourcedSecurityRepository creates a new event-sourced security repository
func NewEventSourcedSecurityRepository(eventStore events.EventStore) *EventSourcedSecurityRepository {
	return &EventSourcedSecurityRepository{
		eventStore: eventStore,
	}
}

// FindByID finds a security by ID by replaying events
func (r *EventSourcedSecurityRepository) FindByID(securityID string) (*SecurityAggregate, error) {
	// Check for snapshot first
	snapshot, err := r.eventStore.GetSnapshot(securityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	security := NewSecurityAggregate(securityID)
	fromVersion := 0

	// Load from snapshot if available
	if snapshot != nil {
		err = r.loadFromSnapshot(security, snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to load from snapshot: %w", err)
		}
		fromVersion = snapshot.AggregateVersion + 1
	}

	// Load events after snapshot
	eventRecords, err := r.eventStore.GetEvents(securityID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(eventRecords) == 0 && snapshot == nil {
		return nil, NewNotFoundError("security", securityID)
	}

	// Convert event records to domain events and apply them
	domainEvents, err := r.convertEventRecordsToDomainEvents(eventRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to convert events: %w", err)
	}

	err = security.LoadFromHistory(domainEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to load from history: %w", err)
	}

	return security, nil
}

// FindBySymbol finds a security by symbol
func (r *EventSourcedSecurityRepository) FindBySymbol(symbol string) (*SecurityAggregate, error) {
	// Get all SecurityListed events and find matching symbol
	securityListedEvents, err := r.eventStore.GetEventsByType("SecurityListed", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get security listing events: %w", err)
	}

	for _, eventRecord := range securityListedEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if secListEvent, ok := domainEvent.(*SecurityListed); ok {
			if secListEvent.Symbol == symbol {
				// Found the security, now load the full aggregate
				return r.FindByID(eventRecord.AggregateID)
			}
		}
	}

	return nil, NewNotFoundError("security", symbol)
}

// FindByIssuer finds all securities for a given issuer
func (r *EventSourcedSecurityRepository) FindByIssuer(issuerID string) ([]*SecurityAggregate, error) {
	// Get all SecurityListed events and find matching issuer
	securityListedEvents, err := r.eventStore.GetEventsByType("SecurityListed", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get security listing events: %w", err)
	}

	var securities []*SecurityAggregate
	for _, eventRecord := range securityListedEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if secListEvent, ok := domainEvent.(*SecurityListed); ok {
			if secListEvent.IssuerID == issuerID {
				// Found a security, load the full aggregate
				security, err := r.FindByID(eventRecord.AggregateID)
				if err != nil {
					continue // Skip securities that can't be loaded
				}
				securities = append(securities, security)
			}
		}
	}

	return securities, nil
}

// FindByType finds all securities of a given type
func (r *EventSourcedSecurityRepository) FindByType(securityType SecurityType) ([]*SecurityAggregate, error) {
	// Get all SecurityListed events and find matching type
	securityListedEvents, err := r.eventStore.GetEventsByType("SecurityListed", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get security listing events: %w", err)
	}

	var securities []*SecurityAggregate
	for _, eventRecord := range securityListedEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if secListEvent, ok := domainEvent.(*SecurityListed); ok {
			if SecurityType(secListEvent.SecurityType) == securityType {
				// Found a security, load the full aggregate
				security, err := r.FindByID(eventRecord.AggregateID)
				if err != nil {
					continue // Skip securities that can't be loaded
				}
				securities = append(securities, security)
			}
		}
	}

	return securities, nil
}

// FindByStatus finds all securities with a given status
func (r *EventSourcedSecurityRepository) FindByStatus(status SecurityStatus) ([]*SecurityAggregate, error) {
	// This is inefficient for event sourcing - would need projection in real system
	// For now, load all securities and filter by status
	securityListedEvents, err := r.eventStore.GetEventsByType("SecurityListed", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get security listing events: %w", err)
	}

	var securities []*SecurityAggregate
	for _, eventRecord := range securityListedEvents {
		// Load the full aggregate to check current status
		security, err := r.FindByID(eventRecord.AggregateID)
		if err != nil {
			continue // Skip securities that can't be loaded
		}

		if security.Status == status {
			securities = append(securities, security)
		}
	}

	return securities, nil
}

// FindByOwner finds all securities owned by a user
func (r *EventSourcedSecurityRepository) FindByOwner(ownerID string) ([]*SecurityAggregate, error) {
	// This is very inefficient for event sourcing - would need projection in real system
	// For now, load all securities and check ownership
	securityListedEvents, err := r.eventStore.GetEventsByType("SecurityListed", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get security listing events: %w", err)
	}

	var securities []*SecurityAggregate
	for _, eventRecord := range securityListedEvents {
		// Load the full aggregate to check ownership
		security, err := r.FindByID(eventRecord.AggregateID)
		if err != nil {
			continue // Skip securities that can't be loaded
		}

		if security.GetSharesOwned(ownerID) > 0 {
			securities = append(securities, security)
		}
	}

	return securities, nil
}

// Save saves a security aggregate (used for snapshots)
func (r *EventSourcedSecurityRepository) Save(security *SecurityAggregate) error {
	return r.createSnapshot(security)
}

// createSnapshot creates a snapshot of the security aggregate
func (r *EventSourcedSecurityRepository) createSnapshot(security *SecurityAggregate) error {
	if security.GetVersion() < 10 {
		// Only create snapshots for aggregates with significant event history
		return nil
	}

	snapshotData, err := r.serializeAggregate(security)
	if err != nil {
		return fmt.Errorf("failed to serialize aggregate: %w", err)
	}

	snapshot := &events.Snapshot{
		AggregateID:      security.GetID(),
		AggregateType:    security.GetType(),
		AggregateVersion: security.GetVersion(),
		SnapshotData:     snapshotData,
	}

	return r.eventStore.SaveSnapshot(snapshot)
}

// loadFromSnapshot loads security state from a snapshot
func (r *EventSourcedSecurityRepository) loadFromSnapshot(security *SecurityAggregate, snapshot *events.Snapshot) error {
	return r.deserializeAggregate(security, snapshot.SnapshotData)
}

// serializeAggregate serializes the security aggregate to bytes
func (r *EventSourcedSecurityRepository) serializeAggregate(security *SecurityAggregate) ([]byte, error) {
	// Implementation would serialize the aggregate state
	// For now, return empty bytes
	return []byte{}, nil
}

// deserializeAggregate deserializes bytes to security aggregate
func (r *EventSourcedSecurityRepository) deserializeAggregate(security *SecurityAggregate, data []byte) error {
	// Implementation would deserialize the aggregate state
	// For now, do nothing
	return nil
}

// convertEventRecordsToDomainEvents converts event store records to domain events
func (r *EventSourcedSecurityRepository) convertEventRecordsToDomainEvents(eventRecords []*events.Event) ([]events.DomainEvent, error) {
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
func (r *EventSourcedSecurityRepository) convertEventRecordToDomainEvent(eventRecord *events.Event) (events.DomainEvent, error) {
	switch eventRecord.EventType {
	case "SecurityListed":
		return r.deserializeSecurityListed(eventRecord.EventData)
	case "SecurityDocumentAdded":
		return r.deserializeSecurityDocumentAdded(eventRecord.EventData)
	case "SecurityUpdated":
		return r.deserializeSecurityUpdated(eventRecord.EventData)
	case "SecuritySuspended":
		return r.deserializeSecuritySuspended(eventRecord.EventData)
	case "SecurityReinstated":
		return r.deserializeSecurityReinstated(eventRecord.EventData)
	case "SecurityDelisted":
		return r.deserializeSecurityDelisted(eventRecord.EventData)
	case "SecurityOwnershipChanged":
		return r.deserializeSecurityOwnershipChanged(eventRecord.EventData)
	case "SecurityDividendDeclared":
		return r.deserializeSecurityDividendDeclared(eventRecord.EventData)
	case "SecuritySplitAnnounced":
		return r.deserializeSecuritySplitAnnounced(eventRecord.EventData)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventRecord.EventType)
	}
}

// Event deserialization methods (simplified implementations)
func (r *EventSourcedSecurityRepository) deserializeSecurityListed(data []byte) (*SecurityListed, error) {
	// Implementation would deserialize JSON to SecurityListed
	// For now, return a basic event
	return &SecurityListed{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityDocumentAdded(data []byte) (*SecurityDocumentAdded, error) {
	return &SecurityDocumentAdded{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityUpdated(data []byte) (*SecurityUpdated, error) {
	return &SecurityUpdated{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecuritySuspended(data []byte) (*SecuritySuspended, error) {
	return &SecuritySuspended{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityReinstated(data []byte) (*SecurityReinstated, error) {
	return &SecurityReinstated{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityDelisted(data []byte) (*SecurityDelisted, error) {
	return &SecurityDelisted{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityOwnershipChanged(data []byte) (*SecurityOwnershipChanged, error) {
	return &SecurityOwnershipChanged{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecurityDividendDeclared(data []byte) (*SecurityDividendDeclared, error) {
	return &SecurityDividendDeclared{}, nil
}

func (r *EventSourcedSecurityRepository) deserializeSecuritySplitAnnounced(data []byte) (*SecuritySplitAnnounced, error) {
	return &SecuritySplitAnnounced{}, nil
}

// ProjectionSecurityRepository implements SecurityRepository using read model projections
type ProjectionSecurityRepository struct {
	db *sql.DB
}

// NewProjectionSecurityRepository creates a new projection-based security repository
func NewProjectionSecurityRepository(db *sql.DB) *ProjectionSecurityRepository {
	return &ProjectionSecurityRepository{db: db}
}

// FindByID finds a security by ID from the projection
func (r *ProjectionSecurityRepository) FindByID(securityID string) (*SecurityAggregate, error) {
	// This would query the securities projection table
	// For now, return not implemented
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindBySymbol finds a security by symbol from the projection
func (r *ProjectionSecurityRepository) FindBySymbol(symbol string) (*SecurityAggregate, error) {
	// This would query the securities projection table
	// For now, return not implemented
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByIssuer finds securities by issuer from the projection
func (r *ProjectionSecurityRepository) FindByIssuer(issuerID string) ([]*SecurityAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByType finds securities by type from the projection
func (r *ProjectionSecurityRepository) FindByType(securityType SecurityType) ([]*SecurityAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByStatus finds securities by status from the projection
func (r *ProjectionSecurityRepository) FindByStatus(status SecurityStatus) ([]*SecurityAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByOwner finds securities by owner from the projection
func (r *ProjectionSecurityRepository) FindByOwner(ownerID string) ([]*SecurityAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// Save saves a security aggregate (not applicable for read-only projections)
func (r *ProjectionSecurityRepository) Save(security *SecurityAggregate) error {
	return fmt.Errorf("save operation not supported for projection repository")
}