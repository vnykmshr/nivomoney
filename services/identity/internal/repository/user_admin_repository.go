package repository

import (
	"context"
	"database/sql"

	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/errors"
)

// UserAdminRepository handles database operations for user-admin pairings.
type UserAdminRepository struct {
	db *database.DB
}

// NewUserAdminRepository creates a new user-admin repository.
func NewUserAdminRepository(db *database.DB) *UserAdminRepository {
	return &UserAdminRepository{db: db}
}

// CreatePairing creates a pairing between a regular user and their User-Admin account.
func (r *UserAdminRepository) CreatePairing(ctx context.Context, userID, adminUserID string) *errors.Error {
	query := `
		INSERT INTO user_admin_pairs (user_id, admin_user_id)
		VALUES ($1, $2)
	`
	_, err := r.db.ExecContext(ctx, query, userID, adminUserID)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return errors.Conflict("user-admin pairing already exists")
		}
		return errors.DatabaseWrap(err, "failed to create user-admin pairing")
	}
	return nil
}

// CreatePairingWithTx creates a pairing within a transaction.
func (r *UserAdminRepository) CreatePairingWithTx(ctx context.Context, tx *sql.Tx, userID, adminUserID string) *errors.Error {
	query := `
		INSERT INTO user_admin_pairs (user_id, admin_user_id)
		VALUES ($1, $2)
	`
	_, err := tx.ExecContext(ctx, query, userID, adminUserID)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return errors.Conflict("user-admin pairing already exists")
		}
		return errors.DatabaseWrap(err, "failed to create user-admin pairing")
	}
	return nil
}

// GetPairedUserID returns the regular user ID for a given User-Admin ID.
func (r *UserAdminRepository) GetPairedUserID(ctx context.Context, adminUserID string) (string, *errors.Error) {
	var userID string
	query := `SELECT user_id FROM user_admin_pairs WHERE admin_user_id = $1`
	err := r.db.QueryRowContext(ctx, query, adminUserID).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", errors.NotFound("no paired user found for this admin account")
	}
	if err != nil {
		return "", errors.DatabaseWrap(err, "failed to get paired user")
	}
	return userID, nil
}

// GetAdminUserID returns the User-Admin ID for a given regular user ID.
func (r *UserAdminRepository) GetAdminUserID(ctx context.Context, userID string) (string, *errors.Error) {
	var adminUserID string
	query := `SELECT admin_user_id FROM user_admin_pairs WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&adminUserID)
	if err == sql.ErrNoRows {
		return "", errors.NotFound("no user-admin account found for this user")
	}
	if err != nil {
		return "", errors.DatabaseWrap(err, "failed to get user-admin account")
	}
	return adminUserID, nil
}

// IsUserAdmin checks if a user ID belongs to a User-Admin account.
func (r *UserAdminRepository) IsUserAdmin(ctx context.Context, userID string) (bool, *errors.Error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_admin_pairs WHERE admin_user_id = $1)`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, errors.DatabaseWrap(err, "failed to check user-admin status")
	}
	return exists, nil
}

// ValidatePairing checks if the adminUserID is authorized to act on userID.
func (r *UserAdminRepository) ValidatePairing(ctx context.Context, adminUserID, userID string) (bool, *errors.Error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_admin_pairs WHERE admin_user_id = $1 AND user_id = $2)`
	err := r.db.QueryRowContext(ctx, query, adminUserID, userID).Scan(&exists)
	if err != nil {
		return false, errors.DatabaseWrap(err, "failed to validate pairing")
	}
	return exists, nil
}

// GetPairingByUserID retrieves the full pairing record for a user.
func (r *UserAdminRepository) GetPairingByUserID(ctx context.Context, userID string) (*UserAdminPair, *errors.Error) {
	pair := &UserAdminPair{}
	query := `
		SELECT id, user_id, admin_user_id, created_at, updated_at
		FROM user_admin_pairs
		WHERE user_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&pair.ID,
		&pair.UserID,
		&pair.AdminUserID,
		&pair.CreatedAt,
		&pair.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("no user-admin pairing found")
	}
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get user-admin pairing")
	}
	return pair, nil
}

// GetPairingByAdminUserID retrieves the full pairing record for an admin user.
func (r *UserAdminRepository) GetPairingByAdminUserID(ctx context.Context, adminUserID string) (*UserAdminPair, *errors.Error) {
	pair := &UserAdminPair{}
	query := `
		SELECT id, user_id, admin_user_id, created_at, updated_at
		FROM user_admin_pairs
		WHERE admin_user_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, adminUserID).Scan(
		&pair.ID,
		&pair.UserID,
		&pair.AdminUserID,
		&pair.CreatedAt,
		&pair.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("no user-admin pairing found")
	}
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get user-admin pairing")
	}
	return pair, nil
}

// DeletePairing removes a user-admin pairing.
func (r *UserAdminRepository) DeletePairing(ctx context.Context, userID string) *errors.Error {
	query := `DELETE FROM user_admin_pairs WHERE user_id = $1`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete user-admin pairing")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("no user-admin pairing found to delete")
	}
	return nil
}

// UserAdminPair represents a pairing between a user and their admin account.
type UserAdminPair struct {
	ID          string
	UserID      string
	AdminUserID string
	CreatedAt   string
	UpdatedAt   string
}
