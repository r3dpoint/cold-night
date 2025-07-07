package execution

import (
	"fmt"
	"sort"
	"time"

	"securities-marketplace/domains/shared/events"
)

// MatchingAlgorithm represents different matching algorithms
type MatchingAlgorithm string

const (
	PriceTimePriority  MatchingAlgorithm = "price_time_priority"
	UniformPriceAuction MatchingAlgorithm = "uniform_price_auction"
	NegotiatedTrading  MatchingAlgorithm = "negotiated_trading"
)

// OrderBookEntry represents an entry in the order book
type OrderBookEntry struct {
	ListingID     string
	BidID         *string
	UserID        string
	SecurityID    string
	OrderType     string // "sell" or "buy"
	Quantity      int64
	Price         *float64 // nil for market orders
	Timestamp     time.Time
	IsAccredited  bool
	ExpiresAt     *time.Time
}

// MatchResult represents the result of a matching operation
type MatchResult struct {
	TradeID          string
	ListingID        string
	BidID            *string
	BuyerID          string
	SellerID         string
	SecurityID       string
	SharesTraded     int64
	TradePrice       float64
	TotalAmount      float64
	SettlementDate   time.Time
	MatchingAlgorithm string
}

// OrderMatchingEngine handles order matching for securities trading
type OrderMatchingEngine struct {
	eventStore events.EventStore
	eventBus   events.EventBus
}

// NewOrderMatchingEngine creates a new order matching engine
func NewOrderMatchingEngine(eventStore events.EventStore, eventBus events.EventBus) *OrderMatchingEngine {
	return &OrderMatchingEngine{
		eventStore: eventStore,
		eventBus:   eventBus,
	}
}

