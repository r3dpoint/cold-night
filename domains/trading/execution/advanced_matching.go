package execution

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// AdvancedMatchingEngine extends the basic matching engine with sophisticated algorithms
type AdvancedMatchingEngine struct {
	*OrderMatchingEngine
	marketDataProvider MarketDataProvider
	riskEngine         RiskEngine
}

// MarketDataProvider interface for getting market data
type MarketDataProvider interface {
	GetLastTradePrice(securityID string) (float64, error)
	GetMarketHours() (open, close time.Time, isOpen bool)
	GetVolatility(securityID string, period time.Duration) (float64, error)
	GetReferencePrice(securityID string) (float64, error)
}

// RiskEngine interface for risk assessment
type RiskEngine interface {
	AssessTradeRisk(trade *MatchResult) (*RiskAssessment, error)
	CheckPositionLimits(userID, securityID string, quantity int64) error
	ValidateCounterparty(buyerID, sellerID string) error
}

// RiskAssessment represents trade risk evaluation
type RiskAssessment struct {
	RiskLevel      string  `json:"riskLevel"`      // low, medium, high, extreme
	RiskScore      float64 `json:"riskScore"`      // 0-100
	RequiresReview bool    `json:"requiresReview"`
	RiskFactors    []string `json:"riskFactors"`
	MaxAllowedSize int64   `json:"maxAllowedSize"`
}

// NewAdvancedMatchingEngine creates an enhanced matching engine
func NewAdvancedMatchingEngine(
	basic *OrderMatchingEngine,
	marketData MarketDataProvider,
	riskEngine RiskEngine,
) *AdvancedMatchingEngine {
	return &AdvancedMatchingEngine{
		OrderMatchingEngine: basic,
		marketDataProvider:  marketData,
		riskEngine:         riskEngine,
	}
}

// MatchOrdersWithRisk performs order matching with risk assessment
func (e *AdvancedMatchingEngine) MatchOrdersWithRisk(securityID string, algorithm MatchingAlgorithm) ([]*MatchResult, []*RiskAssessment, error) {
	// Get basic matches
	matches, err := e.MatchOrders(securityID, algorithm)
	if err != nil {
		return nil, nil, err
	}

	var riskAssessments []*RiskAssessment
	var validMatches []*MatchResult

	// Assess risk for each match
	for _, match := range matches {
		assessment, err := e.riskEngine.AssessTradeRisk(match)
		if err != nil {
			// Log error but continue
			fmt.Printf("Risk assessment failed for trade %s: %v\n", match.TradeID, err)
			continue
		}

		riskAssessments = append(riskAssessments, assessment)

		// Filter out high-risk trades if needed
		if assessment.RiskLevel != "extreme" {
			validMatches = append(validMatches, match)
		}
	}

	return validMatches, riskAssessments, nil
}

