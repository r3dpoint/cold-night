package execution

import (
	"encoding/json"
	"time"

	"securities-marketplace/domains/shared/events"
)

// Trade Execution Domain Events

// TradeMatched event is emitted when a trade is initially matched
type TradeMatched struct {
	events.BaseEvent
	ListingID         string  `json:"listingId"`
	BidID             *string `json:"bidId,omitempty"`
	BuyerID           string  `json:"buyerId"`
	SellerID          string  `json:"sellerId"`
	SecurityID        string  `json:"securityId"`
	SharesTraded      int64   `json:"sharesTraded"`
	TradePrice        float64 `json:"tradePrice"`
	TotalAmount       float64 `json:"totalAmount"`
	SettlementDate    time.Time `json:"settlementDate"`
	MatchingAlgorithm string  `json:"matchingAlgorithm"`
}

// NewTradeMatched creates a new TradeMatched event
func NewTradeMatched(tradeID, listingID string, bidID *string, buyerID, sellerID, securityID string, sharesTraded int64, tradePrice, totalAmount float64, settlementDate time.Time, matchingAlgorithm string) *TradeMatched {
	return &TradeMatched{
		BaseEvent:         events.NewBaseEvent(tradeID, "Trade"),
		ListingID:         listingID,
		BidID:             bidID,
		BuyerID:           buyerID,
		SellerID:          sellerID,
		SecurityID:        securityID,
		SharesTraded:      sharesTraded,
		TradePrice:        tradePrice,
		TotalAmount:       totalAmount,
		SettlementDate:    settlementDate,
		MatchingAlgorithm: matchingAlgorithm,
	}
}

func (e *TradeMatched) GetEventType() string     { return "TradeMatched" }
func (e *TradeMatched) GetAggregateID() string   { return e.AggregateID }
func (e *TradeMatched) GetAggregateType() string { return e.AggregateType }

func (e *TradeMatched) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeMatched) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// TradeConfirmed event is emitted when both parties confirm the trade
type TradeConfirmed struct {
	events.BaseEvent
	BuyerConfirmed  bool      `json:"buyerConfirmed"`
	SellerConfirmed bool      `json:"sellerConfirmed"`
	ConfirmedBy     string    `json:"confirmedBy"`
	ConfirmedAt     time.Time `json:"confirmedAt"`
}

func NewTradeConfirmed(tradeID string, buyerConfirmed, sellerConfirmed bool, confirmedBy string, confirmedAt time.Time) *TradeConfirmed {
	return &TradeConfirmed{
		BaseEvent:       events.NewBaseEvent(tradeID, "Trade"),
		BuyerConfirmed:  buyerConfirmed,
		SellerConfirmed: sellerConfirmed,
		ConfirmedBy:     confirmedBy,
		ConfirmedAt:     confirmedAt,
	}
}

func (e *TradeConfirmed) GetEventType() string     { return "TradeConfirmed" }
func (e *TradeConfirmed) GetAggregateID() string   { return e.AggregateID }
func (e *TradeConfirmed) GetAggregateType() string { return e.AggregateType }

func (e *TradeConfirmed) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeConfirmed) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// TradeSettlementInitiated event is emitted when settlement process begins
type TradeSettlementInitiated struct {
	events.BaseEvent
	EscrowAccountID string    `json:"escrowAccountId"`
	InitiatedBy     string    `json:"initiatedBy"`
	InitiatedAt     time.Time `json:"initiatedAt"`
}

func NewTradeSettlementInitiated(tradeID, escrowAccountID, initiatedBy string, initiatedAt time.Time) *TradeSettlementInitiated {
	return &TradeSettlementInitiated{
		BaseEvent:       events.NewBaseEvent(tradeID, "Trade"),
		EscrowAccountID: escrowAccountID,
		InitiatedBy:     initiatedBy,
		InitiatedAt:     initiatedAt,
	}
}

func (e *TradeSettlementInitiated) GetEventType() string     { return "TradeSettlementInitiated" }
func (e *TradeSettlementInitiated) GetAggregateID() string   { return e.AggregateID }
func (e *TradeSettlementInitiated) GetAggregateType() string { return e.AggregateType }

