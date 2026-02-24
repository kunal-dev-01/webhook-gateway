package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"billing/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// WEBHOOK EVENT REPOSITORY
// ============================================================================

type WebhookEventRepository struct {
	db *db.DB
}

func NewWebhookEventRepository(database *db.DB) *WebhookEventRepository {
	return &WebhookEventRepository{db: database}
}

func (r *WebhookEventRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

const webhookEventCols = `
	id, lemon_squeezy_event_id, event_type, status,
	organization_id, subscription_id, payment_id,
	payload, processed_at, failed_at, error_message, retry_count,
	created_at, updated_at`

func (r *WebhookEventRepository) scanEvent(scanner interface{ Scan(...any) error }) (*models.WebhookEvent, error) {
	e := &models.WebhookEvent{}
	err := scanner.Scan(
		&e.ID, &e.LemonSqueezyEventID, &e.EventType, &e.Status,
		&e.OrganizationID, &e.SubscriptionID, &e.PaymentID,
		&e.Payload, &e.ProcessedAt, &e.FailedAt, &e.ErrorMessage, &e.RetryCount,
		&e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *WebhookEventRepository) Create(ctx context.Context, event *models.WebhookEvent) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO webhook_events (
			id, lemon_squeezy_event_id, event_type, status,
			organization_id, subscription_id, payment_id,
			payload, processed_at, failed_at, error_message, retry_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)`

	_, err := r.db.Exec(writeCtx, query,
		event.ID, event.LemonSqueezyEventID, event.EventType, event.Status,
		event.OrganizationID, event.SubscriptionID, event.PaymentID,
		event.Payload, event.ProcessedAt, event.FailedAt, event.ErrorMessage, event.RetryCount,
		event.CreatedAt, event.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create webhook event", err, "event_type", event.EventType)
		return errors.Wrap(err, 500, "Failed to create webhook event")
	}

	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *WebhookEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WebhookEvent, error) {
	query := `SELECT ` + webhookEventCols + ` FROM webhook_events WHERE id = $1`

	e, err := r.scanEvent(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Webhook event not found")
	}
	if err != nil {
		return nil, errors.Internal(err)
	}
	return e, nil
}

// ExistsByLemonSqueezyEventID — idempotency check
func (r *WebhookEventRepository) ExistsByLemonSqueezyEventID(ctx context.Context, lsEventID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM webhook_events WHERE lemon_squeezy_event_id = $1 AND status = 'processed')`,
		lsEventID,
	).Scan(&exists); err != nil {
		return false, errors.Internal(err)
	}
	return exists, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *WebhookEventRepository) Update(ctx context.Context, event *models.WebhookEvent) error {
	event.UpdatedAt = time.Now()
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE webhook_events SET
			status = $2,
			organization_id = $3,
			subscription_id = $4,
			payment_id = $5,
			processed_at = $6,
			failed_at = $7,
			error_message = $8,
			retry_count = $9,
			updated_at = $10
		WHERE id = $1`

	_, err := r.db.Exec(writeCtx, query,
		event.ID, event.Status,
		event.OrganizationID, event.SubscriptionID, event.PaymentID,
		event.ProcessedAt, event.FailedAt, event.ErrorMessage, event.RetryCount,
		event.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to update webhook event", err, "event_id", event.ID)
		return errors.Wrap(err, 500, "Failed to update webhook event")
	}

	return nil
}