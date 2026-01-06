package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/errors"
)

const (
	// VerificationTTL is the default time-to-live for verification requests.
	VerificationTTL = 10 * time.Minute
	// MaxVerificationAttempts is the maximum number of OTP verification attempts.
	MaxVerificationAttempts = 5
	// VerificationRateLimit is the minimum time between verification requests of same type.
	VerificationRateLimit = 1 * time.Minute
)

// VerificationRepository handles verification request database operations.
type VerificationRepository struct {
	db *database.DB
}

// NewVerificationRepository creates a new verification repository.
func NewVerificationRepository(db *database.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// Create creates a new verification request.
func (r *VerificationRepository) Create(ctx context.Context, req *models.VerificationRequest) *errors.Error {
	query := `
		INSERT INTO verification_requests
			(id, user_id, operation_type, otp_code, status, metadata, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.UserID,
		req.OperationType,
		req.OTPCode,
		req.Status,
		req.Metadata,
		req.ExpiresAt.Time,
		req.CreatedAt.Time,
	)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to create verification request")
	}
	return nil
}

// GetByID retrieves a verification request by ID.
func (r *VerificationRepository) GetByID(ctx context.Context, id string) (*models.VerificationRequest, *errors.Error) {
	query := `
		SELECT id, user_id, operation_type, otp_code, status, metadata,
		       expires_at, created_at, verified_at, attempt_count, last_attempt_at
		FROM verification_requests
		WHERE id = $1
	`
	req := &models.VerificationRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.UserID,
		&req.OperationType,
		&req.OTPCode,
		&req.Status,
		&req.Metadata,
		&req.ExpiresAt,
		&req.CreatedAt,
		&req.VerifiedAt,
		&req.AttemptCount,
		&req.LastAttemptAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("verification request not found")
	}
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get verification request")
	}
	return req, nil
}

// GetPendingByUserID retrieves all pending verifications for a user.
func (r *VerificationRepository) GetPendingByUserID(ctx context.Context, userID string) ([]*models.VerificationRequest, *errors.Error) {
	// First, expire old requests
	r.expireOldRequests(ctx)

	query := `
		SELECT id, user_id, operation_type, otp_code, status, metadata,
		       expires_at, created_at, verified_at, attempt_count, last_attempt_at
		FROM verification_requests
		WHERE user_id = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get pending verifications")
	}
	defer func() { _ = rows.Close() }()

	var requests []*models.VerificationRequest
	for rows.Next() {
		req := &models.VerificationRequest{}
		if err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.OperationType,
			&req.OTPCode,
			&req.Status,
			&req.Metadata,
			&req.ExpiresAt,
			&req.CreatedAt,
			&req.VerifiedAt,
			&req.AttemptCount,
			&req.LastAttemptAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan verification request")
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to iterate verification requests")
	}

	return requests, nil
}

// GetByUserID retrieves verifications for a user with optional status filter.
func (r *VerificationRepository) GetByUserID(ctx context.Context, userID string, status string, limit int) ([]*models.VerificationRequest, *errors.Error) {
	var query string
	var args []interface{}

	if status == "" || status == "all" {
		query = `
			SELECT id, user_id, operation_type, otp_code, status, metadata,
			       expires_at, created_at, verified_at, attempt_count, last_attempt_at
			FROM verification_requests
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{userID, limit}
	} else {
		query = `
			SELECT id, user_id, operation_type, otp_code, status, metadata,
			       expires_at, created_at, verified_at, attempt_count, last_attempt_at
			FROM verification_requests
			WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{userID, status, limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get user verifications")
	}
	defer func() { _ = rows.Close() }()

	var requests []*models.VerificationRequest
	for rows.Next() {
		req := &models.VerificationRequest{}
		if err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.OperationType,
			&req.OTPCode,
			&req.Status,
			&req.Metadata,
			&req.ExpiresAt,
			&req.CreatedAt,
			&req.VerifiedAt,
			&req.AttemptCount,
			&req.LastAttemptAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan verification request")
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to iterate verification requests")
	}

	return requests, nil
}

// UpdateStatus updates the status of a verification request.
func (r *VerificationRepository) UpdateStatus(ctx context.Context, id string, status models.VerificationStatus) *errors.Error {
	var query string
	if status == models.VerificationStatusVerified {
		query = `
			UPDATE verification_requests
			SET status = $2, verified_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE verification_requests
			SET status = $2
			WHERE id = $1
		`
	}
	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to update verification status")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("verification request not found")
	}
	return nil
}

// IncrementAttempts increments the attempt counter and returns the new count.
func (r *VerificationRepository) IncrementAttempts(ctx context.Context, id string) (int, *errors.Error) {
	query := `
		UPDATE verification_requests
		SET attempt_count = attempt_count + 1, last_attempt_at = NOW()
		WHERE id = $1
		RETURNING attempt_count
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, errors.NotFound("verification request not found")
	}
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to increment attempts")
	}
	return count, nil
}

// CancelPendingForUser cancels all pending verifications for a user.
func (r *VerificationRepository) CancelPendingForUser(ctx context.Context, userID string) *errors.Error {
	query := `
		UPDATE verification_requests
		SET status = 'cancelled'
		WHERE user_id = $1 AND status = 'pending'
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to cancel pending verifications")
	}
	return nil
}

// HasRecentVerification checks if user has a recent verification of same type.
func (r *VerificationRepository) HasRecentVerification(ctx context.Context, userID string, operationType models.OperationType, within time.Duration) (bool, *errors.Error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM verification_requests
			WHERE user_id = $1
			  AND operation_type = $2
			  AND status = 'pending'
			  AND created_at > $3
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, operationType, time.Now().Add(-within)).Scan(&exists)
	if err != nil {
		return false, errors.DatabaseWrap(err, "failed to check recent verification")
	}
	return exists, nil
}

// expireOldRequests marks expired requests as expired.
func (r *VerificationRepository) expireOldRequests(ctx context.Context) {
	query := `
		UPDATE verification_requests
		SET status = 'expired'
		WHERE status = 'pending' AND expires_at < NOW()
	`
	_, _ = r.db.ExecContext(ctx, query)
}

// CountPendingByUserID returns the count of pending verifications for a user.
func (r *VerificationRepository) CountPendingByUserID(ctx context.Context, userID string) (int, *errors.Error) {
	query := `
		SELECT COUNT(*) FROM verification_requests
		WHERE user_id = $1 AND status = 'pending' AND expires_at > NOW()
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to count pending verifications")
	}
	return count, nil
}
