package repository

import (
	"context"
	"database/sql"

	"billing/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"

	"github.com/google/uuid"
)

// ============================================================================
// PLAN REPOSITORY — Read-only (plans seeded via migrations)
// ============================================================================

type PlanRepository struct {
	db *db.DB
}

func NewPlanRepository(database *db.DB) *PlanRepository {
	return &PlanRepository{db: database}
}

const planCols = `
	id, tier, name, description, billing_period,
	price, price_per_month, currency,
	discount_percentage, savings_amount,
	lemon_squeezy_variant_id,
	monthly_events_limit, retention_days, team_seats_limit,
	rate_limit_per_second, projects_limit,
	features, display_order, is_active, is_popular,
	created_at, updated_at`

func (r *PlanRepository) scanPlan(scanner interface{ Scan(...any) error }) (*models.SubscriptionPlan, error) {
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

func (r *PlanRepository) GetByTierAndBilling(ctx context.Context, tier models.PlanTier, period models.BillingPeriod) (*models.SubscriptionPlan, error) {
	query := `SELECT ` + planCols + ` FROM subscription_plans WHERE tier = $1 AND billing_period = $2 AND is_active = TRUE LIMIT 1`

	plan, err := r.scanPlan(r.db.QueryRow(ctx, query, tier, period))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Plan not found")
	}
	if err != nil {
		logger.Error("Failed to get plan", err, "tier", tier, "billing", period)
		return nil, errors.Internal(err)
	}
	return plan, nil
}

func (r *PlanRepository) GetByVariantID(ctx context.Context, variantID string) (*models.SubscriptionPlan, error) {
	query := `SELECT ` + planCols + ` FROM subscription_plans WHERE lemon_squeezy_variant_id = $1 AND is_active = TRUE LIMIT 1`

	plan, err := r.scanPlan(r.db.QueryRow(ctx, query, variantID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Plan not found")
	}
	if err != nil {
		logger.Error("Failed to get plan by variant", err, "variant_id", variantID)
		return nil, errors.Internal(err)
	}
	return plan, nil
}

func (r *PlanRepository) ListAll(ctx context.Context) ([]*models.SubscriptionPlan, error) {
	query := `SELECT ` + planCols + ` FROM subscription_plans WHERE is_active = TRUE ORDER BY display_order ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.Error("Failed to list plans", err)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var plans []*models.SubscriptionPlan
	for rows.Next() {
		plan, err := r.scanPlan(rows)
		if err != nil {
			logger.Error("Failed to scan plan", err)
			return nil, errors.Internal(err)
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	query := `SELECT ` + planCols + ` FROM subscription_plans WHERE id = $1`
	plan, err := r.scanPlan(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Plan not found")
	}
	if err != nil {
		return nil, errors.Internal(err)
	}
	return plan, nil
}