package repository

import (
	"authentication/internal/db"
	"context"
	"database/sql"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

type SubscriptionPlanRepository struct {
	db *db.DB
}

func NewSubscriptionPlanRepository(database *db.DB) *SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{db: database}
}

// ───────────────── Helpers (Optimization) ─────────────────
// Centralized column list to avoid typos and make queries cleaner
const planCols = `
    id, tier, name, description, billing_period,
    price, price_per_month, currency,
    discount_percentage, savings_amount,
    lemon_squeezy_variant_id,
    monthly_events_limit, retention_days, team_seats_limit,
    rate_limit_per_second, projects_limit,
    features, display_order, is_active, is_popular,
    created_at, updated_at`

// scanPlan handles scanning a single row to avoid code duplication across methods
func (r *SubscriptionPlanRepository) scanPlan(scanner interface {
	Scan(dest ...any) error
}) (*models.SubscriptionPlan, error) {
	plan := &models.SubscriptionPlan{}
	err := scanner.Scan(
		&plan.ID, &plan.Tier, &plan.Name, &plan.Description, &plan.BillingPeriod,
		&plan.Price, &plan.PricePerMonth, &plan.Currency,
		&plan.DiscountPercentage, &plan.SavingsAmount,
		&plan.LemonSqueezyVariantID,
		&plan.MonthlyEventsLimit, &plan.RetentionDays, &plan.TeamSeatsLimit,
		&plan.RateLimitPerSecond, &plan.ProjectsLimit,
		&plan.Features, &plan.DisplayOrder, &plan.IsActive, &plan.IsPopular,
		&plan.CreatedAt, &plan.UpdatedAt,
	)
	return plan, err
}

// ============================================================================
// READ OPERATIONS
// ============================================================================
// GetByTierAndBilling retrieves a plan by tier and billing period
func (r *SubscriptionPlanRepository) GetByTierAndBilling(
	ctx context.Context,
	tier models.PlanTier,
	billingPeriod models.BillingPeriod,
) (*models.SubscriptionPlan, error) {
	query := `SELECT ` + planCols + ` FROM subscription_plans WHERE tier = $1 AND billing_period = $2 AND is_active = TRUE LIMIT 1`
	plan, err := r.scanPlan(r.db.QueryRow(ctx, query, tier, billingPeriod))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Subscription plan not found")
	}
	if err != nil {
		logger.Error("Failed to get subscription plan", err, "tier", tier, "billing", billingPeriod)
		return nil, errors.Internal(err)
	}
	return plan, nil
}
