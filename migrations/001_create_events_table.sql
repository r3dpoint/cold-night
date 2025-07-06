-- Create events table for event sourcing
CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    event_version INTEGER NOT NULL DEFAULT 1,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Audit and correlation fields
    user_id UUID,
    correlation_id UUID,
    causation_id UUID,
    
    -- Integrity and security
    checksum VARCHAR(64) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Constraints
    UNIQUE(aggregate_id, aggregate_version),
    CHECK (aggregate_version > 0)
);

-- Indexes for performance
CREATE INDEX idx_events_aggregate_id ON events(aggregate_id);
CREATE INDEX idx_events_aggregate_type ON events(aggregate_type);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_correlation_id ON events(correlation_id);
CREATE INDEX idx_events_user_id ON events(user_id);

-- Index for event replay and projections
CREATE INDEX idx_events_aggregate_version ON events(aggregate_id, aggregate_version);
CREATE INDEX idx_events_global_sequence ON events(timestamp, event_id);