package projections

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/users"
)

// ComplianceProjection maintains read models for compliance records
type ComplianceProjection struct {
	db *sql.DB
}

// NewComplianceProjection creates a new compliance projection
func NewComplianceProjection(db *sql.DB) *ComplianceProjection {
	return &ComplianceProjection{
		db: db,
	}
}

// Handle processes domain events to update compliance read models
func (p *ComplianceProjection) Handle(event events.DomainEvent) error {
	switch e := event.(type) {
	case *users.UserRegistered:
		return p.handleUserRegistered(e)
	case *users.ComplianceCheckPerformed:
		return p.handleComplianceCheckPerformed(e)
	case *users.UserSuspended:
		return p.handleUserSuspended(e)
	case *users.UserReinstated:
		return p.handleUserReinstated(e)
	default:
		// Ignore events we don't handle
		return nil
	}
}

// GetProjectionName returns the name of this projection
func (p *ComplianceProjection) GetProjectionName() string {
	return "compliance_projection"
}

// handleUserRegistered creates a new compliance record
func (p *ComplianceProjection) handleUserRegistered(event *users.UserRegistered) error {
	recordID := fmt.Sprintf("compliance-%s", event.AggregateID)
	
	query := `
		INSERT INTO compliance_records (
			record_id, user_id, risk_score, risk_factors,
			overall_status, watchlist_status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	riskFactors, _ := json.Marshal([]string{})

	_, err := p.db.Exec(query,
		recordID,
		event.AggregateID,
		0,              // initial risk score
		riskFactors,    // empty risk factors
		"clear",        // initial status
		"none",         // watchlist status
		event.Timestamp,
		event.Timestamp,
	)

	return err
}

// handleComplianceCheckPerformed updates compliance records
func (p *ComplianceProjection) handleComplianceCheckPerformed(event *users.ComplianceCheckPerformed) error {
	// Update the main compliance record
	query := `
		UPDATE compliance_records 
		SET last_risk_assessment = $1,
		    updated_at = $2
		WHERE user_id = $3
	`

	_, err := p.db.Exec(query,
		event.Timestamp,
		event.Timestamp,
		event.AggregateID,
	)

	if err != nil {
		return err
	}

	// Update specific compliance fields based on check type
	switch event.CheckType {
	case "kyc":
		if event.Status == "clear" {
			_, err = p.db.Exec(
				"UPDATE compliance_records SET kyc_completed_at = $1 WHERE user_id = $2",
				event.Timestamp,
				event.AggregateID,
			)
		}
	case "aml":
		if event.Status == "clear" {
			_, err = p.db.Exec(
				"UPDATE compliance_records SET aml_completed_at = $1 WHERE user_id = $2",
				event.Timestamp,
				event.AggregateID,
			)
		}
	}

	if err != nil {
		return err
	}

	// Update next review date if provided
	if event.NextReview != nil {
		_, err = p.db.Exec(
			"UPDATE compliance_records SET next_review_due = $1 WHERE user_id = $2",
			*event.NextReview,
			event.AggregateID,
		)
	}

	// Calculate and update overall status
	return p.updateOverallComplianceStatus(event.AggregateID, event.Status)
}

// handleUserSuspended updates compliance status when user is suspended
func (p *ComplianceProjection) handleUserSuspended(event *users.UserSuspended) error {
	// Add to watchlist and update status
	query := `
		UPDATE compliance_records 
		SET overall_status = 'blocked',
		    watchlist_status = 'suspended',
		    watchlist_reason = $1,
		    watchlist_added_at = $2,
		    updated_at = $3
		WHERE user_id = $4
	`

	_, err := p.db.Exec(query,
		event.Reason,
		event.Timestamp,
		event.Timestamp,
		event.AggregateID,
	)

	return err
}

// handleUserReinstated updates compliance status when user is reinstated
func (p *ComplianceProjection) handleUserReinstated(event *users.UserReinstated) error {
	// Remove from watchlist and restore status
	query := `
		UPDATE compliance_records 
		SET watchlist_status = 'none',
		    watchlist_reason = NULL,
		    watchlist_added_at = NULL,
		    updated_at = $1
		WHERE user_id = $2
	`

	_, err := p.db.Exec(query,
		event.Timestamp,
		event.AggregateID,
	)

	if err != nil {
		return err
	}

	// Recalculate overall status
	return p.updateOverallComplianceStatus(event.AggregateID, "clear")
}

// updateOverallComplianceStatus calculates and updates overall compliance status
func (p *ComplianceProjection) updateOverallComplianceStatus(userID, latestStatus string) error {
	// Get current KYC and AML completion status
	var kycCompleted, amlCompleted sql.NullTime
	query := `
		SELECT kyc_completed_at, aml_completed_at 
		FROM compliance_records 
		WHERE user_id = $1
	`

	err := p.db.QueryRow(query, userID).Scan(&kycCompleted, &amlCompleted)
	if err != nil {
		return err
	}

	// Determine overall status
	var overallStatus string
	if kycCompleted.Valid && amlCompleted.Valid {
		overallStatus = "clear"
	} else if latestStatus == "blocked" {
		overallStatus = "blocked"
	} else if latestStatus == "review" {
		overallStatus = "review"
	} else {
		overallStatus = "clear"
	}

	// Update overall status
	_, err = p.db.Exec(
		"UPDATE compliance_records SET overall_status = $1 WHERE user_id = $2",
		overallStatus,
		userID,
	)

	return err
}

// GetComplianceRecord retrieves a compliance record by user ID
func (p *ComplianceProjection) GetComplianceRecord(userID string) (*ComplianceRecord, error) {
	query := `
		SELECT record_id, user_id, risk_score, risk_factors,
		       last_risk_assessment, overall_status,
		       kyc_completed_at, aml_completed_at, next_review_due,
		       watchlist_status, watchlist_reason, watchlist_added_at,
		       created_at, updated_at
		FROM compliance_records
		WHERE user_id = $1
	`

	var record ComplianceRecord
	var riskFactors []byte
	var lastRiskAssessment, kycCompleted, amlCompleted, nextReviewDue sql.NullTime
	var watchlistReason sql.NullString
	var watchlistAddedAt sql.NullTime

	err := p.db.QueryRow(query, userID).Scan(
		&record.RecordID,
		&record.UserID,
		&record.RiskScore,
		&riskFactors,
		&lastRiskAssessment,
		&record.OverallStatus,
		&kycCompleted,
		&amlCompleted,
		&nextReviewDue,
		&record.WatchlistStatus,
		&watchlistReason,
		&watchlistAddedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if lastRiskAssessment.Valid {
		record.LastRiskAssessment = &lastRiskAssessment.Time
	}
	if kycCompleted.Valid {
		record.KYCCompletedAt = &kycCompleted.Time
	}
	if amlCompleted.Valid {
		record.AMLCompletedAt = &amlCompleted.Time
	}
	if nextReviewDue.Valid {
		record.NextReviewDue = &nextReviewDue.Time
	}
	if watchlistReason.Valid {
		record.WatchlistReason = watchlistReason.String
	}
	if watchlistAddedAt.Valid {
		record.WatchlistAddedAt = &watchlistAddedAt.Time
	}

	// Unmarshal risk factors
	if len(riskFactors) > 0 {
		json.Unmarshal(riskFactors, &record.RiskFactors)
	}

	return &record, nil
}

// ListComplianceRecords retrieves compliance records with filters
func (p *ComplianceProjection) ListComplianceRecords(limit, offset int, filters map[string]interface{}) ([]*ComplianceRecord, error) {
	baseQuery := `
		SELECT record_id, user_id, risk_score, overall_status,
		       last_risk_assessment, kyc_completed_at, aml_completed_at,
		       watchlist_status, created_at, updated_at
		FROM compliance_records
	`

	// Build WHERE clause based on filters
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if status, ok := filters["status"]; ok {
		whereClause = fmt.Sprintf(" WHERE overall_status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if watchlist, ok := filters["watchlist"]; ok {
		if whereClause == "" {
			whereClause = fmt.Sprintf(" WHERE watchlist_status = $%d", argIndex)
		} else {
			whereClause += fmt.Sprintf(" AND watchlist_status = $%d", argIndex)
		}
		args = append(args, watchlist)
		argIndex++
	}

	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*ComplianceRecord
	for rows.Next() {
		var record ComplianceRecord
		var lastRiskAssessment, kycCompleted, amlCompleted sql.NullTime

		err := rows.Scan(
			&record.RecordID,
			&record.UserID,
			&record.RiskScore,
			&record.OverallStatus,
			&lastRiskAssessment,
			&kycCompleted,
			&amlCompleted,
			&record.WatchlistStatus,
			&record.CreatedAt,
			&record.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if lastRiskAssessment.Valid {
			record.LastRiskAssessment = &lastRiskAssessment.Time
		}
		if kycCompleted.Valid {
			record.KYCCompletedAt = &kycCompleted.Time
		}
		if amlCompleted.Valid {
			record.AMLCompletedAt = &amlCompleted.Time
		}

		records = append(records, &record)
	}

	return records, rows.Err()
}

// GetComplianceStatistics returns compliance statistics
func (p *ComplianceProjection) GetComplianceStatistics() (*ComplianceStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN overall_status = 'clear' THEN 1 END) as compliant_users,
			COUNT(CASE WHEN overall_status = 'pending' THEN 1 END) as pending_users,
			COUNT(CASE WHEN overall_status = 'review' THEN 1 END) as review_users,
			COUNT(CASE WHEN overall_status = 'blocked' THEN 1 END) as blocked_users,
			COUNT(CASE WHEN watchlist_status != 'none' THEN 1 END) as watchlist_users,
			AVG(risk_score) as avg_risk_score
		FROM compliance_records
	`

	var stats ComplianceStatistics
	var avgRiskScore sql.NullFloat64

	err := p.db.QueryRow(query).Scan(
		&stats.TotalUsers,
		&stats.CompliantUsers,
		&stats.PendingUsers,
		&stats.ReviewUsers,
		&stats.BlockedUsers,
		&stats.WatchlistUsers,
		&avgRiskScore,
	)

	if err != nil {
		return nil, err
	}

	if avgRiskScore.Valid {
		stats.AverageRiskScore = avgRiskScore.Float64
	}

	return &stats, nil
}

