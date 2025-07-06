package bidding

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// BidType represents the type of bid
type BidType string

const (
	BidTypeMarket     BidType = "market"     // Market bid (accept current price)
	BidTypeLimit      BidType = "limit"      // Limit bid (maximum price)
	BidTypeStop       BidType = "stop"       // Stop bid (trigger price)
	BidTypeStopLimit  BidType = "stop_limit" // Stop-limit bid
	BidTypeAllOrNone  BidType = "all_or_none" // All-or-none bid
)

// BidStatus represents the current status of a bid
type BidStatus string

const (
	BidStatusActive         BidStatus = "active"
	BidStatusPartiallyFilled BidStatus = "partially_filled"
	BidStatusFilled         BidStatus = "filled"
	BidStatusWithdrawn      BidStatus = "withdrawn"
	BidStatusExpired        BidStatus = "expired"
	BidStatusRejected       BidStatus = "rejected"
)

// FillRecord represents a partial fill of the bid
type FillRecord struct {
	TradeID      string    `json:"tradeId"`
	SharesFilled int64     `json:"sharesFilled"`
	FillPrice    float64   `json:"fillPrice"`
	SellerID     string    `json:"sellerId"`
	FilledAt     time.Time `json:"filledAt"`
}

// BidAggregate represents a bid to purchase securities
type BidAggregate struct {
	events.AggregateRoot
	
	// Basic bid information
	ListingID       string    `json:"listingId"`
	BidderID        string    `json:"bidderId"`
	SharesRequested int64     `json:"sharesRequested"`
	SharesRemaining int64     `json:"sharesRemaining"`
	BidPrice        float64   `json:"bidPrice"`
	BidType         BidType   `json:"bidType"`
	
	// Status and lifecycle
	Status    BidStatus `json:"status"`
	PlacedAt  time.Time `json:"placedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	
	// Fill tracking
	SharesFilled    int64        `json:"sharesFilled"`
	AverageFillPrice float64     `json:"averageFillPrice"`
	FillRecords     []FillRecord `json:"fillRecords"`
	
	// Completion details
	FilledAt    *time.Time `json:"filledAt,omitempty"`
	WithdrawnAt *time.Time `json:"withdrawnAt,omitempty"`
	ExpiredAt   *time.Time `json:"expiredAt,omitempty"`
	RejectedAt  *time.Time `json:"rejectedAt,omitempty"`
	
	// Withdrawal/rejection details
	WithdrawalReason string `json:"withdrawalReason,omitempty"`
	WithdrawnBy      string `json:"withdrawnBy,omitempty"`
	RejectionReason  string `json:"rejectionReason,omitempty"`
	RejectedBy       string `json:"rejectedBy,omitempty"`
}

// NewBidAggregate creates a new bid aggregate
func NewBidAggregate(bidID string) *BidAggregate {
	return &BidAggregate{
		AggregateRoot: events.NewAggregateRoot(bidID, "Bid"),
		Status:        BidStatusActive,
		FillRecords:   make([]FillRecord, 0),
		SharesFilled:  0,
	}
}

// PlaceBid places a new bid
func (b *BidAggregate) PlaceBid(listingID, bidderID string, sharesRequested int64, bidPrice float64, bidType BidType, expiresAt *time.Time) error {
	if b.Version > 0 {
		return fmt.Errorf("bid already exists")
	}

	// Validate basic requirements
	if sharesRequested <= 0 {
		return fmt.Errorf("shares requested must be greater than zero")
	}

	if bidPrice <= 0 {
		return fmt.Errorf("bid price must be greater than zero")
	}

	// Validate expiration
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration date cannot be in the past")
	}

	event := NewBidPlaced(b.ID, listingID, bidderID, sharesRequested, bidPrice, string(bidType), expiresAt)
	b.AddEvent(event)
	return b.ApplyEvent(event)
}

// ModifyBid modifies an existing bid
func (b *BidAggregate) ModifyBid(newSharesRequested int64, newBidPrice float64, modifiedBy, reason string) error {
	if b.Status != BidStatusActive && b.Status != BidStatusPartiallyFilled {
		return fmt.Errorf("can only modify active or partially filled bids")
	}

	if newSharesRequested <= 0 {
		return fmt.Errorf("shares requested must be greater than zero")
	}

	if newBidPrice <= 0 {
		return fmt.Errorf("bid price must be greater than zero")
	}

	// Cannot reduce shares below what has already been filled
	if newSharesRequested < b.SharesFilled {
		return fmt.Errorf("cannot reduce shares below filled amount (%d < %d)", newSharesRequested, b.SharesFilled)
	}

	event := NewBidModified(b.ID, b.SharesRequested, newSharesRequested, b.BidPrice, newBidPrice, modifiedBy, reason)
	b.AddEvent(event)
	return b.ApplyEvent(event)
}

// PartiallyFill partially fills the bid
func (b *BidAggregate) PartiallyFill(sharesFilled int64, fillPrice float64, tradeID, sellerID string) error {
	if b.Status != BidStatusActive && b.Status != BidStatusPartiallyFilled {
		return fmt.Errorf("can only fill active or partially filled bids")
	}

	if sharesFilled <= 0 {
		return fmt.Errorf("shares filled must be greater than zero")
	}

	if sharesFilled > b.SharesRemaining {
		return fmt.Errorf("cannot fill more shares than remaining (%d > %d)", sharesFilled, b.SharesRemaining)
	}

	if fillPrice <= 0 {
		return fmt.Errorf("fill price must be greater than zero")
	}

	// For limit bids, check price constraint
	if b.BidType == BidTypeLimit && fillPrice > b.BidPrice {
		return fmt.Errorf("fill price %.2f exceeds bid limit of %.2f", fillPrice, b.BidPrice)
	}

	newSharesRemaining := b.SharesRemaining - sharesFilled

	event := NewBidPartiallyFilled(b.ID, sharesFilled, newSharesRemaining, fillPrice, tradeID, sellerID)
	b.AddEvent(event)
	
	// If all shares are filled, mark as filled
	if newSharesRemaining == 0 {
		totalFilled := b.SharesFilled + sharesFilled
		filledEvent := NewBidFilled(b.ID, totalFilled, fillPrice, time.Now(), tradeID, sellerID)
		b.AddEvent(filledEvent)
		if err := b.ApplyEvent(event); err != nil {
			return err
		}
		return b.ApplyEvent(filledEvent)
	}

	return b.ApplyEvent(event)
}

// Withdraw withdraws the bid
func (b *BidAggregate) Withdraw(reason, withdrawnBy string) error {
	if b.Status != BidStatusActive && b.Status != BidStatusPartiallyFilled {
		return fmt.Errorf("can only withdraw active or partially filled bids")
	}

	if reason == "" {
		return fmt.Errorf("withdrawal reason is required")
	}

	event := NewBidWithdrawn(b.ID, reason, withdrawnBy)
	b.AddEvent(event)
	return b.ApplyEvent(event)
}

// Expire marks the bid as expired
func (b *BidAggregate) Expire() error {
	if b.Status != BidStatusActive && b.Status != BidStatusPartiallyFilled {
		return fmt.Errorf("can only expire active or partially filled bids")
	}

	if b.ExpiresAt == nil {
		return fmt.Errorf("bid has no expiration date")
	}

	if time.Now().Before(*b.ExpiresAt) {
		return fmt.Errorf("bid has not yet expired")
	}

	event := NewBidExpired(b.ID, time.Now())
	b.AddEvent(event)
	return b.ApplyEvent(event)
}

// Reject rejects the bid
func (b *BidAggregate) Reject(reason, rejectedBy string) error {
	if b.Status != BidStatusActive {
		return fmt.Errorf("can only reject active bids")
	}

	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	event := NewBidRejected(b.ID, reason, rejectedBy)
	b.AddEvent(event)
	return b.ApplyEvent(event)
}

// ApplyEvent applies an event to the aggregate
func (b *BidAggregate) ApplyEvent(event events.DomainEvent) error {
	switch e := event.(type) {
	case *BidPlaced:
		return b.applyBidPlaced(e)
	case *BidModified:
		return b.applyBidModified(e)
	case *BidPartiallyFilled:
		return b.applyBidPartiallyFilled(e)
	case *BidFilled:
		return b.applyBidFilled(e)
	case *BidWithdrawn:
		return b.applyBidWithdrawn(e)
	case *BidExpired:
		return b.applyBidExpired(e)
	case *BidRejected:
		return b.applyBidRejected(e)
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// LoadFromHistory loads the aggregate from a sequence of events
func (b *BidAggregate) LoadFromHistory(events []events.DomainEvent) error {
	for _, event := range events {
		if err := b.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.GetEventType(), err)
		}
		b.IncrementVersion()
	}
	return nil
}

// Event application methods

func (b *BidAggregate) applyBidPlaced(event *BidPlaced) error {
	b.ListingID = event.ListingID
	b.BidderID = event.BidderID
	b.SharesRequested = event.SharesRequested
	b.SharesRemaining = event.SharesRequested
	b.BidPrice = event.BidPrice
	b.BidType = BidType(event.BidType)
	b.Status = BidStatusActive
	b.PlacedAt = event.Timestamp
	b.ExpiresAt = event.ExpiresAt
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidModified(event *BidModified) error {
	b.SharesRequested = event.NewSharesRequested
	b.SharesRemaining = event.NewSharesRequested - b.SharesFilled
	b.BidPrice = event.NewBidPrice
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidPartiallyFilled(event *BidPartiallyFilled) error {
	b.SharesRemaining = event.SharesRemaining
	b.SharesFilled += event.SharesFilled
	
	// Update average fill price
	totalValue := b.AverageFillPrice * float64(b.SharesFilled-event.SharesFilled)
	totalValue += event.FillPrice * float64(event.SharesFilled)
	b.AverageFillPrice = totalValue / float64(b.SharesFilled)
	
	// Add fill record
	fillRecord := FillRecord{
		TradeID:      event.TradeID,
		SharesFilled: event.SharesFilled,
		FillPrice:    event.FillPrice,
		SellerID:     event.SellerID,
		FilledAt:     event.Timestamp,
	}
	b.FillRecords = append(b.FillRecords, fillRecord)
	
	b.Status = BidStatusPartiallyFilled
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidFilled(event *BidFilled) error {
	b.Status = BidStatusFilled
	b.FilledAt = &event.FilledAt
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidWithdrawn(event *BidWithdrawn) error {
	b.Status = BidStatusWithdrawn
	b.WithdrawalReason = event.Reason
	b.WithdrawnBy = event.WithdrawnBy
	b.WithdrawnAt = &event.Timestamp
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidExpired(event *BidExpired) error {
	b.Status = BidStatusExpired
	b.ExpiredAt = &event.ExpiredAt
	
	b.IncrementVersion()
	return nil
}

func (b *BidAggregate) applyBidRejected(event *BidRejected) error {
	b.Status = BidStatusRejected
	b.RejectionReason = event.Reason
	b.RejectedBy = event.RejectedBy
	b.RejectedAt = &event.Timestamp
	
	b.IncrementVersion()
	return nil
}

// Helper methods

// IsActive returns true if the bid is active and can be filled
func (b *BidAggregate) IsActive() bool {
	if b.Status != BidStatusActive && b.Status != BidStatusPartiallyFilled {
		return false
	}
	
	if b.SharesRemaining <= 0 {
		return false
	}
	
	if b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt) {
		return false
	}
	
	return true
}

// IsExpired returns true if the bid has expired
func (b *BidAggregate) IsExpired() bool {
	return b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt)
}

// GetFillPercentage returns the percentage of the bid that has been filled
func (b *BidAggregate) GetFillPercentage() float64 {
	if b.SharesRequested == 0 {
		return 0
	}
	return float64(b.SharesFilled) / float64(b.SharesRequested) * 100
}

// GetTimeRemaining returns the time remaining until expiration
func (b *BidAggregate) GetTimeRemaining() *time.Duration {
	if b.ExpiresAt == nil {
		return nil
	}
	
	remaining := time.Until(*b.ExpiresAt)
	if remaining < 0 {
		remaining = 0
	}
	
	return &remaining
}

// CanBeFilled returns true if the bid can accept fills
func (b *BidAggregate) CanBeFilled(fillPrice float64) bool {
	if !b.IsActive() {
		return false
	}
	
	// Check price constraints based on bid type
	switch b.BidType {
	case BidTypeLimit:
		return fillPrice <= b.BidPrice
	case BidTypeMarket:
		return true // Market bids accept any price
	default:
		return fillPrice <= b.BidPrice
	}
}

// GetMaxFillShares returns the maximum number of shares that can be filled
func (b *BidAggregate) GetMaxFillShares() int64 {
	if !b.IsActive() {
		return 0
	}
	return b.SharesRemaining
}

// GetTotalValue returns the total value of the bid at the bid price
func (b *BidAggregate) GetTotalValue() float64 {
	return float64(b.SharesRequested) * b.BidPrice
}

// GetRemainingValue returns the remaining value of unfilled shares
func (b *BidAggregate) GetRemainingValue() float64 {
	return float64(b.SharesRemaining) * b.BidPrice
}