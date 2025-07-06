package securities

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// SecurityType represents the type of security
type SecurityType string

const (
	SecurityTypeStock     SecurityType = "stock"
	SecurityTypeBond      SecurityType = "bond"
	SecurityTypePreferred SecurityType = "preferred"
	SecurityTypeWarrant   SecurityType = "warrant"
	SecurityTypeOption    SecurityType = "option"
)

// SecurityStatus represents the trading status of a security
type SecurityStatus string

const (
	SecurityStatusActive    SecurityStatus = "active"
	SecurityStatusSuspended SecurityStatus = "suspended"
	SecurityStatusDelisted  SecurityStatus = "delisted"
	SecurityStatusPending   SecurityStatus = "pending"
)

// DividendInfo holds dividend information
type DividendInfo struct {
	DividendPerShare float64   `json:"dividendPerShare"`
	ExDividendDate   time.Time `json:"exDividendDate"`
	PaymentDate      time.Time `json:"paymentDate"`
	RecordDate       time.Time `json:"recordDate"`
	DeclaredBy       string    `json:"declaredBy"`
	DeclaredAt       time.Time `json:"declaredAt"`
}

// SplitInfo holds stock split information
type SplitInfo struct {
	SplitRatio   string    `json:"splitRatio"`
	EffectiveAt  time.Time `json:"effectiveAt"`
	AnnouncedBy  string    `json:"announcedBy"`
	AnnouncedAt  time.Time `json:"announcedAt"`
	Description  string    `json:"description"`
	Applied      bool      `json:"applied"`
}

// OwnershipRecord tracks ownership of shares
type OwnershipRecord struct {
	OwnerID     string `json:"ownerId"`
	SharesOwned int64  `json:"sharesOwned"`
}

// SecurityAggregate represents a security in the marketplace
type SecurityAggregate struct {
	events.AggregateRoot
	
	// Basic security information
	IssuerID     string            `json:"issuerId"`
	SecurityType SecurityType      `json:"securityType"`
	Name         string            `json:"name"`
	Symbol       string            `json:"symbol"`
	TotalShares  int64             `json:"totalShares"`
	ParValue     *float64          `json:"parValue,omitempty"`
	Details      map[string]string `json:"details"`
	
	// Status and lifecycle
	Status      SecurityStatus `json:"status"`
	ListedAt    time.Time      `json:"listedAt"`
	DelistedAt  *time.Time     `json:"delistedAt,omitempty"`
	
	// Suspension details
	SuspendedAt     *time.Time `json:"suspendedAt,omitempty"`
	SuspensionUntil *time.Time `json:"suspensionUntil,omitempty"`
	SuspensionReason string    `json:"suspensionReason,omitempty"`
	
	// Documents and compliance
	Documents       []SecurityDocument `json:"documents"`
	ProspectusHash  string             `json:"prospectusHash,omitempty"`
	
	// Corporate actions
	Dividends []DividendInfo `json:"dividends"`
	Splits    []SplitInfo    `json:"splits"`
	
	// Ownership tracking
	Ownership map[string]*OwnershipRecord `json:"ownership"`
	
	// Market data cache (for performance)
	LastTradePrice *float64   `json:"lastTradePrice,omitempty"`
	MarketCap      *float64   `json:"marketCap,omitempty"`
	LastUpdated    *time.Time `json:"lastUpdated,omitempty"`
}

// NewSecurityAggregate creates a new security aggregate
func NewSecurityAggregate(securityID string) *SecurityAggregate {
	return &SecurityAggregate{
		AggregateRoot: events.NewAggregateRoot(securityID, "Security"),
		Status:        SecurityStatusPending,
		Details:       make(map[string]string),
		Documents:     make([]SecurityDocument, 0),
		Dividends:     make([]DividendInfo, 0),
		Splits:        make([]SplitInfo, 0),
		Ownership:     make(map[string]*OwnershipRecord),
	}
}

