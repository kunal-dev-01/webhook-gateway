package repository

import (
	"context"
	"database/sql"

	"authentication/internal/db"
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

// ============================================================================
// READ OPERATIONS
// ============================================================================

// GetByTierAndBilling retrieves a plan by tier and billing period
func (r *SubscriptionPlanRepository) GetByTierAndBilling(
	ctx context.Context,
	tier models.PlanTier,
	billingPeriod models.BillingPeriod,
) (*models.SubscriptionPlan, error) {
	// Log the request
	logger.Debug("Fetching subscription plan",
		"tier", tier,
		"billing", billingPeriod,
	)

	plan := &models.SubscriptionPlan{}
	query := `
		SELECT 
			id, tier, name, description, billing_period,
			price, price_per_month, currency,
			discount_percentage, savings_amount,
			lemon_squeezy_variant_id,
			monthly_events_limit, retention_days, team_seats_limit,
			rate_limit_per_second, projects_limit,
			features, display_order, is_active, is_popular,
			created_at, updated_at
		FROM subscription_plans
		WHERE tier = $1 AND billing_period = $2 AND is_active = TRUE
		LIMIT 1`

	// DB wrapper handles timeout internally
	err := r.db.QueryRow(ctx, query, tier, billingPeriod).Scan(
		&plan.ID, &plan.Tier, &plan.Name, &plan.Description, &plan.BillingPeriod,
		&plan.Price, &plan.PricePerMonth, &plan.Currency,
		&plan.DiscountPercentage, &plan.SavingsAmount,
		&plan.LemonSqueezyVariantID,
		&plan.MonthlyEventsLimit, &plan.RetentionDays, &plan.TeamSeatsLimit,
		&plan.RateLimitPerSecond, &plan.ProjectsLimit,
		&plan.Features, &plan.DisplayOrder, &plan.IsActive, &plan.IsPopular,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		logger.Warn("Subscription plan not found",
			"tier", tier,
			"billing", billingPeriod,
		)
		return nil, errors.NotFound("Subscription plan not found")
	}
	
	if err != nil {
		logger.Error("Failed to get subscription plan", err,
			"tier", tier,
			"billing", billingPeriod,
		)
		return nil, errors.Internal(err)
	}

	logger.Debug("Successfully fetched subscription plan",
		"plan_id", plan.ID,
		"plan_name", plan.Name,
	)

	return plan, nil
}

// GetByVariantID retrieves a plan by LemonSqueezy variant ID
func (r *SubscriptionPlanRepository) GetByVariantID(
	ctx context.Context, 
	variantID string,
) (*models.SubscriptionPlan, error) {
	plan := &models.SubscriptionPlan{}
	query := `
		SELECT 
			id, tier, name, description, billing_period,
			price, price_per_month, currency,
			discount_percentage, savings_amount,
			lemon_squeezy_variant_id,
			monthly_events_limit, retention_days, team_seats_limit,
			rate_limit_per_second, projects_limit,
			features, display_order, is_active, is_popular,
			created_at, updated_at
		FROM subscription_plans
		WHERE lemon_squeezy_variant_id = $1 AND is_active = TRUE
		LIMIT 1`

	err := r.db.QueryRow(ctx, query, variantID).Scan(
		&plan.ID, &plan.Tier, &plan.Name, &plan.Description, &plan.BillingPeriod,
		&plan.Price, &plan.PricePerMonth, &plan.Currency,
		&plan.DiscountPercentage, &plan.SavingsAmount,
		&plan.LemonSqueezyVariantID,
		&plan.MonthlyEventsLimit, &plan.RetentionDays, &plan.TeamSeatsLimit,
		&plan.RateLimitPerSecond, &plan.ProjectsLimit,
		&plan.Features, &plan.DisplayOrder, &plan.IsActive, &plan.IsPopular,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Subscription plan not found")
	}
	if err != nil {
		logger.Error("Failed to get subscription plan by variant", err, "variant_id", variantID)
		return nil, errors.Internal(err)
	}

	return plan, nil
}

// ListAll retrieves all active plans (for pricing page)
func (r *SubscriptionPlanRepository) ListAll(ctx context.Context) ([]*models.SubscriptionPlan, error) {
	query := `
		SELECT 
			id, tier, name, description, billing_period,
			price, price_per_month, currency,
			discount_percentage, savings_amount,
			lemon_squeezy_variant_id,
			monthly_events_limit, retention_days, team_seats_limit,
			rate_limit_per_second, projects_limit,
			features, display_order, is_active, is_popular,
			created_at, updated_at
		FROM subscription_plans
		WHERE is_active = TRUE
		ORDER BY display_order ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.Error("Failed to list subscription plans", err)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var plans []*models.SubscriptionPlan
	for rows.Next() {
		plan := &models.SubscriptionPlan{}
		if err := rows.Scan(
			&plan.ID, &plan.Tier, &plan.Name, &plan.Description, &plan.BillingPeriod,
			&plan.Price, &plan.PricePerMonth, &plan.Currency,
			&plan.DiscountPercentage, &plan.SavingsAmount,
			&plan.LemonSqueezyVariantID,
			&plan.MonthlyEventsLimit, &plan.RetentionDays, &plan.TeamSeatsLimit,
			&plan.RateLimitPerSecond, &plan.ProjectsLimit,
			&plan.Features, &plan.DisplayOrder, &plan.IsActive, &plan.IsPopular,
			&plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			logger.Error("Failed to scan subscription plan", err)
			return nil, errors.Internal(err)
		}
		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating subscription plans", err)
		return nil, errors.Internal(err)
	}

	return plans, nil
}

// ListByTier retrieves all plans for a specific tier (monthly + annual)
func (r *SubscriptionPlanRepository) ListByTier(
	ctx context.Context, 
	tier models.PlanTier,
) ([]*models.SubscriptionPlan, error) {
	query := `
		SELECT 
			id, tier, name, description, billing_period,
			price, price_per_month, currency,
			discount_percentage, savings_amount,
			lemon_squeezy_variant_id,
			monthly_events_limit, retention_days, team_seats_limit,
			rate_limit_per_second, projects_limit,
			features, display_order, is_active, is_popular,
			created_at, updated_at
		FROM subscription_plans
		WHERE tier = $1 AND is_active = TRUE
		ORDER BY billing_period ASC`

	rows, err := r.db.Query(ctx, query, tier)
	if err != nil {
		logger.Error("Failed to list plans by tier", err, "tier", tier)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var plans []*models.SubscriptionPlan
	for rows.Next() {
		plan := &models.SubscriptionPlan{}
		if err := rows.Scan(
			&plan.ID, &plan.Tier, &plan.Name, &plan.Description, &plan.BillingPeriod,
			&plan.Price, &plan.PricePerMonth, &plan.Currency,
			&plan.DiscountPercentage, &plan.SavingsAmount,
			&plan.LemonSqueezyVariantID,
			&plan.MonthlyEventsLimit, &plan.RetentionDays, &plan.TeamSeatsLimit,
			&plan.RateLimitPerSecond, &plan.ProjectsLimit,
			&plan.Features, &plan.DisplayOrder, &plan.IsActive, &plan.IsPopular,
			&plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			logger.Error("Failed to scan subscription plan", err)
			return nil, errors.Internal(err)
		}
		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating subscription plans", err)
		return nil, errors.Internal(err)
	}

	return plans, nil
}