// ComplianceRecord represents the read model for compliance records
type ComplianceRecord struct {
	RecordID             string    `json:"recordId"`
	UserID               string    `json:"userId"`
	RiskScore            int       `json:"riskScore"`
	RiskFactors          []string  `json:"riskFactors"`
	LastRiskAssessment   *time.Time `json:"lastRiskAssessment,omitempty"`
	OverallStatus        string    `json:"overallStatus"`
	KYCCompletedAt       *time.Time `json:"kycCompletedAt,omitempty"`
	AMLCompletedAt       *time.Time `json:"amlCompletedAt,omitempty"`
	NextReviewDue        *time.Time `json:"nextReviewDue,omitempty"`
	WatchlistStatus      string    `json:"watchlistStatus"`
	WatchlistReason      string    `json:"watchlistReason,omitempty"`
	WatchlistAddedAt     *time.Time `json:"watchlistAddedAt,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// ComplianceStatistics provides summary statistics for compliance
type ComplianceStatistics struct {
	TotalUsers        int     `json:"totalUsers"`
	CompliantUsers    int     `json:"compliantUsers"`
	PendingUsers      int     `json:"pendingUsers"`
	ReviewUsers       int     `json:"reviewUsers"`
	BlockedUsers      int     `json:"blockedUsers"`
	WatchlistUsers    int     `json:"watchlistUsers"`
	AverageRiskScore  float64 `json:"averageRiskScore"`
}