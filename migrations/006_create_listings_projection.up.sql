-- Create listings projection table for read models
CREATE TABLE listings_projection (
    listing_id UUID PRIMARY KEY,
    seller_id UUID NOT NULL,
    security_id UUID NOT NULL,
    
    -- Listing details
    listing_type VARCHAR(50) NOT NULL,
    shares_offered BIGINT NOT NULL,
    shares_remaining BIGINT NOT NULL,
    
    -- Pricing information
    listing_price DECIMAL(15,2),
    minimum_price DECIMAL(15,2),
    current_price DECIMAL(15,2),
    price_type VARCHAR(20) NOT NULL,
    
    -- Timing
    listed_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    
    -- Status
    status VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Restrictions
    accredited_only BOOLEAN NOT NULL DEFAULT false,
    minimum_investment DECIMAL(15,2),
    maximum_investment DECIMAL(15,2),
    
    -- Market data
    bid_count INTEGER DEFAULT 0,
    highest_bid DECIMAL(15,2),
    total_bid_volume BIGINT DEFAULT 0,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Foreign key constraints
    FOREIGN KEY (seller_id) REFERENCES users_projection(user_id),
    FOREIGN KEY (security_id) REFERENCES securities_projection(security_id),
    
    -- Constraints
    CHECK (shares_offered > 0),
    CHECK (shares_remaining >= 0),
    CHECK (shares_remaining <= shares_offered)
);

-- Indexes for performance
CREATE INDEX idx_listings_projection_seller ON listings_projection(seller_id);
CREATE INDEX idx_listings_projection_security ON listings_projection(security_id);
CREATE INDEX idx_listings_projection_type ON listings_projection(listing_type);
CREATE INDEX idx_listings_projection_status ON listings_projection(status);
CREATE INDEX idx_listings_projection_active ON listings_projection(is_active);
CREATE INDEX idx_listings_projection_price ON listings_projection(current_price);
CREATE INDEX idx_listings_projection_expires_at ON listings_projection(expires_at);
CREATE INDEX idx_listings_projection_accredited_only ON listings_projection(accredited_only);
CREATE INDEX idx_listings_projection_updated_at ON listings_projection(updated_at);