package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// VirtualCardRepository handles database operations for virtual cards.
type VirtualCardRepository struct {
	db *sql.DB
}

// NewVirtualCardRepository creates a new virtual card repository.
func NewVirtualCardRepository(db *sql.DB) *VirtualCardRepository {
	return &VirtualCardRepository{db: db}
}

// Create creates a new virtual card.
func (r *VirtualCardRepository) Create(ctx context.Context, card *models.VirtualCard) *errors.Error {
	// Generate card number and CVV
	cardNumber := generateCardNumber()
	cvv := generateCVV()
	hashedCVV := hashCVV(cvv)

	// Set expiry (3 years from now)
	now := time.Now()
	card.ExpiryMonth = int(now.Month())
	card.ExpiryYear = now.Year() + 3

	query := `
		INSERT INTO virtual_cards (
			wallet_id, user_id, card_number, card_holder_name,
			expiry_month, expiry_year, cvv, card_type, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, daily_limit, monthly_limit, per_transaction_limit, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		card.WalletID,
		card.UserID,
		cardNumber,
		card.CardHolderName,
		card.ExpiryMonth,
		card.ExpiryYear,
		hashedCVV,
		models.CardTypeVirtual,
		models.CardStatusActive,
	).Scan(
		&card.ID,
		&card.DailyLimit,
		&card.MonthlyLimit,
		&card.PerTransactionLimit,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to create virtual card")
	}

	// Set generated values
	card.CardNumber = cardNumber
	card.CVV = cvv // Return plain CVV only during creation
	card.CardType = models.CardTypeVirtual
	card.Status = models.CardStatusActive

	return nil
}

// GetByID retrieves a virtual card by ID.
func (r *VirtualCardRepository) GetByID(ctx context.Context, id string) (*models.VirtualCard, *errors.Error) {
	card := &models.VirtualCard{}

	query := `
		SELECT id, wallet_id, user_id, card_number, card_holder_name,
		       expiry_month, expiry_year, cvv, card_type, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_spent, monthly_spent, last_used_at,
		       frozen_at, frozen_reason, cancelled_at, cancelled_reason,
		       created_at, updated_at
		FROM virtual_cards
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&card.ID,
		&card.WalletID,
		&card.UserID,
		&card.CardNumber,
		&card.CardHolderName,
		&card.ExpiryMonth,
		&card.ExpiryYear,
		&card.CVV,
		&card.CardType,
		&card.Status,
		&card.DailyLimit,
		&card.MonthlyLimit,
		&card.PerTransactionLimit,
		&card.DailySpent,
		&card.MonthlySpent,
		&card.LastUsedAt,
		&card.FrozenAt,
		&card.FrozenReason,
		&card.CancelledAt,
		&card.CancelledReason,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("virtual card", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get virtual card")
	}

	return card, nil
}