// MatchOrders attempts to match buy and sell orders for a security
func (e *OrderMatchingEngine) MatchOrders(securityID string, algorithm MatchingAlgorithm) ([]*MatchResult, error) {
	// Get current order book for the security
	orderBook, err := e.buildOrderBook(securityID)
	if err != nil {
		return nil, fmt.Errorf("failed to build order book: %w", err)
	}

	// Apply matching algorithm
	var matches []*MatchResult
	switch algorithm {
	case PriceTimePriority:
		matches, err = e.matchPriceTimePriority(orderBook)
	case UniformPriceAuction:
		matches, err = e.matchUniformPriceAuction(orderBook)
	case NegotiatedTrading:
		matches, err = e.matchNegotiated(orderBook)
	default:
		return nil, fmt.Errorf("unknown matching algorithm: %s", algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("matching failed: %w", err)
	}

	return matches, nil
}

// MatchSpecificOrders attempts to match a specific bid against a specific listing
func (e *OrderMatchingEngine) MatchSpecificOrders(listingID, bidID string) (*MatchResult, error) {
	// This would be used for direct negotiations or manual matching
	// Implementation would load the specific listing and bid, validate compatibility,
	// and create a match result if valid
	
	// For now, return a placeholder implementation
	return nil, fmt.Errorf("specific order matching not yet implemented")
}

// buildOrderBook constructs the current order book for a security
func (e *OrderMatchingEngine) buildOrderBook(securityID string) (*OrderBook, error) {
	orderBook := NewOrderBook(securityID)

	// Get all active listings for the security
	listings, err := e.getActiveListings(securityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active listings: %w", err)
	}

	// Add sell orders from listings
	for _, listing := range listings {
		entry := &OrderBookEntry{
			ListingID:    listing.ListingID,
			UserID:       listing.SellerID,
			SecurityID:   securityID,
			OrderType:    "sell",
			Quantity:     listing.SharesRemaining,
			Price:        listing.CurrentPrice,
			Timestamp:    listing.CreatedAt,
			IsAccredited: true, // Assume sellers are accredited
			ExpiresAt:    listing.ExpiresAt,
		}
		orderBook.AddSellOrder(entry)
	}

	// Get all active bids for the security
	bids, err := e.getActiveBids(securityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active bids: %w", err)
	}

	// Add buy orders from bids
	for _, bid := range bids {
		entry := &OrderBookEntry{
			BidID:        &bid.BidID,
			UserID:       bid.BidderID,
			SecurityID:   securityID,
			OrderType:    "buy",
			Quantity:     bid.SharesRemaining,
			Price:        &bid.BidPrice,
			Timestamp:    bid.PlacedAt,
			IsAccredited: bid.IsAccredited,
			ExpiresAt:    bid.ExpiresAt,
		}
		orderBook.AddBuyOrder(entry)
	}

	return orderBook, nil
}

// matchPriceTimePriority implements price-time priority matching
func (e *OrderMatchingEngine) matchPriceTimePriority(orderBook *OrderBook) ([]*MatchResult, error) {
	var matches []*MatchResult

	// Sort sell orders by price (ascending), then by time (ascending)
	sellOrders := orderBook.GetSellOrders()
	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price == nil && sellOrders[j].Price == nil {
			return sellOrders[i].Timestamp.Before(sellOrders[j].Timestamp)
		}
		if sellOrders[i].Price == nil {
			return true // Market orders have priority
		}
		if sellOrders[j].Price == nil {
			return false
		}
		if *sellOrders[i].Price == *sellOrders[j].Price {
			return sellOrders[i].Timestamp.Before(sellOrders[j].Timestamp)
		}
		return *sellOrders[i].Price < *sellOrders[j].Price
	})

	// Sort buy orders by price (descending), then by time (ascending)
	buyOrders := orderBook.GetBuyOrders()
	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price == nil && buyOrders[j].Price == nil {
			return buyOrders[i].Timestamp.Before(buyOrders[j].Timestamp)
		}
		if buyOrders[i].Price == nil {
			return true // Market orders have priority
		}
		if buyOrders[j].Price == nil {
			return false
		}
		if *buyOrders[i].Price == *buyOrders[j].Price {
			return buyOrders[i].Timestamp.Before(buyOrders[j].Timestamp)
		}
		return *buyOrders[i].Price > *buyOrders[j].Price
	})

	// Match orders
	sellIndex := 0
	buyIndex := 0

	for sellIndex < len(sellOrders) && buyIndex < len(buyOrders) {
		sellOrder := sellOrders[sellIndex]
		buyOrder := buyOrders[buyIndex]

		// Check if orders can be matched
		if !e.canMatch(sellOrder, buyOrder) {
			// Try next buy order
			buyIndex++
			if buyIndex >= len(buyOrders) {
				// No more buy orders, try next sell order
				sellIndex++
				buyIndex = 0
			}
			continue
		}

		// Determine trade price (seller's price takes precedence in price-time priority)
		var tradePrice float64
		if sellOrder.Price != nil {
			tradePrice = *sellOrder.Price
		} else if buyOrder.Price != nil {
			tradePrice = *buyOrder.Price
		} else {
			// Both are market orders - need a reference price
			// In practice, this would use last trade price or opening price
			tradePrice = 100.0 // Placeholder
		}

		// Determine quantity
		quantity := min(sellOrder.Quantity, buyOrder.Quantity)

		// Create match result
		match := &MatchResult{
			TradeID:          e.generateTradeID(),
			ListingID:        sellOrder.ListingID,
			BidID:            buyOrder.BidID,
			BuyerID:          buyOrder.UserID,
			SellerID:         sellOrder.UserID,
			SecurityID:       orderBook.SecurityID,
			SharesTraded:     quantity,
			TradePrice:       tradePrice,
			TotalAmount:      float64(quantity) * tradePrice,
			SettlementDate:   e.calculateSettlementDate(),
			MatchingAlgorithm: string(PriceTimePriority),
		}

		matches = append(matches, match)

		// Update order quantities
		sellOrder.Quantity -= quantity
		buyOrder.Quantity -= quantity

		// Remove fully filled orders
		if sellOrder.Quantity == 0 {
			sellIndex++
		}
		if buyOrder.Quantity == 0 {
			buyIndex++
		}
	}

	return matches, nil
}

// matchUniformPriceAuction implements uniform price auction matching
func (e *OrderMatchingEngine) matchUniformPriceAuction(orderBook *OrderBook) ([]*MatchResult, error) {
	// Simplified implementation - in practice this would be more complex
	// This algorithm finds a single clearing price where supply meets demand

	sellOrders := orderBook.GetSellOrders()
	buyOrders := orderBook.GetBuyOrders()

	// Sort orders
	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price == nil {
			return true
		}
		if sellOrders[j].Price == nil {
			return false
		}
		return *sellOrders[i].Price < *sellOrders[j].Price
	})

	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price == nil {
			return true
		}
		if buyOrders[j].Price == nil {
			return false
		}
		return *buyOrders[i].Price > *buyOrders[j].Price
	})

	// Find clearing price
	clearingPrice, err := e.findClearingPrice(sellOrders, buyOrders)
	if err != nil {
		return nil, fmt.Errorf("no clearing price found: %w", err)
	}

	// Create matches at clearing price
	var matches []*MatchResult
	sellIndex := 0
	buyIndex := 0

	for sellIndex < len(sellOrders) && buyIndex < len(buyOrders) {
		sellOrder := sellOrders[sellIndex]
		buyOrder := buyOrders[buyIndex]

		// Only match orders that would execute at clearing price
		if sellOrder.Price != nil && *sellOrder.Price > clearingPrice {
			break
		}
		if buyOrder.Price != nil && *buyOrder.Price < clearingPrice {
			break
		}

		if !e.canMatch(sellOrder, buyOrder) {
			buyIndex++
			continue
		}

		quantity := min(sellOrder.Quantity, buyOrder.Quantity)

		match := &MatchResult{
			TradeID:          e.generateTradeID(),
			ListingID:        sellOrder.ListingID,
			BidID:            buyOrder.BidID,
			BuyerID:          buyOrder.UserID,
			SellerID:         sellOrder.UserID,
			SecurityID:       orderBook.SecurityID,
			SharesTraded:     quantity,
			TradePrice:       clearingPrice,
			TotalAmount:      float64(quantity) * clearingPrice,
			SettlementDate:   e.calculateSettlementDate(),
			MatchingAlgorithm: string(UniformPriceAuction),
		}

		matches = append(matches, match)

		sellOrder.Quantity -= quantity
		buyOrder.Quantity -= quantity

		if sellOrder.Quantity == 0 {
			sellIndex++
		}
		if buyOrder.Quantity == 0 {
			buyIndex++
		}
	}

	return matches, nil
}

