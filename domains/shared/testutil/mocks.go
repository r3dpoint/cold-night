package testutil

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// MockMarketDataProvider provides mock market data for testing
type MockMarketDataProvider struct {
	LastTradePrices map[string]float64
	Volatilities    map[string]float64
	ReferencePrices map[string]float64
	IsMarketOpen    bool
	MarketOpenTime  time.Time
	MarketCloseTime time.Time
}

func NewMockMarketDataProvider() *MockMarketDataProvider {
	return &MockMarketDataProvider{
		LastTradePrices: make(map[string]float64),
		Volatilities:    make(map[string]float64),
		ReferencePrices: make(map[string]float64),
		IsMarketOpen:    true,
		MarketOpenTime:  time.Now().Add(-4 * time.Hour),
		MarketCloseTime: time.Now().Add(4 * time.Hour),
	}
}

func (m *MockMarketDataProvider) GetLastTradePrice(securityID string) (float64, error) {
	if price, ok := m.LastTradePrices[securityID]; ok {
		return price, nil
	}
	return 100.0, nil // Default price
}

func (m *MockMarketDataProvider) GetMarketHours() (open, close time.Time, isOpen bool) {
	return m.MarketOpenTime, m.MarketCloseTime, m.IsMarketOpen
}

func (m *MockMarketDataProvider) GetVolatility(securityID string, period time.Duration) (float64, error) {
	if volatility, ok := m.Volatilities[securityID]; ok {
		return volatility, nil
	}
	return 0.15, nil // Default 15% volatility
}

func (m *MockMarketDataProvider) GetReferencePrice(securityID string) (float64, error) {
	if price, ok := m.ReferencePrices[securityID]; ok {
		return price, nil
	}
	return 100.0, nil // Default reference price
}

func (m *MockMarketDataProvider) SetLastTradePrice(securityID string, price float64) {
	m.LastTradePrices[securityID] = price
}

func (m *MockMarketDataProvider) SetVolatility(securityID string, volatility float64) {
	m.Volatilities[securityID] = volatility
}

func (m *MockMarketDataProvider) SetReferencePrice(securityID string, price float64) {
	m.ReferencePrices[securityID] = price
}

// Generic interfaces to avoid import cycles

// MatchResult represents a trade match result
type MatchResult struct {
	TradeID      string
	BuyerID      string
	SellerID     string
	SecurityID   string
	SharesTraded int64
	TradePrice   float64
	TotalAmount  float64
}

// RiskAssessment represents trade risk evaluation
type RiskAssessment struct {
	RiskLevel      string
	RiskScore      float64
	RequiresReview bool
	RiskFactors    []string
	MaxAllowedSize int64
}

// MockRiskEngine provides mock risk assessment for testing
type MockRiskEngine struct {
	RiskAssessments map[string]*RiskAssessment
	PositionLimits  map[string]int64
	BlockedPairs    map[string]bool
}

func NewMockRiskEngine() *MockRiskEngine {
	return &MockRiskEngine{
		RiskAssessments: make(map[string]*RiskAssessment),
		PositionLimits:  make(map[string]int64),
		BlockedPairs:    make(map[string]bool),
	}
}

func (m *MockRiskEngine) AssessTradeRisk(trade *MatchResult) (*RiskAssessment, error) {
	key := fmt.Sprintf("%s-%s", trade.BuyerID, trade.SellerID)
	if assessment, ok := m.RiskAssessments[key]; ok {
		return assessment, nil
	}
	
	// Default low-risk assessment
	return &RiskAssessment{
		RiskLevel:      "low",
		RiskScore:      15.0,
		RequiresReview: false,
		RiskFactors:    []string{},
		MaxAllowedSize: trade.SharesTraded * 10, // Allow 10x the current trade size
	}, nil
}

func (m *MockRiskEngine) CheckPositionLimits(userID, securityID string, quantity int64) error {
	key := fmt.Sprintf("%s-%s", userID, securityID)
	if limit, ok := m.PositionLimits[key]; ok {
		if quantity > limit {
			return fmt.Errorf("position limit exceeded: %d > %d", quantity, limit)
		}
	}
	// Default: allow up to 100,000 shares
	if quantity > 100000 {
		return fmt.Errorf("position limit exceeded: %d > 100000", quantity)
	}
	return nil
}

