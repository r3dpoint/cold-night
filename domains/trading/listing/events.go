package listing

import (
	"encoding/json"
	"time"

	"securities-marketplace/domains/shared/events"
)

// Listing Domain Events

// ListingCreated event is emitted when a new listing is created
type ListingCreated struct {
	events.BaseEvent
	SecurityID      string  `json:"securityId"`
	SellerID        string  `json:"sellerId"`
	SharesOffered   int64   `json:"sharesOffered"`
	ListingType     string  `json:"listingType"`
	MinimumPrice    *float64 `json:"minimumPrice,omitempty"`
	ReservePrice    *float64 `json:"reservePrice,omitempty"`
	CurrentPrice    *float64 `json:"currentPrice,omitempty"`
	RestrictionType *string  `json:"restrictionType,omitempty"`
	AccreditedOnly  bool    `json:"accreditedOnly"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}

// NewListingCreated creates a new ListingCreated event
func NewListingCreated(listingID, securityID, sellerID string, sharesOffered int64, listingType string, minimumPrice, reservePrice, currentPrice *float64, restrictionType *string, accreditedOnly bool, expiresAt *time.Time) *ListingCreated {
	return &ListingCreated{
		BaseEvent:       events.NewBaseEvent(listingID, "Listing"),
		SecurityID:      securityID,
		SellerID:        sellerID,
		SharesOffered:   sharesOffered,
		ListingType:     listingType,
		MinimumPrice:    minimumPrice,
		ReservePrice:    reservePrice,
		CurrentPrice:    currentPrice,
		RestrictionType: restrictionType,
		AccreditedOnly:  accreditedOnly,
		ExpiresAt:       expiresAt,
	}
}

func (e *ListingCreated) GetEventType() string     { return "ListingCreated" }
func (e *ListingCreated) GetAggregateID() string   { return e.AggregateID }
func (e *ListingCreated) GetAggregateType() string { return e.AggregateType }

func (e *ListingCreated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingCreated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingPriceUpdated event is emitted when listing price is updated
type ListingPriceUpdated struct {
	events.BaseEvent
	OldPrice    *float64 `json:"oldPrice,omitempty"`
	NewPrice    float64  `json:"newPrice"`
	UpdatedBy   string   `json:"updatedBy"`
	Reason      string   `json:"reason"`
}

func NewListingPriceUpdated(listingID string, oldPrice *float64, newPrice float64, updatedBy, reason string) *ListingPriceUpdated {
	return &ListingPriceUpdated{
		BaseEvent: events.NewBaseEvent(listingID, "Listing"),
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		UpdatedBy: updatedBy,
		Reason:    reason,
	}
}

func (e *ListingPriceUpdated) GetEventType() string     { return "ListingPriceUpdated" }
func (e *ListingPriceUpdated) GetAggregateID() string   { return e.AggregateID }
func (e *ListingPriceUpdated) GetAggregateType() string { return e.AggregateType }

func (e *ListingPriceUpdated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingPriceUpdated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingSharesReduced event is emitted when shares are partially sold
type ListingSharesReduced struct {
	events.BaseEvent
	SharesSold      int64  `json:"sharesSold"`
	SharesRemaining int64  `json:"sharesRemaining"`
	TradeID         string `json:"tradeId"`
	BuyerID         string `json:"buyerId"`
	SalePrice       float64 `json:"salePrice"`
}

func NewListingSharesReduced(listingID string, sharesSold, sharesRemaining int64, tradeID, buyerID string, salePrice float64) *ListingSharesReduced {
	return &ListingSharesReduced{
		BaseEvent:       events.NewBaseEvent(listingID, "Listing"),
		SharesSold:      sharesSold,
		SharesRemaining: sharesRemaining,
		TradeID:         tradeID,
		BuyerID:         buyerID,
		SalePrice:       salePrice,
	}
}

func (e *ListingSharesReduced) GetEventType() string     { return "ListingSharesReduced" }
func (e *ListingSharesReduced) GetAggregateID() string   { return e.AggregateID }
func (e *ListingSharesReduced) GetAggregateType() string { return e.AggregateType }

func (e *ListingSharesReduced) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingSharesReduced) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingCancelled event is emitted when a listing is cancelled
type ListingCancelled struct {
	events.BaseEvent
	Reason      string `json:"reason"`
	CancelledBy string `json:"cancelledBy"`
}

func NewListingCancelled(listingID, reason, cancelledBy string) *ListingCancelled {
	return &ListingCancelled{
		BaseEvent:   events.NewBaseEvent(listingID, "Listing"),
		Reason:      reason,
		CancelledBy: cancelledBy,
	}
}

func (e *ListingCancelled) GetEventType() string     { return "ListingCancelled" }
func (e *ListingCancelled) GetAggregateID() string   { return e.AggregateID }
func (e *ListingCancelled) GetAggregateType() string { return e.AggregateType }

func (e *ListingCancelled) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingCancelled) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingExpired event is emitted when a listing expires
type ListingExpired struct {
	events.BaseEvent
	ExpiredAt time.Time `json:"expiredAt"`
}

func NewListingExpired(listingID string, expiredAt time.Time) *ListingExpired {
	return &ListingExpired{
		BaseEvent: events.NewBaseEvent(listingID, "Listing"),
		ExpiredAt: expiredAt,
	}
}

func (e *ListingExpired) GetEventType() string     { return "ListingExpired" }
func (e *ListingExpired) GetAggregateID() string   { return e.AggregateID }
func (e *ListingExpired) GetAggregateType() string { return e.AggregateType }

func (e *ListingExpired) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingExpired) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingCompleted event is emitted when all shares in a listing are sold
type ListingCompleted struct {
	events.BaseEvent
	TotalSharesSold int64     `json:"totalSharesSold"`
	FinalPrice      float64   `json:"finalPrice"`
	CompletedAt     time.Time `json:"completedAt"`
}

func NewListingCompleted(listingID string, totalSharesSold int64, finalPrice float64, completedAt time.Time) *ListingCompleted {
	return &ListingCompleted{
		BaseEvent:       events.NewBaseEvent(listingID, "Listing"),
		TotalSharesSold: totalSharesSold,
		FinalPrice:      finalPrice,
		CompletedAt:     completedAt,
	}
}

func (e *ListingCompleted) GetEventType() string     { return "ListingCompleted" }
func (e *ListingCompleted) GetAggregateID() string   { return e.AggregateID }
func (e *ListingCompleted) GetAggregateType() string { return e.AggregateType }

func (e *ListingCompleted) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingCompleted) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ListingReactivated event is emitted when a cancelled listing is reactivated
type ListingReactivated struct {
	events.BaseEvent
	ReactivatedBy string `json:"reactivatedBy"`
	Reason        string `json:"reason"`
}

func NewListingReactivated(listingID, reactivatedBy, reason string) *ListingReactivated {
	return &ListingReactivated{
		BaseEvent:     events.NewBaseEvent(listingID, "Listing"),
		ReactivatedBy: reactivatedBy,
		Reason:        reason,
	}
}

func (e *ListingReactivated) GetEventType() string     { return "ListingReactivated" }
func (e *ListingReactivated) GetAggregateID() string   { return e.AggregateID }
func (e *ListingReactivated) GetAggregateType() string { return e.AggregateType }

func (e *ListingReactivated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ListingReactivated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}