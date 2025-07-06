-- Create trades projection table for read models
CREATE TABLE trades_projection (
    trade_id UUID PRIMARY KEY,
    listing_id UUID,
    bid_id UUID,
    buyer_id UUID NOT NULL,
    seller_id UUID NOT NULL,
    security_id UUID NOT NULL,
    
    -- Trade details
    shares_traded BIGINT NOT NULL,
    trade_price DECIMAL(15,2) NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    fees DECIMAL(15,2) DEFAULT 0,
    taxes DECIMAL(15,2) DEFAULT 0,
    net_amount DECIMAL(15,2) NOT NULL,
    
    -- Settlement information
    settlement_date DATE NOT NULL,
    escrow_account_id UUID,
    
    -- Status and lifecycle
    status VARCHAR(50) NOT NULL,
    settlement_stage VARCHAR(50) NOT NULL,
    
    -- Confirmation tracking
    buyer_confirmed BOOLEAN NOT NULL DEFAULT false,
    seller_confirmed BOOLEAN NOT NULL DEFAULT false,
    
    -- Timing
    matched_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    
    -- Settlement details
    payment_amount DECIMAL(15,2),
    payment_currency VARCHAR(3),
    payment_method VARCHAR(50),
    payment_transaction_id VARCHAR(100),
    payment_received_at TIMESTAMPTZ,
    
    -- Share transfer details
    shares_transferred BIGINT,
    transfer_method VARCHAR(50),
    certificate_hash VARCHAR(64),
    shares_transferred_at TIMESTAMPTZ,
    
    -- Matching and algorithm info
    matching_algorithm VARCHAR(50) NOT NULL,
    
    -- Failure and cancellation details
    failure_reason TEXT,
    failure_stage VARCHAR(50),
    cancellation_reason TEXT,
    cancelled_by UUID,
    recovery_action TEXT,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Foreign key constraints
    FOREIGN KEY (buyer_id) REFERENCES users_projection(user_id),
    FOREIGN KEY (seller_id) REFERENCES users_projection(user_id),
    FOREIGN KEY (security_id) REFERENCES securities_projection(security_id),
    FOREIGN KEY (listing_id) REFERENCES listings_projection(listing_id),
    FOREIGN KEY (bid_id) REFERENCES bids_projection(bid_id),
    FOREIGN KEY (cancelled_by) REFERENCES users_projection(user_id),
    
    -- Constraints
    CHECK (shares_traded > 0),
    CHECK (trade_price > 0),
    CHECK (total_amount > 0),
    CHECK (net_amount > 0),
    CHECK (settlement_date >= DATE(matched_at))
);

-- Indexes for performance
CREATE INDEX idx_trades_projection_buyer ON trades_projection(buyer_id);
CREATE INDEX idx_trades_projection_seller ON trades_projection(seller_id);
CREATE INDEX idx_trades_projection_security ON trades_projection(security_id);
CREATE INDEX idx_trades_projection_listing ON trades_projection(listing_id);
CREATE INDEX idx_trades_projection_bid ON trades_projection(bid_id);
CREATE INDEX idx_trades_projection_status ON trades_projection(status);
CREATE INDEX idx_trades_projection_settlement_stage ON trades_projection(settlement_stage);
CREATE INDEX idx_trades_projection_matched_at ON trades_projection(matched_at);
CREATE INDEX idx_trades_projection_settlement_date ON trades_projection(settlement_date);
CREATE INDEX idx_trades_projection_confirmed_at ON trades_projection(confirmed_at);
CREATE INDEX idx_trades_projection_settled_at ON trades_projection(settled_at);
CREATE INDEX idx_trades_projection_algorithm ON trades_projection(matching_algorithm);
CREATE INDEX idx_trades_projection_updated_at ON trades_projection(updated_at);

-- Composite indexes for common queries
CREATE INDEX idx_trades_projection_user_trades ON trades_projection(buyer_id, seller_id);
CREATE INDEX idx_trades_projection_security_period ON trades_projection(security_id, matched_at);
CREATE INDEX idx_trades_projection_pending_settlement ON trades_projection(status, settlement_date) WHERE status IN ('confirmed', 'settlement_initiated', 'payment_received', 'shares_transferred');