// MatchWithProRata implements pro-rata allocation for uniform price auctions
func (e *AdvancedMatchingEngine) MatchWithProRata(orderBook *OrderBook, clearingPrice float64) ([]*MatchResult, error) {
	sellOrders := orderBook.GetSellOrders()
	buyOrders := orderBook.GetBuyOrders()

	// Filter orders that would execute at clearing price
	var eligibleSells []*OrderBookEntry
	var eligibleBuys []*OrderBookEntry

	for _, order := range sellOrders {
		if order.Price == nil || *order.Price <= clearingPrice {
			eligibleSells = append(eligibleSells, order)
		}
	}

	for _, order := range buyOrders {
		if order.Price == nil || *order.Price >= clearingPrice {
			eligibleBuys = append(eligibleBuys, order)
		}
	}

	// Calculate total supply and demand
	var totalSupply, totalDemand int64
	for _, order := range eligibleSells {
		totalSupply += order.Quantity
	}
	for _, order := range eligibleBuys {
		totalDemand += order.Quantity
	}

	if totalSupply == 0 || totalDemand == 0 {
		return []*MatchResult{}, nil
	}

	// Determine tradeable quantity
	tradeableQuantity := min(totalSupply, totalDemand)
	
	var matches []*MatchResult

	// If supply exceeds demand, allocate proportionally among sellers
	if totalSupply > totalDemand {
		// Allocate full demand, pro-rata among sellers
		for _, sellOrder := range eligibleSells {
			allocation := float64(sellOrder.Quantity) / float64(totalSupply) * float64(tradeableQuantity)
			allocatedShares := int64(math.Floor(allocation))

			if allocatedShares > 0 {
				// Find matching buy orders
				buyMatches := e.allocateToBuyers(eligibleBuys, allocatedShares)
				for _, buyMatch := range buyMatches {
					match := &MatchResult{
						TradeID:           e.generateTradeID(),
						ListingID:         sellOrder.ListingID,
						BidID:             buyMatch.BidID,
						BuyerID:           buyMatch.UserID,
						SellerID:          sellOrder.UserID,
						SecurityID:        orderBook.SecurityID,
						SharesTraded:      buyMatch.Quantity,
						TradePrice:        clearingPrice,
						TotalAmount:       float64(buyMatch.Quantity) * clearingPrice,
						SettlementDate:    e.calculateSettlementDate(),
						MatchingAlgorithm: string(UniformPriceAuction),
					}
					matches = append(matches, match)
				}
			}
		}
	} else {
		// Allocate full supply, pro-rata among buyers
		for _, buyOrder := range eligibleBuys {
			allocation := float64(buyOrder.Quantity) / float64(totalDemand) * float64(tradeableQuantity)
			allocatedShares := int64(math.Floor(allocation))

			if allocatedShares > 0 {
				// Find matching sell orders
				sellMatches := e.allocateToSellers(eligibleSells, allocatedShares)
				for _, sellMatch := range sellMatches {
					match := &MatchResult{
						TradeID:           e.generateTradeID(),
						ListingID:         sellMatch.ListingID,
						BidID:             buyOrder.BidID,
						BuyerID:           buyOrder.UserID,
						SellerID:          sellMatch.UserID,
						SecurityID:        orderBook.SecurityID,
						SharesTraded:      sellMatch.Quantity,
						TradePrice:        clearingPrice,
						TotalAmount:       float64(sellMatch.Quantity) * clearingPrice,
						SettlementDate:    e.calculateSettlementDate(),
						MatchingAlgorithm: string(UniformPriceAuction),
					}
					matches = append(matches, match)
				}
			}
		}
	}

	return matches, nil
}

// MatchWithTimeWeightedPriority implements time-weighted priority matching
func (e *AdvancedMatchingEngine) MatchWithTimeWeightedPriority(orderBook *OrderBook) ([]*MatchResult, error) {
	sellOrders := orderBook.GetSellOrders()
	buyOrders := orderBook.GetBuyOrders()

	// Sort with time weighting - older orders get better treatment
	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price == nil && sellOrders[j].Price == nil {
			return sellOrders[i].Timestamp.Before(sellOrders[j].Timestamp)
		}
		if sellOrders[i].Price == nil {
			return true
		}
		if sellOrders[j].Price == nil {
			return false
		}

		priceDiff := *sellOrders[i].Price - *sellOrders[j].Price
		if math.Abs(priceDiff) < 0.01 { // Same price level
			// Apply time weighting
			timeWeight := e.calculateTimeWeight(sellOrders[i].Timestamp)
			return timeWeight > e.calculateTimeWeight(sellOrders[j].Timestamp)
		}
		return *sellOrders[i].Price < *sellOrders[j].Price
	})

	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price == nil && buyOrders[j].Price == nil {
			return buyOrders[i].Timestamp.Before(buyOrders[j].Timestamp)
		}
		if buyOrders[i].Price == nil {
			return true
		}
		if buyOrders[j].Price == nil {
			return false
		}

		priceDiff := *buyOrders[i].Price - *buyOrders[j].Price
		if math.Abs(priceDiff) < 0.01 { // Same price level
			timeWeight := e.calculateTimeWeight(buyOrders[i].Timestamp)
			return timeWeight > e.calculateTimeWeight(buyOrders[j].Timestamp)
		}
		return *buyOrders[i].Price > *buyOrders[j].Price
	})

	// Use standard matching logic with the time-weighted sorted orders
	return e.matchPriceTimePriority(orderBook)
}

