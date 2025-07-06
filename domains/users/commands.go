package users

import (
	"time"
)

// User Domain Commands

// RegisterUserCommand represents a command to register a new user
type RegisterUserCommand struct {
	UserID               string            `json:"userId"`
	Email                string            `json:"email"`
	FirstName            string            `json:"firstName"`
	LastName             string            `json:"lastName"`
	Password             string            `json:"password"`
	AccreditationType    string            `json:"accreditationType"`
	AccreditationDetails map[string]string `json:"accreditationDetails"`
}

// SubmitAccreditationCommand represents a command to submit accreditation documents
type SubmitAccreditationCommand struct {
	UserID               string            `json:"userId"`
	AccreditationType    string            `json:"accreditationType"`
	Documents            []DocumentInfo    `json:"documents"`
	SubmissionDetails    map[string]string `json:"submissionDetails"`
}

// VerifyAccreditationCommand represents a command to verify user accreditation
type VerifyAccreditationCommand struct {
	UserID               string    `json:"userId"`
	AccreditationType    string    `json:"accreditationType"`
	ValidUntil           time.Time `json:"validUntil"`
	VerifiedBy           string    `json:"verifiedBy"`
	VerificationNotes    string    `json:"verificationNotes"`
}

// RevokeAccreditationCommand represents a command to revoke user accreditation
type RevokeAccreditationCommand struct {
	UserID    string `json:"userId"`
	Reason    string `json:"reason"`
	RevokedBy string `json:"revokedBy"`
}

// PerformComplianceCheckCommand represents a command to perform compliance check
type PerformComplianceCheckCommand struct {
	UserID      string            `json:"userId"`
	CheckType   string            `json:"checkType"`
	Status      string            `json:"status"`
	Results     map[string]string `json:"results"`
	PerformedBy string            `json:"performedBy"`
	NextReview  *time.Time        `json:"nextReview,omitempty"`
}

// SuspendUserCommand represents a command to suspend a user
type SuspendUserCommand struct {
	UserID      string     `json:"userId"`
	Reason      string     `json:"reason"`
	SuspendedBy string     `json:"suspendedBy"`
	Duration    *time.Time `json:"duration,omitempty"`
}

// ReinstateUserCommand represents a command to reinstate a suspended user
type ReinstateUserCommand struct {
	UserID       string `json:"userId"`
	ReinstatedBy string `json:"reinstatedBy"`
	Reason       string `json:"reason"`
}

// UpdateUserProfileCommand represents a command to update user profile
type UpdateUserProfileCommand struct {
	UserID        string                 `json:"userId"`
	UpdatedFields map[string]interface{} `json:"updatedFields"`
	UpdatedBy     string                 `json:"updatedBy"`
}

// AuthenticateUserCommand represents a command to authenticate a user
type AuthenticateUserCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordCommand represents a command to change user password
type ChangePasswordCommand struct {
	UserID      string `json:"userId"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// RequestPasswordResetCommand represents a command to request password reset
type RequestPasswordResetCommand struct {
	Email string `json:"email"`
}

// ResetPasswordCommand represents a command to reset password with token
type ResetPasswordCommand struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// Command validation methods

// Validate validates the RegisterUserCommand
func (c *RegisterUserCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.Email == "" {
		return NewValidationError("email", "Email is required")
	}
	if c.FirstName == "" {
		return NewValidationError("firstName", "First name is required")
	}
	if c.LastName == "" {
		return NewValidationError("lastName", "Last name is required")
	}
	if c.Password == "" {
		return NewValidationError("password", "Password is required")
	}
	if len(c.Password) < 8 {
		return NewValidationError("password", "Password must be at least 8 characters")
	}
	if c.AccreditationType == "" {
		return NewValidationError("accreditationType", "Accreditation type is required")
	}
	return nil
}

// Validate validates the SubmitAccreditationCommand
func (c *SubmitAccreditationCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.AccreditationType == "" {
		return NewValidationError("accreditationType", "Accreditation type is required")
	}
	if len(c.Documents) == 0 {
		return NewValidationError("documents", "At least one document is required")
	}
	return nil
}

// Validate validates the VerifyAccreditationCommand
func (c *VerifyAccreditationCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.AccreditationType == "" {
		return NewValidationError("accreditationType", "Accreditation type is required")
	}
	if c.ValidUntil.IsZero() {
		return NewValidationError("validUntil", "Valid until date is required")
	}
	if c.VerifiedBy == "" {
		return NewValidationError("verifiedBy", "Verified by is required")
	}
	return nil
}

// Validate validates the RevokeAccreditationCommand
func (c *RevokeAccreditationCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	if c.RevokedBy == "" {
		return NewValidationError("revokedBy", "Revoked by is required")
	}
	return nil
}

// Validate validates the PerformComplianceCheckCommand
func (c *PerformComplianceCheckCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.CheckType == "" {
		return NewValidationError("checkType", "Check type is required")
	}
	if c.Status == "" {
		return NewValidationError("status", "Status is required")
	}
	if c.PerformedBy == "" {
		return NewValidationError("performedBy", "Performed by is required")
	}
	return nil
}

// Validate validates the SuspendUserCommand
func (c *SuspendUserCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	if c.SuspendedBy == "" {
		return NewValidationError("suspendedBy", "Suspended by is required")
	}
	return nil
}

// Validate validates the ReinstateUserCommand
func (c *ReinstateUserCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.ReinstatedBy == "" {
		return NewValidationError("reinstatedBy", "Reinstated by is required")
	}
	if c.Reason == "" {
		return NewValidationError("reason", "Reason is required")
	}
	return nil
}

// Validate validates the UpdateUserProfileCommand
func (c *UpdateUserProfileCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if len(c.UpdatedFields) == 0 {
		return NewValidationError("updatedFields", "At least one field must be updated")
	}
	if c.UpdatedBy == "" {
		return NewValidationError("updatedBy", "Updated by is required")
	}
	return nil
}

// Validate validates the AuthenticateUserCommand
func (c *AuthenticateUserCommand) Validate() error {
	if c.Email == "" {
		return NewValidationError("email", "Email is required")
	}
	if c.Password == "" {
		return NewValidationError("password", "Password is required")
	}
	return nil
}

// Validate validates the ChangePasswordCommand
func (c *ChangePasswordCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("userId", "User ID is required")
	}
	if c.OldPassword == "" {
		return NewValidationError("oldPassword", "Old password is required")
	}
	if c.NewPassword == "" {
		return NewValidationError("newPassword", "New password is required")
	}
	if len(c.NewPassword) < 8 {
		return NewValidationError("newPassword", "New password must be at least 8 characters")
	}
	return nil
}

// Validate validates the RequestPasswordResetCommand
func (c *RequestPasswordResetCommand) Validate() error {
	if c.Email == "" {
		return NewValidationError("email", "Email is required")
	}
	return nil
}

// Validate validates the ResetPasswordCommand
func (c *ResetPasswordCommand) Validate() error {
	if c.Token == "" {
		return NewValidationError("token", "Token is required")
	}
	if c.NewPassword == "" {
		return NewValidationError("newPassword", "New password is required")
	}
	if len(c.NewPassword) < 8 {
		return NewValidationError("newPassword", "New password must be at least 8 characters")
	}
	return nil
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