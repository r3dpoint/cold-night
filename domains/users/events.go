package users

import (
	"encoding/json"
	"time"

	"securities-marketplace/domains/shared/events"
)

// User Domain Events

// UserRegistered event is emitted when a new user registers
type UserRegistered struct {
	events.BaseEvent
	Email                string            `json:"email"`
	FirstName            string            `json:"firstName"`
	LastName             string            `json:"lastName"`
	PasswordHash         string            `json:"passwordHash"`
	AccreditationType    string            `json:"accreditationType"`
	AccreditationDetails map[string]string `json:"accreditationDetails"`
}

// NewUserRegistered creates a new UserRegistered event
func NewUserRegistered(userID, email, firstName, lastName, passwordHash, accreditationType string, accreditationDetails map[string]string) *UserRegistered {
	return &UserRegistered{
		BaseEvent:            events.NewBaseEvent(userID, "User"),
		Email:                email,
		FirstName:            firstName,
		LastName:             lastName,
		PasswordHash:         passwordHash,
		AccreditationType:    accreditationType,
		AccreditationDetails: accreditationDetails,
	}
}

func (e *UserRegistered) GetEventType() string { return "UserRegistered" }
func (e *UserRegistered) GetAggregateID() string { return e.AggregateID }
func (e *UserRegistered) GetAggregateType() string { return e.AggregateType }

func (e *UserRegistered) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserRegistered) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// AccreditationSubmitted event is emitted when user submits accreditation documents
type AccreditationSubmitted struct {
	events.BaseEvent
	AccreditationType string            `json:"accreditationType"`
	Documents         []DocumentInfo    `json:"documents"`
	SubmissionDetails map[string]string `json:"submissionDetails"`
}

