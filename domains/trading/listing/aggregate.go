package listing

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// ListingType represents the type of listing
type ListingType string

const (
	ListingTypeFixed    ListingType = "fixed"     // Fixed price listing
	ListingTypeAuction  ListingType = "auction"   // Auction-style listing
	ListingTypeMarket   ListingType = "market"    // Market order (immediate execution)
	ListingTypeLimit    ListingType = "limit"     // Limit order (price threshold)
)

// ListingStatus represents the current status of a listing
type ListingStatus string

const (
	ListingStatusActive    ListingStatus = "active"
	ListingStatusCancelled ListingStatus = "cancelled"
	ListingStatusExpired   ListingStatus = "expired"
	ListingStatusCompleted ListingStatus = "completed"
	ListingStatusSuspended ListingStatus = "suspended"
)

// RestrictionType represents trading restrictions
type RestrictionType string

const (
	RestrictionNone        RestrictionType = "none"
	RestrictionAccredited  RestrictionType = "accredited"
	RestrictionInstitutional RestrictionType = "institutional"
	RestrictionGeographic  RestrictionType = "geographic"
)

// ListingAggregate represents a security listing for sale
type ListingAggregate struct {
	events.AggregateRoot
	
	// Basic listing information
	SecurityID      string          `json:"securityId"`
	SellerID        string          `json:"sellerId"`
	SharesOffered   int64           `json:"sharesOffered"`
	SharesRemaining int64           `json:"sharesRemaining"`
	ListingType     ListingType     `json:"listingType"`
	
	// Pricing information
	MinimumPrice    *float64        `json:"minimumPrice,omitempty"`
	ReservePrice    *float64        `json:"reservePrice,omitempty"`
	CurrentPrice    *float64        `json:"currentPrice,omitempty"`
	
	// Restrictions and requirements
	RestrictionType *RestrictionType `json:"restrictionType,omitempty"`
	AccreditedOnly  bool            `json:"accreditedOnly"`
	
	// Status and lifecycle
	Status          ListingStatus   `json:"status"`
	CreatedAt       time.Time       `json:"createdAt"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	CancelledAt     *time.Time      `json:"cancelledAt,omitempty"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
	
	// Trading history
	TotalSharesSold int64           `json:"totalSharesSold"`
	TradeIDs        []string        `json:"tradeIds"`
	
	// Cancellation details
	CancellationReason string       `json:"cancellationReason,omitempty"`
	CancelledBy        string       `json:"cancelledBy,omitempty"`
}

// NewListingAggregate creates a new listing aggregate
func NewListingAggregate(listingID string) *ListingAggregate {
	return &ListingAggregate{
		AggregateRoot:   events.NewAggregateRoot(listingID, "Listing"),
		Status:          ListingStatusActive,
		TradeIDs:        make([]string, 0),
		TotalSharesSold: 0,
	}
}

// CreateListing creates a new listing
func (l *ListingAggregate) CreateListing(securityID, sellerID string, sharesOffered int64, listingType ListingType, minimumPrice, reservePrice, currentPrice *float64, restrictionType *RestrictionType, accreditedOnly bool, expiresAt *time.Time) error {
	if l.Version > 0 {
		return fmt.Errorf("listing already exists")
	}

	// Validate basic requirements
	if sharesOffered <= 0 {
		return fmt.Errorf("shares offered must be greater than zero")
	}

	// Validate pricing based on listing type
	if err := l.validatePricing(listingType, minimumPrice, reservePrice, currentPrice); err != nil {
		return fmt.Errorf("invalid pricing: %w", err)
	}

	// Validate expiration
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration date cannot be in the past")
	}

	var restrictionTypeStr *string
	if restrictionType != nil {
		str := string(*restrictionType)
		restrictionTypeStr = &str
	}

	event := NewListingCreated(l.ID, securityID, sellerID, sharesOffered, string(listingType), minimumPrice, reservePrice, currentPrice, restrictionTypeStr, accreditedOnly, expiresAt)
	l.AddEvent(event)
	return l.ApplyEvent(event)
}

// UpdatePrice updates the listing price
func (l *ListingAggregate) UpdatePrice(newPrice float64, updatedBy, reason string) error {
	if l.Status != ListingStatusActive {
		return fmt.Errorf("cannot update price of non-active listing")
	}

	if l.ListingType == ListingTypeFixed {
		return fmt.Errorf("cannot update price of fixed-price listing")
	}

	if newPrice <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}

	// Validate against minimum price if set
	if l.MinimumPrice != nil && newPrice < *l.MinimumPrice {
		return fmt.Errorf("price cannot be below minimum price of %.2f", *l.MinimumPrice)
	}

	event := NewListingPriceUpdated(l.ID, l.CurrentPrice, newPrice, updatedBy, reason)
	l.AddEvent(event)
	return l.ApplyEvent(event)
}

