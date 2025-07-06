package events

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresEventStore implements EventStore using PostgreSQL
type PostgresEventStore struct {
	db *sql.DB
}

// NewEventStore creates a new PostgreSQL event store
func NewEventStore(db *sql.DB) *PostgresEventStore {
	return &PostgresEventStore{db: db}
}

// SaveEvent saves a single event to the event store
func (es *PostgresEventStore) SaveEvent(event *Event) error {
	return es.SaveEvents([]*Event{event})
}

// SaveEvents saves multiple events to the event store in a transaction
func (es *PostgresEventStore) SaveEvents(events []*Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := es.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO events (
			event_id, event_type, aggregate_id, aggregate_type, aggregate_version, 
			event_version, event_data, metadata, occurred_at, user_id, correlation_id, 
			causation_id, ip_address, user_agent, session_id, checksum
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		// Generate checksum for event integrity
		event.Checksum = es.generateChecksum(event)
		
		// Serialize metadata
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to serialize metadata: %w", err)
		}

		_, err = stmt.Exec(
			event.EventID,
			event.EventType,
			event.AggregateID,
			event.AggregateType,
			event.AggregateVersion,
			event.EventVersion,
			event.EventData,
			metadataJSON,
			event.OccurredAt,
			event.UserID,
			event.CorrelationID,
			event.CausationID,
			event.IPAddress,
			event.UserAgent,
			event.SessionID,
			event.Checksum,
		)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" { // unique_violation
					return fmt.Errorf("event already exists or version conflict: %w", err)
				}
			}
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return tx.Commit()
}

