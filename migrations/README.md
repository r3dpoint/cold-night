# Database Migrations

This directory contains PostgreSQL database migrations for the Securities Trading Platform.

## Migration Order

The migrations must be applied in the following order:

1. **001_create_events_table.sql** - Core event sourcing table with audit trails
2. **002_create_snapshots_table.sql** - Aggregate snapshots for performance
3. **003_create_projections_metadata.sql** - Projection tracking and status
4. **004_create_users_projection.sql** - User read model with compliance fields
5. **005_create_securities_projection.sql** - Securities read model with market data
6. **006_create_listings_projection.sql** - Listings read model with pricing
7. **007_create_bids_projection.sql** - Bids read model with partial fills
8. **008_create_trades_projection.sql** - Trades read model with settlement tracking
9. **009_create_market_data_projection.sql** - Market data aggregations
10. **010_create_user_portfolio_projection.sql** - User portfolio holdings
11. **011_create_audit_log_table.sql** - Compliance and security audit log
12. **012_create_functions_and_triggers.sql** - Database functions and triggers

## Key Features

### Event Sourcing Infrastructure
- **Events table**: Complete audit trail with checksums and correlation IDs
- **Snapshots table**: Performance optimization for aggregate reconstruction
- **Projections metadata**: Tracking of projection rebuild status

### Read Model Projections
- **Users**: Complete user profiles with accreditation and compliance status
- **Securities**: Security details with ownership and valuation data
- **Listings**: Active sell orders with pricing and restrictions
- **Bids**: Buy orders with partial fill tracking
- **Trades**: Complete trade lifecycle from matching to settlement
- **Market Data**: Price, volume, and statistical data by time period
- **User Portfolio**: Holdings, cost basis, and performance metrics

### Compliance and Security
- **Audit Log**: Comprehensive logging for regulatory compliance
- **User Permissions**: Role-based access with accreditation requirements
- **Trade Validation**: Business rule enforcement via triggers
- **Data Integrity**: Foreign key constraints and check constraints

### Performance Optimizations
- **Strategic Indexes**: Optimized for common query patterns
- **Composite Indexes**: Multi-column indexes for complex queries
- **Partial Indexes**: Filtered indexes for specific conditions
- **Automatic Updates**: Triggers for calculated fields and timestamps

## Usage

Apply migrations using your preferred PostgreSQL migration tool:

```bash
# Using psql
psql -d trading_platform -f 001_create_events_table.sql
psql -d trading_platform -f 002_create_snapshots_table.sql
# ... continue with remaining files

# Using flyway, liquibase, or similar migration tools
# Follow tool-specific documentation
```

## Environment Setup

Ensure your PostgreSQL instance has:
- Version 12+ (for generated columns and advanced features)
- UUID extension: `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`
- Proper connection limits and performance tuning
- Regular backup and archival procedures for compliance

## Monitoring

Monitor the following for performance:
- Event table growth rate and partitioning needs
- Projection rebuild times and error rates
- Query performance on indexed columns
- Audit log retention and archival requirements