// ReduceShares reduces the number of available shares after a partial sale
func (l *ListingAggregate) ReduceShares(sharesSold int64, tradeID, buyerID string, salePrice float64) error {
	if l.Status != ListingStatusActive {
		return fmt.Errorf("cannot reduce shares of non-active listing")
	}

	if sharesSold <= 0 {
		return fmt.Errorf("shares sold must be greater than zero")
	}

	if sharesSold > l.SharesRemaining {
		return fmt.Errorf("cannot sell more shares than remaining (%d > %d)", sharesSold, l.SharesRemaining)
	}

	newSharesRemaining := l.SharesRemaining - sharesSold

	event := NewListingSharesReduced(l.ID, sharesSold, newSharesRemaining, tradeID, buyerID, salePrice)
	l.AddEvent(event)
	
	// If all shares are sold, mark as completed
	if newSharesRemaining == 0 {
		completedEvent := NewListingCompleted(l.ID, l.TotalSharesSold+sharesSold, salePrice, time.Now())
		l.AddEvent(completedEvent)
		if err := l.ApplyEvent(event); err != nil {
			return err
		}
		return l.ApplyEvent(completedEvent)
	}

	return l.ApplyEvent(event)
}

// Cancel cancels the listing
func (l *ListingAggregate) Cancel(reason, cancelledBy string) error {
	if l.Status != ListingStatusActive {
		return fmt.Errorf("can only cancel active listings")
	}

	if reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}

	event := NewListingCancelled(l.ID, reason, cancelledBy)
	l.AddEvent(event)
	return l.ApplyEvent(event)
}

// Expire marks the listing as expired
func (l *ListingAggregate) Expire() error {
	if l.Status != ListingStatusActive {
		return fmt.Errorf("can only expire active listings")
	}

	if l.ExpiresAt == nil {
		return fmt.Errorf("listing has no expiration date")
	}

	if time.Now().Before(*l.ExpiresAt) {
		return fmt.Errorf("listing has not yet expired")
	}

	event := NewListingExpired(l.ID, time.Now())
	l.AddEvent(event)
	return l.ApplyEvent(event)
}

// Reactivate reactivates a cancelled listing
func (l *ListingAggregate) Reactivate(reactivatedBy, reason string) error {
	if l.Status != ListingStatusCancelled {
		return fmt.Errorf("can only reactivate cancelled listings")
	}

	if l.SharesRemaining <= 0 {
		return fmt.Errorf("cannot reactivate listing with no remaining shares")
	}

	// Check if expired
	if l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt) {
		return fmt.Errorf("cannot reactivate expired listing")
	}

	event := NewListingReactivated(l.ID, reactivatedBy, reason)
	l.AddEvent(event)
	return l.ApplyEvent(event)
}

// ApplyEvent applies an event to the aggregate
func (l *ListingAggregate) ApplyEvent(event events.DomainEvent) error {
	switch e := event.(type) {
	case *ListingCreated:
		return l.applyListingCreated(e)
	case *ListingPriceUpdated:
		return l.applyListingPriceUpdated(e)
	case *ListingSharesReduced:
		return l.applyListingSharesReduced(e)
	case *ListingCancelled:
		return l.applyListingCancelled(e)
	case *ListingExpired:
		return l.applyListingExpired(e)
	case *ListingCompleted:
		return l.applyListingCompleted(e)
	case *ListingReactivated:
		return l.applyListingReactivated(e)
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// LoadFromHistory loads the aggregate from a sequence of events
func (l *ListingAggregate) LoadFromHistory(events []events.DomainEvent) error {
	for _, event := range events {
		if err := l.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.GetEventType(), err)
		}
		l.IncrementVersion()
	}
	return nil
}

// Event application methods