// matchNegotiated implements negotiated trading matching
func (e *OrderMatchingEngine) matchNegotiated(orderBook *OrderBook) ([]*MatchResult, error) {
	// Negotiated trading allows for more flexible matching rules
	// This is a simplified implementation
	return []*MatchResult{}, nil
}

// Helper methods

func (e *OrderMatchingEngine) canMatch(sellOrder, buyOrder *OrderBookEntry) bool {
	// Check accreditation requirements
	if sellOrder.IsAccredited && !buyOrder.IsAccredited {
		// This would check if the listing requires accredited buyers
		// For now, assume all trades require accreditation
		return false
	}

	// Check price compatibility
	if sellOrder.Price != nil && buyOrder.Price != nil {
		return *buyOrder.Price >= *sellOrder.Price
	}

	// Market orders can always match
	return true
}

func (e *OrderMatchingEngine) findClearingPrice(sellOrders, buyOrders []*OrderBookEntry) (float64, error) {
	// Simplified clearing price calculation
	// In practice, this would analyze supply and demand curves
	
	if len(sellOrders) == 0 || len(buyOrders) == 0 {
		return 0, fmt.Errorf("insufficient orders")
	}

	// Use midpoint of best bid and ask as clearing price
	var bestAsk, bestBid float64
	
	if sellOrders[0].Price != nil {
		bestAsk = *sellOrders[0].Price
	} else {
		bestAsk = 100.0 // Default price for market orders
	}
	
	if buyOrders[0].Price != nil {
		bestBid = *buyOrders[0].Price
	} else {
		bestBid = 100.0 // Default price for market orders
	}

	if bestBid >= bestAsk {
		return (bestBid + bestAsk) / 2, nil
	}

	return 0, fmt.Errorf("no overlap between bid and ask")
}

func (e *OrderMatchingEngine) generateTradeID() string {
	// Generate unique trade ID
	return fmt.Sprintf("trade_%d", time.Now().UnixNano())
}

func (e *OrderMatchingEngine) calculateSettlementDate() time.Time {
	// T+2 settlement (trade date + 2 business days)
	return time.Now().AddDate(0, 0, 2)
}

func (e *OrderMatchingEngine) getActiveListings(securityID string) ([]ListingSummary, error) {
	// This would query the listing repository/projection
	// For now, return empty slice
	return []ListingSummary{}, nil
}

func (e *OrderMatchingEngine) getActiveBids(securityID string) ([]BidSummary, error) {
	// This would query the bid repository/projection
	// For now, return empty slice
	return []BidSummary{}, nil
}

// Helper types for order matching

type ListingSummary struct {
	ListingID       string
	SellerID        string
	SharesRemaining int64
	CurrentPrice    *float64
	CreatedAt       time.Time
	ExpiresAt       *time.Time
}

type BidSummary struct {
	BidID           string
	BidderID        string
	SharesRemaining int64
	BidPrice        float64
	PlacedAt        time.Time
	ExpiresAt       *time.Time
	IsAccredited    bool
}

// OrderBook represents the current order book for a security
type OrderBook struct {
	SecurityID string
	BuyOrders  []*OrderBookEntry
	SellOrders []*OrderBookEntry
}

func NewOrderBook(securityID string) *OrderBook {
	return &OrderBook{
		SecurityID: securityID,
		BuyOrders:  make([]*OrderBookEntry, 0),
		SellOrders: make([]*OrderBookEntry, 0),
	}
}

func (ob *OrderBook) AddBuyOrder(order *OrderBookEntry) {
	ob.BuyOrders = append(ob.BuyOrders, order)
}

func (ob *OrderBook) AddSellOrder(order *OrderBookEntry) {
	ob.SellOrders = append(ob.SellOrders, order)
}

func (ob *OrderBook) GetBuyOrders() []*OrderBookEntry {
	return ob.BuyOrders
}

func (ob *OrderBook) GetSellOrders() []*OrderBookEntry {
	return ob.SellOrders
}

// Utility function
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}