// MatchBulkOrders handles large orders with special treatment
func (e *AdvancedMatchingEngine) MatchBulkOrders(orderBook *OrderBook, minBulkSize int64) ([]*MatchResult, error) {
	var bulkMatches []*MatchResult
	var regularMatches []*MatchResult

	// Separate bulk orders
	sellOrders := orderBook.GetSellOrders()
	buyOrders := orderBook.GetBuyOrders()

	var bulkSells, regularSells []*OrderBookEntry
	var bulkBuys, regularBuys []*OrderBookEntry

	for _, order := range sellOrders {
		if order.Quantity >= minBulkSize {
			bulkSells = append(bulkSells, order)
		} else {
			regularSells = append(regularSells, order)
		}
	}

	for _, order := range buyOrders {
		if order.Quantity >= minBulkSize {
			bulkBuys = append(bulkBuys, order)
		} else {
			regularBuys = append(regularBuys, order)
		}
	}

	// Match bulk orders first with preferred pricing
	bulkOrderBook := NewOrderBook(orderBook.SecurityID)
	for _, order := range bulkSells {
		bulkOrderBook.AddSellOrder(order)
	}
	for _, order := range bulkBuys {
		bulkOrderBook.AddBuyOrder(order)
	}

	if len(bulkSells) > 0 && len(bulkBuys) > 0 {
		matches, err := e.matchBulkWithDiscounts(bulkOrderBook)
		if err != nil {
			return nil, err
		}
		bulkMatches = append(bulkMatches, matches...)
	}

	// Match remaining regular orders
	regularOrderBook := NewOrderBook(orderBook.SecurityID)
	for _, order := range regularSells {
		regularOrderBook.AddSellOrder(order)
	}
	for _, order := range regularBuys {
		regularOrderBook.AddBuyOrder(order)
	}

	if len(regularSells) > 0 && len(regularBuys) > 0 {
		matches, err := e.matchPriceTimePriority(regularOrderBook)
		if err != nil {
			return nil, err
		}
		regularMatches = append(regularMatches, matches...)
	}

	// Combine results
	allMatches := append(bulkMatches, regularMatches...)
	return allMatches, nil
}

// Enhanced negotiated trading with AI-assisted pricing
func (e *AdvancedMatchingEngine) MatchNegotiatedAdvanced(orderBook *OrderBook) ([]*MatchResult, error) {
	sellOrders := orderBook.GetSellOrders()
	buyOrders := orderBook.GetBuyOrders()

	var matches []*MatchResult

	// For each sell order, find the best compatible buy orders
	for _, sellOrder := range sellOrders {
		// Get reference price for fair value assessment
		referencePrice, err := e.marketDataProvider.GetReferencePrice(sellOrder.SecurityID)
		if err != nil {
			referencePrice = 100.0 // Fallback
		}

		// Find compatible buy orders
		compatibleBuys := e.findCompatibleOrders(sellOrder, buyOrders, referencePrice)
		
		for _, buyOrder := range compatibleBuys {
			// Calculate negotiated price based on multiple factors
			negotiatedPrice := e.calculateNegotiatedPrice(sellOrder, buyOrder, referencePrice)
			
			// Check if both parties would accept this price
			if e.wouldAcceptPrice(sellOrder, negotiatedPrice) && e.wouldAcceptPrice(buyOrder, negotiatedPrice) {
				quantity := min(sellOrder.Quantity, buyOrder.Quantity)
				
				match := &MatchResult{
					TradeID:           e.generateTradeID(),
					ListingID:         sellOrder.ListingID,
					BidID:             buyOrder.BidID,
					BuyerID:           buyOrder.UserID,
					SellerID:          sellOrder.UserID,
					SecurityID:        orderBook.SecurityID,
					SharesTraded:      quantity,
					TradePrice:        negotiatedPrice,
					TotalAmount:       float64(quantity) * negotiatedPrice,
					SettlementDate:    e.calculateSettlementDate(),
					MatchingAlgorithm: string(NegotiatedTrading),
				}
				
				matches = append(matches, match)
				
				// Update order quantities
				sellOrder.Quantity -= quantity
				buyOrder.Quantity -= quantity
				
				if sellOrder.Quantity == 0 {
					break
				}
			}
		}
	}

	return matches, nil
}

// Helper methods for advanced matching

func (e *AdvancedMatchingEngine) calculateTimeWeight(timestamp time.Time) float64 {
	// Orders older than 1 hour get increasing weight
	hoursSinceOrder := time.Since(timestamp).Hours()
	if hoursSinceOrder < 1 {
		return 1.0
	}
	// Logarithmic time weight up to a maximum of 2x
	return math.Min(2.0, 1.0+math.Log10(hoursSinceOrder))
}

