package securities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"securities-marketplace/domains/shared/events"
)

// SecurityService provides application services for security domain
type SecurityService struct {
	repository SecurityRepository
	eventStore events.EventStore
	eventBus   events.EventBus
}

// NewSecurityService creates a new security service
func NewSecurityService(repository SecurityRepository, eventStore events.EventStore, eventBus events.EventBus) *SecurityService {
	return &SecurityService{
		repository: repository,
		eventStore: eventStore,
		eventBus:   eventBus,
	}
}

// ListSecurity handles security listing
func (s *SecurityService) ListSecurity(cmd *ListSecurityCommand) (*SecurityAggregate, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if security with same symbol already exists
	existingSecurity, err := s.repository.FindBySymbol(cmd.Symbol)
	if err != nil && !IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing security: %w", err)
	}
	if existingSecurity != nil {
		return nil, fmt.Errorf("security with symbol %s already exists", cmd.Symbol)
	}

	// Create new security aggregate
	securityID := cmd.SecurityID
	if securityID == "" {
		securityID = uuid.New().String()
	}

	security := NewSecurityAggregate(securityID)
	err = security.ListSecurity(
		cmd.IssuerID,
		SecurityType(cmd.SecurityType),
		cmd.Name,
		cmd.Symbol,
		cmd.TotalShares,
		cmd.ParValue,
		cmd.Details,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list security: %w", err)
	}

	// Save events
	err = s.saveAggregateEvents(security, cmd.IssuerID)
	if err != nil {
		return nil, fmt.Errorf("failed to save security events: %w", err)
	}

	return security, nil
}

// AddSecurityDocument handles adding documents to a security
func (s *SecurityService) AddSecurityDocument(cmd *AddSecurityDocumentCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	// Set upload timestamp
	cmd.DocumentInfo.UploadedAt = time.Now()

	err = security.AddDocument(cmd.DocumentInfo, cmd.AddedBy)
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.AddedBy)
}

