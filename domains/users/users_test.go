package users

import (
	"testing"
	"time"

	"securities-marketplace/domains/shared/testutil"
)

func TestUserAggregate_RegisterUser(t *testing.T) {
	// Arrange
	user := NewTestUser()

	// Act
	accreditationDetails := map[string]string{
		"type":                 "individual",
		"netWorth":            "1000000",
		"annualIncome":        "200000",
		"investmentExperience": "experienced",
		"riskTolerance":       "moderate",
	}
	err := user.Register(
		"test@example.com",
		"John",
		"Doe",
		"password123",
		"individual",
		accreditationDetails,
	)

	// Assert
	testutil.AssertNoError(t, err, "User registration should succeed")
	testutil.AssertEqual(t, "test@example.com", user.Email, "Email should be set")
	testutil.AssertEqual(t, "John", user.FirstName, "First name should be set")
	testutil.AssertEqual(t, "Doe", user.LastName, "Last name should be set")
	testutil.AssertEqual(t, UserStatusActive, user.Status, "User should be active")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "UserRegistered", events[0].GetEventType(), "Should be UserRegistered event")
}

func TestUserAggregate_SubmitAccreditation(t *testing.T) {
	// Arrange
	user := NewTestUserWithEmail("test@example.com")
	user.MarkEventsAsCommitted() // Clear initial events

	submissionDetails := map[string]string{
		"type":                 "qualified_purchaser",
		"netWorth":            "5000000",
		"annualIncome":        "500000",
		"investmentExperience": "expert",
		"riskTolerance":       "high",
	}

	documents := []DocumentInfo{
		{
			Type:       "financial_statement",
			Name:       "statement.pdf",
			Size:       1024,
			Hash:       "abcd1234",
			UploadedAt: time.Now(),
		},
		{
			Type:       "tax_return",
			Name:       "taxes.pdf",
			Size:       2048,
			Hash:       "efgh5678",
			UploadedAt: time.Now(),
		},
	}

	// Act
	err := user.SubmitAccreditation("qualified_purchaser", documents, submissionDetails)

	// Assert
	testutil.AssertNoError(t, err, "Accreditation submission should succeed")
	testutil.AssertEqual(t, AccreditationStatusPending, user.Accreditation.Status, "Status should be pending")
	testutil.AssertNotNil(t, user.Accreditation.Details, "Accreditation details should be set")
	testutil.AssertLengthEqual(t, 2, user.Accreditation.Documents, "Should have two documents")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "AccreditationSubmitted", events[0].GetEventType(), "Should be AccreditationSubmitted event")
}

func TestUserAggregate_VerifyAccreditation(t *testing.T) {
	// Arrange
	user := NewTestUserWithEmail("test@example.com")
	submissionDetails := map[string]string{"type": "individual", "netWorth": "1000000"}
	user.SubmitAccreditation("individual", []DocumentInfo{{Type: "test", Name: "test.pdf", Size: 1024, Hash: "abc123", UploadedAt: time.Now()}}, submissionDetails)
	user.MarkEventsAsCommitted() // Clear initial events

	validUntil := time.Now().AddDate(1, 0, 0)

	// Act
	err := user.VerifyAccreditation("individual", validUntil, "admin-user", "Documents verified")

	// Assert
	testutil.AssertNoError(t, err, "Accreditation verification should succeed")
	testutil.AssertEqual(t, AccreditationStatusVerified, user.Accreditation.Status, "Status should be verified")
	testutil.AssertTimeEqual(t, validUntil, *user.Accreditation.ValidUntil, "Valid until date should be set")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "AccreditationVerified", events[0].GetEventType(), "Should be AccreditationVerified event")
}

func TestUserAggregate_RevokeAccreditation(t *testing.T) {
	// Arrange
	user := NewTestAccreditedUser()
	user.MarkEventsAsCommitted() // Clear initial events

	// Act
	err := user.RevokeAccreditation("compliance_violation", "compliance-officer")

	// Assert
	testutil.AssertNoError(t, err, "Accreditation revocation should succeed")
	testutil.AssertEqual(t, AccreditationStatusRevoked, user.Accreditation.Status, "Status should be revoked")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "AccreditationRevoked", events[0].GetEventType(), "Should be AccreditationRevoked event")
}

func TestUserAggregate_PerformComplianceCheck(t *testing.T) {
	// Arrange
	user := NewTestUserWithEmail("test@example.com")
	user.MarkEventsAsCommitted() // Clear initial events

	// Act
	err := user.PerformComplianceCheck("kyc", "clear", map[string]string{"result": "passed"}, "compliance-system", nil)

	// Assert
	testutil.AssertNoError(t, err, "Compliance check should succeed")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "ComplianceCheckPerformed", events[0].GetEventType(), "Should be ComplianceCheckPerformed event")
}

