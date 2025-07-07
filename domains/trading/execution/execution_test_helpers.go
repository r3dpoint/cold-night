package execution

import (
	"time"

	"github.com/google/uuid"
)

// Test helpers for execution domain

// NewTestTrade creates a trade for testing
func NewTestTrade() *TradeAggregate {
	tradeID := uuid.New().String()
	trade := NewTradeAggregate(tradeID)
	return trade
}

// NewTestTradeWithMatch creates a trade with basic match for testing
func NewTestTradeWithMatch() *TradeAggregate {
	trade := NewTestTrade()
	trade.MatchTrade(
		"listing-123",
		stringPtr("bid-456"),
		"buyer-789",
		"seller-101",
		"security-001",
		100,
		50.00,
		5000.00,
		time.Now().AddDate(0, 0, 2),
		"price_time_priority",
	)
	return trade
}

// NewTestConfirmedTrade creates a confirmed trade for testing
func NewTestConfirmedTrade() *TradeAggregate {
	trade := NewTestTradeWithMatch()
	trade.MarkEventsAsCommitted()
	trade.ConfirmTrade("buyer-789")
	trade.ConfirmTrade("seller-101")
	trade.MarkEventsAsCommitted()
	return trade
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

// Test constants
const (
	TestTradeID     = "test-trade-123"
	TestSecurityID  = "test-security-456"
	TestListingID   = "test-listing-789"
	TestBidID       = "test-bid-101"
	TestBuyerID     = "test-buyer-123"
	TestSellerID    = "test-seller-456"
)