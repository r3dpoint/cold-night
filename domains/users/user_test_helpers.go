package users

import (
	"time"

	"github.com/google/uuid"
)

// Test helpers for users domain

// NewTestUser creates a user for testing
func NewTestUser() *UserAggregate {
	userID := uuid.New().String()
	user := NewUserAggregate(userID)
	return user
}

// NewTestUserWithEmail creates a user with email for testing
func NewTestUserWithEmail(email string) *UserAggregate {
	user := NewTestUser()
	accreditationDetails := map[string]string{
		"type":                 "individual",
		"netWorth":            "1000000",
		"annualIncome":        "200000",
		"investmentExperience": "experienced",
		"riskTolerance":       "moderate",
	}
	user.Register(email, "Test", "User", "password123", "individual", accreditationDetails)
	return user
}

// NewTestAccreditedUser creates an accredited user for testing
func NewTestAccreditedUser() *UserAggregate {
	user := NewTestUserWithEmail("test@example.com")
	user.MarkEventsAsCommitted()
	
	// Submit accreditation
	submissionDetails := map[string]string{
		"netWorth":            "1000000",
		"annualIncome":        "200000",
		"investmentExperience": "experienced",
		"riskTolerance":       "moderate",
	}
	user.SubmitAccreditation("individual", []DocumentInfo{
		{
			Type:       "financial_statement",
			Name:       "statement.pdf",
			Size:       1024,
			Hash:       "abcd1234",
			UploadedAt: time.Now(),
		},
	}, submissionDetails)
	user.MarkEventsAsCommitted()
	
	// Verify accreditation
	user.VerifyAccreditation("individual", time.Now().AddDate(1, 0, 0), "test-admin", "Documents verified")
	user.MarkEventsAsCommitted()
	
	return user
}

// Test constants
const (
	TestUserID    = "test-user-123"
	TestEmail     = "test@example.com"
	TestAdminID   = "admin-user-999"
)