func (e *TradeSettlementInitiated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeSettlementInitiated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// PaymentReceived event is emitted when payment is received in escrow
type PaymentReceived struct {
	events.BaseEvent
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	PaymentMethod   string    `json:"paymentMethod"`
	TransactionID   string    `json:"transactionId"`
	ReceivedAt      time.Time `json:"receivedAt"`
}

func NewPaymentReceived(tradeID string, amount float64, currency, paymentMethod, transactionID string, receivedAt time.Time) *PaymentReceived {
	return &PaymentReceived{
		BaseEvent:     events.NewBaseEvent(tradeID, "Trade"),
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		TransactionID: transactionID,
		ReceivedAt:    receivedAt,
	}
}

func (e *PaymentReceived) GetEventType() string     { return "PaymentReceived" }
func (e *PaymentReceived) GetAggregateID() string   { return e.AggregateID }
func (e *PaymentReceived) GetAggregateType() string { return e.AggregateType }

func (e *PaymentReceived) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *PaymentReceived) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SharesTransferred event is emitted when shares are transferred
type SharesTransferred struct {
	events.BaseEvent
	SharesCount     int64     `json:"sharesCount"`
	FromOwner       string    `json:"fromOwner"`
	ToOwner         string    `json:"toOwner"`
	TransferMethod  string    `json:"transferMethod"`
	CertificateHash string    `json:"certificateHash"`
	TransferredAt   time.Time `json:"transferredAt"`
}

func NewSharesTransferred(tradeID string, sharesCount int64, fromOwner, toOwner, transferMethod, certificateHash string, transferredAt time.Time) *SharesTransferred {
	return &SharesTransferred{
		BaseEvent:       events.NewBaseEvent(tradeID, "Trade"),
		SharesCount:     sharesCount,
		FromOwner:       fromOwner,
		ToOwner:         toOwner,
		TransferMethod:  transferMethod,
		CertificateHash: certificateHash,
		TransferredAt:   transferredAt,
	}
}

func (e *SharesTransferred) GetEventType() string     { return "SharesTransferred" }
func (e *SharesTransferred) GetAggregateID() string   { return e.AggregateID }
func (e *SharesTransferred) GetAggregateType() string { return e.AggregateType }

func (e *SharesTransferred) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SharesTransferred) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// TradeSettled event is emitted when the trade is fully settled
type TradeSettled struct {
	events.BaseEvent
	SettledAt         time.Time `json:"settledAt"`
	FinalAmount       float64   `json:"finalAmount"`
	Fees              float64   `json:"fees"`
	Taxes             float64   `json:"taxes"`
	SettlementMethod  string    `json:"settlementMethod"`
}

func NewTradeSettled(tradeID string, settledAt time.Time, finalAmount, fees, taxes float64, settlementMethod string) *TradeSettled {
	return &TradeSettled{
		BaseEvent:        events.NewBaseEvent(tradeID, "Trade"),
		SettledAt:        settledAt,
		FinalAmount:      finalAmount,
		Fees:             fees,
		Taxes:            taxes,
		SettlementMethod: settlementMethod,
	}
}

func (e *TradeSettled) GetEventType() string     { return "TradeSettled" }
func (e *TradeSettled) GetAggregateID() string   { return e.AggregateID }
func (e *TradeSettled) GetAggregateType() string { return e.AggregateType }

func (e *TradeSettled) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeSettled) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// TradeFailed event is emitted when a trade fails during settlement
type TradeFailed struct {
	events.BaseEvent
	FailureReason   string    `json:"failureReason"`
	FailureStage    string    `json:"failureStage"`
	FailedAt        time.Time `json:"failedAt"`
	RecoveryAction  string    `json:"recoveryAction"`
}

func NewTradeFailed(tradeID, failureReason, failureStage string, failedAt time.Time, recoveryAction string) *TradeFailed {
	return &TradeFailed{
		BaseEvent:      events.NewBaseEvent(tradeID, "Trade"),
		FailureReason:  failureReason,
		FailureStage:   failureStage,
		FailedAt:       failedAt,
		RecoveryAction: recoveryAction,
	}
}

func (e *TradeFailed) GetEventType() string     { return "TradeFailed" }
func (e *TradeFailed) GetAggregateID() string   { return e.AggregateID }
func (e *TradeFailed) GetAggregateType() string { return e.AggregateType }

func (e *TradeFailed) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeFailed) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// TradeCancelled event is emitted when a trade is cancelled before settlement
type TradeCancelled struct {
	events.BaseEvent
	CancellationReason string    `json:"cancellationReason"`
	CancelledBy        string    `json:"cancelledBy"`
	CancelledAt        time.Time `json:"cancelledAt"`
}

func NewTradeCancelled(tradeID, cancellationReason, cancelledBy string, cancelledAt time.Time) *TradeCancelled {
	return &TradeCancelled{
		BaseEvent:          events.NewBaseEvent(tradeID, "Trade"),
		CancellationReason: cancellationReason,
		CancelledBy:        cancelledBy,
		CancelledAt:        cancelledAt,
	}
}

func (e *TradeCancelled) GetEventType() string     { return "TradeCancelled" }
func (e *TradeCancelled) GetAggregateID() string   { return e.AggregateID }
func (e *TradeCancelled) GetAggregateType() string { return e.AggregateType }

func (e *TradeCancelled) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *TradeCancelled) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}