-- Create user portfolio projection table for read models
CREATE TABLE user_portfolio_projection (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    security_id UUID NOT NULL,
    
    -- Holdings information
    shares_owned BIGINT NOT NULL DEFAULT 0,
    average_cost_basis DECIMAL(15,2),
    total_cost_basis DECIMAL(15,2),
    current_market_value DECIMAL(15,2),
    
    -- Performance metrics
    unrealized_gain_loss DECIMAL(15,2),
    unrealized_gain_loss_percent DECIMAL(8,4),
    realized_gain_loss DECIMAL(15,2),
    total_return DECIMAL(15,2),
    total_return_percent DECIMAL(8,4),
    
    -- Transaction summary
    total_purchases BIGINT DEFAULT 0,
    total_sales BIGINT DEFAULT 0,
    total_dividends DECIMAL(15,2) DEFAULT 0,
    
    -- Timing
    first_purchase_date DATE,
    last_transaction_date DATE,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users_projection(user_id),
    FOREIGN KEY (security_id) REFERENCES securities_projection(security_id),
    
    -- Constraints
    UNIQUE(user_id, security_id),
    CHECK (shares_owned >= 0)
);

-- Indexes for performance
CREATE INDEX idx_user_portfolio_projection_user ON user_portfolio_projection(user_id);
CREATE INDEX idx_user_portfolio_projection_security ON user_portfolio_projection(security_id);
CREATE INDEX idx_user_portfolio_projection_shares_owned ON user_portfolio_projection(shares_owned);
CREATE INDEX idx_user_portfolio_projection_market_value ON user_portfolio_projection(current_market_value);
CREATE INDEX idx_user_portfolio_projection_updated_at ON user_portfolio_projection(updated_at);

-- Composite indexes for common queries
CREATE INDEX idx_user_portfolio_projection_user_value ON user_portfolio_projection(user_id, current_market_value DESC);
CREATE INDEX idx_user_portfolio_projection_user_return ON user_portfolio_projection(user_id, total_return_percent DESC);