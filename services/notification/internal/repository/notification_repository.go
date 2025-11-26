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

// NotificationRepository handles database operations for notifications.
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification.
func (r *NotificationRepository) Create(ctx context.Context, notif *models.Notification) *errors.Error {
	var metadataJSON []byte
	var err error

	if notif.Metadata != nil {
		metadataJSON, err = json.Marshal(notif.Metadata)
		if err != nil {
			return errors.Internal("failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO notifications (
			user_id, channel, type, priority, recipient, subject, body,
			template_id, status, correlation_id, source_service, metadata,
			retry_count, queued_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		notif.UserID,
		notif.Channel,
		notif.Type,
		notif.Priority,
		notif.Recipient,
		notif.Subject,
		notif.Body,
		notif.TemplateID,
		notif.Status,
		notif.CorrelationID,
		notif.SourceService,
		metadataJSON,
		notif.RetryCount,
		notif.QueuedAt,
	).Scan(&notif.ID, &notif.CreatedAt, &notif.UpdatedAt)

	if err != nil {
		// Check for duplicate correlation_id (idempotency)
		if strings.Contains(err.Error(), "idx_notifications_correlation_id_unique") {
			return errors.Conflict("notification with this correlation_id already exists")
		}
		return errors.DatabaseWrap(err, "failed to create notification")
	}

	return nil
}

// GetByID retrieves a notification by ID.
func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, *errors.Error) {
	notif := &models.Notification{}
	var metadataJSON []byte

	query := `
		SELECT id, user_id, channel, type, priority, recipient, subject, body,
		       template_id, status, correlation_id, source_service, metadata,
		       retry_count, failure_reason, queued_at, sent_at, delivered_at,
		       failed_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notif.ID,
		&notif.UserID,
		&notif.Channel,
		&notif.Type,
		&notif.Priority,
		&notif.Recipient,
		&notif.Subject,
		&notif.Body,
		&notif.TemplateID,
		&notif.Status,
		&notif.CorrelationID,
		&notif.SourceService,
		&metadataJSON,
		&notif.RetryCount,
		&notif.FailureReason,
		&notif.QueuedAt,
		&notif.SentAt,
		&notif.DeliveredAt,
		&notif.FailedAt,
		&notif.CreatedAt,
		&notif.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("notification", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get notification")
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
			return nil, errors.Internal("failed to unmarshal metadata")
		}
	}

	return notif, nil
}

// GetByCorrelationID retrieves a notification by correlation ID (for idempotency check).
func (r *NotificationRepository) GetByCorrelationID(ctx context.Context, correlationID string) (*models.Notification, *errors.Error) {
	notif := &models.Notification{}
	var metadataJSON []byte

	query := `
		SELECT id, user_id, channel, type, priority, recipient, subject, body,
		       template_id, status, correlation_id, source_service, metadata,
		       retry_count, failure_reason, queued_at, sent_at, delivered_at,
		       failed_at, created_at, updated_at
		FROM notifications
		WHERE correlation_id = $1
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&notif.ID,
		&notif.UserID,
		&notif.Channel,
		&notif.Type,
		&notif.Priority,
		&notif.Recipient,
		&notif.Subject,
		&notif.Body,
		&notif.TemplateID,
		&notif.Status,
		&notif.CorrelationID,
		&notif.SourceService,
		&metadataJSON,
		&notif.RetryCount,
		&notif.FailureReason,
		&notif.QueuedAt,
		&notif.SentAt,
		&notif.DeliveredAt,
		&notif.FailedAt,
		&notif.CreatedAt,
		&notif.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("notification")
		}
		return nil, errors.DatabaseWrap(err, "failed to get notification by correlation_id")
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
			return nil, errors.Internal("failed to unmarshal metadata")
		}
	}

	return notif, nil
}

// List retrieves notifications with optional filters.
func (r *NotificationRepository) List(ctx context.Context, req *models.ListNotificationsRequest) ([]*models.Notification, int64, *errors.Error) {
	// Build dynamic WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.Channel != nil {
		conditions = append(conditions, fmt.Sprintf("channel = $%d", argIndex))
		args = append(args, *req.Channel)
		argIndex++
	}

	if req.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *req.Type)
		argIndex++
	}

	if req.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.SourceService != nil {
		conditions = append(conditions, fmt.Sprintf("source_service = $%d", argIndex))
		args = append(args, *req.SourceService)
		argIndex++
	}

	if req.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, req.StartDate)
		argIndex++
	}

	if req.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, req.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	//nolint:gosec // whereClause is built from controlled filter values, not user input
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to count notifications")
	}

	// Get notifications with pagination
	limit := req.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}

	//nolint:gosec // whereClause is built from controlled filter values, not user input
	query := fmt.Sprintf(`
		SELECT id, user_id, channel, type, priority, recipient, subject, body,
		       template_id, status, correlation_id, source_service, metadata,
		       retry_count, failure_reason, queued_at, sent_at, delivered_at,
		       failed_at, created_at, updated_at
		FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, req.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.DatabaseWrap(err, "failed to list notifications")
	}
	defer func() {
		_ = rows.Close()
	}()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		notif := &models.Notification{}
		var metadataJSON []byte

		if err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Channel,
			&notif.Type,
			&notif.Priority,
			&notif.Recipient,
			&notif.Subject,
			&notif.Body,
			&notif.TemplateID,
			&notif.Status,
			&notif.CorrelationID,
			&notif.SourceService,
			&metadataJSON,
			&notif.RetryCount,
			&notif.FailureReason,
			&notif.QueuedAt,
			&notif.SentAt,
			&notif.DeliveredAt,
			&notif.FailedAt,
			&notif.CreatedAt,
			&notif.UpdatedAt,
		); err != nil {
			return nil, 0, errors.DatabaseWrap(err, "failed to scan notification")
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
				return nil, 0, errors.Internal("failed to unmarshal metadata")
			}
		}

		notifications = append(notifications, notif)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.DatabaseWrap(err, "error iterating notifications")
	}

	return notifications, total, nil
}

// UpdateStatus updates the status of a notification.
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, failureReason *string) *errors.Error {
	query := `
		UPDATE notifications
		SET status = $1::text,
		    failure_reason = $2,
		    sent_at = CASE WHEN $1::text = 'sent' AND sent_at IS NULL THEN NOW() ELSE sent_at END,
		    delivered_at = CASE WHEN $1::text = 'delivered' AND delivered_at IS NULL THEN NOW() ELSE delivered_at END,
		    failed_at = CASE WHEN $1::text = 'failed' AND failed_at IS NULL THEN NOW() ELSE failed_at END,
		    updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, string(status), failureReason, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to update notification status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFoundWithID("notification", id)
	}

	return nil
}

