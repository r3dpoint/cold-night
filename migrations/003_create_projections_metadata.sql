-- Create projections metadata table for tracking projection state
CREATE TABLE projections_metadata (
    projection_name VARCHAR(100) PRIMARY KEY,
    last_processed_event_id UUID,
    last_processed_position BIGINT NOT NULL DEFAULT 0,
    last_processed_timestamp TIMESTAMPTZ,
    projection_version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_rebuilding BOOLEAN NOT NULL DEFAULT false,
    error_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_projections_metadata_active ON projections_metadata(is_active);
CREATE INDEX idx_projections_metadata_rebuilding ON projections_metadata(is_rebuilding);
CREATE INDEX idx_projections_metadata_updated ON projections_metadata(updated_at);