// ListByWallet retrieves all virtual cards for a wallet.
func (r *VirtualCardRepository) ListByWallet(ctx context.Context, walletID string) ([]*models.VirtualCard, *errors.Error) {
	query := `
		SELECT id, wallet_id, user_id, card_number, card_holder_name,
		       expiry_month, expiry_year, cvv, card_type, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_spent, monthly_spent, last_used_at,
		       frozen_at, frozen_reason, cancelled_at, cancelled_reason,
		       created_at, updated_at
		FROM virtual_cards
		WHERE wallet_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list virtual cards")
	}
	defer func() { _ = rows.Close() }()

	cards := make([]*models.VirtualCard, 0)
	for rows.Next() {
		card := &models.VirtualCard{}
		err := rows.Scan(
			&card.ID,
			&card.WalletID,
			&card.UserID,
			&card.CardNumber,
			&card.CardHolderName,
			&card.ExpiryMonth,
			&card.ExpiryYear,
			&card.CVV,
			&card.CardType,
			&card.Status,
			&card.DailyLimit,
			&card.MonthlyLimit,
			&card.PerTransactionLimit,
			&card.DailySpent,
			&card.MonthlySpent,
			&card.LastUsedAt,
			&card.FrozenAt,
			&card.FrozenReason,
			&card.CancelledAt,
			&card.CancelledReason,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan virtual card")
		}
		cards = append(cards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating virtual cards")
	}

	return cards, nil
}

// Freeze freezes a virtual card.
func (r *VirtualCardRepository) Freeze(ctx context.Context, id, reason string) *errors.Error {
	query := `
		UPDATE virtual_cards
		SET status = $1, frozen_at = NOW(), frozen_reason = $2
		WHERE id = $3 AND status = $4
		RETURNING id
	`

	var cardID string
	err := r.db.QueryRowContext(ctx, query,
		models.CardStatusFrozen,
		reason,
		id,
		models.CardStatusActive,
	).Scan(&cardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("card not found or not active")
		}
		return errors.DatabaseWrap(err, "failed to freeze card")
	}

	return nil
}

// Unfreeze unfreezes a virtual card.
func (r *VirtualCardRepository) Unfreeze(ctx context.Context, id string) *errors.Error {
	query := `
		UPDATE virtual_cards
		SET status = $1, frozen_at = NULL, frozen_reason = NULL
		WHERE id = $2 AND status = $3
		RETURNING id
	`

	var cardID string
	err := r.db.QueryRowContext(ctx, query,
		models.CardStatusActive,
		id,
		models.CardStatusFrozen,
	).Scan(&cardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("card not found or not frozen")
		}
		return errors.DatabaseWrap(err, "failed to unfreeze card")
	}

	return nil
}

// Cancel cancels a virtual card.
func (r *VirtualCardRepository) Cancel(ctx context.Context, id, reason string) *errors.Error {
	query := `
		UPDATE virtual_cards
		SET status = $1, cancelled_at = NOW(), cancelled_reason = $2
		WHERE id = $3 AND status != $1
		RETURNING id
	`

	var cardID string
	err := r.db.QueryRowContext(ctx, query,
		models.CardStatusCancelled,
		reason,
		id,
	).Scan(&cardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("card not found or already cancelled")
		}
		return errors.DatabaseWrap(err, "failed to cancel card")
	}

	return nil
}

// UpdateLimits updates the spending limits for a virtual card.
func (r *VirtualCardRepository) UpdateLimits(ctx context.Context, id string, dailyLimit, monthlyLimit, perTxLimit *int64) *errors.Error {
	// Build dynamic update query
	query := "UPDATE virtual_cards SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 0

	if dailyLimit != nil {
		argCount++
		query += fmt.Sprintf(", daily_limit = $%d", argCount)
		args = append(args, *dailyLimit)
	}

	if monthlyLimit != nil {
		argCount++
		query += fmt.Sprintf(", monthly_limit = $%d", argCount)
		args = append(args, *monthlyLimit)
	}

	if perTxLimit != nil {
		argCount++
		query += fmt.Sprintf(", per_transaction_limit = $%d", argCount)
		args = append(args, *perTxLimit)
	}

	if argCount == 0 {
		return errors.Validation("no limits to update")
	}

	argCount++
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id", argCount)
	args = append(args, id)

	var cardID string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&cardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("virtual card", id)
		}
		return errors.DatabaseWrap(err, "failed to update card limits")
	}

	return nil
}

// RecordUsage records a card usage and updates spent amounts.
func (r *VirtualCardRepository) RecordUsage(ctx context.Context, id string, amount int64) *errors.Error {
	now := sharedModels.Now()

	query := `
		UPDATE virtual_cards
		SET daily_spent = daily_spent + $1,
		    monthly_spent = monthly_spent + $1,
		    last_used_at = $2
		WHERE id = $3 AND status = $4
		RETURNING id
	`

	var cardID string
	err := r.db.QueryRowContext(ctx, query,
		amount,
		now,
		id,
		models.CardStatusActive,
	).Scan(&cardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("card not found or not active")
		}
		return errors.DatabaseWrap(err, "failed to record card usage")
	}

	return nil
}

// ResetDailySpent resets the daily spent amount for all cards.
func (r *VirtualCardRepository) ResetDailySpent(ctx context.Context) *errors.Error {
	query := `UPDATE virtual_cards SET daily_spent = 0 WHERE daily_spent > 0`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to reset daily spent")
	}

	return nil
}

// ResetMonthlySpent resets the monthly spent amount for all cards.
func (r *VirtualCardRepository) ResetMonthlySpent(ctx context.Context) *errors.Error {
	query := `UPDATE virtual_cards SET monthly_spent = 0 WHERE monthly_spent > 0`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to reset monthly spent")
	}

	return nil
}

// generateCardNumber generates a random 16-digit card number with valid Luhn checksum.
// Uses 4 prefix for Visa-like cards (for simulation purposes).
func generateCardNumber() string {
	// Generate first 15 digits (4 for prefix + 11 random)
	digits := "4" // Visa prefix
	for i := 0; i < 14; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits += fmt.Sprintf("%d", n.Int64())
	}

	// Calculate Luhn checksum
	checksum := luhnChecksum(digits)
	digits += fmt.Sprintf("%d", checksum)

	return digits
}

// luhnChecksum calculates the Luhn checksum digit.
func luhnChecksum(digits string) int {
	sum := 0
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if (len(digits)-i)%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return (10 - (sum % 10)) % 10
}

// generateCVV generates a random 3-digit CVV.
func generateCVV() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900))
	return fmt.Sprintf("%03d", n.Int64()+100)
}

// hashCVV creates a hash of the CVV for storage.
func hashCVV(cvv string) string {
	hash := sha256.Sum256([]byte(cvv))
	return hex.EncodeToString(hash[:])
}
