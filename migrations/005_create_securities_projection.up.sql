-- Create securities projection table for read models
CREATE TABLE securities_projection (
    security_id UUID PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL UNIQUE,
    company_name VARCHAR(255) NOT NULL,
    
    -- Security details
    security_type VARCHAR(50) NOT NULL,
    description TEXT,
    sector VARCHAR(100),
    industry VARCHAR(100),
    
    -- Ownership and supply
    current_owner UUID NOT NULL,
    total_shares BIGINT NOT NULL,
    outstanding_shares BIGINT NOT NULL,
    treasury_shares BIGINT NOT NULL DEFAULT 0,
    
    -- Valuation
    par_value DECIMAL(15,2),
    book_value DECIMAL(15,2),
    market_cap DECIMAL(15,2),
    last_valuation_at TIMESTAMPTZ,
    
    -- Trading information
    is_tradeable BOOLEAN NOT NULL DEFAULT true,
    min_trade_amount DECIMAL(15,2) DEFAULT 1000,
    accredited_only BOOLEAN NOT NULL DEFAULT false,
    
    -- Compliance and regulatory
    is_compliant BOOLEAN NOT NULL DEFAULT false,
    regulatory_status VARCHAR(50) DEFAULT 'pending',
    compliance_notes TEXT,
    
    -- Corporate actions
    dividend_yield DECIMAL(5,4),
    last_dividend_date DATE,
    next_dividend_date DATE,
    
    -- Document tracking
    documents_count INTEGER DEFAULT 0,
    last_document_upload_at TIMESTAMPTZ,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    listed_at TIMESTAMPTZ,
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Foreign key constraint
    FOREIGN KEY (current_owner) REFERENCES users_projection(user_id)
);

-- Indexes for performance
CREATE INDEX idx_securities_projection_symbol ON securities_projection(symbol);
CREATE INDEX idx_securities_projection_company_name ON securities_projection(company_name);
CREATE INDEX idx_securities_projection_type ON securities_projection(security_type);
CREATE INDEX idx_securities_projection_owner ON securities_projection(current_owner);
CREATE INDEX idx_securities_projection_tradeable ON securities_projection(is_tradeable);
CREATE INDEX idx_securities_projection_accredited_only ON securities_projection(accredited_only);
CREATE INDEX idx_securities_projection_sector ON securities_projection(sector);
CREATE INDEX idx_securities_projection_updated_at ON securities_projection(updated_at);