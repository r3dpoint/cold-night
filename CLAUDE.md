# Securities Marketplace Implementation Guide

## Overview

This document contains the complete architecture and implementation plan for a securities marketplace built with Go backend and event sourcing, designed for implementation with Claude Code.

## Project Requirements

- **Backend**: Go for maximum resource efficiency
- **Architecture**: Event sourcing with fractal domain structure
- **Testing**: Fast unit tests with comprehensive coverage
- **Frontend**: Multiple role-based UIs (admin, compliance, brokers, clients)
- **UI Strategy**: Server-rendered HTML with progressive enhancement, real-time updates
- **Development**: DevContainer with open source tools (Podman, no Docker dependency)
- **Platform**: WSL-optimized

## Event Sourcing & Domain Design

### Core Event Schema

```typescript
// Event Metadata Structure
EventMetadata {
  eventId: string;           // UUID
  eventType: string;         // Event class name
  aggregateId: string;       // Aggregate identifier
  aggregateType: string;     // Aggregate type
  eventVersion: number;      // For schema evolution
  timestamp: Date;           // Event occurrence time
  userId: string;            // Acting user
  correlationId: string;     // Link related events
  causationId: string;       // Parent event that caused this
  ipAddress: string;         // Source IP for audit
  userAgent: string;         // Client information
  sessionId: string;         // User session
  checksum: string;          // Event integrity hash
}
```

### Aggregate Boundaries

**User Aggregate** (`user-{userId}`)

- User lifecycle, accreditation, and compliance status
- Events: UserRegistered, AccreditationSubmitted, AccreditationVerified, AccreditationRevoked, ComplianceCheckPerformed, UserSuspended

**Security Aggregate** (`security-{securityId}`)

- Security listings, documentation, and lifecycle
- Events: SecurityListed, SecurityDocumentAdded, SecurityUpdated, SecuritySuspended, SecurityDelisted

**Listing Aggregate** (`listing-{listingId}`)

- Individual sell orders and their lifecycle
- Events: ListingCreated, ListingPriceUpdated, ListingCancelled, ListingExpired

**Bid Aggregate** (`bid-{bidId}`)

- Bid lifecycle and modifications
- Events: BidPlaced, BidModified, BidWithdrawn, BidExpired, BidRejected

**Trade Aggregate** (`trade-{tradeId}`)

- Trade execution and settlement
- Events: TradeMatched, TradeConfirmed, TradeSettlementInitiated, PaymentReceived, SharesTransferred, TradeSettled, TradeFailed

**Compliance Aggregate** (`compliance-{complianceId}`)

- Regulatory compliance events and reporting
- Events: SuspiciousActivityDetected, ComplianceReportGenerated, RegulatoryInquiryReceived, ComplianceActionTaken

## Fractal Directory Structure

