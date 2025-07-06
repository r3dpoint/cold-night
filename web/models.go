package web

import (
	"time"
)

// User represents a user in the web context
type User struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	Username             string    `json:"username"`
	FullName             string    `json:"fullName"`
	Role                 string    `json:"role"`
	IsActive             bool      `json:"isActive"`
	IsEmailVerified      bool      `json:"isEmailVerified"`
	IsAccredited         bool      `json:"isAccredited"`
	IsCompliant          bool      `json:"isCompliant"`
	CanTrade             bool      `json:"canTrade"`
	MaxInvestmentLimit   float64   `json:"maxInvestmentLimit"`
	CurrentInvestment    float64   `json:"currentInvestment"`
	ComplianceLevel      string    `json:"complianceLevel"`
	RiskTolerance        string    `json:"riskTolerance"`
	CreatedAt            time.Time `json:"createdAt"`
	LastLoginAt          *time.Time `json:"lastLoginAt"`
	LoginCount           int       `json:"loginCount"`
}

// Security represents a security in the web context
type Security struct {
	ID                string    `json:"id"`
	Symbol            string    `json:"symbol"`
	CompanyName       string    `json:"companyName"`
	SecurityType      string    `json:"securityType"`
	Description       string    `json:"description"`
	Sector            string    `json:"sector"`
	Industry          string    `json:"industry"`
	CurrentOwner      string    `json:"currentOwner"`
	TotalShares       int64     `json:"totalShares"`
	OutstandingShares int64     `json:"outstandingShares"`
	ParValue          float64   `json:"parValue"`
	BookValue         float64   `json:"bookValue"`
	MarketCap         float64   `json:"marketCap"`
	IsTradeable       bool      `json:"isTradeable"`
	MinTradeAmount    float64   `json:"minTradeAmount"`
	AccreditedOnly    bool      `json:"accreditedOnly"`
	IsCompliant       bool      `json:"isCompliant"`
	RegulatoryStatus  string    `json:"regulatoryStatus"`
	DividendYield     float64   `json:"dividendYield"`
	LastDividendDate  *time.Time `json:"lastDividendDate"`
	NextDividendDate  *time.Time `json:"nextDividendDate"`
	CreatedAt         time.Time `json:"createdAt"`
	ListedAt          *time.Time `json:"listedAt"`
}

