package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// PLAN FEATURES
// ============================================================================

// PlanFeatures represents plan-specific features stored as JSONB
type PlanFeatures struct {
	RetryPolicy    string `json:"retry_policy"`     // "manual", "auto_3x", "smart_backoff", "smart_custom"
	Support        string `json:"support"`          // "community", "email_48h", "priority_email", "slack_priority"
	CustomWebhooks bool   `json:"custom_webhooks"`  // Custom webhook endpoints
	Analytics      bool   `json:"analytics"`        // Advanced analytics
	CustomDomains  bool   `json:"custom_domains"`   // Custom domain support
	DedicatedIP    bool   `json:"dedicated_ip"`     // Dedicated IP address
}

// Value implements driver.Valuer for PlanFeatures
func (f PlanFeatures) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements sql.Scanner for PlanFeatures
func (f *PlanFeatures) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal PlanFeatures value")
	}
	return json.Unmarshal(bytes, f)
}

// ============================================================================
// SUBSCRIPTION PLAN MODEL
// ============================================================================

// SubscriptionPlan represents a pricing plan
type SubscriptionPlan struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Plan identification
	Tier        PlanTier      `db:"tier" json:"tier"`                 // 'developer', 'starter', 'growth', 'scale'
	Name        string        `db:"name" json:"name"`                 // Display name
	Description string        `db:"description" json:"description"`   // Plan description

	// Billing period
	BillingPeriod BillingPeriod `db:"billing_period" json:"billing_period"` // 'monthly', 'annual'

	// Pricing
	Price         float64 `db:"price" json:"price"`                   // Total price for billing period
	PricePerMonth float64 `db:"price_per_month" json:"price_per_month"` // Effective monthly cost
	Currency      string  `db:"currency" json:"currency"`             // 'USD'

	// Discount info (for annual plans)
	DiscountPercentage int     `db:"discount_percentage" json:"discount_percentage"` // Annual discount %
	SavingsAmount      float64 `db:"savings_amount" json:"savings_amount"`           // Annual savings vs monthly

	// LemonSqueezy integration
	LemonSqueezyVariantID string `db:"lemon_squeezy_variant_id" json:"lemon_squeezy_variant_id"`

	// Plan limits
	MonthlyEventsLimit int64 `db:"monthly_events_limit" json:"monthly_events_limit"`
	RetentionDays      int   `db:"retention_days" json:"retention_days"`
	TeamSeatsLimit     int   `db:"team_seats_limit" json:"team_seats_limit"`
	RateLimitPerSecond int   `db:"rate_limit_per_second" json:"rate_limit_per_second"`
	ProjectsLimit      int   `db:"projects_limit" json:"projects_limit"`

	// Features
	Features PlanFeatures `db:"features" json:"features"`

	// Display settings
	DisplayOrder int  `db:"display_order" json:"display_order"`
	IsActive     bool `db:"is_active" json:"is_active"`
	IsPopular    bool `db:"is_popular" json:"is_popular"` // Badge for popular plans

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// PLAN HELPERS
// ============================================================================

// IsFreePlan checks if this is the free developer plan
func (p *SubscriptionPlan) IsFreePlan() bool {
	return p.Tier == PlanDeveloper
}

// IsAnnualPlan checks if this is an annual billing plan
func (p *SubscriptionPlan) IsAnnualPlan() bool {
	return p.BillingPeriod == BillingAnnual
}

// GetFormattedPrice returns formatted price string
func (p *SubscriptionPlan) GetFormattedPrice() string {
	if p.IsFreePlan() {
		return "Free"
	}
	if p.IsAnnualPlan() {
		return fmt.Sprintf("$%.2f/year", p.Price)
	}
	return fmt.Sprintf("$%.2f/month", p.Price)
}

// GetFormattedPricePerMonth returns formatted monthly price
func (p *SubscriptionPlan) GetFormattedPricePerMonth() string {
	if p.IsFreePlan() {
		return "Free"
	}
	return fmt.Sprintf("$%.2f/month", p.PricePerMonth)
}

// GetSavingsText returns savings text for annual plans
func (p *SubscriptionPlan) GetSavingsText() string {
	if !p.IsAnnualPlan() || p.SavingsAmount == 0 {
		return ""
	}
	return fmt.Sprintf("Save $%.2f/year", p.SavingsAmount)
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the subscription plan struct
func (p *SubscriptionPlan) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("plan name is required")
	}
	if p.LemonSqueezyVariantID == "" {
		return fmt.Errorf("LemonSqueezy variant ID is required")
	}
	if p.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if p.MonthlyEventsLimit <= 0 {
		return fmt.Errorf("monthly events limit must be positive")
	}
	if p.RetentionDays <= 0 {
		return fmt.Errorf("retention days must be positive")
	}
	if p.TeamSeatsLimit <= 0 {
		return fmt.Errorf("team seats limit must be positive")
	}
	if p.ProjectsLimit <= 0 {
		return fmt.Errorf("projects limit must be positive")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicSubscriptionPlan represents safe plan data for API responses (pricing page)
type PublicSubscriptionPlan struct {
	ID                 uuid.UUID     `json:"id"`
	Tier               PlanTier      `json:"tier"`
	Name               string        `json:"name"`
	Description        string        `json:"description"`
	BillingPeriod      BillingPeriod `json:"billing_period"`
	Price              float64       `json:"price"`
	PricePerMonth      float64       `json:"price_per_month"`
	Currency           string        `json:"currency"`
	DiscountPercentage int           `json:"discount_percentage"`
	SavingsAmount      float64       `json:"savings_amount"`
	MonthlyEventsLimit int64         `json:"monthly_events_limit"`
	RetentionDays      int           `json:"retention_days"`
	TeamSeatsLimit     int           `json:"team_seats_limit"`
	RateLimitPerSecond int           `json:"rate_limit_per_second"`
	ProjectsLimit      int           `json:"projects_limit"`
	Features           PlanFeatures  `json:"features"`
	IsPopular          bool          `json:"is_popular"`
}

// ToPublic converts SubscriptionPlan to PublicSubscriptionPlan
func (p *SubscriptionPlan) ToPublic() *PublicSubscriptionPlan {
	return &PublicSubscriptionPlan{
		ID:                 p.ID,
		Tier:               p.Tier,
		Name:               p.Name,
		Description:        p.Description,
		BillingPeriod:      p.BillingPeriod,
		Price:              p.Price,
		PricePerMonth:      p.PricePerMonth,
		Currency:           p.Currency,
		DiscountPercentage: p.DiscountPercentage,
		SavingsAmount:      p.SavingsAmount,
		MonthlyEventsLimit: p.MonthlyEventsLimit,
		RetentionDays:      p.RetentionDays,
		TeamSeatsLimit:     p.TeamSeatsLimit,
		RateLimitPerSecond: p.RateLimitPerSecond,
		ProjectsLimit:      p.ProjectsLimit,
		Features:           p.Features,
		IsPopular:          p.IsPopular,
	}
}