func TestUserAggregate_SuspendUser(t *testing.T) {
	// Arrange
	user := NewTestUserWithEmail("test@example.com")
	user.MarkEventsAsCommitted() // Clear initial events

	// Act
	err := user.Suspend("suspicious_activity", "security-team", nil)

	// Assert
	testutil.AssertNoError(t, err, "User suspension should succeed")
	testutil.AssertEqual(t, UserStatusSuspended, user.Status, "User should be suspended")

	// Check events
	events := user.GetUncommittedEvents()
	testutil.AssertLengthEqual(t, 1, events, "Should have one event")
	testutil.AssertEqual(t, "UserSuspended", events[0].GetEventType(), "Should be UserSuspended event")
}

func TestUserAggregate_ValidationRules(t *testing.T) {
	tests := []testutil.TestCase{
		{
			Name:          "empty email should fail",
			ExpectedError: true,
			SetupFunc: func() interface{} {
				user := NewUserAggregate("test-user")
				return func() error {
					return user.Register("", "John", "Doe", "password123", "individual", nil)
				}
			},
		},
		{
			Name:          "empty first name should fail",
			ExpectedError: true,
			SetupFunc: func() interface{} {
				user := NewUserAggregate("test-user")
				return func() error {
					return user.Register("test@example.com", "", "Doe", "password123", "individual", nil)
				}
			},
		},
		{
			Name:          "empty last name should fail",
			ExpectedError: true,
			SetupFunc: func() interface{} {
				user := NewUserAggregate("test-user")
				return func() error {
					return user.Register("test@example.com", "John", "", "password123", "individual", nil)
				}
			},
		},
		{
			Name:          "weak password should fail",
			ExpectedError: true,
			SetupFunc: func() interface{} {
				user := NewUserAggregate("test-user")
				return func() error {
					return user.Register("test@example.com", "John", "Doe", "123", "individual", nil)
				}
			},
		},
	}

	testutil.RunTableTests(t, tests, func(t *testing.T, tc testutil.TestCase) {
		fn := tc.SetupFunc().(func() error)
		err := fn()

		if tc.ExpectedError {
			testutil.AssertError(t, err, "Expected validation error")
		} else {
			testutil.AssertNoError(t, err, "Expected no error")
		}
	})
}

func TestUserAggregate_BusinessRules(t *testing.T) {
	t.Run("cannot verify accreditation without submission", func(t *testing.T) {
		// Arrange - create user without submitting accreditation
		user := NewTestUser()
		user.Register("test@example.com", "Test", "User", "password123", "", nil)

		// Act
		err := user.VerifyAccreditation("individual", time.Now().AddDate(1, 0, 0), "admin", "test verification")

		// Assert
		testutil.AssertError(t, err, "Should not allow verification without submission")
		if err != nil {
			testutil.AssertContains(t, err.Error(), "no accreditation", "Error should mention missing accreditation")
		}
	})

	t.Run("cannot revoke accreditation that is not verified", func(t *testing.T) {
		// Arrange
		user := NewTestUserWithEmail("test@example.com")
		user.SubmitAccreditation("individual", []DocumentInfo{{Type: "test", Name: "test.pdf", Size: 1024, Hash: "abc123", UploadedAt: time.Now()}}, map[string]string{"type": "individual"})

		// Act
		err := user.RevokeAccreditation("compliance_issue", "admin")

		// Assert
		testutil.AssertError(t, err, "Should not allow revocation of unverified accreditation")
	})

	t.Run("cannot register user twice", func(t *testing.T) {
		// Arrange
		user := NewTestUserWithEmail("test@example.com")

		// Act
		err := user.Register("another@example.com", "Jane", "Smith", "password456", "individual", nil)

		// Assert
		testutil.AssertError(t, err, "Should not allow double registration")
		testutil.AssertContains(t, err.Error(), "already exists", "Error should mention existing registration")
	})
}

// Benchmark tests for performance
func BenchmarkUserRegistration(b *testing.B) {
	testutil.BenchmarkFunction(b, func() {
		user := NewUserAggregate("test-user")
		user.Register("test@example.com", "John", "Doe", "password123", "individual", nil)
	})
}

func BenchmarkAccreditationSubmission(b *testing.B) {
	user := NewTestUserWithEmail("test@example.com")

	accreditationDetails := map[string]string{
		"type":                 "individual",
		"netWorth":            "1000000",
		"annualIncome":        "200000",
		"investmentExperience": "experienced",
		"riskTolerance":       "moderate",
	}

	documents := []DocumentInfo{
		{
			Type:       "financial_statement",
			Name:       "statement.pdf",
			Size:       1024,
			Hash:       "abcd1234",
			UploadedAt: time.Now(),
		},
	}

	testutil.BenchmarkMemory(b, func() {
		user.SubmitAccreditation("individual", documents, accreditationDetails)
	})
}