package projections

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/users"
)

// UserProfileProjection maintains read models for user profiles
type UserProfileProjection struct {
	db *sql.DB
}

// NewUserProfileProjection creates a new user profile projection
func NewUserProfileProjection(db *sql.DB) *UserProfileProjection {
	return &UserProfileProjection{
		db: db,
	}
}

// Handle processes domain events to update read models
func (p *UserProfileProjection) Handle(event events.DomainEvent) error {
	switch e := event.(type) {
	case *users.UserRegistered:
		return p.handleUserRegistered(e)
	case *users.AccreditationSubmitted:
		return p.handleAccreditationSubmitted(e)
	case *users.AccreditationVerified:
		return p.handleAccreditationVerified(e)
	case *users.AccreditationRevoked:
		return p.handleAccreditationRevoked(e)
	case *users.ComplianceCheckPerformed:
		return p.handleComplianceCheckPerformed(e)
	case *users.UserSuspended:
		return p.handleUserSuspended(e)
	case *users.UserReinstated:
		return p.handleUserReinstated(e)
	case *users.UserProfileUpdated:
		return p.handleUserProfileUpdated(e)
	default:
		// Ignore events we don't handle
		return nil
	}
}

// GetProjectionName returns the name of this projection
func (p *UserProfileProjection) GetProjectionName() string {
	return "user_profile_projection"
}