// Listing represents a listing in the web context
type Listing struct {
	ID                string     `json:"id"`
	SellerID          string     `json:"sellerId"`
	SellerName        string     `json:"sellerName"`
	SecurityID        string     `json:"securityId"`
	SecuritySymbol    string     `json:"securitySymbol"`
	SecurityName      string     `json:"securityName"`
	ListingType       string     `json:"listingType"`
	SharesOffered     int64      `json:"sharesOffered"`
	SharesRemaining   int64      `json:"sharesRemaining"`
	ListingPrice      *float64   `json:"listingPrice"`
	MinimumPrice      *float64   `json:"minimumPrice"`
	CurrentPrice      *float64   `json:"currentPrice"`
	PriceType         string     `json:"priceType"`
	Status            string     `json:"status"`
	IsActive          bool       `json:"isActive"`
	AccreditedOnly    bool       `json:"accreditedOnly"`
	MinimumInvestment *float64   `json:"minimumInvestment"`
	MaximumInvestment *float64   `json:"maximumInvestment"`
	BidCount          int        `json:"bidCount"`
	HighestBid        *float64   `json:"highestBid"`
	TotalBidVolume    int64      `json:"totalBidVolume"`
	ListedAt          time.Time  `json:"listedAt"`
	ExpiresAt         *time.Time `json:"expiresAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// Bid represents a bid in the web context
type Bid struct {
	ID               string     `json:"id"`
	BidderID         string     `json:"bidderId"`
	BidderName       string     `json:"bidderName"`
	ListingID        string     `json:"listingId"`
	SecurityID       string     `json:"securityId"`
	SecuritySymbol   string     `json:"securitySymbol"`
	SecurityName     string     `json:"securityName"`
	BidType          string     `json:"bidType"`
	SharesRequested  int64      `json:"sharesRequested"`
	SharesRemaining  int64      `json:"sharesRemaining"`
	BidPrice         float64    `json:"bidPrice"`
	TotalBidAmount   float64    `json:"totalBidAmount"`
	Status           string     `json:"status"`
	IsActive         bool       `json:"isActive"`
	IsAccredited     bool       `json:"isAccredited"`
	SharesFilled     int64      `json:"sharesFilled"`
	AmountFilled     float64    `json:"amountFilled"`
	AverageFillPrice *float64   `json:"averageFillPrice"`
	PlacedAt         time.Time  `json:"placedAt"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// Trade represents a trade in the web context
type Trade struct {
	ID                 string     `json:"id"`
	ListingID          *string    `json:"listingId"`
	BidID              *string    `json:"bidId"`
	BuyerID            string     `json:"buyerId"`
	BuyerName          string     `json:"buyerName"`
	SellerID           string     `json:"sellerId"`
	SellerName         string     `json:"sellerName"`
	SecurityID         string     `json:"securityId"`
	SecuritySymbol     string     `json:"securitySymbol"`
	SecurityName       string     `json:"securityName"`
	SharesTraded       int64      `json:"sharesTraded"`
	TradePrice         float64    `json:"tradePrice"`
	TotalAmount        float64    `json:"totalAmount"`
	Fees               float64    `json:"fees"`
	Taxes              float64    `json:"taxes"`
	NetAmount          float64    `json:"netAmount"`
	SettlementDate     time.Time  `json:"settlementDate"`
	EscrowAccountID    *string    `json:"escrowAccountId"`
	Status             string     `json:"status"`
	SettlementStage    string     `json:"settlementStage"`
	BuyerConfirmed     bool       `json:"buyerConfirmed"`
	SellerConfirmed    bool       `json:"sellerConfirmed"`
	MatchedAt          time.Time  `json:"matchedAt"`
	ConfirmedAt        *time.Time `json:"confirmedAt"`
	SettledAt          *time.Time `json:"settledAt"`
	FailedAt           *time.Time `json:"failedAt"`
	CancelledAt        *time.Time `json:"cancelledAt"`
	MatchingAlgorithm  string     `json:"matchingAlgorithm"`
	FailureReason      string     `json:"failureReason"`
	CancellationReason string     `json:"cancellationReason"`
	CancelledBy        string     `json:"cancelledBy"`
	ProgressPercent    float64    `json:"progressPercent"`
	DaysToSettlement   int        `json:"daysToSettlement"`
	IsOverdue          bool       `json:"isOverdue"`
}

// MarketStats represents market statistics
type MarketStats struct {
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
	Change       float64       `json:"change"`
	ChangePercent float64      `json:"changePercent"`
}

// PortfolioSummary represents a user's portfolio summary
type PortfolioSummary struct {
	UserID             string  `json:"userId"`
	TotalValue         float64 `json:"totalValue"`
	TotalCostBasis     float64 `json:"totalCostBasis"`
	TotalGainLoss      float64 `json:"totalGainLoss"`
	TotalGainPercent   float64 `json:"totalGainPercent"`
	DayGainLoss        float64 `json:"dayGainLoss"`
	DayGainPercent     float64 `json:"dayGainPercent"`
	PositionsCount     int     `json:"positionsCount"`
	CashBalance        float64 `json:"cashBalance"`
	AvailableToBuy     float64 `json:"availableToBuy"`
	PendingTrades      int     `json:"pendingTrades"`
	DividendsYTD       float64 `json:"dividendsYTD"`
}

// PortfolioPosition represents a single position in a portfolio
type PortfolioPosition struct {
	SecurityID           string    `json:"securityId"`
	SecuritySymbol       string    `json:"securitySymbol"`
	SecurityName         string    `json:"securityName"`
	SharesOwned          int64     `json:"sharesOwned"`
	AverageCostBasis     float64   `json:"averageCostBasis"`
	TotalCostBasis       float64   `json:"totalCostBasis"`
	CurrentPrice         float64   `json:"currentPrice"`
	CurrentMarketValue   float64   `json:"currentMarketValue"`
	UnrealizedGainLoss   float64   `json:"unrealizedGainLoss"`
	UnrealizedGainPercent float64  `json:"unrealizedGainPercent"`
	DayGainLoss          float64   `json:"dayGainLoss"`
	DayGainPercent       float64   `json:"dayGainPercent"`
	TotalPurchases       int64     `json:"totalPurchases"`
	TotalSales           int64     `json:"totalSales"`
	TotalDividends       float64   `json:"totalDividends"`
	FirstPurchaseDate    *time.Time `json:"firstPurchaseDate"`
	LastTransactionDate  *time.Time `json:"lastTransactionDate"`
}

// MarketSummary represents overall market summary
type MarketSummary struct {
	TotalSecurities    int     `json:"totalSecurities"`
	TradableSecurities int     `json:"tradableSecurities"`
	ActiveListings     int     `json:"activeListings"`
	ActiveBids         int     `json:"activeBids"`
	TodayTrades        int     `json:"todayTrades"`
	TodayVolume        int64   `json:"todayVolume"`
	TodayValue         float64 `json:"todayValue"`
	ActiveUsers        int     `json:"activeUsers"`
	TotalMarketCap     float64 `json:"totalMarketCap"`
}

// Notification represents a user notification
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"` // "trade", "listing", "bid", "compliance", "system"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"isRead"`
	Priority  string    `json:"priority"` // "low", "medium", "high", "urgent"
	ActionURL *string   `json:"actionUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID               string            `json:"id"`
	EventType        string            `json:"eventType"`
	EventCategory    string            `json:"eventCategory"`
	EventDescription string            `json:"eventDescription"`
	UserID           *string           `json:"userId"`
	Username         *string           `json:"username"`
	UserRole         *string           `json:"userRole"`
	IPAddress        *string           `json:"ipAddress"`
	ResourceType     *string           `json:"resourceType"`
	ResourceID       *string           `json:"resourceId"`
	OldValues        map[string]interface{} `json:"oldValues"`
	NewValues        map[string]interface{} `json:"newValues"`
	Status           string            `json:"status"`
	ErrorMessage     *string           `json:"errorMessage"`
	RiskScore        *int              `json:"riskScore"`
	ComplianceFlags  []string          `json:"complianceFlags"`
	EventTimestamp   time.Time         `json:"eventTimestamp"`
	Metadata         map[string]interface{} `json:"metadata"`
}