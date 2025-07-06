-- Create helper functions and triggers for the trading platform

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at columns
CREATE TRIGGER trigger_users_projection_updated_at
    BEFORE UPDATE ON users_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_securities_projection_updated_at
    BEFORE UPDATE ON securities_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_listings_projection_updated_at
    BEFORE UPDATE ON listings_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_bids_projection_updated_at
    BEFORE UPDATE ON bids_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_trades_projection_updated_at
    BEFORE UPDATE ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_market_data_projection_updated_at
    BEFORE UPDATE ON market_data_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_user_portfolio_projection_updated_at
    BEFORE UPDATE ON user_portfolio_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_projections_metadata_updated_at
    BEFORE UPDATE ON projections_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate net amount for trades
CREATE OR REPLACE FUNCTION calculate_trade_net_amount()
RETURNS TRIGGER AS $$
BEGIN
    NEW.net_amount = NEW.total_amount - COALESCE(NEW.fees, 0) - COALESCE(NEW.taxes, 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for trade net amount calculation
CREATE TRIGGER trigger_trades_projection_calculate_net_amount
    BEFORE INSERT OR UPDATE OF total_amount, fees, taxes ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION calculate_trade_net_amount();

-- Function to validate trade settlement date
CREATE OR REPLACE FUNCTION validate_trade_settlement_date()
RETURNS TRIGGER AS $$
BEGIN
    -- Settlement date must be at least T+1 (next business day)
    IF NEW.settlement_date < DATE(NEW.matched_at) + INTERVAL '1 day' THEN
        RAISE EXCEPTION 'Settlement date cannot be before T+1';
    END IF;
    
    -- Settlement date should not be more than T+5
    IF NEW.settlement_date > DATE(NEW.matched_at) + INTERVAL '5 days' THEN
        RAISE EXCEPTION 'Settlement date cannot be more than T+5';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for trade settlement date validation
CREATE TRIGGER trigger_trades_projection_validate_settlement_date
    BEFORE INSERT OR UPDATE OF settlement_date, matched_at ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION validate_trade_settlement_date();

-- Function to update portfolio on trade completion
CREATE OR REPLACE FUNCTION update_portfolio_on_trade()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update portfolio when trade is settled
    IF NEW.status = 'settled' AND (OLD.status IS NULL OR OLD.status != 'settled') THEN
        -- Update buyer's portfolio
        INSERT INTO user_portfolio_projection (user_id, security_id, shares_owned, total_cost_basis, average_cost_basis, total_purchases, last_transaction_date)
        VALUES (NEW.buyer_id, NEW.security_id, NEW.shares_traded, NEW.net_amount, NEW.trade_price, 1, DATE(NEW.settled_at))
        ON CONFLICT (user_id, security_id) DO UPDATE SET
            shares_owned = user_portfolio_projection.shares_owned + NEW.shares_traded,
            total_cost_basis = user_portfolio_projection.total_cost_basis + NEW.net_amount,
            average_cost_basis = (user_portfolio_projection.total_cost_basis + NEW.net_amount) / (user_portfolio_projection.shares_owned + NEW.shares_traded),
            total_purchases = user_portfolio_projection.total_purchases + 1,
            last_transaction_date = DATE(NEW.settled_at);
        
        -- Update seller's portfolio
        INSERT INTO user_portfolio_projection (user_id, security_id, shares_owned, total_sales, last_transaction_date)
        VALUES (NEW.seller_id, NEW.security_id, -NEW.shares_traded, 1, DATE(NEW.settled_at))
        ON CONFLICT (user_id, security_id) DO UPDATE SET
            shares_owned = user_portfolio_projection.shares_owned - NEW.shares_traded,
            total_sales = user_portfolio_projection.total_sales + 1,
            last_transaction_date = DATE(NEW.settled_at);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for portfolio updates
CREATE TRIGGER trigger_trades_projection_update_portfolio
    AFTER UPDATE OF status ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_portfolio_on_trade();

-- Function to update listing remaining shares
CREATE OR REPLACE FUNCTION update_listing_on_trade()
RETURNS TRIGGER AS $$
BEGIN
    -- Update listing remaining shares when trade is matched
    IF NEW.status = 'matched' AND (OLD.status IS NULL OR OLD.status != 'matched') THEN
        UPDATE listings_projection 
        SET shares_remaining = shares_remaining - NEW.shares_traded
        WHERE listing_id = NEW.listing_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for listing updates
CREATE TRIGGER trigger_trades_projection_update_listing
    AFTER UPDATE OF status ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION update_listing_on_trade();

-- Function to validate user trading permissions
CREATE OR REPLACE FUNCTION validate_user_trading_permissions()
RETURNS TRIGGER AS $$
DECLARE
    buyer_can_trade BOOLEAN;
    seller_can_trade BOOLEAN;
    buyer_accredited BOOLEAN;
    seller_accredited BOOLEAN;
    security_accredited_only BOOLEAN;
BEGIN
    -- Check buyer permissions
    SELECT can_trade, is_accredited INTO buyer_can_trade, buyer_accredited
    FROM users_projection WHERE user_id = NEW.buyer_id;
    
    -- Check seller permissions
    SELECT can_trade, is_accredited INTO seller_can_trade, seller_accredited
    FROM users_projection WHERE user_id = NEW.seller_id;
    
    -- Check security requirements
    SELECT accredited_only INTO security_accredited_only
    FROM securities_projection WHERE security_id = NEW.security_id;
    
    -- Validate permissions
    IF NOT buyer_can_trade THEN
        RAISE EXCEPTION 'Buyer is not authorized to trade';
    END IF;
    
    IF NOT seller_can_trade THEN
        RAISE EXCEPTION 'Seller is not authorized to trade';
    END IF;
    
    IF security_accredited_only AND NOT buyer_accredited THEN
        RAISE EXCEPTION 'Buyer must be accredited for this security';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for trading permission validation
CREATE TRIGGER trigger_trades_projection_validate_permissions
    BEFORE INSERT ON trades_projection
    FOR EACH ROW
    EXECUTE FUNCTION validate_user_trading_permissions();