func (l *ListingAggregate) applyListingCreated(event *ListingCreated) error {
	l.SecurityID = event.SecurityID
	l.SellerID = event.SellerID
	l.SharesOffered = event.SharesOffered
	l.SharesRemaining = event.SharesOffered
	l.ListingType = ListingType(event.ListingType)
	l.MinimumPrice = event.MinimumPrice
	l.ReservePrice = event.ReservePrice
	l.CurrentPrice = event.CurrentPrice
	
	if event.RestrictionType != nil {
		restrictionType := RestrictionType(*event.RestrictionType)
		l.RestrictionType = &restrictionType
	}
	
	l.AccreditedOnly = event.AccreditedOnly
	l.Status = ListingStatusActive
	l.CreatedAt = event.Timestamp
	l.ExpiresAt = event.ExpiresAt
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingPriceUpdated(event *ListingPriceUpdated) error {
	l.CurrentPrice = &event.NewPrice
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingSharesReduced(event *ListingSharesReduced) error {
	l.SharesRemaining = event.SharesRemaining
	l.TotalSharesSold += event.SharesSold
	l.TradeIDs = append(l.TradeIDs, event.TradeID)
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingCancelled(event *ListingCancelled) error {
	l.Status = ListingStatusCancelled
	l.CancellationReason = event.Reason
	l.CancelledBy = event.CancelledBy
	l.CancelledAt = &event.Timestamp
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingExpired(event *ListingExpired) error {
	l.Status = ListingStatusExpired
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingCompleted(event *ListingCompleted) error {
	l.Status = ListingStatusCompleted
	l.CompletedAt = &event.CompletedAt
	
	l.IncrementVersion()
	return nil
}

func (l *ListingAggregate) applyListingReactivated(event *ListingReactivated) error {
	l.Status = ListingStatusActive
	l.CancellationReason = ""
	l.CancelledBy = ""
	l.CancelledAt = nil
	
	l.IncrementVersion()
	return nil
}

// Helper methods

// IsActive returns true if the listing is active and can accept trades
func (l *ListingAggregate) IsActive() bool {
	if l.Status != ListingStatusActive {
		return false
	}
	
	if l.SharesRemaining <= 0 {
		return false
	}
	
	if l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt) {
		return false
	}
	
	return true
}

// IsExpired returns true if the listing has expired
func (l *ListingAggregate) IsExpired() bool {
	return l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt)
}

// GetCurrentPrice returns the current asking price
func (l *ListingAggregate) GetCurrentPrice() *float64 {
	return l.CurrentPrice
}

// GetSharesAvailable returns the number of shares available for purchase
func (l *ListingAggregate) GetSharesAvailable() int64 {
	if !l.IsActive() {
		return 0
	}
	return l.SharesRemaining
}

// CanAcceptBid returns true if the listing can accept a bid from the given user
func (l *ListingAggregate) CanAcceptBid(bidderID string, isAccredited bool) bool {
	if !l.IsActive() {
		return false
	}
	
	// Seller cannot bid on their own listing
	if bidderID == l.SellerID {
		return false
	}
	
	// Check accreditation requirement
	if l.AccreditedOnly && !isAccredited {
		return false
	}
	
	return true
}

// GetFillPercentage returns the percentage of shares that have been sold
func (l *ListingAggregate) GetFillPercentage() float64 {
	if l.SharesOffered == 0 {
		return 0
	}
	return float64(l.TotalSharesSold) / float64(l.SharesOffered) * 100
}

// GetTimeRemaining returns the time remaining until expiration
func (l *ListingAggregate) GetTimeRemaining() *time.Duration {
	if l.ExpiresAt == nil {
		return nil
	}
	
	remaining := time.Until(*l.ExpiresAt)
	if remaining < 0 {
		remaining = 0
	}
	
	return &remaining
}

// validatePricing validates pricing based on listing type
func (l *ListingAggregate) validatePricing(listingType ListingType, minimumPrice, reservePrice, currentPrice *float64) error {
	switch listingType {
	case ListingTypeFixed:
		if currentPrice == nil {
			return fmt.Errorf("fixed price listing requires current price")
		}
		if *currentPrice <= 0 {
			return fmt.Errorf("current price must be greater than zero")
		}
		
	case ListingTypeAuction:
		if minimumPrice != nil && *minimumPrice <= 0 {
			return fmt.Errorf("minimum price must be greater than zero")
		}
		if reservePrice != nil && *reservePrice <= 0 {
			return fmt.Errorf("reserve price must be greater than zero")
		}
		if minimumPrice != nil && reservePrice != nil && *minimumPrice > *reservePrice {
			return fmt.Errorf("minimum price cannot be greater than reserve price")
		}
		
	case ListingTypeLimit:
		if currentPrice == nil {
			return fmt.Errorf("limit listing requires current price")
		}
		if *currentPrice <= 0 {
			return fmt.Errorf("current price must be greater than zero")
		}
		
	case ListingTypeMarket:
		// Market orders don't require specific pricing
		
	default:
		return fmt.Errorf("unknown listing type: %s", listingType)
	}
	
	return nil
}