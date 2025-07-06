package execution

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"securities-marketplace/domains/shared/events"
)

// ExecutionService provides application services for trade execution
type ExecutionService struct {
	repository     TradeRepository
	eventStore     events.EventStore
	eventBus       events.EventBus
	matchingEngine *OrderMatchingEngine
}

// NewExecutionService creates a new execution service
func NewExecutionService(repository TradeRepository, eventStore events.EventStore, eventBus events.EventBus) *ExecutionService {
	matchingEngine := NewOrderMatchingEngine(eventStore, eventBus)
	
	return &ExecutionService{
		repository:     repository,
		eventStore:     eventStore,
		eventBus:       eventBus,
		matchingEngine: matchingEngine,
	}
}

// ExecuteTradeMatch creates a new trade from a match result
func (s *ExecutionService) ExecuteTradeMatch(match *MatchResult) (*TradeAggregate, error) {
	// Create new trade aggregate
	trade := NewTradeAggregate(match.TradeID)
	
	err := trade.MatchTrade(
		match.ListingID,
		match.BidID,
		match.BuyerID,
		match.SellerID,
		match.SecurityID,
		match.SharesTraded,
		match.TradePrice,
		match.TotalAmount,
		match.SettlementDate,
		match.MatchingAlgorithm,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to match trade: %w", err)
	}

	// Save events
	err = s.saveAggregateEvents(trade, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to save trade events: %w", err)
	}

	return trade, nil
}

// ConfirmTrade handles trade confirmation by parties
func (s *ExecutionService) ConfirmTrade(tradeID, confirmedBy string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.ConfirmTrade(confirmedBy)
	if err != nil {
		return fmt.Errorf("failed to confirm trade: %w", err)
	}

	return s.saveAggregateEvents(trade, confirmedBy)
}

// InitiateSettlement starts the settlement process for a trade
func (s *ExecutionService) InitiateSettlement(tradeID, escrowAccountID, initiatedBy string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.InitiateSettlement(escrowAccountID, initiatedBy)
	if err != nil {
		return fmt.Errorf("failed to initiate settlement: %w", err)
	}

	return s.saveAggregateEvents(trade, initiatedBy)
}

// RecordPayment records payment received for a trade
func (s *ExecutionService) RecordPayment(tradeID string, amount float64, currency, paymentMethod, transactionID string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.ReceivePayment(amount, currency, paymentMethod, transactionID)
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	return s.saveAggregateEvents(trade, "system")
}

// RecordShareTransfer records the transfer of shares for a trade
func (s *ExecutionService) RecordShareTransfer(tradeID string, sharesCount int64, fromOwner, toOwner, transferMethod, certificateHash string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.TransferShares(sharesCount, fromOwner, toOwner, transferMethod, certificateHash)
	if err != nil {
		return fmt.Errorf("failed to record share transfer: %w", err)
	}

	return s.saveAggregateEvents(trade, "system")
}

// SettleTrade completes the settlement of a trade
func (s *ExecutionService) SettleTrade(tradeID string, finalAmount, fees, taxes float64, settlementMethod string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.SettleTrade(finalAmount, fees, taxes, settlementMethod)
	if err != nil {
		return fmt.Errorf("failed to settle trade: %w", err)
	}

	return s.saveAggregateEvents(trade, "system")
}

// FailTrade marks a trade as failed
func (s *ExecutionService) FailTrade(tradeID, failureReason, failureStage, recoveryAction string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.FailTrade(failureReason, failureStage, recoveryAction)
	if err != nil {
		return fmt.Errorf("failed to fail trade: %w", err)
	}

	return s.saveAggregateEvents(trade, "system")
}

