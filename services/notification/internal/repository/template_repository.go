package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vnykmshr/nivo/services/notification/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// TemplateRepository handles database operations for notification templates.
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new template repository.
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create creates a new notification template.
func (r *TemplateRepository) Create(ctx context.Context, template *models.NotificationTemplate) *errors.Error {
	var metadataJSON []byte
	var err error

	if template.Metadata != nil {
		metadataJSON, err = json.Marshal(template.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO notification_templates (
			name, channel, subject_template, body_template, version, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		template.Name,
		template.Channel,
		template.SubjectTemplate,
		template.BodyTemplate,
		template.Version,
		metadataJSON,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		// Check for duplicate name
		if strings.Contains(err.Error(), "notification_templates_name_key") {
			return errors.Conflict("template with this name already exists")
		}
		return errors.DatabaseWrap(err, "failed to create template")
	}

	return nil
}

// GetByID retrieves a template by ID.
func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.NotificationTemplate, *errors.Error) {
	template := &models.NotificationTemplate{}
	var metadataJSON []byte

	query := `
		SELECT id, name, channel, subject_template, body_template, version,
		       metadata, created_at, updated_at
		FROM notification_templates
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Channel,
		&template.SubjectTemplate,
		&template.BodyTemplate,
		&template.Version,
		&metadataJSON,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("template", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get template")
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
			return nil, errors.Internal("failed to unmarshal metadata")
		}
	}

	return template, nil
}

// GetByName retrieves a template by name.
func (r *TemplateRepository) GetByName(ctx context.Context, name string) (*models.NotificationTemplate, *errors.Error) {
	template := &models.NotificationTemplate{}
	var metadataJSON []byte

	query := `
		SELECT id, name, channel, subject_template, body_template, version,
		       metadata, created_at, updated_at
		FROM notification_templates
		WHERE name = $1
	`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Channel,
		&template.SubjectTemplate,
		&template.BodyTemplate,
		&template.Version,
		&metadataJSON,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("template")
		}
		return nil, errors.DatabaseWrap(err, "failed to get template by name")
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
			return nil, errors.Internal("failed to unmarshal metadata")
		}
	}

	return template, nil
}

// List retrieves all templates, optionally filtered by channel.
func (r *TemplateRepository) List(ctx context.Context, channel *models.NotificationChannel) ([]*models.NotificationTemplate, *errors.Error) {
	var query string
	var rows *sql.Rows
	var err error

	if channel != nil {
		query = `
			SELECT id, name, channel, subject_template, body_template, version,
			       metadata, created_at, updated_at
			FROM notification_templates
			WHERE channel = $1
			ORDER BY name ASC
		`
		rows, err = r.db.QueryContext(ctx, query, *channel)
	} else {
		query = `
			SELECT id, name, channel, subject_template, body_template, version,
			       metadata, created_at, updated_at
			FROM notification_templates
			ORDER BY name ASC
		`
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list templates")
	}
	defer func() {
		_ = rows.Close()
	}()

	templates := make([]*models.NotificationTemplate, 0)
	for rows.Next() {
		template := &models.NotificationTemplate{}
		var metadataJSON []byte

		if err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Channel,
			&template.SubjectTemplate,
			&template.BodyTemplate,
			&template.Version,
			&metadataJSON,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan template")
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &template.Metadata); err != nil {
				return nil, errors.Internal("failed to unmarshal metadata")
			}
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating templates")
	}

	return templates, nil
}

// Update updates a notification template.
func (r *TemplateRepository) Update(ctx context.Context, id string, req *models.UpdateTemplateRequest) *errors.Error {
	// Build dynamic update
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if req.SubjectTemplate != nil {
		setClauses = append(setClauses, "subject_template = $"+fmt.Sprint(argIndex))
		args = append(args, *req.SubjectTemplate)
		argIndex++
	}

	if req.BodyTemplate != nil {
		setClauses = append(setClauses, "body_template = $"+fmt.Sprint(argIndex))
		args = append(args, *req.BodyTemplate)
		argIndex++
		// Increment version when body is updated
		setClauses = append(setClauses, "version = version + 1")
	}

	if len(req.MetadataRaw) > 0 {
		metadata, err := req.GetMetadata()
		if err != nil {
			return errors.Validation("invalid metadata JSON")
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
		setClauses = append(setClauses, "metadata = $"+fmt.Sprint(argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(setClauses) == 0 {
		return errors.Validation("no fields to update")
	}

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	// Add ID to args
	args = append(args, id)

	//nolint:gosec // setClauses is built from controlled field names, not user input
	query := fmt.Sprintf(`
		UPDATE notification_templates
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to update template")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFoundWithID("template", id)
	}

	return nil
}

// Delete deletes a notification template.
func (r *TemplateRepository) Delete(ctx context.Context, id string) *errors.Error {
	query := "DELETE FROM notification_templates WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete template")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFoundWithID("template", id)
	}

	return nil
}
