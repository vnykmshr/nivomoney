package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// BeneficiaryRepository handles database operations for beneficiaries.
type BeneficiaryRepository struct {
	db *sql.DB
}

// NewBeneficiaryRepository creates a new beneficiary repository.
func NewBeneficiaryRepository(db *sql.DB) *BeneficiaryRepository {
	return &BeneficiaryRepository{db: db}
}

// Create creates a new beneficiary.
func (r *BeneficiaryRepository) Create(ctx context.Context, beneficiary *models.Beneficiary) *errors.Error {
	var metadataJSON []byte
	var err error

	if beneficiary.Metadata != nil {
		metadataJSON, err = json.Marshal(beneficiary.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO beneficiaries (
			owner_user_id, beneficiary_user_id, beneficiary_wallet_id,
			nickname, beneficiary_phone, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		beneficiary.OwnerUserID,
		beneficiary.BeneficiaryUserID,
		beneficiary.BeneficiaryWalletID,
		beneficiary.Nickname,
		beneficiary.BeneficiaryPhone,
		metadataJSON,
	).Scan(&beneficiary.ID, &beneficiary.CreatedAt, &beneficiary.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			// Check which constraint was violated
			if isDuplicateNickname(err) {
				return errors.Conflict("a beneficiary with this nickname already exists")
			}
			return errors.Conflict("this user is already in your beneficiaries")
		}
		return errors.DatabaseWrap(err, "failed to create beneficiary")
	}

	return nil
}

// GetByID retrieves a beneficiary by ID.
func (r *BeneficiaryRepository) GetByID(ctx context.Context, id, ownerUserID string) (*models.Beneficiary, *errors.Error) {
	beneficiary := &models.Beneficiary{}
	var metadataJSON []byte

	query := `
		SELECT id, owner_user_id, beneficiary_user_id, beneficiary_wallet_id,
		       nickname, beneficiary_phone, metadata, created_at, updated_at
		FROM beneficiaries
		WHERE id = $1 AND owner_user_id = $2
	`

	err := r.db.QueryRowContext(ctx, query, id, ownerUserID).Scan(
		&beneficiary.ID,
		&beneficiary.OwnerUserID,
		&beneficiary.BeneficiaryUserID,
		&beneficiary.BeneficiaryWalletID,
		&beneficiary.Nickname,
		&beneficiary.BeneficiaryPhone,
		&metadataJSON,
		&beneficiary.CreatedAt,
		&beneficiary.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("beneficiary", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get beneficiary")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &beneficiary.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return beneficiary, nil
}

// ListByOwner retrieves all beneficiaries for a user.
func (r *BeneficiaryRepository) ListByOwner(ctx context.Context, ownerUserID string) ([]*models.Beneficiary, *errors.Error) {
	query := `
		SELECT id, owner_user_id, beneficiary_user_id, beneficiary_wallet_id,
		       nickname, beneficiary_phone, metadata, created_at, updated_at
		FROM beneficiaries
		WHERE owner_user_id = $1
		ORDER BY nickname ASC
	`

	rows, err := r.db.QueryContext(ctx, query, ownerUserID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list beneficiaries")
	}
	defer func() { _ = rows.Close() }()

	beneficiaries := make([]*models.Beneficiary, 0)
	for rows.Next() {
		beneficiary := &models.Beneficiary{}
		var metadataJSON []byte

		err := rows.Scan(
			&beneficiary.ID,
			&beneficiary.OwnerUserID,
			&beneficiary.BeneficiaryUserID,
			&beneficiary.BeneficiaryWalletID,
			&beneficiary.Nickname,
			&beneficiary.BeneficiaryPhone,
			&metadataJSON,
			&beneficiary.CreatedAt,
			&beneficiary.UpdatedAt,
		)
		if err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan beneficiary")
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &beneficiary.Metadata); err != nil {
				return nil, errors.Internal("failed to parse metadata")
			}
		}

		beneficiaries = append(beneficiaries, beneficiary)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating beneficiaries")
	}

	return beneficiaries, nil
}

// UpdateNickname updates a beneficiary's nickname.
func (r *BeneficiaryRepository) UpdateNickname(ctx context.Context, id, ownerUserID, nickname string) *errors.Error {
	query := `
		UPDATE beneficiaries
		SET nickname = $1, updated_at = NOW()
		WHERE id = $2 AND owner_user_id = $3
		RETURNING id
	`

	var beneficiaryID string
	err := r.db.QueryRowContext(ctx, query, nickname, id, ownerUserID).Scan(&beneficiaryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("beneficiary", id)
		}
		if isUniqueViolation(err) {
			return errors.Conflict("a beneficiary with this nickname already exists")
		}
		return errors.DatabaseWrap(err, "failed to update beneficiary")
	}

	return nil
}

// Delete deletes a beneficiary.
func (r *BeneficiaryRepository) Delete(ctx context.Context, id, ownerUserID string) *errors.Error {
	query := `
		DELETE FROM beneficiaries
		WHERE id = $1 AND owner_user_id = $2
		RETURNING id
	`

	var beneficiaryID string
	err := r.db.QueryRowContext(ctx, query, id, ownerUserID).Scan(&beneficiaryID)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NotFoundWithID("beneficiary", id)
		}
		return errors.DatabaseWrap(err, "failed to delete beneficiary")
	}

	return nil
}

// GetByBeneficiaryUser retrieves a beneficiary by beneficiary user ID for a given owner.
func (r *BeneficiaryRepository) GetByBeneficiaryUser(ctx context.Context, ownerUserID, beneficiaryUserID string) (*models.Beneficiary, *errors.Error) {
	beneficiary := &models.Beneficiary{}
	var metadataJSON []byte

	query := `
		SELECT id, owner_user_id, beneficiary_user_id, beneficiary_wallet_id,
		       nickname, beneficiary_phone, metadata, created_at, updated_at
		FROM beneficiaries
		WHERE owner_user_id = $1 AND beneficiary_user_id = $2
	`

	err := r.db.QueryRowContext(ctx, query, ownerUserID, beneficiaryUserID).Scan(
		&beneficiary.ID,
		&beneficiary.OwnerUserID,
		&beneficiary.BeneficiaryUserID,
		&beneficiary.BeneficiaryWalletID,
		&beneficiary.Nickname,
		&beneficiary.BeneficiaryPhone,
		&metadataJSON,
		&beneficiary.CreatedAt,
		&beneficiary.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("beneficiary not found")
		}
		return nil, errors.DatabaseWrap(err, "failed to get beneficiary")
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &beneficiary.Metadata); err != nil {
			return nil, errors.Internal("failed to parse metadata")
		}
	}

	return beneficiary, nil
}

// isDuplicateNickname checks if the error is a duplicate nickname violation.
func isDuplicateNickname(err error) bool {
	// This is a simplified check; in production, use pq.Error to check constraint name
	return err != nil && (err.Error() == "idx_beneficiaries_unique_nickname" || false)
}
