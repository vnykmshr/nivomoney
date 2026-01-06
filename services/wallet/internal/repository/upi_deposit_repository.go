package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// UPIDepositRepository handles database operations for UPI deposits.
type UPIDepositRepository struct {
	db *sql.DB
}

// NewUPIDepositRepository creates a new UPI deposit repository.
func NewUPIDepositRepository(db *sql.DB) *UPIDepositRepository {
	return &UPIDepositRepository{db: db}
}

// Create creates a new UPI deposit request.
func (r *UPIDepositRepository) Create(ctx context.Context, deposit *models.UPIDeposit) *errors.Error {
	query := `
		INSERT INTO upi_deposits (wallet_id, user_id, amount, upi_reference, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, expires_at
	`

	err := r.db.QueryRowContext(ctx, query,
		deposit.WalletID,
		deposit.UserID,
		deposit.Amount,
		deposit.UPIReference,
		deposit.Status,
	).Scan(&deposit.ID, &deposit.CreatedAt, &deposit.ExpiresAt)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to create UPI deposit")
	}

	return nil
}

// GetByID retrieves a UPI deposit by ID.
func (r *UPIDepositRepository) GetByID(ctx context.Context, id string) (*models.UPIDeposit, *errors.Error) {
	deposit := &models.UPIDeposit{}

	query := `
		SELECT id, wallet_id, user_id, amount, upi_reference, status,
		       created_at, expires_at, completed_at, failed_reason
		FROM upi_deposits
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deposit.ID,
		&deposit.WalletID,
		&deposit.UserID,
		&deposit.Amount,
		&deposit.UPIReference,
		&deposit.Status,
		&deposit.CreatedAt,
		&deposit.ExpiresAt,
		&deposit.CompletedAt,
		&deposit.FailedReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("upi_deposit", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get UPI deposit")
	}

	return deposit, nil
}

// GetByReference retrieves a UPI deposit by reference.
func (r *UPIDepositRepository) GetByReference(ctx context.Context, reference string) (*models.UPIDeposit, *errors.Error) {
	deposit := &models.UPIDeposit{}

	query := `
		SELECT id, wallet_id, user_id, amount, upi_reference, status,
		       created_at, expires_at, completed_at, failed_reason
		FROM upi_deposits
		WHERE upi_reference = $1
	`

	err := r.db.QueryRowContext(ctx, query, reference).Scan(
		&deposit.ID,
		&deposit.WalletID,
		&deposit.UserID,
		&deposit.Amount,
		&deposit.UPIReference,
		&deposit.Status,
		&deposit.CreatedAt,
		&deposit.ExpiresAt,
		&deposit.CompletedAt,
		&deposit.FailedReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("UPI deposit not found")
		}
		return nil, errors.DatabaseWrap(err, "failed to get UPI deposit")
	}

	return deposit, nil
}

// ListByUserID retrieves all UPI deposits for a user.
func (r *UPIDepositRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*models.UPIDeposit, *errors.Error) {
	query := `
		SELECT id, wallet_id, user_id, amount, upi_reference, status,
		       created_at, expires_at, completed_at, failed_reason
		FROM upi_deposits
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list UPI deposits")
	}
	defer func() { _ = rows.Close() }()

	deposits := make([]*models.UPIDeposit, 0)
	for rows.Next() {
		deposit := &models.UPIDeposit{}
		err := rows.Scan(
			&deposit.ID,
			&deposit.WalletID,
			&deposit.UserID,
			&deposit.Amount,
			&deposit.UPIReference,
			&deposit.Status,
			&deposit.CreatedAt,
			&deposit.ExpiresAt,
			&deposit.CompletedAt,
			&deposit.FailedReason,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan UPI deposit")
		}
		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating UPI deposits")
	}

	return deposits, nil
}

// Complete marks a UPI deposit as completed.
func (r *UPIDepositRepository) Complete(ctx context.Context, id string) *errors.Error {
	query := `
		UPDATE upi_deposits
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id
	`

	var depositID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&depositID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("UPI deposit not found or already processed")
		}
		return errors.DatabaseWrap(err, "failed to complete UPI deposit")
	}

	return nil
}

// Fail marks a UPI deposit as failed.
func (r *UPIDepositRepository) Fail(ctx context.Context, id, reason string) *errors.Error {
	query := `
		UPDATE upi_deposits
		SET status = 'failed', failed_reason = $1
		WHERE id = $2 AND status = 'pending'
		RETURNING id
	`

	var depositID string
	err := r.db.QueryRowContext(ctx, query, reason, id).Scan(&depositID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("UPI deposit not found or already processed")
		}
		return errors.DatabaseWrap(err, "failed to mark UPI deposit as failed")
	}

	return nil
}

// ExpirePending marks expired pending deposits as expired.
func (r *UPIDepositRepository) ExpirePending(ctx context.Context) (int64, *errors.Error) {
	query := `
		UPDATE upi_deposits
		SET status = 'expired'
		WHERE status = 'pending' AND expires_at < NOW()
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, errors.DatabaseWrap(err, "failed to expire pending deposits")
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// GetWalletUPIVPA retrieves the UPI VPA for a wallet.
func (r *UPIDepositRepository) GetWalletUPIVPA(ctx context.Context, walletID string) (string, *errors.Error) {
	var upiVPA sql.NullString

	query := `SELECT upi_vpa FROM wallets WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&upiVPA)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.NotFoundWithID("wallet", walletID)
		}
		return "", errors.DatabaseWrap(err, "failed to get wallet UPI VPA")
	}

	if !upiVPA.Valid || upiVPA.String == "" {
		// Generate and update UPI VPA if not set
		generatedVPA := generateUPIVPA(walletID)
		updateQuery := `UPDATE wallets SET upi_vpa = $1 WHERE id = $2`
		_, err = r.db.ExecContext(ctx, updateQuery, generatedVPA, walletID)
		if err != nil {
			return "", errors.DatabaseWrap(err, "failed to set wallet UPI VPA")
		}
		return generatedVPA, nil
	}

	return upiVPA.String, nil
}

// generateUPIVPA generates a UPI VPA from wallet ID.
func generateUPIVPA(walletID string) string {
	// Remove dashes and take first 8 characters
	shortID := ""
	for _, c := range walletID {
		if c != '-' {
			shortID += string(c)
			if len(shortID) >= 8 {
				break
			}
		}
	}
	return fmt.Sprintf("%s@nivo", shortID)
}

// GenerateUPIReference generates a unique UPI reference.
func GenerateUPIReference() string {
	return fmt.Sprintf("NIVO%d%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}
