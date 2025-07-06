package securities

import (
	"time"
)

// Security Domain Commands

// ListSecurityCommand represents a command to list a new security
type ListSecurityCommand struct {
	SecurityID   string            `json:"securityId"`
	IssuerID     string            `json:"issuerId"`
	SecurityType string            `json:"securityType"`
	Name         string            `json:"name"`
	Symbol       string            `json:"symbol"`
	TotalShares  int64             `json:"totalShares"`
	ParValue     *float64          `json:"parValue,omitempty"`
	Details      map[string]string `json:"details"`
}

// AddSecurityDocumentCommand represents a command to add a document to a security
type AddSecurityDocumentCommand struct {
	SecurityID   string           `json:"securityId"`
	DocumentInfo SecurityDocument `json:"documentInfo"`
	AddedBy      string           `json:"addedBy"`
}

// UpdateSecurityCommand represents a command to update security information
type UpdateSecurityCommand struct {
	SecurityID    string                 `json:"securityId"`
	UpdatedFields map[string]interface{} `json:"updatedFields"`
	UpdatedBy     string                 `json:"updatedBy"`
	Reason        string                 `json:"reason"`
}

// SuspendSecurityCommand represents a command to suspend trading of a security
type SuspendSecurityCommand struct {
	SecurityID  string     `json:"securityId"`
	Reason      string     `json:"reason"`
	SuspendedBy string     `json:"suspendedBy"`
	Duration    *time.Time `json:"duration,omitempty"`
}

// ReinstateSecurityCommand represents a command to reinstate trading of a security
type ReinstateSecurityCommand struct {
	SecurityID   string `json:"securityId"`
	ReinstatedBy string `json:"reinstatedBy"`
	Reason       string `json:"reason"`
}

// DelistSecurityCommand represents a command to delist a security
type DelistSecurityCommand struct {
	SecurityID  string    `json:"securityId"`
	Reason      string    `json:"reason"`
	DelistedBy  string    `json:"delistedBy"`
	EffectiveAt time.Time `json:"effectiveAt"`
}

// TransferOwnershipCommand represents a command to transfer security ownership
type TransferOwnershipCommand struct {
	SecurityID  string `json:"securityId"`
	FromOwner   string `json:"fromOwner"`
	ToOwner     string `json:"toOwner"`
	SharesCount int64  `json:"sharesCount"`
	TradeID     string `json:"tradeId"`
}

// DeclareDividendCommand represents a command to declare a dividend
type DeclareDividendCommand struct {
	SecurityID       string    `json:"securityId"`
	DividendPerShare float64   `json:"dividendPerShare"`
	ExDividendDate   time.Time `json:"exDividendDate"`
	PaymentDate      time.Time `json:"paymentDate"`
	RecordDate       time.Time `json:"recordDate"`
	DeclaredBy       string    `json:"declaredBy"`
}

// AnnounceSplitCommand represents a command to announce a stock split
type AnnounceSplitCommand struct {
	SecurityID  string    `json:"securityId"`
	SplitRatio  string    `json:"splitRatio"`
	EffectiveAt time.Time `json:"effectiveAt"`
	AnnouncedBy string    `json:"announcedBy"`
	Description string    `json:"description"`
}

// Command validation methods

// Validate validates the ListSecurityCommand
func (c *ListSecurityCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.IssuerID == "" {
		return NewValidationError("issuerId", "Issuer ID is required")
	}
	if c.SecurityType == "" {
		return NewValidationError("securityType", "Security type is required")
	}
	if !isValidSecurityType(c.SecurityType) {
		return NewValidationError("securityType", "Invalid security type")
	}
	if c.Name == "" {
		return NewValidationError("name", "Security name is required")
	}
	if c.Symbol == "" {
		return NewValidationError("symbol", "Security symbol is required")
	}
	if len(c.Symbol) > 10 {
		return NewValidationError("symbol", "Security symbol must be 10 characters or less")
	}
	if c.TotalShares <= 0 {
		return NewValidationError("totalShares", "Total shares must be greater than zero")
	}
	if c.ParValue != nil && *c.ParValue < 0 {
		return NewValidationError("parValue", "Par value cannot be negative")
	}
	return nil
}

// Validate validates the AddSecurityDocumentCommand
func (c *AddSecurityDocumentCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.DocumentInfo.DocumentID == "" {
		return NewValidationError("documentInfo.documentId", "Document ID is required")
	}
	if c.DocumentInfo.DocumentType == "" {
		return NewValidationError("documentInfo.documentType", "Document type is required")
	}
	if c.DocumentInfo.Title == "" {
		return NewValidationError("documentInfo.title", "Document title is required")
	}
	if c.DocumentInfo.FileName == "" {
		return NewValidationError("documentInfo.fileName", "File name is required")
	}
	if c.DocumentInfo.FileSize <= 0 {
		return NewValidationError("documentInfo.fileSize", "File size must be greater than zero")
	}
	if c.DocumentInfo.ContentHash == "" {
		return NewValidationError("documentInfo.contentHash", "Content hash is required")
	}
	if c.AddedBy == "" {
		return NewValidationError("addedBy", "Added by is required")
	}
	return nil
}

