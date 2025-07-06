package execution

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// TradeStatus represents the current status of a trade
type TradeStatus string

const (
	TradeStatusMatched              TradeStatus = "matched"
	TradeStatusPendingConfirmation  TradeStatus = "pending_confirmation"
	TradeStatusConfirmed            TradeStatus = "confirmed"
	TradeStatusSettlementInitiated TradeStatus = "settlement_initiated"
	TradeStatusPaymentReceived      TradeStatus = "payment_received"
	TradeStatusSharesTransferred    TradeStatus = "shares_transferred"
	TradeStatusSettled              TradeStatus = "settled"
	TradeStatusFailed               TradeStatus = "failed"
	TradeStatusCancelled            TradeStatus = "cancelled"
)

// SettlementStage represents the current stage of settlement
type SettlementStage string

const (
	SettlementStageNone              SettlementStage = "none"
	SettlementStageEscrowCreated     SettlementStage = "escrow_created"
	SettlementStageAwaitingPayment   SettlementStage = "awaiting_payment"
	SettlementStagePaymentReceived   SettlementStage = "payment_received"
	SettlementStageAwaitingTransfer  SettlementStage = "awaiting_transfer"
	SettlementStageSharesTransferred SettlementStage = "shares_transferred"
	SettlementStageCompleted         SettlementStage = "completed"
	SettlementStageFailed            SettlementStage = "failed"
)

