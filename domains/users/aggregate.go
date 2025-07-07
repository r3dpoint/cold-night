package users

import (
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
)

// AccreditationStatus represents the user's accreditation status
type AccreditationStatus string

const (
	AccreditationStatusPending  AccreditationStatus = "pending"
	AccreditationStatusVerified AccreditationStatus = "verified"
	AccreditationStatusRevoked  AccreditationStatus = "revoked"
	AccreditationStatusExpired  AccreditationStatus = "expired"
)

// ComplianceStatus represents the user's compliance status
type ComplianceStatus string

const (
	ComplianceStatusClear     ComplianceStatus = "clear"
	ComplianceStatusPending   ComplianceStatus = "pending"
	ComplianceStatusReview    ComplianceStatus = "review"
	ComplianceStatusBlocked   ComplianceStatus = "blocked"
)

// UserStatus represents the user's account status
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusInactive  UserStatus = "inactive"
)

// DocumentInfo represents an uploaded document
type DocumentInfo struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Size       int64     `json:"size"`
	Hash       string    `json:"hash"`
	UploadedAt time.Time `json:"uploadedAt"`
	URL        string    `json:"url,omitempty"`
}

// AccreditationInfo holds accreditation details
type AccreditationInfo struct {
	Type         string            `json:"type"`
	Status       AccreditationStatus `json:"status"`
	ValidUntil   *time.Time        `json:"validUntil,omitempty"`
	Documents    []DocumentInfo    `json:"documents"`
	Details      map[string]string `json:"details"`
	VerifiedBy   string            `json:"verifiedBy,omitempty"`
	VerifiedAt   *time.Time        `json:"verifiedAt,omitempty"`
	Notes        string            `json:"notes,omitempty"`
}

// ComplianceInfo holds compliance status details
type ComplianceInfo struct {
	OverallStatus ComplianceStatus  `json:"overallStatus"`
	KYCStatus     ComplianceStatus  `json:"kycStatus"`
	AMLStatus     ComplianceStatus  `json:"amlStatus"`
	SanctionsStatus ComplianceStatus `json:"sanctionsStatus"`
	LastCheck     *time.Time        `json:"lastCheck,omitempty"`
	NextReview    *time.Time        `json:"nextReview,omitempty"`
	RiskScore     int               `json:"riskScore"`
	WatchlistStatus string          `json:"watchlistStatus"`
}

// UserAggregate represents a user in the securities marketplace
type UserAggregate struct {
	events.AggregateRoot
	
	// Basic user information
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	PasswordHash string `json:"passwordHash"`
	
	// Account status
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	
	// Accreditation information
	Accreditation AccreditationInfo `json:"accreditation"`
	
	// Compliance information
	Compliance ComplianceInfo `json:"compliance"`
	
	// Suspension details (if applicable)
	SuspendedAt     *time.Time `json:"suspendedAt,omitempty"`
	SuspensionUntil *time.Time `json:"suspensionUntil,omitempty"`
	SuspensionReason string    `json:"suspensionReason,omitempty"`
}

// NewUserAggregate creates a new user aggregate
func NewUserAggregate(userID string) *UserAggregate {
	return &UserAggregate{
		AggregateRoot: events.NewAggregateRoot(userID, "User"),
		Status:        "", // Will be set when registered
		Accreditation: AccreditationInfo{
			Status:    "", // Will be set when accreditation is submitted
			Documents: make([]DocumentInfo, 0),
			Details:   make(map[string]string),
		},
		Compliance: ComplianceInfo{
			OverallStatus:   ComplianceStatusPending,
			KYCStatus:      ComplianceStatusPending,
			AMLStatus:      ComplianceStatusPending,
			SanctionsStatus: ComplianceStatusPending,
			RiskScore:      0,
			WatchlistStatus: "none",
		},
	}
}

