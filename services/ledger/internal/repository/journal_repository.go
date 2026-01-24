package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vnykmshr/nivo/services/ledger/internal/models"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/errors"
)

// JournalEntryRepository handles database operations for journal entries.
type JournalEntryRepository struct {
	db *database.DB
}

// NewJournalEntryRepository creates a new journal entry repository.
func NewJournalEntryRepository(db *database.DB) *JournalEntryRepository {
	return &JournalEntryRepository{db: db}
}

// Create creates a new journal entry with lines in a transaction.
func (r *JournalEntryRepository) Create(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error {
	// Start transaction
	err := r.db.Transaction(ctx, func(tx *sql.Tx) error {
		// Generate entry number
		var entryNumber string
		err := tx.QueryRowContext(ctx, "SELECT generate_entry_number()").Scan(&entryNumber)
		if err != nil {
			return errors.DatabaseWrap(err, "failed to generate entry number")
		}

		// Serialize metadata
		metadataJSON, err := json.Marshal(entry.Metadata)
		if err != nil {
			return errors.BadRequest("invalid metadata format")
		}

		// Insert journal entry
		query := `
			INSERT INTO journal_entries (entry_number, type, status, description,
			                              reference_type, reference_id, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`

		err = tx.QueryRowContext(ctx, query,
			entryNumber,
			entry.Type,
			entry.Status,
			entry.Description,
			entry.ReferenceType,
			entry.ReferenceID,
			metadataJSON,
		).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

		if err != nil {
			return errors.DatabaseWrap(err, "failed to create journal entry")
		}

		entry.EntryNumber = entryNumber

		// Insert ledger lines
		lineQuery := `
			INSERT INTO ledger_lines (entry_id, account_id, debit_amount, credit_amount, description, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at
		`

		for i := range lines {
			line := &lines[i]
			line.EntryID = entry.ID

			// Serialize line metadata
			lineMetadataJSON, err := json.Marshal(line.Metadata)
			if err != nil {
				return errors.BadRequest("invalid line metadata format")
			}

			err = tx.QueryRowContext(ctx, lineQuery,
				line.EntryID,
				line.AccountID,
				line.DebitAmount,
				line.CreditAmount,
				line.Description,
				lineMetadataJSON,
			).Scan(&line.ID, &line.CreatedAt)

			if err != nil {
				return errors.DatabaseWrap(err, "failed to create ledger line")
			}
		}

		// Update entry with lines
		entry.Lines = lines

		return nil
	})

	if err != nil {
		// Check if it's already an *errors.Error
		if e, ok := err.(*errors.Error); ok {
			return e
		}
		// Otherwise wrap it
		return errors.DatabaseWrap(err, "transaction failed")
	}

	return nil
}

// GetByID retrieves a journal entry with its lines.
func (r *JournalEntryRepository) GetByID(ctx context.Context, id string) (*models.JournalEntry, *errors.Error) {
	entry := &models.JournalEntry{}
	var metadataJSON []byte

	query := `
		SELECT id, entry_number, type, status, description, reference_type, reference_id,
		       posted_at, posted_by, voided_at, voided_by, void_reason, reversal_entry_id,
		       metadata, created_at, updated_at
		FROM journal_entries
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.EntryNumber,
		&entry.Type,
		&entry.Status,
		&entry.Description,
		&entry.ReferenceType,
		&entry.ReferenceID,
		&entry.PostedAt,
		&entry.PostedBy,
		&entry.VoidedAt,
		&entry.VoidedBy,
		&entry.VoidReason,
		&entry.ReversalEntryID,
		&metadataJSON,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("journal entry", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get journal entry")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &entry.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	// Load lines
	lines, linesErr := r.GetLinesByEntryID(ctx, id)
	if linesErr != nil {
		return nil, linesErr
	}
	entry.Lines = lines

	return entry, nil
}

// GetLinesByEntryID retrieves all lines for a journal entry.
func (r *JournalEntryRepository) GetLinesByEntryID(ctx context.Context, entryID string) ([]models.LedgerLine, *errors.Error) {
	query := `
		SELECT id, entry_id, account_id, debit_amount, credit_amount, description, metadata, created_at
		FROM ledger_lines
		WHERE entry_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get ledger lines")
	}
	defer func() { _ = rows.Close() }()

	lines := make([]models.LedgerLine, 0)
	for rows.Next() {
		line := models.LedgerLine{}
		var metadataJSON []byte

		err := rows.Scan(
			&line.ID,
			&line.EntryID,
			&line.AccountID,
			&line.DebitAmount,
			&line.CreditAmount,
			&line.Description,
			&metadataJSON,
			&line.CreatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan ledger line")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &line.Metadata); err != nil {
				return nil, errors.Internal("failed to parse line metadata")
			}
		}

		lines = append(lines, line)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating ledger lines")
	}

	return lines, nil
}

// Post posts a draft journal entry.
func (r *JournalEntryRepository) Post(ctx context.Context, entryID, postedBy string) *errors.Error {
	query := `
		UPDATE journal_entries
		SET status = 'posted', posted_at = NOW(), posted_by = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'draft'
		RETURNING entry_number
	`

	var entryNumber string
	err := r.db.QueryRowContext(ctx, query, entryID, postedBy).Scan(&entryNumber)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("journal entry not found or already posted")
		}
		// Check if error is from validation trigger
		if database.IsCheckViolation(err) {
			return errors.Validation("journal entry validation failed: " + err.Error())
		}
		return errors.DatabaseWrap(err, "failed to post journal entry")
	}

	return nil
}

// Void voids a posted journal entry.
func (r *JournalEntryRepository) Void(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error {
	query := `
		UPDATE journal_entries
		SET status = 'voided', voided_at = NOW(), voided_by = $2,
		    void_reason = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'posted'
		RETURNING entry_number
	`

	var entryNumber string
	err := r.db.QueryRowContext(ctx, query, entryID, voidedBy, voidReason).Scan(&entryNumber)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.BadRequest("journal entry not found or not posted")
		}
		return errors.DatabaseWrap(err, "failed to void journal entry")
	}

	return nil
}

// List retrieves journal entries with filters.
func (r *JournalEntryRepository) List(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error) {
	query := `
		SELECT id, entry_number, type, status, description, reference_type, reference_id,
		       posted_at, posted_by, voided_at, voided_by, void_reason, reversal_entry_id,
		       metadata, created_at, updated_at
		FROM journal_entries
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list journal entries")
	}
	defer func() { _ = rows.Close() }()

	entries := make([]*models.JournalEntry, 0)
	for rows.Next() {
		entry := &models.JournalEntry{}
		var metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.EntryNumber,
			&entry.Type,
			&entry.Status,
			&entry.Description,
			&entry.ReferenceType,
			&entry.ReferenceID,
			&entry.PostedAt,
			&entry.PostedBy,
			&entry.VoidedAt,
			&entry.VoidedBy,
			&entry.VoidReason,
			&entry.ReversalEntryID,
			&metadataJSON,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan journal entry")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &entry.Metadata); err != nil {
				return nil, errors.Internal("failed to parse metadata")
			}
		}

		// Note: Lines are not loaded in list view for performance
		// Call GetByID to get full entry with lines

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating journal entries")
	}

	return entries, nil
}
