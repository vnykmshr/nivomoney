package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

// Create creates a new wallet and its associated transfer limits.
func (r *WalletRepository) Create(ctx context.Context, wallet *models.Wallet) *errors.Error {
	// Start transaction to create wallet and limits atomically
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to begin transaction")
	}

	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var metadataJSON []byte
	if wallet.Metadata != nil {
		metadataJSON, err = json.Marshal(wallet.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	// Create wallet
	query := `
		INSERT INTO wallets (user_id, type, currency, balance, status, ledger_account_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, available_balance, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
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

	// Create default wallet limits (₹10,000/day, ₹100,000/month)
	limitsQuery := `
		INSERT INTO wallet_limits (wallet_id, daily_limit, monthly_limit)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, limitsQuery, wallet.ID, 1000000, 10000000) // Default limits in paise
	if err != nil {
		return errors.DatabaseWrap(err, "failed to create wallet limits")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.DatabaseWrap(err, "failed to commit wallet creation")
	}

	committed = true
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

// ProcessTransferWithinTx processes a wallet-to-wallet transfer atomically within a transaction.
// This checks limits, verifies balance, and updates wallet balances in a single transaction.
// The transactionID is used for idempotency - if this transaction has already been processed,
// the function returns success without re-executing the transfer.
func (r *WalletRepository) ProcessTransferWithinTx(ctx context.Context, sourceWalletID, destWalletID string, amount int64, transactionID string) *errors.Error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to begin transaction")
	}

	// Track whether we need to rollback
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// 1. Idempotency check - has this transaction already been processed?
	var existingTxID string
	err = tx.QueryRowContext(ctx, `
		SELECT transaction_id
		FROM processed_transfers
		WHERE transaction_id = $1
	`, transactionID).Scan(&existingTxID)

	if err == nil {
		// Transaction already processed - return success (idempotent)
		_ = tx.Rollback() // No changes needed
		return nil
	} else if err != sql.ErrNoRows {
		// Unexpected error
		return errors.DatabaseWrap(err, "failed to check idempotency")
	}
	// If ErrNoRows, continue with transfer processing

	// 2. Lock both wallets in deterministic order to prevent deadlocks
	// Always lock wallets in lexicographic order by ID
	firstID, secondID := sourceWalletID, destWalletID
	if sourceWalletID > destWalletID {
		firstID, secondID = destWalletID, sourceWalletID
	}

	var sourceStatus, sourceCurrency string
	var sourceBalance int64
	var destStatus, destCurrency string

	// Lock first wallet
	if firstID == sourceWalletID {
		err = tx.QueryRowContext(ctx, `
			SELECT status, balance, currency
			FROM wallets
			WHERE id = $1
			FOR UPDATE
		`, sourceWalletID).Scan(&sourceStatus, &sourceBalance, &sourceCurrency)
	} else {
		err = tx.QueryRowContext(ctx, `
			SELECT status, currency
			FROM wallets
			WHERE id = $1
			FOR UPDATE
		`, destWalletID).Scan(&destStatus, &destCurrency)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("wallet", firstID)
		}
		return errors.DatabaseWrap(err, "failed to lock first wallet")
	}

	// Lock second wallet
	if secondID == sourceWalletID {
		err = tx.QueryRowContext(ctx, `
			SELECT status, balance, currency
			FROM wallets
			WHERE id = $1
			FOR UPDATE
		`, sourceWalletID).Scan(&sourceStatus, &sourceBalance, &sourceCurrency)
	} else {
		err = tx.QueryRowContext(ctx, `
			SELECT status, currency
			FROM wallets
			WHERE id = $1
			FOR UPDATE
		`, destWalletID).Scan(&destStatus, &destCurrency)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("wallet", secondID)
		}
		return errors.DatabaseWrap(err, "failed to lock second wallet")
	}

	// 3. Validate both wallets are active
	if sourceStatus != string(models.WalletStatusActive) {
		return errors.BadRequest("source wallet is not active")
	}

	if destStatus != string(models.WalletStatusActive) {
		return errors.BadRequest("destination wallet is not active")
	}

	// 4. Validate currency match
	if sourceCurrency != destCurrency {
		return errors.BadRequest(fmt.Sprintf("currency mismatch: source is %s, destination is %s", sourceCurrency, destCurrency))
	}

	// 5. Check if source has sufficient balance
	if sourceBalance < amount {
		shortfall := amount - sourceBalance
		return errors.BadRequest(fmt.Sprintf("insufficient balance (short by: ₹%.2f)", float64(shortfall)/100))
	}

	// 6. Check and reserve limits
	if limitErr := r.CheckAndReserveLimitWithinTx(ctx, tx, sourceWalletID, amount); limitErr != nil {
		return limitErr
	}

	// 7. Update source wallet balance (debit)
	_, err = tx.ExecContext(ctx, `
		UPDATE wallets
		SET balance = balance - $1,
		    available_balance = available_balance - $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, sourceWalletID)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to debit source wallet")
	}

	// 8. Update destination wallet balance (credit)
	_, err = tx.ExecContext(ctx, `
		UPDATE wallets
		SET balance = balance + $1,
		    available_balance = available_balance + $1,
		    updated_at = NOW()
		WHERE id = $2
	`, amount, destWalletID)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to credit destination wallet")
	}

	// 9. Record this transfer as processed for idempotency
	_, err = tx.ExecContext(ctx, `
		INSERT INTO processed_transfers (transaction_id, source_wallet_id, destination_wallet_id, amount)
		VALUES ($1, $2, $3, $4)
	`, transactionID, sourceWalletID, destWalletID, amount)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to record processed transfer")
	}

	// 10. Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.DatabaseWrap(err, "failed to commit transfer transaction")
	}

	// Mark as committed to prevent rollback in defer
	committed = true
	return nil
}

// CheckAndReserveLimitWithinTx checks if a transfer is within limits and reserves the amount atomically.
// This must be called within a transaction to ensure atomic limit checking and reservation.
func (r *WalletRepository) CheckAndReserveLimitWithinTx(ctx context.Context, tx *sql.Tx, walletID string, amount int64) *errors.Error {
	// Get current limits with row lock (FOR UPDATE)
	// Also check and reset if limits have expired using the check function
	limits := &models.WalletLimits{}

	query := `
		WITH current_limits AS (
			SELECT id, wallet_id, daily_limit, daily_spent, daily_reset_at,
			       monthly_limit, monthly_spent, monthly_reset_at
			FROM wallet_limits
			WHERE wallet_id = $1
			FOR UPDATE
		),
		reset_check AS (
			SELECT
				cl.id,
				cl.wallet_id,
				cl.daily_limit,
				cl.monthly_limit,
				cr.new_daily_spent,
				cr.new_daily_reset_at,
				cr.new_monthly_spent,
				cr.new_monthly_reset_at
			FROM current_limits cl,
			LATERAL check_and_reset_wallet_limits(
				cl.wallet_id,
				cl.daily_limit,
				cl.daily_spent,
				cl.daily_reset_at,
				cl.monthly_limit,
				cl.monthly_spent,
				cl.monthly_reset_at
			) cr
		)
		SELECT id, wallet_id, daily_limit, new_daily_spent, new_daily_reset_at,
		       monthly_limit, new_monthly_spent, new_monthly_reset_at
		FROM reset_check
	`

	var dailyResetAt, monthlyResetAt interface{}
	err := tx.QueryRowContext(ctx, query, walletID).Scan(
		&limits.ID,
		&limits.WalletID,
		&limits.DailyLimit,
		&limits.DailySpent,
		&dailyResetAt,
		&limits.MonthlyLimit,
		&limits.MonthlySpent,
		&monthlyResetAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFound("wallet limits not found")
		}
		return errors.DatabaseWrap(err, "failed to get wallet limits")
	}

	// If limits were reset, update them in the database
	needsUpdate := false
	updateQuery := "UPDATE wallet_limits SET "
	updateArgs := []interface{}{}
	argCount := 0

	// Check if daily needs reset
	if limits.DailySpent == 0 && dailyResetAt != nil {
		needsUpdate = true
		argCount++
		updateQuery += fmt.Sprintf("daily_spent = $%d, ", argCount)
		updateArgs = append(updateArgs, 0)
		argCount++
		updateQuery += fmt.Sprintf("daily_reset_at = $%d, ", argCount)
		updateArgs = append(updateArgs, dailyResetAt)
	}

	// Check if monthly needs reset
	if limits.MonthlySpent == 0 && monthlyResetAt != nil {
		needsUpdate = true
		argCount++
		updateQuery += fmt.Sprintf("monthly_spent = $%d, ", argCount)
		updateArgs = append(updateArgs, 0)
		argCount++
		updateQuery += fmt.Sprintf("monthly_reset_at = $%d, ", argCount)
		updateArgs = append(updateArgs, monthlyResetAt)
	}

	if needsUpdate {
		argCount++
		updateQuery += fmt.Sprintf("updated_at = NOW() WHERE id = $%d", argCount)
		updateArgs = append(updateArgs, limits.ID)
		_, err = tx.ExecContext(ctx, updateQuery, updateArgs...)
		if err != nil {
			return errors.DatabaseWrap(err, "failed to reset expired limits")
		}
	}

	// Check if amount exceeds daily limit
	if limits.DailySpent+amount > limits.DailyLimit {
		remaining := limits.DailyLimit - limits.DailySpent
		return errors.BadRequest(fmt.Sprintf("transfer exceeds daily limit (remaining: ₹%.2f)", float64(remaining)/100))
	}

	// Check if amount exceeds monthly limit
	if limits.MonthlySpent+amount > limits.MonthlyLimit {
		remaining := limits.MonthlyLimit - limits.MonthlySpent
		return errors.BadRequest(fmt.Sprintf("transfer exceeds monthly limit (remaining: ₹%.2f)", float64(remaining)/100))
	}

	// Reserve the amount
	reserveQuery := `
		UPDATE wallet_limits
		SET daily_spent = daily_spent + $1,
		    monthly_spent = monthly_spent + $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, reserveQuery, amount, limits.ID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to reserve transfer limit")
	}

	return nil
}
