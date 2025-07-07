package execution

import (
	"testing"
	"time"

	"securities-marketplace/domains/shared/testutil"
)

func TestTradeAggregate_MatchTrade_Simple(t *testing.T) {
	// Arrange
	trade := NewTestTrade()

	// Act
	err := trade.MatchTrade(
		"listing-456",
		stringPtr("bid-789"),
		"buyer-101",
		"seller-202",
		"security-303",
		100,
		50.00,
		5000.00,
		time.Now().AddDate(0, 0, 2),
		"price_time_priority",
	)

	// Assert
	testutil.AssertNoError(t, err, "Trade matching should succeed")
	testutil.AssertEqual(t, TradeStatusMatched, trade.Status, "Trade should be matched")
	testutil.AssertEqual(t, "listing-456", *trade.ListingID, "Listing ID should be set")
	testutil.AssertEqual(t, "buyer-101", trade.BuyerID, "Buyer ID should be set")
	testutil.AssertEqual(t, "seller-202", trade.SellerID, "Seller ID should be set")
	testutil.AssertEqual(t, int64(100), trade.SharesTraded, "Shares traded should be set")
	testutil.AssertEqual(t, 50.00, trade.TradePrice, "Trade price should be set")

	// Check events
	events := trade.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "TradeMatched", events[0].GetEventType(), "Should be TradeMatched event")
}

func TestTradeAggregate_ConfirmTrade_Simple(t *testing.T) {
	// Test buyer confirmation
	t.Run("buyer confirms trade", func(t *testing.T) {
		// Arrange
		trade := NewTestTradeWithMatch()
		trade.MarkEventsAsCommitted() // Clear initial events

		// Act
		err := trade.ConfirmTrade("buyer-789")

		// Assert
		testutil.AssertNoError(t, err, "Trade confirmation should succeed")
		
		// Check events
		events := trade.GetUncommittedEvents()
		testutil.AssertLengthEqual(t, 1, events, "Should have one event")
		testutil.AssertEqual(t, "TradeConfirmed", events[0].GetEventType(), "Should be TradeConfirmed event")
	})

	// Test seller confirmation
	t.Run("seller confirms trade", func(t *testing.T) {
		// Arrange
		trade := NewTestTradeWithMatch()
		trade.MarkEventsAsCommitted() // Clear initial events

		// Act
		err := trade.ConfirmTrade("seller-101")

		// Assert
		testutil.AssertNoError(t, err, "Trade confirmation should succeed")
		
		// Check events
		events := trade.GetUncommittedEvents()
		testutil.AssertLengthEqual(t, 1, events, "Should have one event")
		testutil.AssertEqual(t, "TradeConfirmed", events[0].GetEventType(), "Should be TradeConfirmed event")
	})

	// Test unauthorized confirmation
	t.Run("unauthorized user cannot confirm", func(t *testing.T) {
		// Arrange
		trade := NewTestTradeWithMatch()
		
		// Act
		err := trade.ConfirmTrade("unauthorized-user")

		// Assert
		testutil.AssertError(t, err, "Should reject unauthorized confirmation")
	})
}

func TestTradeAggregate_SettlementWorkflow_Simple(t *testing.T) {
	// Arrange
	trade := NewTestConfirmedTrade()

	// Test settlement initiation
	t.Run("initiate settlement", func(t *testing.T) {
		err := trade.InitiateSettlement("escrow-123", "system")
		
		testutil.AssertNoError(t, err, "Settlement initiation should succeed")
		testutil.AssertEqual(t, TradeStatusSettlementInitiated, trade.Status, "Status should be settlement initiated")
		testutil.AssertEqual(t, "escrow-123", *trade.EscrowAccountID, "Escrow account should be set")
	})

	// Test payment recording
	t.Run("record payment", func(t *testing.T) {
		err := trade.ReceivePayment(5000.00, "USD", "wire_transfer", "txn-456")
		
		testutil.AssertNoError(t, err, "Payment recording should succeed")
		testutil.AssertEqual(t, TradeStatusPaymentReceived, trade.Status, "Status should be payment received")
		testutil.AssertNotNil(t, trade.PaymentInfo, "Payment info should be set")
		testutil.AssertEqual(t, 5000.00, trade.PaymentInfo.Amount, "Payment amount should match")
	})

	// Test share transfer
	t.Run("record share transfer", func(t *testing.T) {
		err := trade.TransferShares(100, "seller-101", "buyer-789", "electronic", "cert-789")
		
		testutil.AssertNoError(t, err, "Share transfer should succeed")
		testutil.AssertEqual(t, TradeStatusSharesTransferred, trade.Status, "Status should be shares transferred")
		testutil.AssertNotNil(t, trade.TransferInfo, "Transfer info should be set")
		testutil.AssertEqual(t, int64(100), trade.TransferInfo.SharesCount, "Share count should match")
	})

	// Test settlement completion
	t.Run("complete settlement", func(t *testing.T) {
		err := trade.SettleTrade(5000.00, 25.00, 15.00, "automated")
		
		testutil.AssertNoError(t, err, "Settlement completion should succeed")
		testutil.AssertEqual(t, TradeStatusSettled, trade.Status, "Status should be settled")
		testutil.AssertEqual(t, 25.00, trade.Fees, "Fees should be set")
		testutil.AssertEqual(t, 15.00, trade.Taxes, "Taxes should be set")
		testutil.AssertTrue(t, trade.IsSettled(), "Trade should be marked as settled")
	})
}

func TestOrderMatchingEngine_Simple(t *testing.T) {
	// Arrange
	setup := testutil.NewTestSetup()
	engine := NewOrderMatchingEngine(setup.EventStore, setup.EventBus)

	// Create order book with test orders
	orderBook := NewOrderBook("TEST-001")
	
	// Add sell orders
	orderBook.AddSellOrder(&OrderBookEntry{
		ListingID:    "listing-1",
		UserID:       "seller-1",
		SecurityID:   "TEST-001",
		OrderType:    "sell",
		Quantity:     100,
		Price:        float64Ptr(50.00),
		Timestamp:    testutil.TestTime,
		IsAccredited: true,
	})

	// Add buy orders
	orderBook.AddBuyOrder(&OrderBookEntry{
		BidID:        stringPtr("bid-1"),
		UserID:       "buyer-1",
		SecurityID:   "TEST-001",
		OrderType:    "buy",
		Quantity:     150,
		Price:        float64Ptr(52.00),
		Timestamp:    testutil.TestTime,
		IsAccredited: true,
	})

	// Act
	matches, err := engine.matchPriceTimePriority(orderBook)

	// Assert
	testutil.AssertNoError(t, err, "Matching should succeed")
	testutil.AssertLengthEqual(t, 1, matches, "Should have one match")
	
	match := matches[0]
	testutil.AssertEqual(t, "seller-1", match.SellerID, "Should match with seller")
	testutil.AssertEqual(t, "buyer-1", match.BuyerID, "Should match with buyer")
	testutil.AssertEqual(t, int64(100), match.SharesTraded, "Should trade 100 shares")
	testutil.AssertEqual(t, 50.00, match.TradePrice, "Should use seller's price")
}

// Benchmark tests
func BenchmarkTradeMatching_Simple(b *testing.B) {
	testutil.BenchmarkFunction(b, func() {
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
	})
}