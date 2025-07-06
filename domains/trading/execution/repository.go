package execution

import (
	"database/sql"
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// TradeRepository defines the interface for trade persistence
type TradeRepository interface {
	FindByID(tradeID string) (*TradeAggregate, error)
	FindByUser(userID string) ([]*TradeAggregate, error)
	FindBySecurity(securityID string) ([]*TradeAggregate, error)
	FindByStatus(status TradeStatus) ([]*TradeAggregate, error)
	FindPendingSettlements() ([]*TradeAggregate, error)
	FindBySecurityAndPeriod(securityID string, from, to time.Time) ([]*TradeAggregate, error)
	Save(trade *TradeAggregate) error
}

// EventSourcedTradeRepository implements TradeRepository using event sourcing
type EventSourcedTradeRepository struct {
	eventStore events.EventStore
}

// NewEventSourcedTradeRepository creates a new event-sourced trade repository
func NewEventSourcedTradeRepository(eventStore events.EventStore) *EventSourcedTradeRepository {
	return &EventSourcedTradeRepository{
		eventStore: eventStore,
	}
}

// FindByID finds a trade by ID by replaying events
func (r *EventSourcedTradeRepository) FindByID(tradeID string) (*TradeAggregate, error) {
	// Check for snapshot first
	snapshot, err := r.eventStore.GetSnapshot(tradeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	trade := NewTradeAggregate(tradeID)
	fromVersion := 0

	// Load from snapshot if available
	if snapshot != nil {
		err = r.loadFromSnapshot(trade, snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to load from snapshot: %w", err)
		}
		fromVersion = snapshot.AggregateVersion + 1
	}

	// Load events after snapshot
	eventRecords, err := r.eventStore.GetEvents(tradeID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(eventRecords) == 0 && snapshot == nil {
		return nil, NewNotFoundError("trade", tradeID)
	}

	// Convert event records to domain events and apply them
	domainEvents, err := r.convertEventRecordsToDomainEvents(eventRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to convert events: %w", err)
	}

	err = trade.LoadFromHistory(domainEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to load from history: %w", err)
	}

	return trade, nil
}

// FindByUser finds trades where the user is buyer or seller
func (r *EventSourcedTradeRepository) FindByUser(userID string) ([]*TradeAggregate, error) {
	// Get all TradeMatched events and filter by user
	tradeMatchedEvents, err := r.eventStore.GetEventsByType("TradeMatched", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade matched events: %w", err)
	}

	var trades []*TradeAggregate
	for _, eventRecord := range tradeMatchedEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if tradeEvent, ok := domainEvent.(*TradeMatched); ok {
			if tradeEvent.BuyerID == userID || tradeEvent.SellerID == userID {
				// Found a trade involving this user, load the full aggregate
				trade, err := r.FindByID(eventRecord.AggregateID)
				if err != nil {
					continue // Skip trades that can't be loaded
				}
				trades = append(trades, trade)
			}
		}
	}

	return trades, nil
}

// FindBySecurity finds all trades for a security
func (r *EventSourcedTradeRepository) FindBySecurity(securityID string) ([]*TradeAggregate, error) {
	// Get all TradeMatched events and filter by security
	tradeMatchedEvents, err := r.eventStore.GetEventsByType("TradeMatched", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade matched events: %w", err)
	}

	var trades []*TradeAggregate
	for _, eventRecord := range tradeMatchedEvents {
		domainEvent, err := r.convertEventRecordToDomainEvent(eventRecord)
		if err != nil {
			continue // Skip invalid events
		}

		if tradeEvent, ok := domainEvent.(*TradeMatched); ok {
			if tradeEvent.SecurityID == securityID {
				// Found a trade for this security, load the full aggregate
				trade, err := r.FindByID(eventRecord.AggregateID)
				if err != nil {
					continue // Skip trades that can't be loaded
				}
				trades = append(trades, trade)
			}
		}
	}

	return trades, nil
}

// FindByStatus finds all trades with a specific status
func (r *EventSourcedTradeRepository) FindByStatus(status TradeStatus) ([]*TradeAggregate, error) {
	// This is inefficient for event sourcing - would need projection in real system
	// For now, load all trades and filter by status
	tradeMatchedEvents, err := r.eventStore.GetEventsByType("TradeMatched", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade matched events: %w", err)
	}

	var trades []*TradeAggregate
	for _, eventRecord := range tradeMatchedEvents {
		// Load the full aggregate to check current status
		trade, err := r.FindByID(eventRecord.AggregateID)
		if err != nil {
			continue // Skip trades that can't be loaded
		}

		if trade.Status == status {
			trades = append(trades, trade)
		}
	}

	return trades, nil
}

// FindPendingSettlements finds all trades that are in settlement process
func (r *EventSourcedTradeRepository) FindPendingSettlements() ([]*TradeAggregate, error) {
	// Find trades with settlement-related statuses
	var pendingTrades []*TradeAggregate

	settlementStatuses := []TradeStatus{
		TradeStatusConfirmed,
		TradeStatusSettlementInitiated,
		TradeStatusPaymentReceived,
		TradeStatusSharesTransferred,
	}

	for _, status := range settlementStatuses {
		trades, err := r.FindByStatus(status)
		if err != nil {
			return nil, fmt.Errorf("failed to get trades with status %s: %w", status, err)
		}
		pendingTrades = append(pendingTrades, trades...)
	}

	return pendingTrades, nil
}

// FindBySecurityAndPeriod finds trades for a security within a time period
func (r *EventSourcedTradeRepository) FindBySecurityAndPeriod(securityID string, from, to time.Time) ([]*TradeAggregate, error) {
	// Get all trades for the security first
	allTrades, err := r.FindBySecurity(securityID)
	if err != nil {
		return nil, err
	}

	// Filter by time period
	var periodTrades []*TradeAggregate
	for _, trade := range allTrades {
		if trade.MatchedAt.After(from) && trade.MatchedAt.Before(to) {
			periodTrades = append(periodTrades, trade)
		}
	}

	return periodTrades, nil
}

// Save saves a trade aggregate (used for snapshots)
func (r *EventSourcedTradeRepository) Save(trade *TradeAggregate) error {
	return r.createSnapshot(trade)
}

// createSnapshot creates a snapshot of the trade aggregate
func (r *EventSourcedTradeRepository) createSnapshot(trade *TradeAggregate) error {
	if trade.GetVersion() < 10 {
		// Only create snapshots for aggregates with significant event history
		return nil
	}

	snapshotData, err := r.serializeAggregate(trade)
	if err != nil {
		return fmt.Errorf("failed to serialize aggregate: %w", err)
	}

	snapshot := &events.Snapshot{
		AggregateID:      trade.GetID(),
		AggregateType:    trade.GetType(),
		AggregateVersion: trade.GetVersion(),
		SnapshotData:     snapshotData,
	}

	return r.eventStore.SaveSnapshot(snapshot)
}

// loadFromSnapshot loads trade state from a snapshot
func (r *EventSourcedTradeRepository) loadFromSnapshot(trade *TradeAggregate, snapshot *events.Snapshot) error {
	return r.deserializeAggregate(trade, snapshot.SnapshotData)
}

// serializeAggregate serializes the trade aggregate to bytes
func (r *EventSourcedTradeRepository) serializeAggregate(trade *TradeAggregate) ([]byte, error) {
	// Implementation would serialize the aggregate state
	// For now, return empty bytes
	return []byte{}, nil
}

// deserializeAggregate deserializes bytes to trade aggregate
func (r *EventSourcedTradeRepository) deserializeAggregate(trade *TradeAggregate, data []byte) error {
	// Implementation would deserialize the aggregate state
	// For now, do nothing
	return nil
}

// convertEventRecordsToDomainEvents converts event store records to domain events
func (r *EventSourcedTradeRepository) convertEventRecordsToDomainEvents(eventRecords []*events.Event) ([]events.DomainEvent, error) {
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
func (r *EventSourcedTradeRepository) convertEventRecordToDomainEvent(eventRecord *events.Event) (events.DomainEvent, error) {
	switch eventRecord.EventType {
	case "TradeMatched":
		return r.deserializeTradeMatched(eventRecord.EventData)
	case "TradeConfirmed":
		return r.deserializeTradeConfirmed(eventRecord.EventData)
	case "TradeSettlementInitiated":
		return r.deserializeTradeSettlementInitiated(eventRecord.EventData)
	case "PaymentReceived":
		return r.deserializePaymentReceived(eventRecord.EventData)
	case "SharesTransferred":
		return r.deserializeSharesTransferred(eventRecord.EventData)
	case "TradeSettled":
		return r.deserializeTradeSettled(eventRecord.EventData)
	case "TradeFailed":
		return r.deserializeTradeFailed(eventRecord.EventData)
	case "TradeCancelled":
		return r.deserializeTradeCancelled(eventRecord.EventData)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventRecord.EventType)
	}
}

// Event deserialization methods (simplified implementations)
func (r *EventSourcedTradeRepository) deserializeTradeMatched(data []byte) (*TradeMatched, error) {
	// Implementation would deserialize JSON to TradeMatched
	// For now, return a basic event
	return &TradeMatched{}, nil
}

func (r *EventSourcedTradeRepository) deserializeTradeConfirmed(data []byte) (*TradeConfirmed, error) {
	return &TradeConfirmed{}, nil
}

func (r *EventSourcedTradeRepository) deserializeTradeSettlementInitiated(data []byte) (*TradeSettlementInitiated, error) {
	return &TradeSettlementInitiated{}, nil
}

func (r *EventSourcedTradeRepository) deserializePaymentReceived(data []byte) (*PaymentReceived, error) {
	return &PaymentReceived{}, nil
}

func (r *EventSourcedTradeRepository) deserializeSharesTransferred(data []byte) (*SharesTransferred, error) {
	return &SharesTransferred{}, nil
}

func (r *EventSourcedTradeRepository) deserializeTradeSettled(data []byte) (*TradeSettled, error) {
	return &TradeSettled{}, nil
}

func (r *EventSourcedTradeRepository) deserializeTradeFailed(data []byte) (*TradeFailed, error) {
	return &TradeFailed{}, nil
}

func (r *EventSourcedTradeRepository) deserializeTradeCancelled(data []byte) (*TradeCancelled, error) {
	return &TradeCancelled{}, nil
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

// ProjectionTradeRepository implements TradeRepository using read model projections
type ProjectionTradeRepository struct {
	db *sql.DB
}

// NewProjectionTradeRepository creates a new projection-based trade repository
func NewProjectionTradeRepository(db *sql.DB) *ProjectionTradeRepository {
	return &ProjectionTradeRepository{db: db}
}

// FindByID finds a trade by ID from the projection
func (r *ProjectionTradeRepository) FindByID(tradeID string) (*TradeAggregate, error) {
	// This would query the trades projection table
	// For now, return not implemented
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByUser finds trades by user from the projection
func (r *ProjectionTradeRepository) FindByUser(userID string) ([]*TradeAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindBySecurity finds trades by security from the projection
func (r *ProjectionTradeRepository) FindBySecurity(securityID string) ([]*TradeAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindByStatus finds trades by status from the projection
func (r *ProjectionTradeRepository) FindByStatus(status TradeStatus) ([]*TradeAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindPendingSettlements finds pending settlements from the projection
func (r *ProjectionTradeRepository) FindPendingSettlements() ([]*TradeAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// FindBySecurityAndPeriod finds trades by security and period from the projection
func (r *ProjectionTradeRepository) FindBySecurityAndPeriod(securityID string, from, to time.Time) ([]*TradeAggregate, error) {
	return nil, fmt.Errorf("projection repository not fully implemented")
}

// Save saves a trade aggregate (not applicable for read-only projections)
func (r *ProjectionTradeRepository) Save(trade *TradeAggregate) error {
	return fmt.Errorf("save operation not supported for projection repository")
}