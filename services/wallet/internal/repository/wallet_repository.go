package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// WalletRepository handles database operations for wallets.
type WalletRepository struct {
	db *sql.DB
}

// NewWalletRepository creates a new wallet repository.
func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Create creates a new wallet.
func (r *WalletRepository) Create(ctx context.Context, wallet *models.Wallet) *errors.Error {
	var metadataJSON []byte
	var err error

	if wallet.Metadata != nil {
		metadataJSON, err = json.Marshal(wallet.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO wallets (user_id, type, currency, balance, status, ledger_account_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, available_balance, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		wallet.UserID,
		wallet.Type,
		wallet.Currency,
		wallet.Balance,
		wallet.Status,
		wallet.LedgerAccountID,
		metadataJSON,
	).Scan(&wallet.ID, &wallet.AvailableBalance, &wallet.CreatedAt, &wallet.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("wallet of this type and currency already exists for user")
		}
		return errors.DatabaseWrap(err, "failed to create wallet")
	}

	return nil
}

// GetByID retrieves a wallet by ID.
func (r *WalletRepository) GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error) {
	wallet := &models.Wallet{}
	var metadataJSON []byte

	query := `
		SELECT id, user_id, type, currency, balance, available_balance, status,
		       ledger_account_id, metadata, created_at, updated_at, closed_at, closed_reason
		FROM wallets
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Type,
		&wallet.Currency,
		&wallet.Balance,
		&wallet.AvailableBalance,
		&wallet.Status,
		&wallet.LedgerAccountID,
		&metadataJSON,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
		&wallet.ClosedAt,
		&wallet.ClosedReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("wallet", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get wallet")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &wallet.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return wallet, nil
}

// ListByUserID retrieves all wallets for a user.
func (r *WalletRepository) ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error) {
	query := `
		SELECT id, user_id, type, currency, balance, available_balance, status,
		       ledger_account_id, metadata, created_at, updated_at, closed_at, closed_reason
		FROM wallets
		WHERE user_id = $1
	`

	args := []interface{}{userID}

	if status != nil {
		query += ` AND status = $2`
		args = append(args, *status)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list wallets")
	}
	defer func() { _ = rows.Close() }()

	wallets := make([]*models.Wallet, 0)
	for rows.Next() {
		wallet := &models.Wallet{}
		var metadataJSON []byte

		err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Type,
			&wallet.Currency,
			&wallet.Balance,
			&wallet.AvailableBalance,
			&wallet.Status,
			&wallet.LedgerAccountID,
			&metadataJSON,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
			&wallet.ClosedAt,
			&wallet.ClosedReason,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan wallet")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &wallet.Metadata); err != nil {
				return nil, errors.Internal("failed to parse metadata")
			}
		}

		wallets = append(wallets, wallet)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating wallets")
	}

	return wallets, nil
}

// UpdateStatus updates the status of a wallet.
func (r *WalletRepository) UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error {
	query := `
		UPDATE wallets
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id
	`

	var walletID string
	err := r.db.QueryRowContext(ctx, query, status, id).Scan(&walletID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("wallet", id)
		}
		return errors.DatabaseWrap(err, "failed to update wallet status")
	}

	return nil
}

// Close closes a wallet permanently.
func (r *WalletRepository) Close(ctx context.Context, id, reason string) *errors.Error {
	query := `
		UPDATE wallets
		SET status = 'closed', closed_at = NOW(), closed_reason = $1, updated_at = NOW()
		WHERE id = $2 AND status != 'closed'
		RETURNING id
	`

	var walletID string
	err := r.db.QueryRowContext(ctx, query, reason, id).Scan(&walletID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("wallet not found or already closed")
		}
		return errors.DatabaseWrap(err, "failed to close wallet")
	}

	return nil
}

// GetBalance retrieves the balance of a wallet.
func (r *WalletRepository) GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error) {
	balance := &models.WalletBalance{WalletID: id}

	query := `
		SELECT balance, available_balance
		FROM wallets
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&balance.Balance, &balance.AvailableBalance)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("wallet", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get wallet balance")
	}

	balance.HeldAmount = balance.Balance - balance.AvailableBalance

	return balance, nil
}

