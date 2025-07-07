package users

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"github.com/google/uuid"
	
	"securities-marketplace/domains/shared/events"
)

// UserService provides application services for user domain
type UserService struct {
	repository UserRepository
	eventStore events.EventStore
	eventBus   events.EventBus
}

// NewUserService creates a new user service
func NewUserService(repository UserRepository, eventStore events.EventStore, eventBus events.EventBus) *UserService {
	return &UserService{
		repository: repository,
		eventStore: eventStore,
		eventBus:   eventBus,
	}
}

// RegisterUser handles user registration
func (s *UserService) RegisterUser(cmd *RegisterUserCommand) (*UserAggregate, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.repository.FindByEmail(cmd.Email)
	if err != nil && !IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", cmd.Email)
	}

	// Hash password
	passwordHash, err := s.hashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user aggregate
	userID := cmd.UserID
	if userID == "" {
		userID = uuid.New().String()
	}

	user := NewUserAggregate(userID)
	err = user.Register(cmd.Email, cmd.FirstName, cmd.LastName, passwordHash, cmd.AccreditationType, cmd.AccreditationDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Save events
	err = s.saveAggregateEvents(user, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to save user events: %w", err)
	}

	return user, nil
}

// SubmitAccreditation handles accreditation submission
func (s *UserService) SubmitAccreditation(cmd *SubmitAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.SubmitAccreditation(cmd.AccreditationType, cmd.Documents, cmd.SubmissionDetails)
	if err != nil {
		return fmt.Errorf("failed to submit accreditation: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.UserID)
}

// VerifyAccreditation handles accreditation verification
func (s *UserService) VerifyAccreditation(cmd *VerifyAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.VerifyAccreditation(cmd.AccreditationType, cmd.ValidUntil, cmd.VerifiedBy, cmd.VerificationNotes)
	if err != nil {
		return fmt.Errorf("failed to verify accreditation: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.VerifiedBy)
}

// RevokeAccreditation handles accreditation revocation
func (s *UserService) RevokeAccreditation(cmd *RevokeAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.RevokeAccreditation(cmd.Reason, cmd.RevokedBy)
	if err != nil {
		return fmt.Errorf("failed to revoke accreditation: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.RevokedBy)
}

// PerformComplianceCheck handles compliance check
func (s *UserService) PerformComplianceCheck(cmd *PerformComplianceCheckCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.PerformComplianceCheck(cmd.CheckType, cmd.Status, cmd.Results, cmd.PerformedBy, cmd.NextReview)
	if err != nil {
		return fmt.Errorf("failed to perform compliance check: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.PerformedBy)
}

// SuspendUser handles user suspension
func (s *UserService) SuspendUser(cmd *SuspendUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.Suspend(cmd.Reason, cmd.SuspendedBy, cmd.Duration)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.SuspendedBy)
}

// ReinstateUser handles user reinstatement
func (s *UserService) ReinstateUser(cmd *ReinstateUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.Reinstate(cmd.ReinstatedBy, cmd.Reason)
	if err != nil {
		return fmt.Errorf("failed to reinstate user: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.ReinstatedBy)
}

// UpdateUserProfile handles user profile updates
func (s *UserService) UpdateUserProfile(cmd *UpdateUserProfileCommand) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	err = user.UpdateProfile(cmd.UpdatedFields, cmd.UpdatedBy)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return s.saveAggregateEvents(user, cmd.UpdatedBy)
}

// AuthenticateUser handles user authentication
func (s *UserService) AuthenticateUser(cmd *AuthenticateUserCommand) (*UserAggregate, error) {
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.repository.FindByEmail(cmd.Email)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	valid, err := s.verifyPassword(cmd.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user can authenticate
	if user.Status == UserStatusSuspended {
		return nil, fmt.Errorf("user account is suspended")
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(userID string) (*UserAggregate, error) {
	return s.repository.FindByID(userID)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*UserAggregate, error) {
	return s.repository.FindByEmail(email)
}

// Password hashing and verification

func (s *UserService) hashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Hash password with Argon2
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode salt and hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", b64Salt, b64Hash), nil
}

func (s *UserService) verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version of argon2")
	}

	var memory, time, parallelism uint32
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, uint8(parallelism), uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}

// saveAggregateEvents saves uncommitted events from an aggregate
func (s *UserService) saveAggregateEvents(user *UserAggregate, userID string) error {
	uncommittedEvents := user.GetUncommittedEvents()
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

		event.AggregateVersion = user.GetVersion() + i + 1
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
	user.MarkEventsAsCommitted()

	return nil
}