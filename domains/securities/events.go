package securities

import (
	"encoding/json"
	"time"

	"securities-marketplace/domains/shared/events"
)

// Security Domain Events

// SecurityListed event is emitted when a new security is listed
type SecurityListed struct {
	events.BaseEvent
	IssuerID     string            `json:"issuerId"`
	SecurityType string            `json:"securityType"`
	Name         string            `json:"name"`
	Symbol       string            `json:"symbol"`
	TotalShares  int64             `json:"totalShares"`
	ParValue     *float64          `json:"parValue,omitempty"`
	Details      map[string]string `json:"details"`
}

// NewSecurityListed creates a new SecurityListed event
func NewSecurityListed(securityID, issuerID, securityType, name, symbol string, totalShares int64, parValue *float64, details map[string]string) *SecurityListed {
	return &SecurityListed{
		BaseEvent:    events.NewBaseEvent(securityID, "Security"),
		IssuerID:     issuerID,
		SecurityType: securityType,
		Name:         name,
		Symbol:       symbol,
		TotalShares:  totalShares,
		ParValue:     parValue,
		Details:      details,
	}
}

func (e *SecurityListed) GetEventType() string     { return "SecurityListed" }
func (e *SecurityListed) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityListed) GetAggregateType() string { return e.AggregateType }

func (e *SecurityListed) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityListed) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityDocumentAdded event is emitted when a document is added to a security
type SecurityDocumentAdded struct {
	events.BaseEvent
	DocumentInfo SecurityDocument `json:"documentInfo"`
	AddedBy      string           `json:"addedBy"`
}

