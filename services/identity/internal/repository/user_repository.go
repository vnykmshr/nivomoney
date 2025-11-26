package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/errors"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, user *models.User) *errors.Error {
	query := `
		INSERT INTO users (email, phone, full_name, password_hash, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Phone,
		user.FullName,
		user.PasswordHash,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return errors.Conflict("user with this email or phone already exists")
		}
		return errors.DatabaseWrap(err, "failed to create user")
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, *errors.Error) {
	user := &models.User{}

	query := `
		SELECT id, email, phone, full_name, password_hash, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("user", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get user")
	}

	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, *errors.Error) {
	user := &models.User{}

	query := `
		SELECT id, email, phone, full_name, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("user")
		}
		return nil, errors.DatabaseWrap(err, "failed to get user by email")
	}

	return user, nil
}

// GetByPhone retrieves a user by phone number.
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, *errors.Error) {
	user := &models.User{}

	query := `
		SELECT id, email, phone, full_name, password_hash, status, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("user")
		}
		return nil, errors.DatabaseWrap(err, "failed to get user by phone")
	}

	return user, nil
}

// Update updates a user's information.
func (r *UserRepository) Update(ctx context.Context, user *models.User) *errors.Error {
	query := `
		UPDATE users
		SET email = $2, phone = $3, full_name = $4, status = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.FullName,
		user.Status,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("user", user.ID)
		}
		if database.IsUniqueViolation(err) {
			return errors.Conflict("user with this email or phone already exists")
		}
		return errors.DatabaseWrap(err, "failed to update user")
	}

	return nil
}

// UpdateStatus updates a user's status.
func (r *UserRepository) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) *errors.Error {
	query := `
		UPDATE users
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, status)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to update user status")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rows == 0 {
		return errors.NotFoundWithID("user", userID)
	}

	return nil
}

// Delete soft-deletes a user by setting status to closed.
func (r *UserRepository) Delete(ctx context.Context, userID string) *errors.Error {
	return r.UpdateStatus(ctx, userID, models.UserStatusClosed)
}

// List retrieves a paginated list of users.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, *errors.Error) {
	query := `
		SELECT id, email, phone, full_name, password_hash, status, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list users")
	}
	defer func() { _ = rows.Close() }()

	users := make([]*models.User, 0)

	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.FullName,
			&user.PasswordHash,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan user")
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating users")
	}

	return users, nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int, *errors.Error) {
	var count int
	query := `SELECT COUNT(*) FROM users`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to count users")
	}

	return count, nil
}

// CountByStatus returns the number of users with a specific status.
func (r *UserRepository) CountByStatus(ctx context.Context, status models.UserStatus) (int, *errors.Error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE status = $1`

	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to count users by status")
	}

	return count, nil
}

// KYCRepository handles database operations for KYC information.
type KYCRepository struct {
	db *database.DB
}

// NewKYCRepository creates a new KYC repository.
func NewKYCRepository(db *database.DB) *KYCRepository {
	return &KYCRepository{db: db}
}

// Create creates or updates KYC information for a user.
func (r *KYCRepository) Create(ctx context.Context, kyc *models.KYCInfo) *errors.Error {
	// Serialize address to JSONB
	addressJSON, err := json.Marshal(kyc.Address)
	if err != nil {
		return errors.BadRequest("invalid address format")
	}

	query := `
		INSERT INTO user_kyc (user_id, status, pan, aadhaar, date_of_birth, address)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET pan = $3, aadhaar = $4, date_of_birth = $5, address = $6,
		    status = 'pending', updated_at = NOW()
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		kyc.UserID,
		kyc.Status,
		kyc.PAN,
		kyc.Aadhaar,
		kyc.DateOfBirth,
		addressJSON,
	).Scan(&kyc.CreatedAt, &kyc.UpdatedAt)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return errors.Conflict("PAN or Aadhaar already registered")
		}
		return errors.DatabaseWrap(err, "failed to create KYC")
	}

	return nil
}

// GetByUserID retrieves KYC information by user ID.
func (r *KYCRepository) GetByUserID(ctx context.Context, userID string) (*models.KYCInfo, *errors.Error) {
	kyc := &models.KYCInfo{}
	var addressJSON []byte

	query := `
		SELECT user_id, status, pan, aadhaar, date_of_birth, address,
		       verified_at, rejected_at, rejection_reason, created_at, updated_at
		FROM user_kyc
		WHERE user_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&kyc.UserID,
		&kyc.Status,
		&kyc.PAN,
		&kyc.Aadhaar,
		&kyc.DateOfBirth,
		&addressJSON,
		&kyc.VerifiedAt,
		&kyc.RejectedAt,
		&kyc.RejectionReason,
		&kyc.CreatedAt,
		&kyc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("KYC information")
		}
		return nil, errors.DatabaseWrap(err, "failed to get KYC")
	}

	// Deserialize address from JSONB
	if err := json.Unmarshal(addressJSON, &kyc.Address); err != nil {
		return nil, errors.Internal("failed to parse address")
	}

	return kyc, nil
}