// handleUserRegistered creates a new user profile record
func (p *UserProfileProjection) handleUserRegistered(event *users.UserRegistered) error {
	query := `
		INSERT INTO user_profiles (
			user_id, email, first_name, last_name,
			accreditation_status, accreditation_type,
			compliance_status, kyc_status, aml_status, sanctions_status,
			created_at, last_updated_at, risk_score, watchlist_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	accreditationDocs, _ := json.Marshal([]interface{}{})

	_, err := p.db.Exec(query,
		event.AggregateID,
		event.Email,
		event.FirstName,
		event.LastName,
		"pending",               // accreditation_status
		event.AccreditationType, // accreditation_type
		"clear",                 // compliance_status
		"pending",               // kyc_status
		"pending",               // aml_status
		"pending",               // sanctions_status
		event.Timestamp,         // created_at
		event.Timestamp,         // last_updated_at
		0,                       // risk_score
		"none",                  // watchlist_status
	)

	if err != nil {
		return fmt.Errorf("failed to insert user profile: %w", err)
	}

	// Also update accreditation_documents as empty array
	_, err = p.db.Exec(
		"UPDATE user_profiles SET accreditation_documents = $1 WHERE user_id = $2",
		accreditationDocs,
		event.AggregateID,
	)

	return err
}

// handleAccreditationSubmitted updates accreditation information
func (p *UserProfileProjection) handleAccreditationSubmitted(event *users.AccreditationSubmitted) error {
	documentsJSON, err := json.Marshal(event.Documents)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	query := `
		UPDATE user_profiles 
		SET accreditation_status = 'pending',
		    accreditation_type = $1,
		    accreditation_documents = $2,
		    last_updated_at = $3
		WHERE user_id = $4
	`

	_, err = p.db.Exec(query,
		event.AccreditationType,
		documentsJSON,
		event.Timestamp,
		event.AggregateID,
	)

	return err
}

// handleAccreditationVerified updates accreditation to verified status
func (p *UserProfileProjection) handleAccreditationVerified(event *users.AccreditationVerified) error {
	query := `
		UPDATE user_profiles 
		SET accreditation_status = 'verified',
		    accreditation_type = $1,
		    accreditation_valid_until = $2,
		    last_updated_at = $3
		WHERE user_id = $4
	`

	_, err := p.db.Exec(query,
		event.AccreditationType,
		event.ValidUntil,
		event.Timestamp,
		event.AggregateID,
	)

	return err
}

// handleAccreditationRevoked updates accreditation to revoked status
func (p *UserProfileProjection) handleAccreditationRevoked(event *users.AccreditationRevoked) error {
	query := `
		UPDATE user_profiles 
		SET accreditation_status = 'revoked',
		    accreditation_valid_until = NULL,
		    last_updated_at = $1
		WHERE user_id = $2
	`

	_, err := p.db.Exec(query,
		event.Timestamp,
		event.AggregateID,
	)

	return err
}

// handleComplianceCheckPerformed updates compliance status
func (p *UserProfileProjection) handleComplianceCheckPerformed(event *users.ComplianceCheckPerformed) error {
	// Update specific compliance status field based on check type
	var fieldName string
	switch event.CheckType {
	case "kyc":
		fieldName = "kyc_status"
	case "aml":
		fieldName = "aml_status"
	case "sanctions":
		fieldName = "sanctions_status"
	default:
		fieldName = "compliance_status"
	}

	query := fmt.Sprintf(`
		UPDATE user_profiles 
		SET %s = $1,
		    last_updated_at = $2
		WHERE user_id = $3
	`, fieldName)

	_, err := p.db.Exec(query,
		event.Status,
		event.Timestamp,
		event.AggregateID,
	)

	if err != nil {
		return err
	}

	// Update overall compliance status based on individual statuses
	return p.updateOverallComplianceStatus(event.AggregateID)
}

// handleUserSuspended updates user status to suspended
func (p *UserProfileProjection) handleUserSuspended(event *users.UserSuspended) error {
	query := `
		UPDATE user_profiles 
		SET compliance_status = 'blocked',
		    last_updated_at = $1
		WHERE user_id = $2
	`

	_, err := p.db.Exec(query,
		event.Timestamp,
		event.AggregateID,
	)

	return err
}

// handleUserReinstated updates user status back to active
func (p *UserProfileProjection) handleUserReinstated(event *users.UserReinstated) error {
	// First restore compliance status, then recalculate overall status
	_, err := p.db.Exec(
		"UPDATE user_profiles SET last_updated_at = $1 WHERE user_id = $2",
		event.Timestamp,
		event.AggregateID,
	)

	if err != nil {
		return err
	}

	return p.updateOverallComplianceStatus(event.AggregateID)
}

// handleUserProfileUpdated updates basic profile information
func (p *UserProfileProjection) handleUserProfileUpdated(event *users.UserProfileUpdated) error {
	// Build dynamic update query based on updated fields
	setParts := []string{"last_updated_at = $1"}
	args := []interface{}{event.Timestamp}
	argIndex := 2

	for field, value := range event.UpdatedFields {
		switch field {
		case "firstName":
			setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "lastName":
			setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "email":
			setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 1 {
		// No fields to update besides timestamp
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE user_profiles SET %s WHERE user_id = $%d",
		fmt.Sprintf("%s", setParts),
		argIndex,
	)
	args = append(args, event.AggregateID)

	_, err := p.db.Exec(query, args...)
	return err
}

// updateOverallComplianceStatus calculates and updates the overall compliance status
func (p *UserProfileProjection) updateOverallComplianceStatus(userID string) error {
	// Get current individual compliance statuses
	var kycStatus, amlStatus, sanctionsStatus string
	query := `
		SELECT kyc_status, aml_status, sanctions_status 
		FROM user_profiles 
		WHERE user_id = $1
	`

	err := p.db.QueryRow(query, userID).Scan(&kycStatus, &amlStatus, &sanctionsStatus)
	if err != nil {
		return err
	}

	// Calculate overall status
	var overallStatus string
	if kycStatus == "clear" && amlStatus == "clear" && sanctionsStatus == "clear" {
		overallStatus = "clear"
	} else if kycStatus == "blocked" || amlStatus == "blocked" || sanctionsStatus == "blocked" {
		overallStatus = "blocked"
	} else if kycStatus == "review" || amlStatus == "review" || sanctionsStatus == "review" {
		overallStatus = "review"
	} else {
		overallStatus = "pending"
	}

	// Update overall status
	_, err = p.db.Exec(
		"UPDATE user_profiles SET compliance_status = $1 WHERE user_id = $2",
		overallStatus,
		userID,
	)

	return err
}

// GetUserProfile retrieves a user profile by ID
func (p *UserProfileProjection) GetUserProfile(userID string) (*UserProfile, error) {
	query := `
		SELECT user_id, email, first_name, last_name,
		       accreditation_status, accreditation_type, accreditation_valid_until, accreditation_documents,
		       compliance_status, kyc_status, aml_status, sanctions_status,
		       created_at, last_updated_at, last_login_at,
		       risk_score, watchlist_status
		FROM user_profiles
		WHERE user_id = $1
	`

	var profile UserProfile
	var accreditationDocs []byte
	var validUntil, lastLogin sql.NullTime

	err := p.db.QueryRow(query, userID).Scan(
		&profile.UserID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.AccreditationStatus,
		&profile.AccreditationType,
		&validUntil,
		&accreditationDocs,
		&profile.ComplianceStatus,
		&profile.KYCStatus,
		&profile.AMLStatus,
		&profile.SanctionsStatus,
		&profile.CreatedAt,
		&profile.LastUpdatedAt,
		&lastLogin,
		&profile.RiskScore,
		&profile.WatchlistStatus,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if validUntil.Valid {
		profile.AccreditationValidUntil = &validUntil.Time
	}
	if lastLogin.Valid {
		profile.LastLoginAt = &lastLogin.Time
	}

	// Unmarshal documents
	if len(accreditationDocs) > 0 {
		json.Unmarshal(accreditationDocs, &profile.AccreditationDocuments)
	}

	return &profile, nil
}

// ListUserProfiles retrieves a paginated list of user profiles
func (p *UserProfileProjection) ListUserProfiles(limit, offset int, filters map[string]interface{}) ([]*UserProfile, error) {
	query := `
		SELECT user_id, email, first_name, last_name,
		       accreditation_status, accreditation_type, accreditation_valid_until,
		       compliance_status, kyc_status, aml_status, sanctions_status,
		       created_at, last_updated_at, last_login_at,
		       risk_score, watchlist_status
		FROM user_profiles
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := p.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*UserProfile
	for rows.Next() {
		var profile UserProfile
		var validUntil, lastLogin sql.NullTime

		err := rows.Scan(
			&profile.UserID,
			&profile.Email,
			&profile.FirstName,
			&profile.LastName,
			&profile.AccreditationStatus,
			&profile.AccreditationType,
			&validUntil,
			&profile.ComplianceStatus,
			&profile.KYCStatus,
			&profile.AMLStatus,
			&profile.SanctionsStatus,
			&profile.CreatedAt,
			&profile.LastUpdatedAt,
			&lastLogin,
			&profile.RiskScore,
			&profile.WatchlistStatus,
		)

		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if validUntil.Valid {
			profile.AccreditationValidUntil = &validUntil.Time
		}
		if lastLogin.Valid {
			profile.LastLoginAt = &lastLogin.Time
		}

		profiles = append(profiles, &profile)
	}

	return profiles, rows.Err()
}

// UserProfile represents the read model for user profiles
type UserProfile struct {
	UserID                     string                 `json:"userId"`
	Email                      string                 `json:"email"`
	FirstName                  string                 `json:"firstName"`
	LastName                   string                 `json:"lastName"`
	AccreditationStatus        string                 `json:"accreditationStatus"`
	AccreditationType          string                 `json:"accreditationType"`
	AccreditationValidUntil    *time.Time             `json:"accreditationValidUntil,omitempty"`
	AccreditationDocuments     []users.DocumentInfo   `json:"accreditationDocuments"`
	ComplianceStatus           string                 `json:"complianceStatus"`
	KYCStatus                  string                 `json:"kycStatus"`
	AMLStatus                  string                 `json:"amlStatus"`
	SanctionsStatus            string                 `json:"sanctionsStatus"`
	CreatedAt                  time.Time              `json:"createdAt"`
	LastUpdatedAt              time.Time              `json:"lastUpdatedAt"`
	LastLoginAt                *time.Time             `json:"lastLoginAt,omitempty"`
	RiskScore                  int                    `json:"riskScore"`
	WatchlistStatus            string                 `json:"watchlistStatus"`
}