// Validate validates the UpdateSecurityCommand
func (c *UpdateSecurityCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if len(c.UpdatedFields) == 0 {
		return NewValidationError("updatedFields", "At least one field must be updated")
	}
	if c.UpdatedBy == "" {
		return NewValidationError("updatedBy", "Updated by is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	
	// Validate updated fields
	for field, value := range c.UpdatedFields {
		switch field {
		case "name":
			if v, ok := value.(string); !ok || v == "" {
				return NewValidationError("updatedFields.name", "Name must be a non-empty string")
			}
		case "totalShares":
			if v, ok := value.(float64); !ok || v <= 0 {
				return NewValidationError("updatedFields.totalShares", "Total shares must be greater than zero")
			}
		case "parValue":
			if v, ok := value.(float64); !ok || v < 0 {
				return NewValidationError("updatedFields.parValue", "Par value cannot be negative")
			}
		}
	}
	
	return nil
}

// Validate validates the SuspendSecurityCommand
func (c *SuspendSecurityCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	if c.SuspendedBy == "" {
		return NewValidationError("suspendedBy", "Suspended by is required")
	}
	if c.Duration != nil && c.Duration.Before(time.Now()) {
		return NewValidationError("duration", "Duration cannot be in the past")
	}
	return nil
}

// Validate validates the ReinstateSecurityCommand
func (c *ReinstateSecurityCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.ReinstatedBy == "" {
		return NewValidationError("reinstatedBy", "Reinstated by is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	return nil
}

// Validate validates the DelistSecurityCommand
func (c *DelistSecurityCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	if c.DelistedBy == "" {
		return NewValidationError("delistedBy", "Delisted by is required")
	}
	if c.EffectiveAt.IsZero() {
		return NewValidationError("effectiveAt", "Effective at date is required")
	}
	return nil
}

// Validate validates the TransferOwnershipCommand
func (c *TransferOwnershipCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.FromOwner == "" {
		return NewValidationError("fromOwner", "From owner is required")
	}
	if c.ToOwner == "" {
		return NewValidationError("toOwner", "To owner is required")
	}
	if c.FromOwner == c.ToOwner {
		return NewValidationError("toOwner", "Cannot transfer to the same owner")
	}
	if c.SharesCount <= 0 {
		return NewValidationError("sharesCount", "Shares count must be greater than zero")
	}
	if c.TradeID == "" {
		return NewValidationError("tradeId", "Trade ID is required")
	}
	return nil
}

// Validate validates the DeclareDividendCommand
func (c *DeclareDividendCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.DividendPerShare <= 0 {
		return NewValidationError("dividendPerShare", "Dividend per share must be greater than zero")
	}
	if c.ExDividendDate.IsZero() {
		return NewValidationError("exDividendDate", "Ex-dividend date is required")
	}
	if c.PaymentDate.IsZero() {
		return NewValidationError("paymentDate", "Payment date is required")
	}
	if c.RecordDate.IsZero() {
		return NewValidationError("recordDate", "Record date is required")
	}
	if c.DeclaredBy == "" {
		return NewValidationError("declaredBy", "Declared by is required")
	}
	
	// Date validation
	now := time.Now()
	if c.ExDividendDate.Before(now) {
		return NewValidationError("exDividendDate", "Ex-dividend date cannot be in the past")
	}
	if c.RecordDate.Before(c.ExDividendDate) {
		return NewValidationError("recordDate", "Record date must be on or after ex-dividend date")
	}
	if c.PaymentDate.Before(c.RecordDate) {
		return NewValidationError("paymentDate", "Payment date must be on or after record date")
	}
	
	return nil
}

// Validate validates the AnnounceSplitCommand
func (c *AnnounceSplitCommand) Validate() error {
	if c.SecurityID == "" {
		return NewValidationError("securityId", "Security ID is required")
	}
	if c.SplitRatio == "" {
		return NewValidationError("splitRatio", "Split ratio is required")
	}
	if !isValidSplitRatio(c.SplitRatio) {
		return NewValidationError("splitRatio", "Invalid split ratio format (e.g., '2:1', '3:2')")
	}
	if c.EffectiveAt.IsZero() {
		return NewValidationError("effectiveAt", "Effective at date is required")
	}
	if c.EffectiveAt.Before(time.Now()) {
		return NewValidationError("effectiveAt", "Effective at date cannot be in the past")
	}
	if c.AnnouncedBy == "" {
		return NewValidationError("announcedBy", "Announced by is required")
	}
	if c.Description == "" {
		return NewValidationError("description", "Description is required")
	}
	return nil
}

// Helper functions

func isValidSecurityType(securityType string) bool {
	validTypes := []string{
		string(SecurityTypeStock),
		string(SecurityTypeBond),
		string(SecurityTypePreferred),
		string(SecurityTypeWarrant),
		string(SecurityTypeOption),
	}
	
	for _, validType := range validTypes {
		if securityType == validType {
			return true
		}
	}
	return false
}

func isValidSplitRatio(ratio string) bool {
	// Simple validation for format like "2:1", "3:2", etc.
	// In a real implementation, you'd want more robust validation
	if len(ratio) < 3 {
		return false
	}
	
	for i, char := range ratio {
		if i == 1 && char != ':' {
			return false
		}
		if i != 1 && (char < '0' || char > '9') {
			return false
		}
	}
	
	return true
}

// ValidationError represents a command validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}