// isUniqueViolation checks if the error is a unique constraint violation.
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code is 23505
	// This is a simplified check; in production, use pq.Error
	return err != nil && (err.Error() == "UNIQUE constraint failed" ||
		// Add PostgreSQL-specific check if using lib/pq
		false)
}

// GetLimits retrieves the transfer limits for a wallet.
func (r *WalletRepository) GetLimits(ctx context.Context, walletID string) (*models.WalletLimits, *errors.Error) {
	limits := &models.WalletLimits{}

	query := `
		SELECT id, wallet_id, daily_limit, daily_spent, daily_reset_at,
		       monthly_limit, monthly_spent, monthly_reset_at, created_at, updated_at
		FROM wallet_limits
		WHERE wallet_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, walletID).Scan(
		&limits.ID,
		&limits.WalletID,
		&limits.DailyLimit,
		&limits.DailySpent,
		&limits.DailyResetAt,
		&limits.MonthlyLimit,
		&limits.MonthlySpent,
		&limits.MonthlyResetAt,
		&limits.CreatedAt,
		&limits.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("wallet limits not found")
		}
		return nil, errors.DatabaseWrap(err, "failed to get wallet limits")
	}

	return limits, nil
}

// UpdateLimits updates the transfer limits for a wallet.
func (r *WalletRepository) UpdateLimits(ctx context.Context, walletID string, dailyLimit, monthlyLimit int64) *errors.Error {
	query := `
		UPDATE wallet_limits
		SET daily_limit = $1, monthly_limit = $2, updated_at = NOW()
		WHERE wallet_id = $3
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query, dailyLimit, monthlyLimit, walletID).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("wallet limits not found")
		}
		return errors.DatabaseWrap(err, "failed to update wallet limits")
	}

	return nil
}

// IncrementSpent increments the daily and monthly spent amounts for a wallet.
// This is called after a successful transfer to track usage against limits.
func (r *WalletRepository) IncrementSpent(ctx context.Context, walletID string, amount int64) *errors.Error {
	query := `
		UPDATE wallet_limits
		SET daily_spent = daily_spent + $1,
		    monthly_spent = monthly_spent + $1,
		    updated_at = NOW()
		WHERE wallet_id = $2
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query, amount, walletID).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("wallet limits not found")
		}
		return errors.DatabaseWrap(err, "failed to increment spent amount")
	}

	return nil
}

// CheckAndReserveLimitWithinTx checks if a transfer is within limits and reserves the amount atomically.
// This must be called within a transaction to ensure atomic limit checking and reservation.
func (r *WalletRepository) CheckAndReserveLimitWithinTx(ctx context.Context, tx *sql.Tx, walletID string, amount int64) *errors.Error {
	// Get current limits with row lock (FOR UPDATE)
	limits := &models.WalletLimits{}

	query := `
		SELECT id, wallet_id, daily_limit, daily_spent, daily_reset_at,
		       monthly_limit, monthly_spent, monthly_reset_at
		FROM wallet_limits
		WHERE wallet_id = $1
		FOR UPDATE
	`

	err := tx.QueryRowContext(ctx, query, walletID).Scan(
		&limits.ID,
		&limits.WalletID,
		&limits.DailyLimit,
		&limits.DailySpent,
		&limits.DailyResetAt,
		&limits.MonthlyLimit,
		&limits.MonthlySpent,
		&limits.MonthlyResetAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("wallet limits not found")
		}
		return errors.DatabaseWrap(err, "failed to get wallet limits")
	}

	// Check if amount exceeds daily limit
	if limits.DailySpent+amount > limits.DailyLimit {
		return errors.BadRequest("transfer exceeds daily limit")
	}

	// Check if amount exceeds monthly limit
	if limits.MonthlySpent+amount > limits.MonthlyLimit {
		return errors.BadRequest("transfer exceeds monthly limit")
	}

	// Reserve the amount
	updateQuery := `
		UPDATE wallet_limits
		SET daily_spent = daily_spent + $1,
		    monthly_spent = monthly_spent + $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, updateQuery, amount, limits.ID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to reserve transfer limit")
	}

	return nil
}