// PaymentInfo holds payment details
type PaymentInfo struct {
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"paymentMethod"`
	TransactionID string    `json:"transactionId"`
	ReceivedAt    time.Time `json:"receivedAt"`
}

// TransferInfo holds share transfer details
type TransferInfo struct {
	SharesCount     int64     `json:"sharesCount"`
	FromOwner       string    `json:"fromOwner"`
	ToOwner         string    `json:"toOwner"`
	TransferMethod  string    `json:"transferMethod"`
	CertificateHash string    `json:"certificateHash"`
	TransferredAt   time.Time `json:"transferredAt"`
}

// TradeAggregate represents a trade execution and settlement
type TradeAggregate struct {
	events.AggregateRoot
	
	// Basic trade information
	ListingID       *string `json:"listingId,omitempty"`
	BidID           *string `json:"bidId,omitempty"`
	BuyerID         string  `json:"buyerId"`
	SellerID        string  `json:"sellerId"`
	SecurityID      string  `json:"securityId"`
	
	// Trade details
	SharesTraded    int64   `json:"sharesTraded"`
	TradePrice      float64 `json:"tradePrice"`
	TotalAmount     float64 `json:"totalAmount"`
	Fees            float64 `json:"fees"`
	Taxes           float64 `json:"taxes"`
	
	// Settlement information
	SettlementDate  time.Time `json:"settlementDate"`
	EscrowAccountID *string   `json:"escrowAccountId,omitempty"`
	
	// Status and lifecycle
	Status           TradeStatus      `json:"status"`
	SettlementStage  SettlementStage  `json:"settlementStage"`
	MatchedAt        time.Time        `json:"matchedAt"`
	ConfirmedAt      *time.Time       `json:"confirmedAt,omitempty"`
	SettledAt        *time.Time       `json:"settledAt,omitempty"`
	FailedAt         *time.Time       `json:"failedAt,omitempty"`
	CancelledAt      *time.Time       `json:"cancelledAt,omitempty"`
	
	// Confirmation tracking
	BuyerConfirmed   bool `json:"buyerConfirmed"`
	SellerConfirmed  bool `json:"sellerConfirmed"`
	
	// Settlement details
	PaymentInfo  *PaymentInfo  `json:"paymentInfo,omitempty"`
	TransferInfo *TransferInfo `json:"transferInfo,omitempty"`
	
	// Matching and algorithm info
	MatchingAlgorithm string `json:"matchingAlgorithm"`
	
	// Failure and cancellation details
	FailureReason      string `json:"failureReason,omitempty"`
	FailureStage       string `json:"failureStage,omitempty"`
	CancellationReason string `json:"cancellationReason,omitempty"`
	CancelledBy        string `json:"cancelledBy,omitempty"`
	RecoveryAction     string `json:"recoveryAction,omitempty"`
}

// NewTradeAggregate creates a new trade aggregate
func NewTradeAggregate(tradeID string) *TradeAggregate {
	return &TradeAggregate{
		AggregateRoot:     events.NewAggregateRoot(tradeID, "Trade"),
		Status:            TradeStatusMatched,
		SettlementStage:   SettlementStageNone,
		BuyerConfirmed:    false,
		SellerConfirmed:   false,
		Fees:              0,
		Taxes:             0,
	}
}

// MatchTrade creates a new trade match
func (t *TradeAggregate) MatchTrade(listingID string, bidID *string, buyerID, sellerID, securityID string, sharesTraded int64, tradePrice, totalAmount float64, settlementDate time.Time, matchingAlgorithm string) error {
	if t.Version > 0 {
		return fmt.Errorf("trade already exists")
	}

	// Validate basic requirements
	if sharesTraded <= 0 {
		return fmt.Errorf("shares traded must be greater than zero")
	}

	if tradePrice <= 0 {
		return fmt.Errorf("trade price must be greater than zero")
	}

	if totalAmount <= 0 {
		return fmt.Errorf("total amount must be greater than zero")
	}

	if settlementDate.Before(time.Now()) {
		return fmt.Errorf("settlement date cannot be in the past")
	}

	event := NewTradeMatched(t.ID, listingID, bidID, buyerID, sellerID, securityID, sharesTraded, tradePrice, totalAmount, settlementDate, matchingAlgorithm)
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// ConfirmTrade confirms the trade by one of the parties
func (t *TradeAggregate) ConfirmTrade(confirmedBy string) error {
	if t.Status != TradeStatusMatched && t.Status != TradeStatusPendingConfirmation {
		return fmt.Errorf("can only confirm matched or pending confirmation trades")
	}

	var buyerConfirmed, sellerConfirmed bool

	if confirmedBy == t.BuyerID {
		buyerConfirmed = true
		sellerConfirmed = t.SellerConfirmed
	} else if confirmedBy == t.SellerID {
		sellerConfirmed = true
		buyerConfirmed = t.BuyerConfirmed
	} else {
		return fmt.Errorf("only buyer or seller can confirm the trade")
	}

	event := NewTradeConfirmed(t.ID, buyerConfirmed, sellerConfirmed, confirmedBy, time.Now())
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// InitiateSettlement initiates the settlement process
func (t *TradeAggregate) InitiateSettlement(escrowAccountID, initiatedBy string) error {
	if t.Status != TradeStatusConfirmed {
		return fmt.Errorf("can only initiate settlement for confirmed trades")
	}

	if escrowAccountID == "" {
		return fmt.Errorf("escrow account ID is required")
	}

	event := NewTradeSettlementInitiated(t.ID, escrowAccountID, initiatedBy, time.Now())
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// ReceivePayment records payment received in escrow
func (t *TradeAggregate) ReceivePayment(amount float64, currency, paymentMethod, transactionID string) error {
	if t.Status != TradeStatusSettlementInitiated {
		return fmt.Errorf("can only receive payment for trades with initiated settlement")
	}

	if amount <= 0 {
		return fmt.Errorf("payment amount must be greater than zero")
	}

	if amount < t.TotalAmount {
		return fmt.Errorf("payment amount %.2f is less than required %.2f", amount, t.TotalAmount)
	}

	event := NewPaymentReceived(t.ID, amount, currency, paymentMethod, transactionID, time.Now())
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// TransferShares records the transfer of shares
func (t *TradeAggregate) TransferShares(sharesCount int64, fromOwner, toOwner, transferMethod, certificateHash string) error {
	if t.Status != TradeStatusPaymentReceived {
		return fmt.Errorf("can only transfer shares after payment is received")
	}

	if sharesCount != t.SharesTraded {
		return fmt.Errorf("shares count %d does not match traded amount %d", sharesCount, t.SharesTraded)
	}

	if fromOwner != t.SellerID {
		return fmt.Errorf("from owner must be the seller")
	}

	if toOwner != t.BuyerID {
		return fmt.Errorf("to owner must be the buyer")
	}

	event := NewSharesTransferred(t.ID, sharesCount, fromOwner, toOwner, transferMethod, certificateHash, time.Now())
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// SettleTrade completes the trade settlement
func (t *TradeAggregate) SettleTrade(finalAmount, fees, taxes float64, settlementMethod string) error {
	if t.Status != TradeStatusSharesTransferred {
		return fmt.Errorf("can only settle after shares are transferred")
	}

	if finalAmount <= 0 {
		return fmt.Errorf("final amount must be greater than zero")
	}

	event := NewTradeSettled(t.ID, time.Now(), finalAmount, fees, taxes, settlementMethod)
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// FailTrade marks the trade as failed
func (t *TradeAggregate) FailTrade(failureReason, failureStage, recoveryAction string) error {
	if t.Status == TradeStatusSettled || t.Status == TradeStatusCancelled {
		return fmt.Errorf("cannot fail already completed or cancelled trade")
	}

	if failureReason == "" {
		return fmt.Errorf("failure reason is required")
	}

	event := NewTradeFailed(t.ID, failureReason, failureStage, time.Now(), recoveryAction)
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// CancelTrade cancels the trade before settlement
func (t *TradeAggregate) CancelTrade(cancellationReason, cancelledBy string) error {
	if t.Status == TradeStatusSettled || t.Status == TradeStatusFailed {
		return fmt.Errorf("cannot cancel already completed or failed trade")
	}

	if t.Status == TradeStatusPaymentReceived || t.Status == TradeStatusSharesTransferred {
		return fmt.Errorf("cannot cancel trade after payment received or shares transferred")
	}

	if cancellationReason == "" {
		return fmt.Errorf("cancellation reason is required")
	}

	event := NewTradeCancelled(t.ID, cancellationReason, cancelledBy, time.Now())
	t.AddEvent(event)
	return t.ApplyEvent(event)
}

// ApplyEvent applies an event to the aggregate
func (t *TradeAggregate) ApplyEvent(event events.DomainEvent) error {
	switch e := event.(type) {
	case *TradeMatched:
		return t.applyTradeMatched(e)
	case *TradeConfirmed:
		return t.applyTradeConfirmed(e)
	case *TradeSettlementInitiated:
		return t.applyTradeSettlementInitiated(e)
	case *PaymentReceived:
		return t.applyPaymentReceived(e)
	case *SharesTransferred:
		return t.applySharesTransferred(e)
	case *TradeSettled:
		return t.applyTradeSettled(e)
	case *TradeFailed:
		return t.applyTradeFailed(e)
	case *TradeCancelled:
		return t.applyTradeCancelled(e)
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// LoadFromHistory loads the aggregate from a sequence of events
func (t *TradeAggregate) LoadFromHistory(events []events.DomainEvent) error {
	for _, event := range events {
		if err := t.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.GetEventType(), err)
		}
		t.IncrementVersion()
	}
	return nil
}

// Event application methods

func (t *TradeAggregate) applyTradeMatched(event *TradeMatched) error {
	t.ListingID = &event.ListingID
	t.BidID = event.BidID
	t.BuyerID = event.BuyerID
	t.SellerID = event.SellerID
	t.SecurityID = event.SecurityID
	t.SharesTraded = event.SharesTraded
	t.TradePrice = event.TradePrice
	t.TotalAmount = event.TotalAmount
	t.SettlementDate = event.SettlementDate
	t.MatchingAlgorithm = event.MatchingAlgorithm
	t.Status = TradeStatusMatched
	t.MatchedAt = event.Timestamp
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyTradeConfirmed(event *TradeConfirmed) error {
	t.BuyerConfirmed = event.BuyerConfirmed
	t.SellerConfirmed = event.SellerConfirmed
	
	if event.BuyerConfirmed && event.SellerConfirmed {
		t.Status = TradeStatusConfirmed
		t.ConfirmedAt = &event.ConfirmedAt
	} else {
		t.Status = TradeStatusPendingConfirmation
	}
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyTradeSettlementInitiated(event *TradeSettlementInitiated) error {
	t.Status = TradeStatusSettlementInitiated
	t.SettlementStage = SettlementStageEscrowCreated
	t.EscrowAccountID = &event.EscrowAccountID
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyPaymentReceived(event *PaymentReceived) error {
	t.Status = TradeStatusPaymentReceived
	t.SettlementStage = SettlementStagePaymentReceived
	t.PaymentInfo = &PaymentInfo{
		Amount:        event.Amount,
		Currency:      event.Currency,
		PaymentMethod: event.PaymentMethod,
		TransactionID: event.TransactionID,
		ReceivedAt:    event.ReceivedAt,
	}
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applySharesTransferred(event *SharesTransferred) error {
	t.Status = TradeStatusSharesTransferred
	t.SettlementStage = SettlementStageSharesTransferred
	t.TransferInfo = &TransferInfo{
		SharesCount:     event.SharesCount,
		FromOwner:       event.FromOwner,
		ToOwner:         event.ToOwner,
		TransferMethod:  event.TransferMethod,
		CertificateHash: event.CertificateHash,
		TransferredAt:   event.TransferredAt,
	}
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyTradeSettled(event *TradeSettled) error {
	t.Status = TradeStatusSettled
	t.SettlementStage = SettlementStageCompleted
	t.SettledAt = &event.SettledAt
	t.Fees = event.Fees
	t.Taxes = event.Taxes
	// Final amount might differ from total amount due to fees/taxes
	if event.FinalAmount != t.TotalAmount {
		t.TotalAmount = event.FinalAmount
	}
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyTradeFailed(event *TradeFailed) error {
	t.Status = TradeStatusFailed
	t.SettlementStage = SettlementStageFailed
	t.FailedAt = &event.FailedAt
	t.FailureReason = event.FailureReason
	t.FailureStage = event.FailureStage
	t.RecoveryAction = event.RecoveryAction
	
	t.IncrementVersion()
	return nil
}

func (t *TradeAggregate) applyTradeCancelled(event *TradeCancelled) error {
	t.Status = TradeStatusCancelled
	t.CancelledAt = &event.CancelledAt
	t.CancellationReason = event.CancellationReason
	t.CancelledBy = event.CancelledBy
	
	t.IncrementVersion()
	return nil
}

// Helper methods

// IsCompleted returns true if the trade is in a final state
func (t *TradeAggregate) IsCompleted() bool {
	return t.Status == TradeStatusSettled || t.Status == TradeStatusFailed || t.Status == TradeStatusCancelled
}

// IsSettled returns true if the trade is successfully settled
func (t *TradeAggregate) IsSettled() bool {
	return t.Status == TradeStatusSettled
}

// CanBeConfirmed returns true if the trade can be confirmed
func (t *TradeAggregate) CanBeConfirmed() bool {
	return t.Status == TradeStatusMatched || t.Status == TradeStatusPendingConfirmation
}

// CanBeCancelled returns true if the trade can be cancelled
func (t *TradeAggregate) CanBeCancelled() bool {
	return t.Status != TradeStatusSettled && t.Status != TradeStatusFailed && 
		   t.Status != TradeStatusPaymentReceived && t.Status != TradeStatusSharesTransferred
}

// GetNetAmount returns the net amount after fees and taxes
func (t *TradeAggregate) GetNetAmount() float64 {
	return t.TotalAmount - t.Fees - t.Taxes
}

// GetDaysToSettlement returns the number of days until settlement
func (t *TradeAggregate) GetDaysToSettlement() int {
	if t.IsCompleted() {
		return 0
	}
	
	days := int(time.Until(t.SettlementDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsOverdue returns true if the settlement date has passed
func (t *TradeAggregate) IsOverdue() bool {
	if t.IsCompleted() {
		return false
	}
	return time.Now().After(t.SettlementDate)
}

// GetProgressPercentage returns the settlement progress as a percentage
func (t *TradeAggregate) GetProgressPercentage() float64 {
	switch t.Status {
	case TradeStatusMatched:
		return 10.0
	case TradeStatusPendingConfirmation:
		return 20.0
	case TradeStatusConfirmed:
		return 30.0
	case TradeStatusSettlementInitiated:
		return 50.0
	case TradeStatusPaymentReceived:
		return 75.0
	case TradeStatusSharesTransferred:
		return 90.0
	case TradeStatusSettled:
		return 100.0
	case TradeStatusFailed, TradeStatusCancelled:
		return 0.0
	default:
		return 0.0
	}
}