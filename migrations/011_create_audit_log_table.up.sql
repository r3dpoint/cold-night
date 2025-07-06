-- Create audit log table for compliance and security
CREATE TABLE audit_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Event identification
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL, -- 'security', 'trading', 'compliance', 'admin'
    event_description TEXT NOT NULL,
    
    -- User and session information
    user_id UUID,
    username VARCHAR(100),
    user_role VARCHAR(50),
    session_id VARCHAR(100),
    
    -- Request information
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    
    -- Resource information
    resource_type VARCHAR(100),
    resource_id UUID,
    
    -- Data changes
    old_values JSONB,
    new_values JSONB,
    
    -- Status and result
    status VARCHAR(20) NOT NULL, -- 'success', 'failure', 'warning'
    error_message TEXT,
    
    -- Compliance and risk
    risk_score INTEGER,
    compliance_flags TEXT[],
    
    -- Timing
    event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Additional metadata
    metadata JSONB,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users_projection(user_id)
);

-- Indexes for performance and compliance queries
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_event_category ON audit_log(event_category);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_username ON audit_log(username);
CREATE INDEX idx_audit_log_timestamp ON audit_log(event_timestamp);
CREATE INDEX idx_audit_log_status ON audit_log(status);
CREATE INDEX idx_audit_log_ip_address ON audit_log(ip_address);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_session_id ON audit_log(session_id);

-- Composite indexes for common compliance queries
CREATE INDEX idx_audit_log_user_period ON audit_log(user_id, event_timestamp DESC);
CREATE INDEX idx_audit_log_category_period ON audit_log(event_category, event_timestamp DESC);
CREATE INDEX idx_audit_log_failures ON audit_log(status, event_timestamp DESC) WHERE status = 'failure';

-- Partial index for high-risk events
CREATE INDEX idx_audit_log_high_risk ON audit_log(event_timestamp DESC, risk_score) WHERE risk_score > 70;