type SecurityDocument struct {
	DocumentID   string    `json:"documentId"`
	DocumentType string    `json:"documentType"`
	Title        string    `json:"title"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	ContentHash  string    `json:"contentHash"`
	UploadedAt   time.Time `json:"uploadedAt"`
	Version      string    `json:"version"`
	IsProspectus bool      `json:"isProspectus"`
}

func NewSecurityDocumentAdded(securityID string, documentInfo SecurityDocument, addedBy string) *SecurityDocumentAdded {
	return &SecurityDocumentAdded{
		BaseEvent:    events.NewBaseEvent(securityID, "Security"),
		DocumentInfo: documentInfo,
		AddedBy:      addedBy,
	}
}

func (e *SecurityDocumentAdded) GetEventType() string     { return "SecurityDocumentAdded" }
func (e *SecurityDocumentAdded) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityDocumentAdded) GetAggregateType() string { return e.AggregateType }

func (e *SecurityDocumentAdded) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityDocumentAdded) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityUpdated event is emitted when security information is updated
type SecurityUpdated struct {
	events.BaseEvent
	UpdatedFields map[string]interface{} `json:"updatedFields"`
	UpdatedBy     string                 `json:"updatedBy"`
	Reason        string                 `json:"reason"`
}

func NewSecurityUpdated(securityID string, updatedFields map[string]interface{}, updatedBy, reason string) *SecurityUpdated {
	return &SecurityUpdated{
		BaseEvent:     events.NewBaseEvent(securityID, "Security"),
		UpdatedFields: updatedFields,
		UpdatedBy:     updatedBy,
		Reason:        reason,
	}
}

func (e *SecurityUpdated) GetEventType() string     { return "SecurityUpdated" }
func (e *SecurityUpdated) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityUpdated) GetAggregateType() string { return e.AggregateType }

func (e *SecurityUpdated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityUpdated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecuritySuspended event is emitted when trading of a security is suspended
type SecuritySuspended struct {
	events.BaseEvent
	Reason      string     `json:"reason"`
	SuspendedBy string     `json:"suspendedBy"`
	Duration    *time.Time `json:"duration,omitempty"` // nil for indefinite suspension
}

func NewSecuritySuspended(securityID, reason, suspendedBy string, duration *time.Time) *SecuritySuspended {
	return &SecuritySuspended{
		BaseEvent:   events.NewBaseEvent(securityID, "Security"),
		Reason:      reason,
		SuspendedBy: suspendedBy,
		Duration:    duration,
	}
}

func (e *SecuritySuspended) GetEventType() string     { return "SecuritySuspended" }
func (e *SecuritySuspended) GetAggregateID() string   { return e.AggregateID }
func (e *SecuritySuspended) GetAggregateType() string { return e.AggregateType }

func (e *SecuritySuspended) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecuritySuspended) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityReinstated event is emitted when a suspended security is reinstated
type SecurityReinstated struct {
	events.BaseEvent
	ReinstatedBy string `json:"reinstatedBy"`
	Reason       string `json:"reason"`
}

func NewSecurityReinstated(securityID, reinstatedBy, reason string) *SecurityReinstated {
	return &SecurityReinstated{
		BaseEvent:    events.NewBaseEvent(securityID, "Security"),
		ReinstatedBy: reinstatedBy,
		Reason:       reason,
	}
}

func (e *SecurityReinstated) GetEventType() string     { return "SecurityReinstated" }
func (e *SecurityReinstated) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityReinstated) GetAggregateType() string { return e.AggregateType }

func (e *SecurityReinstated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityReinstated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityDelisted event is emitted when a security is delisted
type SecurityDelisted struct {
	events.BaseEvent
	Reason      string `json:"reason"`
	DelistedBy  string `json:"delistedBy"`
	EffectiveAt time.Time `json:"effectiveAt"`
}

func NewSecurityDelisted(securityID, reason, delistedBy string, effectiveAt time.Time) *SecurityDelisted {
	return &SecurityDelisted{
		BaseEvent:   events.NewBaseEvent(securityID, "Security"),
		Reason:      reason,
		DelistedBy:  delistedBy,
		EffectiveAt: effectiveAt,
	}
}

func (e *SecurityDelisted) GetEventType() string     { return "SecurityDelisted" }
func (e *SecurityDelisted) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityDelisted) GetAggregateType() string { return e.AggregateType }

func (e *SecurityDelisted) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityDelisted) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityOwnershipChanged event is emitted when ownership of shares changes
type SecurityOwnershipChanged struct {
	events.BaseEvent
	FromOwner   string `json:"fromOwner"`
	ToOwner     string `json:"toOwner"`
	SharesCount int64  `json:"sharesCount"`
	TradeID     string `json:"tradeId"`
}

func NewSecurityOwnershipChanged(securityID, fromOwner, toOwner string, sharesCount int64, tradeID string) *SecurityOwnershipChanged {
	return &SecurityOwnershipChanged{
		BaseEvent:   events.NewBaseEvent(securityID, "Security"),
		FromOwner:   fromOwner,
		ToOwner:     toOwner,
		SharesCount: sharesCount,
		TradeID:     tradeID,
	}
}

func (e *SecurityOwnershipChanged) GetEventType() string     { return "SecurityOwnershipChanged" }
func (e *SecurityOwnershipChanged) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityOwnershipChanged) GetAggregateType() string { return e.AggregateType }

func (e *SecurityOwnershipChanged) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityOwnershipChanged) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecurityDividendDeclared event is emitted when a dividend is declared
type SecurityDividendDeclared struct {
	events.BaseEvent
	DividendPerShare float64   `json:"dividendPerShare"`
	ExDividendDate   time.Time `json:"exDividendDate"`
	PaymentDate      time.Time `json:"paymentDate"`
	RecordDate       time.Time `json:"recordDate"`
	DeclaredBy       string    `json:"declaredBy"`
}

func NewSecurityDividendDeclared(securityID string, dividendPerShare float64, exDividendDate, paymentDate, recordDate time.Time, declaredBy string) *SecurityDividendDeclared {
	return &SecurityDividendDeclared{
		BaseEvent:        events.NewBaseEvent(securityID, "Security"),
		DividendPerShare: dividendPerShare,
		ExDividendDate:   exDividendDate,
		PaymentDate:      paymentDate,
		RecordDate:       recordDate,
		DeclaredBy:       declaredBy,
	}
}

func (e *SecurityDividendDeclared) GetEventType() string     { return "SecurityDividendDeclared" }
func (e *SecurityDividendDeclared) GetAggregateID() string   { return e.AggregateID }
func (e *SecurityDividendDeclared) GetAggregateType() string { return e.AggregateType }

func (e *SecurityDividendDeclared) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecurityDividendDeclared) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// SecuritySplitAnnounced event is emitted when a stock split is announced
type SecuritySplitAnnounced struct {
	events.BaseEvent
	SplitRatio   string    `json:"splitRatio"`   // e.g., "2:1", "3:2"
	EffectiveAt  time.Time `json:"effectiveAt"`
	AnnouncedBy  string    `json:"announcedBy"`
	Description  string    `json:"description"`
}

func NewSecuritySplitAnnounced(securityID, splitRatio string, effectiveAt time.Time, announcedBy, description string) *SecuritySplitAnnounced {
	return &SecuritySplitAnnounced{
		BaseEvent:   events.NewBaseEvent(securityID, "Security"),
		SplitRatio:  splitRatio,
		EffectiveAt: effectiveAt,
		AnnouncedBy: announcedBy,
		Description: description,
	}
}

func (e *SecuritySplitAnnounced) GetEventType() string     { return "SecuritySplitAnnounced" }
func (e *SecuritySplitAnnounced) GetAggregateID() string   { return e.AggregateID }
func (e *SecuritySplitAnnounced) GetAggregateType() string { return e.AggregateType }

func (e *SecuritySplitAnnounced) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *SecuritySplitAnnounced) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}