func (e *AdvancedMatchingEngine) allocateToBuyers(buyers []*OrderBookEntry, totalShares int64) []*OrderBookEntry {
	// Simple FIFO allocation for now - could be enhanced with pro-rata
	var allocations []*OrderBookEntry
	remaining := totalShares

	for _, buyer := range buyers {
		if remaining <= 0 {
			break
		}
		
		allocation := min(buyer.Quantity, remaining)
		if allocation > 0 {
			buyerCopy := *buyer
			buyerCopy.Quantity = allocation
			allocations = append(allocations, &buyerCopy)
			remaining -= allocation
		}
	}

	return allocations
}

func (e *AdvancedMatchingEngine) allocateToSellers(sellers []*OrderBookEntry, totalShares int64) []*OrderBookEntry {
	// Simple FIFO allocation for now - could be enhanced with pro-rata
	var allocations []*OrderBookEntry
	remaining := totalShares

	for _, seller := range sellers {
		if remaining <= 0 {
			break
		}
		
		allocation := min(seller.Quantity, remaining)
		if allocation > 0 {
			sellerCopy := *seller
			sellerCopy.Quantity = allocation
			allocations = append(allocations, &sellerCopy)
			remaining -= allocation
		}
	}

	return allocations
}

func (e *AdvancedMatchingEngine) matchBulkWithDiscounts(orderBook *OrderBook) ([]*MatchResult, error) {
	// Apply bulk discounts to pricing
	matches, err := e.matchPriceTimePriority(orderBook)
	if err != nil {
		return nil, err
	}

	// Apply 0.5% discount to bulk trades
	for _, match := range matches {
		if match.SharesTraded >= 10000 { // Bulk threshold
			discount := 0.005 // 0.5%
			match.TradePrice *= (1 - discount)
			match.TotalAmount = float64(match.SharesTraded) * match.TradePrice
		}
	}

	return matches, nil
}

func (e *AdvancedMatchingEngine) findCompatibleOrders(order *OrderBookEntry, candidates []*OrderBookEntry, referencePrice float64) []*OrderBookEntry {
	var compatible []*OrderBookEntry
	
	for _, candidate := range candidates {
		// Basic compatibility checks
		if !e.canMatch(order, candidate) {
			continue
		}
		
		// Additional checks for negotiated trading
		priceRange := referencePrice * 0.1 // 10% range
		
		if order.Price != nil && candidate.Price != nil {
			// Both have prices - check if they're within reasonable range
			if math.Abs(*order.Price-*candidate.Price) <= priceRange {
				compatible = append(compatible, candidate)
			}
		} else {
			// At least one is a market order - consider compatible
			compatible = append(compatible, candidate)
		}
	}
	
	return compatible
}

func (e *AdvancedMatchingEngine) calculateNegotiatedPrice(sellOrder, buyOrder *OrderBookEntry, referencePrice float64) float64 {
	// Sophisticated price calculation considering multiple factors
	
	var sellPrice, buyPrice float64
	
	if sellOrder.Price != nil {
		sellPrice = *sellOrder.Price
	} else {
		sellPrice = referencePrice * 1.02 // 2% above reference for market sells
	}
	
	if buyOrder.Price != nil {
		buyPrice = *buyOrder.Price
	} else {
		buyPrice = referencePrice * 0.98 // 2% below reference for market buys
	}
	
	// Weight based on order sizes
	sellWeight := float64(sellOrder.Quantity)
	buyWeight := float64(buyOrder.Quantity)
	totalWeight := sellWeight + buyWeight
	
	// Weighted average with slight bias toward the larger order
	weightedPrice := (sellPrice*sellWeight + buyPrice*buyWeight) / totalWeight
	
	// Apply time pressure factor (older orders get slight price preference)
	sellAge := time.Since(sellOrder.Timestamp).Hours()
	buyAge := time.Since(buyOrder.Timestamp).Hours()
	
	if sellAge > buyAge {
		// Favor seller slightly
		weightedPrice = weightedPrice * 1.001
	} else if buyAge > sellAge {
		// Favor buyer slightly
		weightedPrice = weightedPrice * 0.999
	}
	
	return weightedPrice
}

func (e *AdvancedMatchingEngine) wouldAcceptPrice(order *OrderBookEntry, price float64) bool {
	if order.Price == nil {
		// Market orders accept any reasonable price
		return true
	}
	
	if order.OrderType == "sell" {
		// Seller accepts if price >= their ask
		return price >= *order.Price
	} else {
		// Buyer accepts if price <= their bid
		return price <= *order.Price
	}
}