// GetEvents retrieves events for a specific aggregate
func (es *PostgresEventStore) GetEvents(aggregateID string, fromVersion int) ([]*Event, error) {
	query := `
		SELECT event_number, event_id, event_type, aggregate_id, aggregate_type, 
			   aggregate_version, event_version, event_data, metadata, occurred_at, 
			   user_id, correlation_id, causation_id, ip_address, user_agent, 
			   session_id, checksum
		FROM events
		WHERE aggregate_id = $1 AND aggregate_version >= $2
		ORDER BY aggregate_version ASC
	`

	rows, err := es.db.Query(query, aggregateID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		var metadataJSON []byte

		err := rows.Scan(
			&event.EventNumber,
			&event.EventID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.AggregateVersion,
			&event.EventVersion,
			&event.EventData,
			&metadataJSON,
			&event.OccurredAt,
			&event.UserID,
			&event.CorrelationID,
			&event.CausationID,
			&event.IPAddress,
			&event.UserAgent,
			&event.SessionID,
			&event.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Deserialize metadata
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

// GetEventsByType retrieves events by event type
func (es *PostgresEventStore) GetEventsByType(eventType string, limit int) ([]*Event, error) {
	query := `
		SELECT event_number, event_id, event_type, aggregate_id, aggregate_type, 
			   aggregate_version, event_version, event_data, metadata, occurred_at, 
			   user_id, correlation_id, causation_id, ip_address, user_agent, 
			   session_id, checksum
		FROM events
		WHERE event_type = $1
		ORDER BY event_number ASC
		LIMIT $2
	`

	rows, err := es.db.Query(query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}
	defer rows.Close()

	return es.scanEvents(rows)
}

// GetAllEvents retrieves all events from a specific event number
func (es *PostgresEventStore) GetAllEvents(fromEventNumber int64, limit int) ([]*Event, error) {
	query := `
		SELECT event_number, event_id, event_type, aggregate_id, aggregate_type, 
			   aggregate_version, event_version, event_data, metadata, occurred_at, 
			   user_id, correlation_id, causation_id, ip_address, user_agent, 
			   session_id, checksum
		FROM events
		WHERE event_number > $1
		ORDER BY event_number ASC
		LIMIT $2
	`

	rows, err := es.db.Query(query, fromEventNumber, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query all events: %w", err)
	}
	defer rows.Close()

	return es.scanEvents(rows)
}

// GetSnapshot retrieves the latest snapshot for an aggregate
func (es *PostgresEventStore) GetSnapshot(aggregateID string) (*Snapshot, error) {
	query := `
		SELECT aggregate_id, aggregate_type, aggregate_version, snapshot_data, created_at
		FROM snapshots
		WHERE aggregate_id = $1
	`

	snapshot := &Snapshot{}
	err := es.db.QueryRow(query, aggregateID).Scan(
		&snapshot.AggregateID,
		&snapshot.AggregateType,
		&snapshot.AggregateVersion,
		&snapshot.SnapshotData,
		&snapshot.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No snapshot found
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return snapshot, nil
}

// SaveSnapshot saves an aggregate snapshot
func (es *PostgresEventStore) SaveSnapshot(snapshot *Snapshot) error {
	query := `
		INSERT INTO snapshots (aggregate_id, aggregate_type, aggregate_version, snapshot_data, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (aggregate_id)
		DO UPDATE SET
			aggregate_type = EXCLUDED.aggregate_type,
			aggregate_version = EXCLUDED.aggregate_version,
			snapshot_data = EXCLUDED.snapshot_data,
			created_at = EXCLUDED.created_at
	`

	_, err := es.db.Exec(query,
		snapshot.AggregateID,
		snapshot.AggregateType,
		snapshot.AggregateVersion,
		snapshot.SnapshotData,
		snapshot.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	return nil
}

// GetProjectionCheckpoint retrieves the checkpoint for a projection
func (es *PostgresEventStore) GetProjectionCheckpoint(projectionName string) (*ProjectionCheckpoint, error) {
	query := `
		SELECT projection_name, last_processed_event_number, last_processed_at, status
		FROM projection_checkpoints
		WHERE projection_name = $1
	`

	checkpoint := &ProjectionCheckpoint{}
	err := es.db.QueryRow(query, projectionName).Scan(
		&checkpoint.ProjectionName,
		&checkpoint.LastProcessedEventNumber,
		&checkpoint.LastProcessedAt,
		&checkpoint.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default checkpoint if not found
			return &ProjectionCheckpoint{
				ProjectionName:            projectionName,
				LastProcessedEventNumber:  0,
				LastProcessedAt:           time.Now(),
				Status:                    "active",
			}, nil
		}
		return nil, fmt.Errorf("failed to get projection checkpoint: %w", err)
	}

	return checkpoint, nil
}

// SaveProjectionCheckpoint saves a projection checkpoint
func (es *PostgresEventStore) SaveProjectionCheckpoint(checkpoint *ProjectionCheckpoint) error {
	query := `
		INSERT INTO projection_checkpoints (projection_name, last_processed_event_number, last_processed_at, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (projection_name)
		DO UPDATE SET
			last_processed_event_number = EXCLUDED.last_processed_event_number,
			last_processed_at = EXCLUDED.last_processed_at,
			status = EXCLUDED.status
	`

	_, err := es.db.Exec(query,
		checkpoint.ProjectionName,
		checkpoint.LastProcessedEventNumber,
		checkpoint.LastProcessedAt,
		checkpoint.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to save projection checkpoint: %w", err)
	}

	return nil
}

// CreateEventFromDomain creates an Event from a DomainEvent
func (es *PostgresEventStore) CreateEventFromDomain(domainEvent DomainEvent, userID, correlationID string, causationID *string) (*Event, error) {
	eventData, err := domainEvent.GetEventData()
	if err != nil {
		return nil, fmt.Errorf("failed to get event data: %w", err)
	}

	metadata := domainEvent.GetMetadata()
	if correlationID != "" {
		metadata.CorrelationID = correlationID
	}
	if causationID != nil {
		metadata.CausationID = *causationID
	}

	event := &Event{
		EventID:          uuid.New().String(),
		EventType:        domainEvent.GetEventType(),
		AggregateID:      domainEvent.GetAggregateID(),
		AggregateType:    domainEvent.GetAggregateType(),
		AggregateVersion: 0, // Will be set by repository
		EventVersion:     1,
		EventData:        eventData,
		Metadata:         metadata,
		OccurredAt:       time.Now(),
		UserID:           userID,
		CorrelationID:    correlationID,
		CausationID:      causationID,
	}

	return event, nil
}

// scanEvents is a helper function to scan multiple events from database rows
func (es *PostgresEventStore) scanEvents(rows *sql.Rows) ([]*Event, error) {
	var events []*Event
	for rows.Next() {
		event := &Event{}
		var metadataJSON []byte

		err := rows.Scan(
			&event.EventNumber,
			&event.EventID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.AggregateVersion,
			&event.EventVersion,
			&event.EventData,
			&metadataJSON,
			&event.OccurredAt,
			&event.UserID,
			&event.CorrelationID,
			&event.CausationID,
			&event.IPAddress,
			&event.UserAgent,
			&event.SessionID,
			&event.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Deserialize metadata
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

// generateChecksum creates a SHA256 checksum for event integrity
func (es *PostgresEventStore) generateChecksum(event *Event) string {
	data := fmt.Sprintf("%s%s%s%s%d%s",
		event.EventID,
		event.EventType,
		event.AggregateID,
		event.AggregateType,
		event.AggregateVersion,
		string(event.EventData),
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}