// UpdateStatus updates KYC verification status.
func (r *KYCRepository) UpdateStatus(ctx context.Context, userID string, status models.KYCStatus, reason string) *errors.Error {
	var query string

	switch status {
	case models.KYCStatusVerified:
		query = `
			UPDATE user_kyc
			SET status = $2, verified_at = NOW(), rejected_at = NULL,
			    rejection_reason = NULL, updated_at = NOW()
			WHERE user_id = $1
		`
	case models.KYCStatusRejected:
		query = `
			UPDATE user_kyc
			SET status = $2, rejected_at = NOW(), rejection_reason = $3,
			    verified_at = NULL, updated_at = NOW()
			WHERE user_id = $1
		`
	default:
		query = `
			UPDATE user_kyc
			SET status = $2, updated_at = NOW()
			WHERE user_id = $1
		`
	}

	var result sql.Result
	var err error

	if status == models.KYCStatusRejected {
		result, err = r.db.ExecContext(ctx, query, userID, status, reason)
	} else {
		result, err = r.db.ExecContext(ctx, query, userID, status)
	}

	if err != nil {
		return errors.DatabaseWrap(err, "failed to update KYC status")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rows == 0 {
		return errors.NotFound("KYC information")
	}

	return nil
}

// GetByPAN retrieves KYC information by PAN.
func (r *KYCRepository) GetByPAN(ctx context.Context, pan string) (*models.KYCInfo, *errors.Error) {
	kyc := &models.KYCInfo{}
	var addressJSON []byte

	query := `
		SELECT user_id, status, pan, aadhaar, date_of_birth, address,
		       verified_at, rejected_at, rejection_reason, created_at, updated_at
		FROM user_kyc
		WHERE pan = $1
	`

	err := r.db.QueryRowContext(ctx, query, pan).Scan(
		&kyc.UserID,
		&kyc.Status,
		&kyc.PAN,
		&kyc.Aadhaar,
		&kyc.DateOfBirth,
		&addressJSON,
		&kyc.VerifiedAt,
		&kyc.RejectedAt,
		&kyc.RejectionReason,
		&kyc.CreatedAt,
		&kyc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("KYC information")
		}
		return nil, errors.DatabaseWrap(err, "failed to get KYC by PAN")
	}

	// Deserialize address
	if err := json.Unmarshal(addressJSON, &kyc.Address); err != nil {
		return nil, errors.Internal("failed to parse address")
	}

	return kyc, nil
}

// KYCWithUser represents KYC information with user details.
type KYCWithUser struct {
	KYC  models.KYCInfo
	User models.User
}

// ListPending retrieves all KYC submissions with pending status.
func (r *KYCRepository) ListPending(ctx context.Context, limit, offset int) ([]KYCWithUser, *errors.Error) {
	query := `
		SELECT
			k.user_id, k.status, k.pan, k.aadhaar, k.date_of_birth, k.address,
			k.verified_at, k.rejected_at, k.rejection_reason, k.created_at, k.updated_at,
			u.id, u.email, u.phone, u.full_name, u.status, u.created_at, u.updated_at
		FROM user_kyc k
		INNER JOIN users u ON k.user_id = u.id
		WHERE k.status = 'pending'
		ORDER BY k.created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list pending KYCs")
	}
	defer func() { _ = rows.Close() }()

	results := make([]KYCWithUser, 0)

	for rows.Next() {
		var kycWithUser KYCWithUser
		var addressJSON []byte

		err := rows.Scan(
			&kycWithUser.KYC.UserID,
			&kycWithUser.KYC.Status,
			&kycWithUser.KYC.PAN,
			&kycWithUser.KYC.Aadhaar,
			&kycWithUser.KYC.DateOfBirth,
			&addressJSON,
			&kycWithUser.KYC.VerifiedAt,
			&kycWithUser.KYC.RejectedAt,
			&kycWithUser.KYC.RejectionReason,
			&kycWithUser.KYC.CreatedAt,
			&kycWithUser.KYC.UpdatedAt,
			&kycWithUser.User.ID,
			&kycWithUser.User.Email,
			&kycWithUser.User.Phone,
			&kycWithUser.User.FullName,
			&kycWithUser.User.Status,
			&kycWithUser.User.CreatedAt,
			&kycWithUser.User.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan KYC with user")
		}

		// Deserialize address
		if err := json.Unmarshal(addressJSON, &kycWithUser.KYC.Address); err != nil {
			return nil, errors.Internal("failed to parse address")
		}

		results = append(results, kycWithUser)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating KYC records")
	}

	return results, nil
}

// SessionRepository handles database operations for sessions.
type SessionRepository struct {
	db *database.DB
}

// NewSessionRepository creates a new session repository.
func NewSessionRepository(db *database.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session.
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) *errors.Error {
	query := `
		INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		session.UserID,
		session.Token, // This is the token hash
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to create session")
	}

	return nil
}

// GetByTokenHash retrieves a session by token hash.
func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, *errors.Error) {
	session := &models.Session{}

	query := `
		SELECT id, user_id, token_hash, ip_address, user_agent, expires_at, created_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Unauthorized("invalid or expired session")
		}
		return nil, errors.DatabaseWrap(err, "failed to get session")
	}

	return session, nil
}

// DeleteByTokenHash deletes a session by token hash (logout).
func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) *errors.Error {
	query := `DELETE FROM sessions WHERE token_hash = $1`

	result, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete session")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rows == 0 {
		return errors.NotFound("session")
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user (logout all devices).
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) *errors.Error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete user sessions")
	}

	return nil
}

// CleanupExpired deletes all expired sessions.
func (r *SessionRepository) CleanupExpired(ctx context.Context) (int, *errors.Error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to cleanup expired sessions")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to get rows affected")
	}

	return int(rows), nil
}
