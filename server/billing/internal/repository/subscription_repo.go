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
// SUBSCRIPTION REPOSITORY
// ============================================================================

type SubscriptionRepository struct {
	db *db.DB
}

func NewSubscriptionRepository(database *db.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: database}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (r *SubscriptionRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

const subCols = `
	id, organization_id, plan_id,
	lemon_squeezy_subscription_id, lemon_squeezy_customer_id, lemon_squeezy_order_id,
	status, current_period_start, current_period_end,
	cancel_at_period_end, canceled_at, paused_at, resumed_at,
	price_amount, currency, metadata,
	created_at, updated_at`

func (r *SubscriptionRepository) scanSubscription(scanner interface{ Scan(...any) error }) (*models.Subscription, error) {
	sub := &models.Subscription{}
	err := scanner.Scan(
		&sub.ID, &sub.OrganizationID, &sub.PlanID,
		&sub.LemonSqueezySubscriptionID, &sub.LemonSqueezyCustomerID, &sub.LemonSqueezyOrderID,
		&sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.PausedAt, &sub.ResumedAt,
		&sub.PriceAmount, &sub.Currency, &sub.Metadata,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	return sub, err
}

func (r *SubscriptionRepository) performUpdate(query, opName string, id uuid.UUID, args ...any) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to "+opName, err, "id", id)
		return errors.Wrap(err, 500, "Failed to "+opName)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Subscription not found or already in this state")
	}

	logger.Debug("Success: "+opName, "id", id)
	return nil
}

// ============================================================================
// CREATE
// ============================================================================

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO subscriptions (
			id, organization_id, plan_id,
			lemon_squeezy_subscription_id, lemon_squeezy_customer_id, lemon_squeezy_order_id,
			status, current_period_start, current_period_end,
			cancel_at_period_end, canceled_at, paused_at, resumed_at,
			price_amount, currency, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17, $18
		)`

	_, err := r.db.Exec(writeCtx, query,
		sub.ID, sub.OrganizationID, sub.PlanID,
		sub.LemonSqueezySubscriptionID, sub.LemonSqueezyCustomerID, sub.LemonSqueezyOrderID,
		sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CanceledAt, sub.PausedAt, sub.ResumedAt,
		sub.PriceAmount, sub.Currency, sub.Metadata,
		sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create subscription", err, "org_id", sub.OrganizationID)
		return errors.Wrap(err, 500, "Failed to create subscription")
	}

	logger.Debug("Subscription created", "sub_id", sub.ID, "org_id", sub.OrganizationID)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `SELECT ` + subCols + ` FROM subscriptions WHERE id = $1`

	sub, err := r.scanSubscription(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Subscription not found")
	}
	if err != nil {
		logger.Error("Failed to get subscription by ID", err, "sub_id", id)
		return nil, errors.Internal(err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*models.Subscription, error) {
	// Returns the latest active subscription for an org
	query := `
		SELECT ` + subCols + `
		FROM subscriptions
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	sub, err := r.scanSubscription(r.db.QueryRow(ctx, query, orgID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("No subscription found for this organization")
	}
	if err != nil {
		logger.Error("Failed to get subscription by org ID", err, "org_id", orgID)
		return nil, errors.Internal(err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) GetByLemonSqueezyID(ctx context.Context, lsSubID string) (*models.Subscription, error) {
	query := `SELECT ` + subCols + ` FROM subscriptions WHERE lemon_squeezy_subscription_id = $1`

	sub, err := r.scanSubscription(r.db.QueryRow(ctx, query, lsSubID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Subscription not found")
	}
	if err != nil {
		logger.Error("Failed to get subscription by LS ID", err, "ls_sub_id", lsSubID)
		return nil, errors.Internal(err)
	}
	return sub, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions SET
			plan_id = $2,
			status = $3,
			current_period_start = $4,
			current_period_end = $5,
			cancel_at_period_end = $6,
			canceled_at = $7,
			paused_at = $8,
			resumed_at = $9,
			price_amount = $10,
			metadata = $11,
			updated_at = $12
		WHERE id = $1`

	return r.performUpdate(query, "update subscription", sub.ID,
		sub.ID, sub.PlanID, sub.Status,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CanceledAt, sub.PausedAt, sub.ResumedAt,
		sub.PriceAmount, sub.Metadata, sub.UpdatedAt,
	)
}