type DocumentInfo struct {
	DocumentID   string    `json:"documentId"`
	DocumentType string    `json:"documentType"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	ContentHash  string    `json:"contentHash"`
	UploadedAt   time.Time `json:"uploadedAt"`
}

func NewAccreditationSubmitted(userID, accreditationType string, documents []DocumentInfo, submissionDetails map[string]string) *AccreditationSubmitted {
	return &AccreditationSubmitted{
		BaseEvent:         events.NewBaseEvent(userID, "User"),
		AccreditationType: accreditationType,
		Documents:         documents,
		SubmissionDetails: submissionDetails,
	}
}

func (e *AccreditationSubmitted) GetEventType() string { return "AccreditationSubmitted" }
func (e *AccreditationSubmitted) GetAggregateID() string { return e.AggregateID }
func (e *AccreditationSubmitted) GetAggregateType() string { return e.AggregateType }

func (e *AccreditationSubmitted) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *AccreditationSubmitted) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// AccreditationVerified event is emitted when accreditation is approved
type AccreditationVerified struct {
	events.BaseEvent
	AccreditationType string    `json:"accreditationType"`
	ValidUntil        time.Time `json:"validUntil"`
	VerifiedBy        string    `json:"verifiedBy"`
	VerificationNotes string    `json:"verificationNotes"`
}

func NewAccreditationVerified(userID, accreditationType string, validUntil time.Time, verifiedBy, verificationNotes string) *AccreditationVerified {
	return &AccreditationVerified{
		BaseEvent:         events.NewBaseEvent(userID, "User"),
		AccreditationType: accreditationType,
		ValidUntil:        validUntil,
		VerifiedBy:        verifiedBy,
		VerificationNotes: verificationNotes,
	}
}

func (e *AccreditationVerified) GetEventType() string { return "AccreditationVerified" }
func (e *AccreditationVerified) GetAggregateID() string { return e.AggregateID }
func (e *AccreditationVerified) GetAggregateType() string { return e.AggregateType }

func (e *AccreditationVerified) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *AccreditationVerified) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// AccreditationRevoked event is emitted when accreditation is revoked
type AccreditationRevoked struct {
	events.BaseEvent
	Reason    string `json:"reason"`
	RevokedBy string `json:"revokedBy"`
}

func NewAccreditationRevoked(userID, reason, revokedBy string) *AccreditationRevoked {
	return &AccreditationRevoked{
		BaseEvent: events.NewBaseEvent(userID, "User"),
		Reason:    reason,
		RevokedBy: revokedBy,
	}
}

func (e *AccreditationRevoked) GetEventType() string { return "AccreditationRevoked" }
func (e *AccreditationRevoked) GetAggregateID() string { return e.AggregateID }
func (e *AccreditationRevoked) GetAggregateType() string { return e.AggregateType }

func (e *AccreditationRevoked) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *AccreditationRevoked) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// ComplianceCheckPerformed event is emitted when compliance check is completed
type ComplianceCheckPerformed struct {
	events.BaseEvent
	CheckType   string            `json:"checkType"`
	Status      string            `json:"status"`
	Results     map[string]string `json:"results"`
	PerformedBy string            `json:"performedBy"`
	NextReview  *time.Time        `json:"nextReview,omitempty"`
}

func NewComplianceCheckPerformed(userID, checkType, status string, results map[string]string, performedBy string, nextReview *time.Time) *ComplianceCheckPerformed {
	return &ComplianceCheckPerformed{
		BaseEvent:   events.NewBaseEvent(userID, "User"),
		CheckType:   checkType,
		Status:      status,
		Results:     results,
		PerformedBy: performedBy,
		NextReview:  nextReview,
	}
}

func (e *ComplianceCheckPerformed) GetEventType() string { return "ComplianceCheckPerformed" }
func (e *ComplianceCheckPerformed) GetAggregateID() string { return e.AggregateID }
func (e *ComplianceCheckPerformed) GetAggregateType() string { return e.AggregateType }

func (e *ComplianceCheckPerformed) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ComplianceCheckPerformed) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// UserSuspended event is emitted when user account is suspended
type UserSuspended struct {
	events.BaseEvent
	Reason      string     `json:"reason"`
	SuspendedBy string     `json:"suspendedBy"`
	Duration    *time.Time `json:"duration,omitempty"` // nil for indefinite suspension
}

func NewUserSuspended(userID, reason, suspendedBy string, duration *time.Time) *UserSuspended {
	return &UserSuspended{
		BaseEvent:   events.NewBaseEvent(userID, "User"),
		Reason:      reason,
		SuspendedBy: suspendedBy,
		Duration:    duration,
	}
}

func (e *UserSuspended) GetEventType() string { return "UserSuspended" }
func (e *UserSuspended) GetAggregateID() string { return e.AggregateID }
func (e *UserSuspended) GetAggregateType() string { return e.AggregateType }

func (e *UserSuspended) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserSuspended) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// UserReinstated event is emitted when suspended user is reinstated
type UserReinstated struct {
	events.BaseEvent
	ReinstatedBy string `json:"reinstatedBy"`
	Reason       string `json:"reason"`
}

func NewUserReinstated(userID, reinstatedBy, reason string) *UserReinstated {
	return &UserReinstated{
		BaseEvent:    events.NewBaseEvent(userID, "User"),
		ReinstatedBy: reinstatedBy,
		Reason:       reason,
	}
}

func (e *UserReinstated) GetEventType() string { return "UserReinstated" }
func (e *UserReinstated) GetAggregateID() string { return e.AggregateID }
func (e *UserReinstated) GetAggregateType() string { return e.AggregateType }

func (e *UserReinstated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserReinstated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}

// UserProfileUpdated event is emitted when user profile is updated
type UserProfileUpdated struct {
	events.BaseEvent
	UpdatedFields map[string]interface{} `json:"updatedFields"`
	UpdatedBy     string                 `json:"updatedBy"`
}

func NewUserProfileUpdated(userID string, updatedFields map[string]interface{}, updatedBy string) *UserProfileUpdated {
	return &UserProfileUpdated{
		BaseEvent:     events.NewBaseEvent(userID, "User"),
		UpdatedFields: updatedFields,
		UpdatedBy:     updatedBy,
	}
}

func (e *UserProfileUpdated) GetEventType() string { return "UserProfileUpdated" }
func (e *UserProfileUpdated) GetAggregateID() string { return e.AggregateID }
func (e *UserProfileUpdated) GetAggregateType() string { return e.AggregateType }

func (e *UserProfileUpdated) GetEventData() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserProfileUpdated) GetMetadata() events.Metadata {
	return events.Metadata{
		EventID:       e.EventID,
		EventType:     e.GetEventType(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventVersion:  e.Version,
		Timestamp:     e.Timestamp,
	}
}