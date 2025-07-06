package bidding

import (
	"encoding/json"
	"time"

	"securities-marketplace/domains/shared/events"
)

// Bidding Domain Events

// BidPlaced event is emitted when a new bid is placed
type BidPlaced struct {
	events.BaseEvent
	ListingID       string  `json:"listingId"`
	BidderID        string  `json:"bidderId"`
	SharesRequested int64   `json:"sharesRequested"`
	BidPrice        float64 `json:"bidPrice"`
	BidType         string  `json:"bidType"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}

// NewBidPlaced creates a new BidPlaced event
func NewBidPlaced(bidID, listingID, bidderID string, sharesRequested int64, bidPrice float64, bidType string, expiresAt *time.Time) *BidPlaced {
	return &BidPlaced{
		BaseEvent:       events.NewBaseEvent(bidID, "Bid"),
		ListingID:       listingID,
		BidderID:        bidderID,
		SharesRequested: sharesRequested,
		BidPrice:        bidPrice,
		BidType:         bidType,
		ExpiresAt:       expiresAt,
	}
}

func (e *BidPlaced) GetEventType() string     { return "BidPlaced" }
func (e *BidPlaced) GetAggregateID() string   { return e.AggregateID }
func (e *BidPlaced) GetAggregateType() string { return e.AggregateType }

func (e *BidPlaced) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidPlaced) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidModified event is emitted when a bid is modified
type BidModified struct {
	events.BaseEvent
	OldSharesRequested int64   `json:"oldSharesRequested"`
	NewSharesRequested int64   `json:"newSharesRequested"`
	OldBidPrice        float64 `json:"oldBidPrice"`
	NewBidPrice        float64 `json:"newBidPrice"`
	ModifiedBy         string  `json:"modifiedBy"`
	Reason             string  `json:"reason"`
}

func NewBidModified(bidID string, oldSharesRequested, newSharesRequested int64, oldBidPrice, newBidPrice float64, modifiedBy, reason string) *BidModified {
	return &BidModified{
		BaseEvent:          events.NewBaseEvent(bidID, "Bid"),
		OldSharesRequested: oldSharesRequested,
		NewSharesRequested: newSharesRequested,
		OldBidPrice:        oldBidPrice,
		NewBidPrice:        newBidPrice,
		ModifiedBy:         modifiedBy,
		Reason:             reason,
	}
}

func (e *BidModified) GetEventType() string     { return "BidModified" }
func (e *BidModified) GetAggregateID() string   { return e.AggregateID }
func (e *BidModified) GetAggregateType() string { return e.AggregateType }

func (e *BidModified) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidModified) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidPartiallyFilled event is emitted when a bid is partially filled
type BidPartiallyFilled struct {
	events.BaseEvent
	SharesFilled     int64   `json:"sharesFilled"`
	SharesRemaining  int64   `json:"sharesRemaining"`
	FillPrice        float64 `json:"fillPrice"`
	TradeID          string  `json:"tradeId"`
	SellerID         string  `json:"sellerId"`
}

func NewBidPartiallyFilled(bidID string, sharesFilled, sharesRemaining int64, fillPrice float64, tradeID, sellerID string) *BidPartiallyFilled {
	return &BidPartiallyFilled{
		BaseEvent:       events.NewBaseEvent(bidID, "Bid"),
		SharesFilled:    sharesFilled,
		SharesRemaining: sharesRemaining,
		FillPrice:       fillPrice,
		TradeID:         tradeID,
		SellerID:        sellerID,
	}
}

func (e *BidPartiallyFilled) GetEventType() string     { return "BidPartiallyFilled" }
func (e *BidPartiallyFilled) GetAggregateID() string   { return e.AggregateID }
func (e *BidPartiallyFilled) GetAggregateType() string { return e.AggregateType }

func (e *BidPartiallyFilled) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidPartiallyFilled) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidFilled event is emitted when a bid is completely filled
type BidFilled struct {
	events.BaseEvent
	TotalSharesFilled int64     `json:"totalSharesFilled"`
	FinalFillPrice    float64   `json:"finalFillPrice"`
	FilledAt          time.Time `json:"filledAt"`
	TradeID           string    `json:"tradeId"`
	SellerID          string    `json:"sellerId"`
}

func NewBidFilled(bidID string, totalSharesFilled int64, finalFillPrice float64, filledAt time.Time, tradeID, sellerID string) *BidFilled {
	return &BidFilled{
		BaseEvent:         events.NewBaseEvent(bidID, "Bid"),
		TotalSharesFilled: totalSharesFilled,
		FinalFillPrice:    finalFillPrice,
		FilledAt:          filledAt,
		TradeID:           tradeID,
		SellerID:          sellerID,
	}
}

func (e *BidFilled) GetEventType() string     { return "BidFilled" }
func (e *BidFilled) GetAggregateID() string   { return e.AggregateID }
func (e *BidFilled) GetAggregateType() string { return e.AggregateType }

func (e *BidFilled) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidFilled) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidWithdrawn event is emitted when a bid is withdrawn
type BidWithdrawn struct {
	events.BaseEvent
	Reason      string `json:"reason"`
	WithdrawnBy string `json:"withdrawnBy"`
}

func NewBidWithdrawn(bidID, reason, withdrawnBy string) *BidWithdrawn {
	return &BidWithdrawn{
		BaseEvent:   events.NewBaseEvent(bidID, "Bid"),
		Reason:      reason,
		WithdrawnBy: withdrawnBy,
	}
}

func (e *BidWithdrawn) GetEventType() string     { return "BidWithdrawn" }
func (e *BidWithdrawn) GetAggregateID() string   { return e.AggregateID }
func (e *BidWithdrawn) GetAggregateType() string { return e.AggregateType }

func (e *BidWithdrawn) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidWithdrawn) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidExpired event is emitted when a bid expires
type BidExpired struct {
	events.BaseEvent
	ExpiredAt time.Time `json:"expiredAt"`
}

func NewBidExpired(bidID string, expiredAt time.Time) *BidExpired {
	return &BidExpired{
		BaseEvent: events.NewBaseEvent(bidID, "Bid"),
		ExpiredAt: expiredAt,
	}
}

func (e *BidExpired) GetEventType() string     { return "BidExpired" }
func (e *BidExpired) GetAggregateID() string   { return e.AggregateID }
func (e *BidExpired) GetAggregateType() string { return e.AggregateType }

func (e *BidExpired) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidExpired) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// BidRejected event is emitted when a bid is rejected
type BidRejected struct {
	events.BaseEvent
	Reason     string `json:"reason"`
	RejectedBy string `json:"rejectedBy"`
}

func NewBidRejected(bidID, reason, rejectedBy string) *BidRejected {
	return &BidRejected{
		BaseEvent:  events.NewBaseEvent(bidID, "Bid"),
		Reason:     reason,
		RejectedBy: rejectedBy,
	}
}

func (e *BidRejected) GetEventType() string     { return "BidRejected" }
func (e *BidRejected) GetAggregateID() string   { return e.AggregateID }
func (e *BidRejected) GetAggregateType() string { return e.AggregateType }

func (e *BidRejected) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *BidRejected) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}