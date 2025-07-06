-- Create users projection table for read models
CREATE TABLE users_projection (
    user_id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    
    -- Account status
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_email_verified BOOLEAN NOT NULL DEFAULT false,
    is_accredited BOOLEAN NOT NULL DEFAULT false,
    is_compliant BOOLEAN NOT NULL DEFAULT false,
    
    -- Profile information
    phone VARCHAR(20),
    date_of_birth DATE,
    address JSONB,
    
    -- Trading permissions
    can_trade BOOLEAN NOT NULL DEFAULT false,
    max_investment_limit DECIMAL(15,2),
    current_investment_total DECIMAL(15,2) DEFAULT 0,
    
    -- Compliance information
    compliance_level VARCHAR(20) DEFAULT 'basic',
    accreditation_verified_at TIMESTAMPTZ,
    compliance_verified_at TIMESTAMPTZ,
    risk_tolerance VARCHAR(20) DEFAULT 'moderate',
    
    -- Document tracking
    documents_uploaded INTEGER DEFAULT 0,
    documents_verified INTEGER DEFAULT 0,
    last_document_upload_at TIMESTAMPTZ,
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    last_login_at TIMESTAMPTZ,
    login_count INTEGER DEFAULT 0,
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1
);

-- Indexes for performance
CREATE INDEX idx_users_projection_email ON users_projection(email);
CREATE INDEX idx_users_projection_username ON users_projection(username);
CREATE INDEX idx_users_projection_active ON users_projection(is_active);
CREATE INDEX idx_users_projection_accredited ON users_projection(is_accredited);
CREATE INDEX idx_users_projection_compliant ON users_projection(is_compliant);
CREATE INDEX idx_users_projection_can_trade ON users_projection(can_trade);
CREATE INDEX idx_users_projection_updated_at ON users_projection(updated_at);