// UpdateSecurity handles security updates
func (s *SecurityService) UpdateSecurity(cmd *UpdateSecurityCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	err = security.UpdateSecurity(cmd.UpdatedFields, cmd.UpdatedBy, cmd.Reason)
	if err != nil {
		return fmt.Errorf("failed to update security: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.UpdatedBy)
}

// SuspendSecurity handles security suspension
func (s *SecurityService) SuspendSecurity(cmd *SuspendSecurityCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	err = security.SuspendTrading(cmd.Reason, cmd.SuspendedBy, cmd.Duration)
	if err != nil {
		return fmt.Errorf("failed to suspend security: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.SuspendedBy)
}

// ReinstateSecurity handles security reinstatement
func (s *SecurityService) ReinstateSecurity(cmd *ReinstateSecurityCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	err = security.ReinstateTrading(cmd.ReinstatedBy, cmd.Reason)
	if err != nil {
		return fmt.Errorf("failed to reinstate security: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.ReinstatedBy)
}

// DelistSecurity handles security delisting
func (s *SecurityService) DelistSecurity(cmd *DelistSecurityCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	err = security.DelistSecurity(cmd.Reason, cmd.DelistedBy, cmd.EffectiveAt)
	if err != nil {
		return fmt.Errorf("failed to delist security: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.DelistedBy)
}

// TransferOwnership handles ownership transfer
func (s *SecurityService) TransferOwnership(cmd *TransferOwnershipCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	err = security.TransferOwnership(cmd.FromOwner, cmd.ToOwner, cmd.SharesCount, cmd.TradeID)
	if err != nil {
		return fmt.Errorf("failed to transfer ownership: %w", err)
	}

	return s.saveAggregateEvents(security, "system")
}

// DeclareDividend handles dividend declaration
func (s *SecurityService) DeclareDividend(cmd *DeclareDividendCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	// Verify that the declarer is the issuer
	if security.IssuerID != cmd.DeclaredBy {
		return fmt.Errorf("only the issuer can declare dividends")
	}

	err = security.DeclareDividend(
		cmd.DividendPerShare,
		cmd.ExDividendDate,
		cmd.PaymentDate,
		cmd.RecordDate,
		cmd.DeclaredBy,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dividend: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.DeclaredBy)
}

// AnnounceSplit handles stock split announcement
func (s *SecurityService) AnnounceSplit(cmd *AnnounceSplitCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	security, err := s.repository.FindByID(cmd.SecurityID)
	if err != nil {
		return fmt.Errorf("failed to find security: %w", err)
	}

	// Verify that the announcer is the issuer
	if security.IssuerID != cmd.AnnouncedBy {
		return fmt.Errorf("only the issuer can announce stock splits")
	}

	err = security.AnnounceSplit(cmd.SplitRatio, cmd.EffectiveAt, cmd.AnnouncedBy, cmd.Description)
	if err != nil {
		return fmt.Errorf("failed to announce split: %w", err)
	}

	return s.saveAggregateEvents(security, cmd.AnnouncedBy)
}

// GetSecurity retrieves a security by ID
func (s *SecurityService) GetSecurity(securityID string) (*SecurityAggregate, error) {
	return s.repository.FindByID(securityID)
}

// GetSecurityBySymbol retrieves a security by symbol
func (s *SecurityService) GetSecurityBySymbol(symbol string) (*SecurityAggregate, error) {
	return s.repository.FindBySymbol(symbol)
}

// GetSecuritiesByIssuer retrieves all securities for a given issuer
func (s *SecurityService) GetSecuritiesByIssuer(issuerID string) ([]*SecurityAggregate, error) {
	return s.repository.FindByIssuer(issuerID)
}

// GetSecuritiesByType retrieves all securities of a given type
func (s *SecurityService) GetSecuritiesByType(securityType SecurityType) ([]*SecurityAggregate, error) {
	return s.repository.FindByType(securityType)
}

// GetActiveSecurities retrieves all actively trading securities
func (s *SecurityService) GetActiveSecurities() ([]*SecurityAggregate, error) {
	return s.repository.FindByStatus(SecurityStatusActive)
}

// ValidateSecurityExists checks if a security exists and is tradable
func (s *SecurityService) ValidateSecurityExists(securityID string) (*SecurityAggregate, error) {
	security, err := s.repository.FindByID(securityID)
	if err != nil {
		return nil, fmt.Errorf("security not found: %w", err)
	}

	if !security.IsTradable() {
		return nil, fmt.Errorf("security %s is not tradable (status: %s)", securityID, security.Status)
	}

	return security, nil
}

// GetOwnership retrieves ownership information for a security
func (s *SecurityService) GetOwnership(securityID string) ([]OwnershipRecord, error) {
	security, err := s.repository.FindByID(securityID)
	if err != nil {
		return nil, fmt.Errorf("failed to find security: %w", err)
	}

	return security.GetAllOwners(), nil
}

// GetUserSecurities retrieves all securities owned by a user
func (s *SecurityService) GetUserSecurities(userID string) ([]*SecurityAggregate, error) {
	return s.repository.FindByOwner(userID)
}

// CalculateMarketValue calculates the market value of a user's holdings
func (s *SecurityService) CalculateMarketValue(userID string) (float64, error) {
	securities, err := s.GetUserSecurities(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user securities: %w", err)
	}

	var totalValue float64
	for _, security := range securities {
		sharesOwned := security.GetSharesOwned(userID)
		if sharesOwned > 0 && security.LastTradePrice != nil {
			totalValue += float64(sharesOwned) * *security.LastTradePrice
		}
	}

	return totalValue, nil
}

// saveAggregateEvents saves uncommitted events from an aggregate
func (s *SecurityService) saveAggregateEvents(security *SecurityAggregate, userID string) error {
	uncommittedEvents := security.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	// Convert domain events to event store events
	var events []*events.Event
	correlationID := uuid.New().String()

	for i, domainEvent := range uncommittedEvents {
		var causationID *string
		if i > 0 {
			prevEventID := events[i-1].EventID
			causationID = &prevEventID
		}

		event, err := s.eventStore.CreateEventFromDomain(domainEvent, userID, correlationID, causationID)
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		event.AggregateVersion = security.GetVersion() + i + 1
		events = append(events, event)
	}

	// Save events
	err := s.eventStore.SaveEvents(events)
	if err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	// Publish events to event bus
	for _, domainEvent := range uncommittedEvents {
		err = s.eventBus.Publish(domainEvent)
		if err != nil {
			// Log error but don't fail the operation
			// Event bus failures should not prevent event storage
			fmt.Printf("Failed to publish event %s: %v\n", domainEvent.GetEventType(), err)
		}
	}

	// Mark events as committed
	security.MarkEventsAsCommitted()

	return nil
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}