// IncrementRetryCount increments the retry count for a notification.
func (r *NotificationRepository) IncrementRetryCount(ctx context.Context, id string) *errors.Error {
	query := `
		UPDATE notifications
		SET retry_count = retry_count + 1,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to increment retry count")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.DatabaseWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFoundWithID("notification", id)
	}

	return nil
}

// GetQueuedNotifications retrieves queued notifications ordered by priority and creation time.
func (r *NotificationRepository) GetQueuedNotifications(ctx context.Context, limit int) ([]*models.Notification, *errors.Error) {
	if limit == 0 {
		limit = 100 // Default batch size
	}

	query := `
		SELECT id, user_id, channel, type, priority, recipient, subject, body,
		       template_id, status, correlation_id, source_service, metadata,
		       retry_count, failure_reason, queued_at, sent_at, delivered_at,
		       failed_at, created_at, updated_at
		FROM notifications
		WHERE status = 'queued'
		ORDER BY
		    CASE priority
		        WHEN 'critical' THEN 1
		        WHEN 'high' THEN 2
		        WHEN 'normal' THEN 3
		        WHEN 'low' THEN 4
		    END,
		    created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get queued notifications")
	}
	defer func() {
		_ = rows.Close()
	}()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		notif := &models.Notification{}
		var metadataJSON []byte

		if err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Channel,
			&notif.Type,
			&notif.Priority,
			&notif.Recipient,
			&notif.Subject,
			&notif.Body,
			&notif.TemplateID,
			&notif.Status,
			&notif.CorrelationID,
			&notif.SourceService,
			&metadataJSON,
			&notif.RetryCount,
			&notif.FailureReason,
			&notif.QueuedAt,
			&notif.SentAt,
			&notif.DeliveredAt,
			&notif.FailedAt,
			&notif.CreatedAt,
			&notif.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan notification")
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
				return nil, errors.Internal("failed to unmarshal metadata")
			}
		}

		notifications = append(notifications, notif)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.DatabaseWrap(err, "error iterating notifications")
	}

	return notifications, nil
}

// GetStats retrieves notification statistics.
func (r *NotificationRepository) GetStats(ctx context.Context) (*models.NotificationStats, *errors.Error) {
	stats := &models.NotificationStats{
		ByChannel: make(map[models.NotificationChannel]int),
		ByStatus:  make(map[models.NotificationStatus]int),
		ByType:    make(map[models.NotificationType]int),
	}

	// Get total count
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications").Scan(&stats.TotalNotifications); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get total count")
	}

	// Get counts by channel
	channelQuery := "SELECT channel, COUNT(*) FROM notifications GROUP BY channel"
	rows, err := r.db.QueryContext(ctx, channelQuery)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get channel stats")
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var channel models.NotificationChannel
		var count int
		if err := rows.Scan(&channel, &count); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan channel stats")
		}
		stats.ByChannel[channel] = count
	}
	_ = rows.Close()

	// Get counts by status
	statusQuery := "SELECT status, COUNT(*) FROM notifications GROUP BY status"
	rows, err = r.db.QueryContext(ctx, statusQuery)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get status stats")
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var status models.NotificationStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan status stats")
		}
		stats.ByStatus[status] = count
	}
	_ = rows.Close()

	// Get counts by type
	typeQuery := "SELECT type, COUNT(*) FROM notifications GROUP BY type"
	rows, err = r.db.QueryContext(ctx, typeQuery)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get type stats")
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var notifType models.NotificationType
		var count int
		if err := rows.Scan(&notifType, &count); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan type stats")
		}
		stats.ByType[notifType] = count
	}
	_ = rows.Close()

	// Calculate success rate
	delivered := stats.ByStatus[models.StatusDelivered]
	failed := stats.ByStatus[models.StatusFailed]
	if (delivered + failed) > 0 {
		stats.SuccessRate = float64(delivered) / float64(delivered+failed) * 100
	}

	// Get average retries
	var avgRetries sql.NullFloat64
	retryQuery := "SELECT AVG(retry_count) FROM notifications WHERE status IN ('delivered', 'failed')"
	if err := r.db.QueryRowContext(ctx, retryQuery).Scan(&avgRetries); err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get average retries")
	}
	if avgRetries.Valid {
		stats.AverageRetries = avgRetries.Float64
	}

	return stats, nil
}
