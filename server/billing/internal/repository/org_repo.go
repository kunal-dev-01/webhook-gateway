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
// ORG REPOSITORY — Billing reads + subscription sync writes
// ============================================================================

type OrgRepository struct {
	db *db.DB
}

func NewOrgRepository(database *db.DB) *OrgRepository {
	return &OrgRepository{db: database}
}

func (r *OrgRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *OrgRepository) scanOrg(scanner interface{ Scan(...any) error }) (*models.Organization, error) {
	org := &models.Organization{}
	err := scanner.Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID,
		&org.PlanTier, &org.BillingPeriod, &org.SubscriptionStatus,
		&org.LemonSqueezyCustomerID, &org.LemonSqueezySubscriptionID,
		&org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.MonthlyEventsLimit, &org.MonthlyEventsUsed, &org.RetentionDays,
		&org.TeamSeatsLimit, &org.RateLimitPerSecond, &org.ProjectsLimit,
		&org.Metadata, &org.DeletedAt, &org.CreatedAt, &org.UpdatedAt,
	)
	return org, err
}

// ── Read ─────────────────────────────────────────────────────────────────────

func (r *OrgRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT
			id, name, slug, owner_id, plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id,
			current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days,
			team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, deleted_at, created_at, updated_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL`

	org, err := r.scanOrg(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Organization not found")
	}
	if err != nil {
		logger.Error("Failed to get organization", err, "org_id", id)
		return nil, errors.Internal(err)
	}
	return org, nil
}

// ── Write (Subscription Sync) ────────────────────────────────────────────────

// SyncSubscription updates org's billing fields after webhook processing
func (r *OrgRepository) SyncSubscription(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()

	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE organizations SET
			plan_tier = $2,
			billing_period = $3,
			subscription_status = $4,
			lemon_squeezy_customer_id = $5,
			lemon_squeezy_subscription_id = $6,
			current_period_start = $7,
			current_period_end = $8,
			monthly_events_limit = $9,
			retention_days = $10,
			team_seats_limit = $11,
			rate_limit_per_second = $12,
			projects_limit = $13,
			updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(writeCtx, query,
		org.ID, org.PlanTier, org.BillingPeriod, org.SubscriptionStatus,
		org.LemonSqueezyCustomerID, org.LemonSqueezySubscriptionID,
		org.CurrentPeriodStart, org.CurrentPeriodEnd,
		org.MonthlyEventsLimit, org.RetentionDays,
		org.TeamSeatsLimit, org.RateLimitPerSecond, org.ProjectsLimit,
		org.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to sync org subscription", err, "org_id", org.ID)
		return errors.Wrap(err, 500, "Failed to sync organization subscription")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Organization not found")
	}

	logger.Debug("Org subscription synced", "org_id", org.ID, "plan", org.PlanTier, "status", org.SubscriptionStatus)
	return nil
}