// ListSecurity lists a new security
func (s *SecurityAggregate) ListSecurity(issuerID string, securityType SecurityType, name, symbol string, totalShares int64, parValue *float64, details map[string]string) error {
	if s.Version > 0 {
		return fmt.Errorf("security already exists")
	}

	event := NewSecurityListed(s.ID, issuerID, string(securityType), name, symbol, totalShares, parValue, details)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// AddDocument adds a document to the security
func (s *SecurityAggregate) AddDocument(documentInfo SecurityDocument, addedBy string) error {
	if s.Status == SecurityStatusDelisted {
		return fmt.Errorf("cannot add documents to delisted security")
	}

	// Check if document already exists
	for _, doc := range s.Documents {
		if doc.DocumentID == documentInfo.DocumentID {
			return fmt.Errorf("document with ID %s already exists", documentInfo.DocumentID)
		}
	}

	event := NewSecurityDocumentAdded(s.ID, documentInfo, addedBy)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// UpdateSecurity updates security information
func (s *SecurityAggregate) UpdateSecurity(updatedFields map[string]interface{}, updatedBy, reason string) error {
	if s.Status == SecurityStatusDelisted {
		return fmt.Errorf("cannot update delisted security")
	}

	event := NewSecurityUpdated(s.ID, updatedFields, updatedBy, reason)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// SuspendTrading suspends trading of the security
func (s *SecurityAggregate) SuspendTrading(reason, suspendedBy string, duration *time.Time) error {
	if s.Status == SecurityStatusSuspended {
		return fmt.Errorf("security is already suspended")
	}
	if s.Status == SecurityStatusDelisted {
		return fmt.Errorf("cannot suspend delisted security")
	}

	event := NewSecuritySuspended(s.ID, reason, suspendedBy, duration)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// ReinstateTrading reinstates trading of a suspended security
func (s *SecurityAggregate) ReinstateTrading(reinstatedBy, reason string) error {
	if s.Status != SecurityStatusSuspended {
		return fmt.Errorf("security is not suspended")
	}

	event := NewSecurityReinstated(s.ID, reinstatedBy, reason)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// DelistSecurity delists the security
func (s *SecurityAggregate) DelistSecurity(reason, delistedBy string, effectiveAt time.Time) error {
	if s.Status == SecurityStatusDelisted {
		return fmt.Errorf("security is already delisted")
	}

	event := NewSecurityDelisted(s.ID, reason, delistedBy, effectiveAt)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// TransferOwnership transfers ownership of shares
func (s *SecurityAggregate) TransferOwnership(fromOwner, toOwner string, sharesCount int64, tradeID string) error {
	if s.Status == SecurityStatusDelisted {
		return fmt.Errorf("cannot transfer ownership of delisted security")
	}
	if s.Status == SecurityStatusSuspended {
		return fmt.Errorf("cannot transfer ownership of suspended security")
	}

	// Validate ownership
	fromRecord, exists := s.Ownership[fromOwner]
	if !exists || fromRecord.SharesOwned < sharesCount {
		return fmt.Errorf("insufficient shares for transfer")
	}

	event := NewSecurityOwnershipChanged(s.ID, fromOwner, toOwner, sharesCount, tradeID)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// DeclareDividend declares a dividend
func (s *SecurityAggregate) DeclareDividend(dividendPerShare float64, exDividendDate, paymentDate, recordDate time.Time, declaredBy string) error {
	if s.Status != SecurityStatusActive {
		return fmt.Errorf("can only declare dividends for active securities")
	}

	event := NewSecurityDividendDeclared(s.ID, dividendPerShare, exDividendDate, paymentDate, recordDate, declaredBy)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// AnnounceSplit announces a stock split
func (s *SecurityAggregate) AnnounceSplit(splitRatio string, effectiveAt time.Time, announcedBy, description string) error {
	if s.Status != SecurityStatusActive {
		return fmt.Errorf("can only announce splits for active securities")
	}

	event := NewSecuritySplitAnnounced(s.ID, splitRatio, effectiveAt, announcedBy, description)
	s.AddEvent(event)
	return s.ApplyEvent(event)
}

// ApplyEvent applies an event to the aggregate
func (s *SecurityAggregate) ApplyEvent(event events.DomainEvent) error {
	switch e := event.(type) {
	case *SecurityListed:
		return s.applySecurityListed(e)
	case *SecurityDocumentAdded:
		return s.applySecurityDocumentAdded(e)
	case *SecurityUpdated:
		return s.applySecurityUpdated(e)
	case *SecuritySuspended:
		return s.applySecuritySuspended(e)
	case *SecurityReinstated:
		return s.applySecurityReinstated(e)
	case *SecurityDelisted:
		return s.applySecurityDelisted(e)
	case *SecurityOwnershipChanged:
		return s.applySecurityOwnershipChanged(e)
	case *SecurityDividendDeclared:
		return s.applySecurityDividendDeclared(e)
	case *SecuritySplitAnnounced:
		return s.applySecuritySplitAnnounced(e)
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// LoadFromHistory loads the aggregate from a sequence of events
func (s *SecurityAggregate) LoadFromHistory(events []events.DomainEvent) error {
	for _, event := range events {
		if err := s.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.GetEventType(), err)
		}
		s.IncrementVersion()
	}
	return nil
}

// Event application methods

func (s *SecurityAggregate) applySecurityListed(event *SecurityListed) error {
	s.IssuerID = event.IssuerID
	s.SecurityType = SecurityType(event.SecurityType)
	s.Name = event.Name
	s.Symbol = event.Symbol
	s.TotalShares = event.TotalShares
	s.ParValue = event.ParValue
	s.Details = event.Details
	s.Status = SecurityStatusActive
	s.ListedAt = event.Timestamp
	
	// Initialize ownership with issuer owning all shares
	s.Ownership[event.IssuerID] = &OwnershipRecord{
		OwnerID:     event.IssuerID,
		SharesOwned: event.TotalShares,
	}
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityDocumentAdded(event *SecurityDocumentAdded) error {
	s.Documents = append(s.Documents, event.DocumentInfo)
	
	// Update prospectus hash if this is a prospectus
	if event.DocumentInfo.IsProspectus {
		s.ProspectusHash = event.DocumentInfo.ContentHash
	}
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityUpdated(event *SecurityUpdated) error {
	for field, value := range event.UpdatedFields {
		switch field {
		case "name":
			if v, ok := value.(string); ok {
				s.Name = v
			}
		case "totalShares":
			if v, ok := value.(float64); ok { // JSON numbers are float64
				s.TotalShares = int64(v)
			}
		case "parValue":
			if v, ok := value.(float64); ok {
				s.ParValue = &v
			}
		}
	}
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecuritySuspended(event *SecuritySuspended) error {
	s.Status = SecurityStatusSuspended
	s.SuspendedAt = &event.Timestamp
	s.SuspensionUntil = event.Duration
	s.SuspensionReason = event.Reason
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityReinstated(event *SecurityReinstated) error {
	s.Status = SecurityStatusActive
	s.SuspendedAt = nil
	s.SuspensionUntil = nil
	s.SuspensionReason = ""
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityDelisted(event *SecurityDelisted) error {
	s.Status = SecurityStatusDelisted
	s.DelistedAt = &event.EffectiveAt
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityOwnershipChanged(event *SecurityOwnershipChanged) error {
	// Update from owner
	fromRecord := s.Ownership[event.FromOwner]
	fromRecord.SharesOwned -= event.SharesCount
	if fromRecord.SharesOwned == 0 {
		delete(s.Ownership, event.FromOwner)
	}
	
	// Update to owner
	toRecord, exists := s.Ownership[event.ToOwner]
	if !exists {
		toRecord = &OwnershipRecord{
			OwnerID:     event.ToOwner,
			SharesOwned: 0,
		}
		s.Ownership[event.ToOwner] = toRecord
	}
	toRecord.SharesOwned += event.SharesCount
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecurityDividendDeclared(event *SecurityDividendDeclared) error {
	dividend := DividendInfo{
		DividendPerShare: event.DividendPerShare,
		ExDividendDate:   event.ExDividendDate,
		PaymentDate:      event.PaymentDate,
		RecordDate:       event.RecordDate,
		DeclaredBy:       event.DeclaredBy,
		DeclaredAt:       event.Timestamp,
	}
	
	s.Dividends = append(s.Dividends, dividend)
	
	s.IncrementVersion()
	return nil
}

func (s *SecurityAggregate) applySecuritySplitAnnounced(event *SecuritySplitAnnounced) error {
	split := SplitInfo{
		SplitRatio:  event.SplitRatio,
		EffectiveAt: event.EffectiveAt,
		AnnouncedBy: event.AnnouncedBy,
		AnnouncedAt: event.Timestamp,
		Description: event.Description,
		Applied:     false,
	}
	
	s.Splits = append(s.Splits, split)
	
	s.IncrementVersion()
	return nil
}

// Helper methods

// IsActive returns true if the security is actively trading
func (s *SecurityAggregate) IsActive() bool {
	return s.Status == SecurityStatusActive
}

// IsTradable returns true if the security can be traded
func (s *SecurityAggregate) IsTradable() bool {
	return s.Status == SecurityStatusActive
}

// GetOwnershipPercentage returns the ownership percentage for a given owner
func (s *SecurityAggregate) GetOwnershipPercentage(ownerID string) float64 {
	if record, exists := s.Ownership[ownerID]; exists {
		return float64(record.SharesOwned) / float64(s.TotalShares) * 100
	}
	return 0
}

// GetSharesOwned returns the number of shares owned by a given owner
func (s *SecurityAggregate) GetSharesOwned(ownerID string) int64 {
	if record, exists := s.Ownership[ownerID]; exists {
		return record.SharesOwned
	}
	return 0
}

// GetAllOwners returns all current owners of the security
func (s *SecurityAggregate) GetAllOwners() []OwnershipRecord {
	var owners []OwnershipRecord
	for _, record := range s.Ownership {
		owners = append(owners, *record)
	}
	return owners
}

// HasProspectus returns true if the security has a prospectus
func (s *SecurityAggregate) HasProspectus() bool {
	return s.ProspectusHash != ""
}

// GetLatestDividend returns the most recent dividend declaration
func (s *SecurityAggregate) GetLatestDividend() *DividendInfo {
	if len(s.Dividends) == 0 {
		return nil
	}
	return &s.Dividends[len(s.Dividends)-1]
}

// GetLatestSplit returns the most recent split announcement
func (s *SecurityAggregate) GetLatestSplit() *SplitInfo {
	if len(s.Splits) == 0 {
		return nil
	}
	return &s.Splits[len(s.Splits)-1]
}