// Register registers a new user
func (u *UserAggregate) Register(email, firstName, lastName, passwordHash, accreditationType string, accreditationDetails map[string]string) error {
	if u.Version > 0 {
		return fmt.Errorf("user already exists")
	}

	// Validate required fields
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if firstName == "" {
		return fmt.Errorf("first name is required")
	}
	if lastName == "" {
		return fmt.Errorf("last name is required")
	}
	if len(passwordHash) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	event := NewUserRegistered(u.ID, email, firstName, lastName, passwordHash, accreditationType, accreditationDetails)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// SubmitAccreditation submits accreditation documents
func (u *UserAggregate) SubmitAccreditation(accreditationType string, documents []DocumentInfo, submissionDetails map[string]string) error {
	if u.Status == UserStatusSuspended {
		return fmt.Errorf("cannot submit accreditation for suspended user")
	}

	event := NewAccreditationSubmitted(u.ID, accreditationType, documents, submissionDetails)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// VerifyAccreditation marks the user's accreditation as verified
func (u *UserAggregate) VerifyAccreditation(accreditationType string, validUntil time.Time, verifiedBy, verificationNotes string) error {
	if u.Accreditation.Status == "" {
		return fmt.Errorf("no accreditation submitted to verify")
	}
	if u.Accreditation.Status != AccreditationStatusPending {
		return fmt.Errorf("accreditation is not in pending status")
	}

	event := NewAccreditationVerified(u.ID, accreditationType, validUntil, verifiedBy, verificationNotes)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// RevokeAccreditation revokes the user's accreditation
func (u *UserAggregate) RevokeAccreditation(reason, revokedBy string) error {
	if u.Accreditation.Status != AccreditationStatusVerified {
		return fmt.Errorf("accreditation is not verified")
	}

	event := NewAccreditationRevoked(u.ID, reason, revokedBy)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// PerformComplianceCheck performs a compliance check
func (u *UserAggregate) PerformComplianceCheck(checkType, status string, results map[string]string, performedBy string, nextReview *time.Time) error {
	event := NewComplianceCheckPerformed(u.ID, checkType, status, results, performedBy, nextReview)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// Suspend suspends the user account
func (u *UserAggregate) Suspend(reason, suspendedBy string, duration *time.Time) error {
	if u.Status == UserStatusSuspended {
		return fmt.Errorf("user is already suspended")
	}

	event := NewUserSuspended(u.ID, reason, suspendedBy, duration)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// Reinstate reinstates a suspended user
func (u *UserAggregate) Reinstate(reinstatedBy, reason string) error {
	if u.Status != UserStatusSuspended {
		return fmt.Errorf("user is not suspended")
	}

	event := NewUserReinstated(u.ID, reinstatedBy, reason)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// UpdateProfile updates user profile information
func (u *UserAggregate) UpdateProfile(updatedFields map[string]interface{}, updatedBy string) error {
	if u.Status == UserStatusSuspended {
		return fmt.Errorf("cannot update profile for suspended user")
	}

	event := NewUserProfileUpdated(u.ID, updatedFields, updatedBy)
	u.AddEvent(event)
	return u.ApplyEvent(event)
}

// ApplyEvent applies an event to the aggregate
func (u *UserAggregate) ApplyEvent(event events.DomainEvent) error {
	switch e := event.(type) {
	case *UserRegistered:
		return u.applyUserRegistered(e)
	case *AccreditationSubmitted:
		return u.applyAccreditationSubmitted(e)
	case *AccreditationVerified:
		return u.applyAccreditationVerified(e)
	case *AccreditationRevoked:
		return u.applyAccreditationRevoked(e)
	case *ComplianceCheckPerformed:
		return u.applyComplianceCheckPerformed(e)
	case *UserSuspended:
		return u.applyUserSuspended(e)
	case *UserReinstated:
		return u.applyUserReinstated(e)
	case *UserProfileUpdated:
		return u.applyUserProfileUpdated(e)
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// LoadFromHistory loads the aggregate from a sequence of events
func (u *UserAggregate) LoadFromHistory(events []events.DomainEvent) error {
	for _, event := range events {
		if err := u.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.GetEventType(), err)
		}
		u.IncrementVersion()
	}
	return nil
}

// Event application methods

func (u *UserAggregate) applyUserRegistered(event *UserRegistered) error {
	u.Email = event.Email
	u.FirstName = event.FirstName
	u.LastName = event.LastName
	u.PasswordHash = event.PasswordHash
	u.CreatedAt = event.Timestamp
	u.Status = UserStatusActive
	
	// Only set up accreditation if provided
	if event.AccreditationType != "" || len(event.AccreditationDetails) > 0 {
		u.Accreditation.Type = event.AccreditationType
		u.Accreditation.Details = event.AccreditationDetails
		u.Accreditation.Status = AccreditationStatusPending
	}
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyAccreditationSubmitted(event *AccreditationSubmitted) error {
	u.Accreditation.Type = event.AccreditationType
	u.Accreditation.Documents = event.Documents
	u.Accreditation.Details = event.SubmissionDetails
	u.Accreditation.Status = AccreditationStatusPending
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyAccreditationVerified(event *AccreditationVerified) error {
	u.Accreditation.Status = AccreditationStatusVerified
	u.Accreditation.ValidUntil = &event.ValidUntil
	u.Accreditation.VerifiedBy = event.VerifiedBy
	u.Accreditation.VerifiedAt = &event.Timestamp
	u.Accreditation.Notes = event.VerificationNotes
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyAccreditationRevoked(event *AccreditationRevoked) error {
	u.Accreditation.Status = AccreditationStatusRevoked
	u.Accreditation.ValidUntil = nil
	u.Accreditation.Notes = event.Reason
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyComplianceCheckPerformed(event *ComplianceCheckPerformed) error {
	now := event.Timestamp
	u.Compliance.LastCheck = &now
	
	if event.NextReview != nil {
		u.Compliance.NextReview = event.NextReview
	}
	
	// Update specific compliance status based on check type
	status := ComplianceStatus(event.Status)
	switch event.CheckType {
	case "kyc":
		u.Compliance.KYCStatus = status
	case "aml":
		u.Compliance.AMLStatus = status
	case "sanctions":
		u.Compliance.SanctionsStatus = status
	}
	
	// Update overall status based on individual statuses
	u.updateOverallComplianceStatus()
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyUserSuspended(event *UserSuspended) error {
	u.Status = UserStatusSuspended
	u.SuspendedAt = &event.Timestamp
	u.SuspensionUntil = event.Duration
	u.SuspensionReason = event.Reason
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyUserReinstated(event *UserReinstated) error {
	u.Status = UserStatusActive
	u.SuspendedAt = nil
	u.SuspensionUntil = nil
	u.SuspensionReason = ""
	
	u.IncrementVersion()
	return nil
}

func (u *UserAggregate) applyUserProfileUpdated(event *UserProfileUpdated) error {
	for field, value := range event.UpdatedFields {
		switch field {
		case "firstName":
			if v, ok := value.(string); ok {
				u.FirstName = v
			}
		case "lastName":
			if v, ok := value.(string); ok {
				u.LastName = v
			}
		case "email":
			if v, ok := value.(string); ok {
				u.Email = v
			}
		}
	}
	
	u.IncrementVersion()
	return nil
}

// Helper methods

func (u *UserAggregate) updateOverallComplianceStatus() {
	if u.Compliance.KYCStatus == ComplianceStatusClear && 
	   u.Compliance.AMLStatus == ComplianceStatusClear && 
	   u.Compliance.SanctionsStatus == ComplianceStatusClear {
		u.Compliance.OverallStatus = ComplianceStatusClear
	} else if u.Compliance.KYCStatus == ComplianceStatusBlocked || 
	          u.Compliance.AMLStatus == ComplianceStatusBlocked || 
	          u.Compliance.SanctionsStatus == ComplianceStatusBlocked {
		u.Compliance.OverallStatus = ComplianceStatusBlocked
	} else if u.Compliance.KYCStatus == ComplianceStatusReview || 
	          u.Compliance.AMLStatus == ComplianceStatusReview || 
	          u.Compliance.SanctionsStatus == ComplianceStatusReview {
		u.Compliance.OverallStatus = ComplianceStatusReview
	} else {
		u.Compliance.OverallStatus = ComplianceStatusPending
	}
}

// IsAccredited returns true if the user is currently accredited
func (u *UserAggregate) IsAccredited() bool {
	if u.Accreditation.Status != AccreditationStatusVerified {
		return false
	}
	
	if u.Accreditation.ValidUntil != nil && time.Now().After(*u.Accreditation.ValidUntil) {
		return false
	}
	
	return true
}

// IsCompliant returns true if the user is compliant for trading
func (u *UserAggregate) IsCompliant() bool {
	return u.Compliance.OverallStatus == ComplianceStatusClear
}

// CanTrade returns true if the user can participate in trading
func (u *UserAggregate) CanTrade() bool {
	return u.Status == UserStatusActive && u.IsAccredited() && u.IsCompliant()
}