-- Create bids projection table for read models
CREATE TABLE bids_projection (
    bid_id UUID PRIMARY KEY,
    bidder_id UUID NOT NULL,
    listing_id UUID NOT NULL,
    security_id UUID NOT NULL,
    
    -- Bid details
    bid_type VARCHAR(50) NOT NULL,
    shares_requested BIGINT NOT NULL,
    shares_remaining BIGINT NOT NULL,
    
    -- Pricing information
    bid_price DECIMAL(15,2) NOT NULL,
    total_bid_amount DECIMAL(15,2) NOT NULL,
    
    -- Timing
    placed_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    
    -- Status
    status VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Bidder information
    is_accredited BOOLEAN NOT NULL DEFAULT false,
    
    -- Partial fills
    shares_filled BIGINT DEFAULT 0,
    amount_filled DECIMAL(15,2) DEFAULT 0,
    average_fill_price DECIMAL(15,2),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Foreign key constraints
    FOREIGN KEY (bidder_id) REFERENCES users_projection(user_id),
    FOREIGN KEY (listing_id) REFERENCES listings_projection(listing_id),
    FOREIGN KEY (security_id) REFERENCES securities_projection(security_id),
    
    -- Constraints
    CHECK (shares_requested > 0),
    CHECK (shares_remaining >= 0),
    CHECK (shares_remaining <= shares_requested),
    CHECK (shares_filled >= 0),
    CHECK (shares_filled <= shares_requested),
    CHECK (bid_price > 0),
    CHECK (total_bid_amount > 0)
);

-- Indexes for performance
CREATE INDEX idx_bids_projection_bidder ON bids_projection(bidder_id);
CREATE INDEX idx_bids_projection_listing ON bids_projection(listing_id);
CREATE INDEX idx_bids_projection_security ON bids_projection(security_id);
CREATE INDEX idx_bids_projection_type ON bids_projection(bid_type);
CREATE INDEX idx_bids_projection_status ON bids_projection(status);
CREATE INDEX idx_bids_projection_active ON bids_projection(is_active);
CREATE INDEX idx_bids_projection_price ON bids_projection(bid_price);
CREATE INDEX idx_bids_projection_expires_at ON bids_projection(expires_at);
CREATE INDEX idx_bids_projection_accredited ON bids_projection(is_accredited);
CREATE INDEX idx_bids_projection_updated_at ON bids_projection(updated_at);