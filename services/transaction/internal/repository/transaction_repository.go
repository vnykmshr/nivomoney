package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// TransactionRepository handles database operations for transactions.
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository.
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction.
func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) *errors.Error {
	var metadataJSON []byte
	var err error

	if tx.Metadata != nil {
		metadataJSON, err = json.Marshal(tx.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO transactions (
			type, status, source_wallet_id, destination_wallet_id,
			amount, currency, description, reference, parent_transaction_id, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		tx.Type,
		tx.Status,
		tx.SourceWalletID,
		tx.DestinationWalletID,
		tx.Amount,
		tx.Currency,
		tx.Description,
		tx.Reference,
		tx.ParentTransactionID,
		metadataJSON,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		return errors.DatabaseWrap(err, "failed to create transaction")
	}

	return nil
}

// GetByID retrieves a transaction by ID.
func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, *errors.Error) {
	tx := &models.Transaction{}
	var metadataJSON []byte

	query := `
		SELECT id, type, status, source_wallet_id, destination_wallet_id,
		       amount, currency, description, reference, ledger_entry_id,
		       parent_transaction_id, metadata, failure_reason,
		       processed_at, completed_at, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.Type,
		&tx.Status,
		&tx.SourceWalletID,
		&tx.DestinationWalletID,
		&tx.Amount,
		&tx.Currency,
		&tx.Description,
		&tx.Reference,
		&tx.LedgerEntryID,
		&tx.ParentTransactionID,
		&metadataJSON,
		&tx.FailureReason,
		&tx.ProcessedAt,
		&tx.CompletedAt,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("transaction", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get transaction")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return tx, nil
}

// ListByWallet retrieves transactions for a wallet (both source and destination).
func (r *TransactionRepository) ListByWallet(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	query := `
		SELECT id, type, status, source_wallet_id, destination_wallet_id,
		       amount, currency, description, reference, ledger_entry_id,
		       parent_transaction_id, metadata, failure_reason,
		       processed_at, completed_at, created_at, updated_at
		FROM transactions
		WHERE (source_wallet_id = $1 OR destination_wallet_id = $1)
	`

	args := []interface{}{walletID}
	argCount := 1

	// Add filters
	if filter != nil {
		if filter.Status != nil {
			argCount++
			query += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, *filter.Status)
		}

		if filter.Type != nil {
			argCount++
			query += fmt.Sprintf(" AND type = $%d", argCount)
			args = append(args, *filter.Type)
		}

		if filter.StartDate != nil {
			argCount++
			query += fmt.Sprintf(" AND created_at >= $%d", argCount)
			args = append(args, filter.StartDate)
		}

		if filter.EndDate != nil {
			argCount++
			query += fmt.Sprintf(" AND created_at <= $%d", argCount)
			args = append(args, filter.EndDate)
		}
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter != nil && filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list transactions")
	}
	defer rows.Close()

	transactions := make([]*models.Transaction, 0)
	for rows.Next() {
		tx := &models.Transaction{}
		var metadataJSON []byte

		err := rows.Scan(
			&tx.ID,
			&tx.Type,
			&tx.Status,
			&tx.SourceWalletID,
			&tx.DestinationWalletID,
			&tx.Amount,
			&tx.Currency,
			&tx.Description,
			&tx.Reference,
			&tx.LedgerEntryID,
			&tx.ParentTransactionID,
			&metadataJSON,
			&tx.FailureReason,
			&tx.ProcessedAt,
			&tx.CompletedAt,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan transaction")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &tx.Metadata); err != nil {
				return nil, errors.Internal("failed to parse metadata")
			}
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating transactions")
	}

	return transactions, nil
}

// UpdateStatus updates the status of a transaction.
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id string, status models.TransactionStatus, failureReason *string) *errors.Error {
	query := `
		UPDATE transactions
		SET status = $1, failure_reason = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id
	`

	var txID string
	err := r.db.QueryRowContext(ctx, query, status, failureReason, id).Scan(&txID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("transaction", id)
		}
		return errors.DatabaseWrap(err, "failed to update transaction status")
	}

	return nil
}

// UpdateLedgerEntry updates the ledger entry ID for a transaction.
func (r *TransactionRepository) UpdateLedgerEntry(ctx context.Context, id, ledgerEntryID string) *errors.Error {
	query := `
		UPDATE transactions
		SET ledger_entry_id = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id
	`

	var txID string
	err := r.db.QueryRowContext(ctx, query, ledgerEntryID, id).Scan(&txID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("transaction", id)
		}
		return errors.DatabaseWrap(err, "failed to update ledger entry")
	}

	return nil
}

// MarkProcessed marks a transaction as processed.
func (r *TransactionRepository) MarkProcessed(ctx context.Context, id string) *errors.Error {
	query := `
		UPDATE transactions
		SET status = $1, processed_at = NOW(), updated_at = NOW()
		WHERE id = $2
		RETURNING id
	`

	var txID string
	err := r.db.QueryRowContext(ctx, query, models.TransactionStatusProcessing, id).Scan(&txID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("transaction", id)
		}
		return errors.DatabaseWrap(err, "failed to mark transaction as processed")
	}

	return nil
}

// MarkCompleted marks a transaction as completed.
func (r *TransactionRepository) MarkCompleted(ctx context.Context, id string) *errors.Error {
	query := `
		UPDATE transactions
		SET status = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status != $1
		RETURNING id
	`

	var txID string
	err := r.db.QueryRowContext(ctx, query, models.TransactionStatusCompleted, id).Scan(&txID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("transaction", id)
		}
		return errors.DatabaseWrap(err, "failed to mark transaction as completed")
	}

	return nil
}