// CancelTrade cancels a trade before settlement
func (s *ExecutionService) CancelTrade(tradeID, cancellationReason, cancelledBy string) error {
	trade, err := s.repository.FindByID(tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	err = trade.CancelTrade(cancellationReason, cancelledBy)
	if err != nil {
		return fmt.Errorf("failed to cancel trade: %w", err)
	}

	return s.saveAggregateEvents(trade, cancelledBy)
}

// RunMatching executes order matching for a security
func (s *ExecutionService) RunMatching(securityID string, algorithm MatchingAlgorithm) ([]*TradeAggregate, error) {
	// Get match results from matching engine
	matches, err := s.matchingEngine.MatchOrders(securityID, algorithm)
	if err != nil {
		return nil, fmt.Errorf("matching failed: %w", err)
	}

	var trades []*TradeAggregate
	
	// Create trades from matches
	for _, match := range matches {
		trade, err := s.ExecuteTradeMatch(match)
		if err != nil {
			// Log error but continue with other matches
			fmt.Printf("Failed to execute trade match %s: %v\n", match.TradeID, err)
			continue
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTrade retrieves a trade by ID
func (s *ExecutionService) GetTrade(tradeID string) (*TradeAggregate, error) {
	return s.repository.FindByID(tradeID)
}

// GetTradesByUser retrieves all trades for a user (buyer or seller)
func (s *ExecutionService) GetTradesByUser(userID string) ([]*TradeAggregate, error) {
	return s.repository.FindByUser(userID)
}

// GetTradesBySecurity retrieves all trades for a security
func (s *ExecutionService) GetTradesBySecurity(securityID string) ([]*TradeAggregate, error) {
	return s.repository.FindBySecurity(securityID)
}

// GetTradesByStatus retrieves all trades with a specific status
func (s *ExecutionService) GetTradesByStatus(status TradeStatus) ([]*TradeAggregate, error) {
	return s.repository.FindByStatus(status)
}

// GetPendingSettlements retrieves all trades pending settlement
func (s *ExecutionService) GetPendingSettlements() ([]*TradeAggregate, error) {
	return s.repository.FindPendingSettlements()
}

// GetOverdueTrades retrieves all trades that are overdue for settlement
func (s *ExecutionService) GetOverdueTrades() ([]*TradeAggregate, error) {
	trades, err := s.repository.FindPendingSettlements()
	if err != nil {
		return nil, err
	}

	var overdueTrades []*TradeAggregate
	for _, trade := range trades {
		if trade.IsOverdue() {
			overdueTrades = append(overdueTrades, trade)
		}
	}

	return overdueTrades, nil
}

// ProcessSettlements automatically processes settlements that are ready
func (s *ExecutionService) ProcessSettlements() error {
	// Get trades ready for settlement
	trades, err := s.repository.FindByStatus(TradeStatusConfirmed)
	if err != nil {
		return fmt.Errorf("failed to get confirmed trades: %w", err)
	}

	for _, trade := range trades {
		// Check if settlement date has arrived
		if time.Now().After(trade.SettlementDate) || time.Now().Equal(trade.SettlementDate) {
			// Initiate settlement
			escrowAccountID := s.generateEscrowAccountID()
			err := s.InitiateSettlement(trade.ID, escrowAccountID, "system")
			if err != nil {
				fmt.Printf("Failed to initiate settlement for trade %s: %v\n", trade.ID, err)
				continue
			}
		}
	}

	return nil
}

// AutoConfirmTrades automatically confirms trades that don't require manual confirmation
func (s *ExecutionService) AutoConfirmTrades() error {
	// Get trades that are matched but not confirmed
	trades, err := s.repository.FindByStatus(TradeStatusMatched)
	if err != nil {
		return fmt.Errorf("failed to get matched trades: %w", err)
	}

	for _, trade := range trades {
		// Auto-confirm after a certain period (e.g., 1 hour)
		if time.Since(trade.MatchedAt) > time.Hour {
			// Confirm on behalf of both parties
			err := s.ConfirmTrade(trade.ID, trade.BuyerID)
			if err != nil {
				fmt.Printf("Failed to auto-confirm trade %s for buyer: %v\n", trade.ID, err)
				continue
			}

			err = s.ConfirmTrade(trade.ID, trade.SellerID)
			if err != nil {
				fmt.Printf("Failed to auto-confirm trade %s for seller: %v\n", trade.ID, err)
				continue
			}
		}
	}

	return nil
}

// GetMarketStatistics calculates market statistics for a security
func (s *ExecutionService) GetMarketStatistics(securityID string, period time.Duration) (*MarketStatistics, error) {
	// Get trades for the security within the period
	trades, err := s.repository.FindBySecurityAndPeriod(securityID, time.Now().Add(-period), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	if len(trades) == 0 {
		return &MarketStatistics{
			SecurityID: securityID,
			Period:     period,
		}, nil
	}

	// Calculate statistics
	stats := &MarketStatistics{
		SecurityID:  securityID,
		Period:      period,
		TradeCount:  len(trades),
	}

	var totalVolume int64
	var totalValue float64
	var prices []float64

	for _, trade := range trades {
		if trade.IsSettled() {
			totalVolume += trade.SharesTraded
			totalValue += trade.TotalAmount
			prices = append(prices, trade.TradePrice)
		}
	}

	stats.TotalVolume = totalVolume
	stats.TotalValue = totalValue

	if len(prices) > 0 {
		// Calculate price statistics
		stats.HighPrice = prices[0]
		stats.LowPrice = prices[0]
		var priceSum float64

		for _, price := range prices {
			if price > stats.HighPrice {
				stats.HighPrice = price
			}
			if price < stats.LowPrice {
				stats.LowPrice = price
			}
			priceSum += price
		}

		stats.AveragePrice = priceSum / float64(len(prices))
		stats.LastPrice = prices[len(prices)-1]

		if stats.TotalVolume > 0 {
			stats.VWAP = totalValue / float64(totalVolume)
		}
	}

	return stats, nil
}

// MarketStatistics holds market data for a security
type MarketStatistics struct {
	SecurityID   string        `json:"securityId"`
	Period       time.Duration `json:"period"`
	TradeCount   int           `json:"tradeCount"`
	TotalVolume  int64         `json:"totalVolume"`
	TotalValue   float64       `json:"totalValue"`
	HighPrice    float64       `json:"highPrice"`
	LowPrice     float64       `json:"lowPrice"`
	LastPrice    float64       `json:"lastPrice"`
	AveragePrice float64       `json:"averagePrice"`
	VWAP         float64       `json:"vwap"` // Volume Weighted Average Price
}

// saveAggregateEvents saves uncommitted events from an aggregate
func (s *ExecutionService) saveAggregateEvents(trade *TradeAggregate, userID string) error {
	uncommittedEvents := trade.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	// Convert domain events to event store events
	var events []*events.Event
	correlationID := uuid.New().String()

	for i, domainEvent := range uncommittedEvents {
		var causationID *string
		if i > 0 {
			prevEventID := events[i-1].EventID
			causationID = &prevEventID
		}

		event, err := s.eventStore.CreateEventFromDomain(domainEvent, userID, correlationID, causationID)
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		event.AggregateVersion = trade.GetVersion() + i + 1
		events = append(events, event)
	}

	// Save events
	err := s.eventStore.SaveEvents(events)
	if err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	// Publish events to event bus
	for _, domainEvent := range uncommittedEvents {
		err = s.eventBus.Publish(domainEvent)
		if err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish event %s: %v\n", domainEvent.GetEventType(), err)
		}
	}

	// Mark events as committed
	trade.MarkEventsAsCommitted()

	return nil
}

// generateEscrowAccountID generates a unique escrow account ID
func (s *ExecutionService) generateEscrowAccountID() string {
	return fmt.Sprintf("escrow_%s", uuid.New().String())
}