```
securities-marketplace/
├── cmd/
│   ├── api/           # Main API server
│   ├── worker/        # Background job processor
│   └── migrate/       # Database migration tool
├── domains/
│   ├── users/
│   │   ├── aggregate.go       # User aggregate
│   │   ├── commands.go        # User commands
│   │   ├── events.go          # User events
│   │   ├── handlers/          # HTTP handlers for users
│   │   │   ├── registration.go
│   │   │   ├── accreditation.go
│   │   │   └── profile.go
│   │   ├── projections/       # User read models
│   │   │   ├── profile.go
│   │   │   └── compliance.go
│   │   ├── repository.go      # User data access
│   │   ├── service.go         # User application service
│   │   ├── templates/         # User-specific templates
│   │   │   ├── registration.html
│   │   │   ├── profile.html
│   │   │   └── accreditation.html
│   │   └── users_test.go      # Domain tests
│   ├── securities/
│   │   ├── aggregate.go       # Security aggregate
│   │   ├── commands.go        # Security commands
│   │   ├── events.go          # Security events
│   │   ├── handlers/
│   │   │   ├── listing.go
│   │   │   ├── management.go
│   │   │   └── documentation.go
│   │   ├── projections/
│   │   │   ├── catalog.go
│   │   │   └── market_data.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   ├── templates/
│   │   │   ├── listing.html
│   │   │   ├── details.html
│   │   │   └── market.html
│   │   └── securities_test.go
│   ├── trading/
│   │   ├── listing/
│   │   │   ├── aggregate.go   # Listing aggregate
│   │   │   ├── commands.go
│   │   │   ├── events.go
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   ├── bidding/
│   │   │   ├── aggregate.go   # Bid aggregate
│   │   │   ├── commands.go
│   │   │   ├── events.go
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   ├── execution/
│   │   │   ├── aggregate.go   # Trade aggregate
│   │   │   ├── engine.go      # Matching engine
│   │   │   ├── settlement.go  # Settlement logic
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   ├── market/
│   │   │   ├── data.go        # Market data calculation
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   └── trading_test.go
│   ├── compliance/
│   │   ├── monitoring/
│   │   │   ├── aggregate.go   # Compliance record aggregate
│   │   │   ├── detection.go   # Pattern detection
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   ├── reporting/
│   │   │   ├── generators.go  # Report generators
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   ├── risk/
│   │   │   ├── assessment.go  # Risk scoring
│   │   │   ├── handlers/
│   │   │   ├── projections/
│   │   │   └── templates/
│   │   └── compliance_test.go
│   └── shared/
│       ├── events/            # Shared event infrastructure
│       │   ├── store.go
│       │   ├── bus.go
│       │   └── types.go
│       ├── auth/              # Authentication/authorization
│       │   ├── middleware.go
│       │   ├── jwt.go
│       │   └── rbac.go
│       ├── web/               # Shared web infrastructure
│       │   ├── renderer.go
│       │   ├── sse.go
│       │   ├── middleware.go
│       │   └── router.go
│       ├── storage/           # Database utilities
│       │   ├── postgres.go
│       │   ├── migrations.go
│       │   └── cache.go
│       └── testutil/          # Shared test utilities
│           ├── fixtures.go
│           ├── mocks.go
│           └── helpers.go
├── web/
│   ├── static/                # Static assets
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   ├── layouts/               # Shared layouts
│   │   ├── admin.html
│   │   ├── client.html
│   │   ├── broker.html
│   │   └── compliance.html
│   └── components/            # Reusable components
│       ├── navigation.html
│       ├── forms.html
│       ├── tables.html
│       └── widgets.html
├── migrations/                # Database migrations
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Development Environment Setup

### DevContainer Configuration (Podman + WSL)

```json
// .devcontainer/devcontainer.json
{
  "name": "Securities Marketplace",
  "dockerFile": "Dockerfile",
  "runArgs": [
    "--runtime", "crun",
    "--security-opt", "label=disable"
  ],
  "containerEnv": {
    "CGO_ENABLED": "1",
    "GOOS": "linux"
  },
  "features": {
    "ghcr.io/devcontainers/features/go:1": {
      "version": "1.21"
    },
    "ghcr.io/devcontainers/features/postgres:1": {
      "version": "15"
    }
  },
  "forwardPorts": [8080, 5432, 6379],
  "postCreateCommand": "make setup",
  "customizations": {
    "vscode": {
      "extensions": [
        "golang.go",
        "ms-vscode.vscode-json",
        "bradlc.vscode-tailwindcss"
      ]
    }
  }
}
```

```dockerfile
# .devcontainer/Dockerfile
FROM registry.fedoraproject.org/fedora:38

# Install development tools
RUN dnf update -y && dnf install -y \
    golang \
    postgresql \
    postgresql-server \
    redis \
    git \
    make \
    gcc \
    sqlite-devel \
    && dnf clean all

# Install Air for live reloading
RUN go install github.com/air-verse/air@latest

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install testing tools
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest
RUN go install gotest.tools/gotestsum@latest

WORKDIR /workspace
```

## Database Schemas

### Event Store Schema

```sql
-- Core event storage table
CREATE TABLE events (
    event_number BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    event_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_version INTEGER NOT NULL,
    event_version INTEGER NOT NULL DEFAULT 1,
    event_data JSONB NOT NULL,
    metadata JSONB NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Audit fields
    user_id VARCHAR(255) NOT NULL,
    correlation_id UUID NOT NULL,
    causation_id UUID,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    checksum VARCHAR(64) NOT NULL,
    
    CONSTRAINT events_aggregate_version_unique 
        UNIQUE (aggregate_id, aggregate_version)
);

