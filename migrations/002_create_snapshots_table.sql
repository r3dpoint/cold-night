-- Create snapshots table for aggregate snapshots
CREATE TABLE snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_version INTEGER NOT NULL,
    snapshot_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(aggregate_id, aggregate_version),
    CHECK (aggregate_version > 0)
);

-- Indexes for performance
CREATE INDEX idx_snapshots_aggregate_id ON snapshots(aggregate_id);
CREATE INDEX idx_snapshots_aggregate_type ON snapshots(aggregate_type);
CREATE INDEX idx_snapshots_created_at ON snapshots(created_at);

-- Index for finding latest snapshot
CREATE INDEX idx_snapshots_latest ON snapshots(aggregate_id, aggregate_version DESC);