func (m *MockRiskEngine) ValidateCounterparty(buyerID, sellerID string) error {
	key := fmt.Sprintf("%s-%s", buyerID, sellerID)
	if blocked, ok := m.BlockedPairs[key]; ok && blocked {
		return fmt.Errorf("counterparty validation failed: %s and %s are blocked", buyerID, sellerID)
	}
	return nil
}

func (m *MockRiskEngine) SetRiskAssessment(buyerID, sellerID string, assessment *RiskAssessment) {
	key := fmt.Sprintf("%s-%s", buyerID, sellerID)
	m.RiskAssessments[key] = assessment
}

func (m *MockRiskEngine) SetPositionLimit(userID, securityID string, limit int64) {
	key := fmt.Sprintf("%s-%s", userID, securityID)
	m.PositionLimits[key] = limit
}

func (m *MockRiskEngine) BlockCounterparty(buyerID, sellerID string) {
	key := fmt.Sprintf("%s-%s", buyerID, sellerID)
	m.BlockedPairs[key] = true
}

// Simplified mock repository interface for testing
type MockRepository struct {
	data map[string]interface{}
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		data: make(map[string]interface{}),
	}
}

func (m *MockRepository) Store(key string, value interface{}) {
	m.data[key] = value
}

func (m *MockRepository) Get(key string) (interface{}, bool) {
	value, exists := m.data[key]
	return value, exists
}

func (m *MockRepository) Delete(key string) {
	delete(m.data, key)
}

func (m *MockRepository) Clear() {
	m.data = make(map[string]interface{})
}

// MockHTTPClient provides mock HTTP responses for testing
type MockHTTPClient struct {
	Responses map[string]MockHTTPResponse
}

type MockHTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Error      error
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Responses: make(map[string]MockHTTPResponse),
	}
}

func (m *MockHTTPClient) SetResponse(url string, response MockHTTPResponse) {
	m.Responses[url] = response
}

// Additional test doubles

// SpyEventBus records method calls for verification
type SpyEventBus struct {
	*TestEventBus
	PublishCalls    []events.DomainEvent
	SubscribeCalls  []string
}

func NewSpyEventBus() *SpyEventBus {
	return &SpyEventBus{
		TestEventBus:   NewTestEventBus(),
		PublishCalls:   make([]events.DomainEvent, 0),
		SubscribeCalls: make([]string, 0),
	}
}

func (s *SpyEventBus) Publish(event events.DomainEvent) error {
	s.PublishCalls = append(s.PublishCalls, event)
	return s.TestEventBus.Publish(event)
}

func (s *SpyEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	s.SubscribeCalls = append(s.SubscribeCalls, eventType)
	return s.TestEventBus.Subscribe(eventType, handler)
}

func (s *SpyEventBus) GetPublishCallCount() int {
	return len(s.PublishCalls)
}

func (s *SpyEventBus) GetSubscribeCallCount() int {
	return len(s.SubscribeCalls)
}

func (s *SpyEventBus) WasEventPublished(eventType string) bool {
	for _, evt := range s.PublishCalls {
		if evt.GetEventType() == eventType {
			return true
		}
	}
	return false
}

// StubMarketDataProvider provides predefined responses
type StubMarketDataProvider struct {
	LastTradePrice  float64
	Volatility      float64
	ReferencePrice  float64
	IsOpen          bool
	ShouldError     bool
	ErrorMessage    string
}

func NewStubMarketDataProvider() *StubMarketDataProvider {
	return &StubMarketDataProvider{
		LastTradePrice: 100.0,
		Volatility:     0.15,
		ReferencePrice: 100.0,
		IsOpen:         true,
		ShouldError:    false,
	}
}

func (s *StubMarketDataProvider) GetLastTradePrice(securityID string) (float64, error) {
	if s.ShouldError {
		return 0, fmt.Errorf("%s", s.ErrorMessage)
	}
	return s.LastTradePrice, nil
}

func (s *StubMarketDataProvider) GetMarketHours() (open, close time.Time, isOpen bool) {
	return time.Now().Add(-4*time.Hour), time.Now().Add(4*time.Hour), s.IsOpen
}

func (s *StubMarketDataProvider) GetVolatility(securityID string, period time.Duration) (float64, error) {
	if s.ShouldError {
		return 0, fmt.Errorf("%s", s.ErrorMessage)
	}
	return s.Volatility, nil
}

func (s *StubMarketDataProvider) GetReferencePrice(securityID string) (float64, error) {
	if s.ShouldError {
		return 0, fmt.Errorf("%s", s.ErrorMessage)
	}
	return s.ReferencePrice, nil
}