-- Indexes for performance
CREATE INDEX idx_events_aggregate_id ON events (aggregate_id);
CREATE INDEX idx_events_aggregate_type ON events (aggregate_type);
CREATE INDEX idx_events_event_type ON events (event_type);
CREATE INDEX idx_events_occurred_at ON events (occurred_at);
CREATE INDEX idx_events_correlation_id ON events (correlation_id);
CREATE INDEX idx_events_user_id ON events (user_id);

-- Snapshots for performance optimization
CREATE TABLE snapshots (
    aggregate_id VARCHAR(255) PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_version INTEGER NOT NULL,
    snapshot_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Projection checkpoints to track processing
CREATE TABLE projection_checkpoints (
    projection_name VARCHAR(100) PRIMARY KEY,
    last_processed_event_number BIGINT NOT NULL DEFAULT 0,
    last_processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active' -- active, rebuilding, failed
);
```

### User Profile Projection Schema

```sql
CREATE TABLE user_profiles (
    user_id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    
    -- Accreditation status
    accreditation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    accreditation_type VARCHAR(20),
    accreditation_valid_until TIMESTAMP WITH TIME ZONE,
    accreditation_documents JSONB DEFAULT '[]',
    
    -- Compliance status
    compliance_status VARCHAR(20) NOT NULL DEFAULT 'clear',
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    aml_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    sanctions_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Audit trail
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Risk assessment
    risk_score INTEGER DEFAULT 0,
    watchlist_status VARCHAR(20) DEFAULT 'none'
);

CREATE INDEX idx_user_profiles_accreditation_status ON user_profiles (accreditation_status);
CREATE INDEX idx_user_profiles_compliance_status ON user_profiles (compliance_status);
CREATE INDEX idx_user_profiles_risk_score ON user_profiles (risk_score);
```

### Securities and Listings Schema

```sql
CREATE TABLE securities (
    security_id VARCHAR(255) PRIMARY KEY,
    issuer_id VARCHAR(255) NOT NULL,
    security_type VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    total_shares BIGINT NOT NULL,
    par_value DECIMAL(15,4),
    
    -- Documentation
    prospectus_hash VARCHAR(64),
    documents JSONB DEFAULT '[]',
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    listed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    delisted_at TIMESTAMP WITH TIME ZONE,
    
    -- Market data cache
    last_trade_price DECIMAL(15,4),
    market_cap DECIMAL(20,2),
    
    FOREIGN KEY (issuer_id) REFERENCES user_profiles(user_id)
);

CREATE TABLE listings (
    listing_id VARCHAR(255) PRIMARY KEY,
    security_id VARCHAR(255) NOT NULL,
    seller_id VARCHAR(255) NOT NULL,
    
    -- Listing details
    shares_offered BIGINT NOT NULL,
    shares_remaining BIGINT NOT NULL,
    listing_type VARCHAR(20) NOT NULL,
    minimum_price DECIMAL(15,4),
    reserve_price DECIMAL(15,4),
    current_price DECIMAL(15,4),
    
    -- Restrictions
    restriction_type VARCHAR(30),
    accredited_only BOOLEAN DEFAULT true,
    
    -- Timing
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    FOREIGN KEY (security_id) REFERENCES securities(security_id),
    FOREIGN KEY (seller_id) REFERENCES user_profiles(user_id)
);

CREATE INDEX idx_listings_security_id ON listings (security_id);
CREATE INDEX idx_listings_seller_id ON listings (seller_id);
CREATE INDEX idx_listings_status ON listings (status);
CREATE INDEX idx_listings_expires_at ON listings (expires_at);
```

### Bidding and Trading Schema

```sql
CREATE TABLE bids (
    bid_id VARCHAR(255) PRIMARY KEY,
    listing_id VARCHAR(255) NOT NULL,
    bidder_id VARCHAR(255) NOT NULL,
    
    -- Bid details
    shares_requested BIGINT NOT NULL,
    shares_remaining BIGINT NOT NULL,
    bid_price DECIMAL(15,4) NOT NULL,
    bid_type VARCHAR(20) NOT NULL,
    
    -- Timing
    placed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    withdrawn_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    
    FOREIGN KEY (listing_id) REFERENCES listings(listing_id),
    FOREIGN KEY (bidder_id) REFERENCES user_profiles(user_id)
);

CREATE INDEX idx_bids_listing_id ON bids (listing_id);
CREATE INDEX idx_bids_bidder_id ON bids (bidder_id);
CREATE INDEX idx_bids_status ON bids (status);
CREATE INDEX idx_bids_bid_price ON bids (bid_price DESC);

CREATE TABLE trades (
    trade_id VARCHAR(255) PRIMARY KEY,
    listing_id VARCHAR(255) NOT NULL,
    bid_id VARCHAR(255),
    buyer_id VARCHAR(255) NOT NULL,
    seller_id VARCHAR(255) NOT NULL,
    security_id VARCHAR(255) NOT NULL,
    
    -- Trade details
    shares_traded BIGINT NOT NULL,
    trade_price DECIMAL(15,4) NOT NULL,
    total_amount DECIMAL(20,2) NOT NULL,
    fees DECIMAL(10,2) DEFAULT 0,
    taxes DECIMAL(10,2) DEFAULT 0,
    
    -- Settlement
    settlement_date DATE NOT NULL,
    escrow_account_id VARCHAR(255),
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'matched',
    buyer_confirmed BOOLEAN DEFAULT false,
    seller_confirmed BOOLEAN DEFAULT false,
    payment_received BOOLEAN DEFAULT false,
    shares_transferred BOOLEAN DEFAULT false,
    
    -- Timing
    matched_at TIMESTAMP WITH TIME ZONE NOT NULL,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    settled_at TIMESTAMP WITH TIME ZONE,
    
    -- Algorithm info
    matching_algorithm VARCHAR(50),
    
    FOREIGN KEY (listing_id) REFERENCES listings(listing_id),
    FOREIGN KEY (bid_id) REFERENCES bids(bid_id),
    FOREIGN KEY (buyer_id) REFERENCES user_profiles(user_id),
    FOREIGN KEY (seller_id) REFERENCES user_profiles(user_id),
    FOREIGN KEY (security_id) REFERENCES securities(security_id)
);

CREATE INDEX idx_trades_buyer_id ON trades (buyer_id);
CREATE INDEX idx_trades_seller_id ON trades (seller_id);
CREATE INDEX idx_trades_security_id ON trades (security_id);
CREATE INDEX idx_trades_status ON trades (status);
CREATE INDEX idx_trades_matched_at ON trades (matched_at);
CREATE INDEX idx_trades_settlement_date ON trades (settlement_date);
```

### Market Data Schema

```sql
CREATE TABLE market_data (
    security_id VARCHAR(255) PRIMARY KEY,
    
    -- Current prices
    current_price DECIMAL(15,4),
    last_trade_price DECIMAL(15,4),
    last_trade_at TIMESTAMP WITH TIME ZONE,
    
    -- 24h statistics
    volume_24h BIGINT DEFAULT 0,
    high_price_24h DECIMAL(15,4),
    low_price_24h DECIMAL(15,4),
    price_change_24h DECIMAL(15,4),
    price_change_pct_24h DECIMAL(8,4),
    
    -- Market depth
    active_listings INTEGER DEFAULT 0,
    total_bids INTEGER DEFAULT 0,
    best_bid DECIMAL(15,4),
    best_ask DECIMAL(15,4),
    bid_ask_spread DECIMAL(15,4),
    
    -- Volume indicators
    avg_volume_30d BIGINT,
    volume_trend VARCHAR(10), -- up, down, stable
    
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (security_id) REFERENCES securities(security_id)
);

-- Historical price data for charts
CREATE TABLE price_history (
    id BIGSERIAL PRIMARY KEY,
    security_id VARCHAR(255) NOT NULL,
    price DECIMAL(15,4) NOT NULL,
    volume BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    FOREIGN KEY (security_id) REFERENCES securities(security_id)
);

CREATE INDEX idx_price_history_security_timestamp 
    ON price_history (security_id, timestamp DESC);
```

### Compliance and Risk Schema

```sql
CREATE TABLE compliance_records (
    record_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    
    -- Risk assessment
    risk_score INTEGER NOT NULL DEFAULT 0,
    risk_factors JSONB DEFAULT '[]',
    last_risk_assessment TIMESTAMP WITH TIME ZONE,
    
    -- Compliance status
    overall_status VARCHAR(20) NOT NULL DEFAULT 'clear',
    kyc_completed_at TIMESTAMP WITH TIME ZONE,
    aml_completed_at TIMESTAMP WITH TIME ZONE,
    next_review_due TIMESTAMP WITH TIME ZONE,
    
    -- Watchlist status
    watchlist_status VARCHAR(20) DEFAULT 'none',
    watchlist_reason TEXT,
    watchlist_added_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES user_profiles(user_id)
);

CREATE TABLE suspicious_activities (
    activity_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    
    -- Activity details
    activity_type VARCHAR(50) NOT NULL,
    severity VARCHAR(10) NOT NULL,
    description TEXT,
    
    -- Detection info
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL,
    detection_method VARCHAR(50), -- automated, manual, external
    auto_generated BOOLEAN DEFAULT true,
    
    -- Investigation
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    assigned_to VARCHAR(255),
    investigated_at TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    
    -- Related data
    related_trade_ids JSONB DEFAULT '[]',
    related_user_ids JSONB DEFAULT '[]',
    
    FOREIGN KEY (user_id) REFERENCES user_profiles(user_id)
);

CREATE INDEX idx_suspicious_activities_user_id ON suspicious_activities (user_id);
CREATE INDEX idx_suspicious_activities_status ON suspicious_activities (status);
CREATE INDEX idx_suspicious_activities_severity ON suspicious_activities (severity);
CREATE INDEX idx_suspicious_activities_detected_at ON suspicious_activities (detected_at);

CREATE TABLE regulatory_reports (
    report_id VARCHAR(255) PRIMARY KEY,
    report_type VARCHAR(20) NOT NULL,
    
    -- Reporting period
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Report data
    data JSONB NOT NULL,
    summary JSONB,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    generated_by VARCHAR(255) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE,
    submitted_by VARCHAR(255),
    
    -- External tracking
    external_reference VARCHAR(255),
    acknowledgment_received BOOLEAN DEFAULT false
);

CREATE INDEX idx_regulatory_reports_type ON regulatory_reports (report_type);
CREATE INDEX idx_regulatory_reports_period ON regulatory_reports (period_start, period_end);
CREATE INDEX idx_regulatory_reports_status ON regulatory_reports (status);
```

### User Dashboard Projection Schema

```sql
CREATE TABLE user_dashboards (
    user_id VARCHAR(255) PRIMARY KEY,
    
    -- Portfolio summary
    portfolio_value DECIMAL(20,2) DEFAULT 0,
    available_cash DECIMAL(20,2) DEFAULT 0,
    total_invested DECIMAL(20,2) DEFAULT 0,
    unrealized_pnl DECIMAL(20,2) DEFAULT 0,
    realized_pnl DECIMAL(20,2) DEFAULT 0,
    
    -- Activity counts
    active_listings_count INTEGER DEFAULT 0,
    active_bids_count INTEGER DEFAULT 0,
    pending_trades_count INTEGER DEFAULT 0,
    
    -- Recent activity summary
    last_trade_date TIMESTAMP WITH TIME ZONE,
    last_listing_date TIMESTAMP WITH TIME ZONE,
    last_bid_date TIMESTAMP WITH TIME ZONE,
    
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES user_profiles(user_id)
);

-- Detailed holdings
CREATE TABLE user_holdings (
    user_id VARCHAR(255),
    security_id VARCHAR(255),
    
    -- Position details
    shares_owned BIGINT NOT NULL DEFAULT 0,
    average_cost DECIMAL(15,4) NOT NULL DEFAULT 0,
    total_cost DECIMAL(20,2) NOT NULL DEFAULT 0,
    current_value DECIMAL(20,2) NOT NULL DEFAULT 0,
    unrealized_pnl DECIMAL(20,2) NOT NULL DEFAULT 0,
    
    -- Acquisition info
    first_acquired_at TIMESTAMP WITH TIME ZONE,
    last_acquired_at TIMESTAMP WITH TIME ZONE,
    
    PRIMARY KEY (user_id, security_id),
    FOREIGN KEY (user_id) REFERENCES user_profiles(user_id),
    FOREIGN KEY (security_id) REFERENCES securities(security_id)
);
```

## Trading Engine Algorithms

### Order Matching Engine

The trading engine implements multiple matching algorithms:

1. **Price-Time Priority**: Fixed price listings match highest bid first, then earliest timestamp
2. **Uniform Price Auction**: Auction listings find clearing price where supply meets demand
3. **Negotiated Trading**: Custom pricing logic for private negotiations

### Risk Management Engine

Real-time pattern detection:

- **Wash Trading**: Detect circular trading between related accounts
- **Unusual Volume**: Flag trades exceeding normal volume thresholds
- **Price Manipulation**: Identify potential market manipulation patterns
- **AML Compliance**: Continuous monitoring for suspicious activity

### Settlement Engine

Automated settlement with escrow:

- T+2 settlement for most securities
- Escrow account creation and management
- Payment verification and fund transfer
- Share certificate transfer and recording
- Automated compliance reporting

## Multi-Frontend Architecture

### Role-Based UI Components

**Client Portal**: Portfolio management, market data, order placement
**Broker Dashboard**: Client management, order execution, market making
**Admin Console**: User management, system configuration, monitoring
**Compliance Center**: Risk monitoring, regulatory reporting, investigations

### Progressive Enhancement Strategy

- **Base Layer**: Server-rendered HTML that works without JavaScript
- **Enhancement Layer**: HTMX for dynamic interactions
- **Real-time Layer**: Server-sent events for live updates
- **Optimization Layer**: Response caching and efficient rendering

### Template Architecture

```html
<!-- Role-specific layouts with shared components -->
web/layouts/client.html      # Client portal layout
web/layouts/admin.html       # Admin dashboard layout
web/layouts/compliance.html  # Compliance center layout
web/layouts/broker.html      # Broker dashboard layout

<!-- Reusable components -->
web/components/navigation.html
web/components/forms.html
web/components/tables.html
web/components/widgets.html

<!-- Domain-specific templates -->
domains/users/templates/registration.html
domains/trading/templates/market-data.html
domains/compliance/templates/risk-dashboard.html
```

## Makefile for Development

```makefile
# Makefile
.PHONY: setup test build run clean

# Setup development environment
setup:
 go mod download
 go install github.com/air-verse/air@latest
 go install gotest.tools/gotestsum@latest
 make migrate-up

# Run tests with coverage
test:
 gotestsum --format testname -- -race -coverprofile=coverage.out ./...
 go tool cover -html=coverage.out -o coverage.html

# Fast unit tests only
test-unit:
 gotestsum --format testname -- -short -race ./domains/...

# Integration tests
test-integration:
 gotestsum --format testname -- -run Integration ./...

# Database migrations
migrate-up:
 migrate -path migrations -database "postgres://user:pass@localhost:5432/securities?sslmode=disable" up

migrate-down:
 migrate -path migrations -database "postgres://user:pass@localhost:5432/securities?sslmode=disable" down

# Development server with live reload
dev:
 air -c .air.toml

# Build production binary
build:
 CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/api cmd/api/main.go
 CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/worker cmd/worker/main.go

# Run with Podman
podman-dev:
 podman-compose up -d postgres redis
 make dev

# Performance testing
perf-test:
 go test -bench=. -benchmem ./domains/...

clean:
 rm -rf bin/ coverage.out coverage.html
```

## Implementation Strategy

### Phase 1: Foundation (4-6 weeks)

1. Set up development environment and project structure
2. Implement event store and basic domain aggregates
3. Create user registration and authentication
4. Build basic admin interface
5. Set up testing framework

### Phase 2: Core Trading (6-8 weeks)

1. Implement securities and listings domain
2. Build bidding system with real-time updates
3. Create trading engine with matching algorithms
4. Add market data projections and dashboards
5. Implement client and broker interfaces

### Phase 3: Compliance & Polish (4-6 weeks)

1. Add compliance monitoring and risk management
2. Implement regulatory reporting features
3. Build compliance officer interface
4. Add advanced market features (auctions, etc.)
5. Performance optimization and production readiness

### Testing Strategy

- **Unit Tests**: Fast, isolated tests for each domain aggregate
- **Integration Tests**: End-to-end workflows across domains
- **Performance Tests**: Benchmark trading engine throughput
- **Compliance Tests**: Verify regulatory requirements

### Deployment Considerations

- **Event Store**: PostgreSQL with proper indexing for performance
- **Caching**: Redis for market data and session storage
- **Monitoring**: Comprehensive logging and metrics
- **Security**: TLS encryption, JWT authentication, input validation
- **Compliance**: Complete audit trails and regulatory reporting

This architecture provides a solid foundation for building a scalable, compliant securities marketplace with excellent developer experience and regulatory compliance built-in from day one.
