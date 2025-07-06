-- Create market data projection table for read models
CREATE TABLE market_data_projection (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    security_id UUID NOT NULL,
    
    -- Price information
    last_price DECIMAL(15,2),
    high_price DECIMAL(15,2),
    low_price DECIMAL(15,2),
    open_price DECIMAL(15,2),
    average_price DECIMAL(15,2),
    vwap DECIMAL(15,2), -- Volume Weighted Average Price
    
    -- Volume information
    volume BIGINT DEFAULT 0,
    trade_count INTEGER DEFAULT 0,
    total_value DECIMAL(15,2) DEFAULT 0,
    
    -- Market depth
    best_bid DECIMAL(15,2),
    best_ask DECIMAL(15,2),
    bid_volume BIGINT DEFAULT 0,
    ask_volume BIGINT DEFAULT 0,
    spread DECIMAL(15,2),
    
    -- Period information
    period_type VARCHAR(10) NOT NULL, -- 'daily', 'weekly', 'monthly'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Statistical data
    volatility DECIMAL(8,6),
    beta DECIMAL(8,6),
    market_cap DECIMAL(15,2),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (security_id) REFERENCES securities_projection(security_id),
    
    -- Constraints
    UNIQUE(security_id, period_type, period_start),
    CHECK (period_type IN ('daily', 'weekly', 'monthly')),
    CHECK (period_end >= period_start)
);

-- Indexes for performance
CREATE INDEX idx_market_data_projection_security ON market_data_projection(security_id);
CREATE INDEX idx_market_data_projection_period ON market_data_projection(period_type, period_start);
CREATE INDEX idx_market_data_projection_price ON market_data_projection(last_price);
CREATE INDEX idx_market_data_projection_volume ON market_data_projection(volume);
CREATE INDEX idx_market_data_projection_updated_at ON market_data_projection(updated_at);

-- Composite indexes for common queries
CREATE INDEX idx_market_data_projection_security_period ON market_data_projection(security_id, period_type, period_start DESC);