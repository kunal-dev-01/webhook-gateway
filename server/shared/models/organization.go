package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// PlanTier represents subscription plan tiers
type PlanTier string

const (
	PlanDeveloper PlanTier = "developer"
	PlanStarter   PlanTier = "starter"
	PlanGrowth    PlanTier = "growth"
	PlanScale     PlanTier = "scale"
)

// BillingPeriod represents billing cycle
type BillingPeriod string

const (
	BillingMonthly BillingPeriod = "monthly"
	BillingAnnual  BillingPeriod = "annual"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionPastDue  SubscriptionStatus = "past_due"
	SubscriptionCanceled SubscriptionStatus = "canceled"
	SubscriptionPaused   SubscriptionStatus = "paused"
	SubscriptionExpired  SubscriptionStatus = "expired"
)

// ============================================================================
// ORGANIZATION METADATA
// ============================================================================

// OrganizationMetadata represents organization metadata stored as JSONB
type OrganizationMetadata struct {
	WebhookEndpoints []string          `json:"webhook_endpoints,omitempty"`
	CustomSettings   map[string]string `json:"custom_settings,omitempty"`
}

// Value implements driver.Valuer for OrganizationMetadata
func (m OrganizationMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for OrganizationMetadata
func (m *OrganizationMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal OrganizationMetadata value")
	}
	return json.Unmarshal(bytes, m)
}

// ============================================================================
// ORGANIZATION MODEL
// ============================================================================

// Organization represents an organization/team in Upnixo
type Organization struct {
	// Primary identification
	ID   uuid.UUID `db:"id" json:"id"`
	Name string    `db:"name" json:"name"`
	Slug string    `db:"slug" json:"slug"` // URL-friendly identifier

	// Owner
	OwnerID uuid.UUID `db:"owner_id" json:"owner_id"`

	// Subscription Plan
	PlanTier           PlanTier           `db:"plan_tier" json:"plan_tier"`
	BillingPeriod      BillingPeriod      `db:"billing_period" json:"billing_period"`
	SubscriptionStatus SubscriptionStatus `db:"subscription_status" json:"subscription_status"`

	// LemonSqueezy payment info
	LemonSqueezyCustomerID     *string `db:"lemon_squeezy_customer_id" json:"lemon_squeezy_customer_id,omitempty"`
	LemonSqueezySubscriptionID *string `db:"lemon_squeezy_subscription_id" json:"lemon_squeezy_subscription_id,omitempty"`

	// Subscription period
	CurrentPeriodStart *time.Time `db:"current_period_start" json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `db:"current_period_end" json:"current_period_end,omitempty"`

	// Plan limits (synced from subscription_plans)
	MonthlyEventsLimit  int64 `db:"monthly_events_limit" json:"monthly_events_limit"`
	MonthlyEventsUsed   int64 `db:"monthly_events_used" json:"monthly_events_used"`
	RetentionDays       int   `db:"retention_days" json:"retention_days"`
	TeamSeatsLimit      int   `db:"team_seats_limit" json:"team_seats_limit"`
	RateLimitPerSecond  int   `db:"rate_limit_per_second" json:"rate_limit_per_second"`
	ProjectsLimit       int   `db:"projects_limit" json:"projects_limit"`

	// Metadata
	Metadata OrganizationMetadata `db:"metadata" json:"metadata"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the organization struct
func (o *Organization) Validate() error {
	if err := utils.ValidateRequired(o.Name, "organization name"); err != nil {
		return err
	}
	if err := utils.ValidateLength(o.Name, 2, 255, "organization name"); err != nil {
		return err
	}
	if err := utils.ValidateRequired(o.Slug, "slug"); err != nil {
		return err
	}
	if err := utils.ValidateSlug(o.Slug); err != nil {
		return err
	}
	if o.OwnerID == uuid.Nil {
		return fmt.Errorf("owner ID is required")
	}
	return nil
}

// ============================================================================
// USAGE TRACKING
// ============================================================================

// IncrementEventsUsed increments the monthly events counter
func (o *Organization) IncrementEventsUsed(count int64) {
	o.MonthlyEventsUsed += count
}

// ResetMonthlyUsage resets the monthly usage counter (called by cron)
func (o *Organization) ResetMonthlyUsage() {
	o.MonthlyEventsUsed = 0
}

// IsWithinEventsLimit checks if organization is within monthly events limit
func (o *Organization) IsWithinEventsLimit() bool {
	return o.MonthlyEventsUsed < o.MonthlyEventsLimit
}

// GetRemainingEvents returns remaining events for current month
func (o *Organization) GetRemainingEvents() int64 {
	remaining := o.MonthlyEventsLimit - o.MonthlyEventsUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetUsagePercentage returns usage percentage (0-100)
func (o *Organization) GetUsagePercentage() float64 {
	if o.MonthlyEventsLimit == 0 {
		return 0
	}
	percentage := (float64(o.MonthlyEventsUsed) / float64(o.MonthlyEventsLimit)) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}

// ============================================================================
// SUBSCRIPTION HELPERS
// ============================================================================

// IsSubscriptionActive checks if subscription is active
func (o *Organization) IsSubscriptionActive() bool {
	return o.SubscriptionStatus == SubscriptionActive
}

// IsPaidPlan checks if organization is on a paid plan
func (o *Organization) IsPaidPlan() bool {
	return o.PlanTier != PlanDeveloper
}

// IsFreePlan checks if organization is on free plan
func (o *Organization) IsFreePlan() bool {
	return o.PlanTier == PlanDeveloper
}

// ============================================================================
// SOFT DELETE
// ============================================================================

// SoftDelete soft deletes the organization
func (o *Organization) SoftDelete() {
	now := time.Now()
	o.DeletedAt = &now
	// Append timestamp to slug to allow recreation
	o.Slug = fmt.Sprintf("%s-deleted-%d", o.Slug, now.Unix())
}

// IsDeleted checks if organization is soft deleted
func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicOrganization represents safe organization data for API responses
type PublicOrganization struct {
	ID                 uuid.UUID          `json:"id"`
	Name               string             `json:"name"`
	Slug               string             `json:"slug"`
	OwnerID            uuid.UUID          `json:"owner_id"`
	PlanTier           PlanTier           `json:"plan_tier"`
	BillingPeriod      BillingPeriod      `json:"billing_period"`
	SubscriptionStatus SubscriptionStatus `json:"subscription_status"`
	CurrentPeriodEnd   *time.Time         `json:"current_period_end,omitempty"`
	MonthlyEventsLimit int64              `json:"monthly_events_limit"`
	MonthlyEventsUsed  int64              `json:"monthly_events_used"`
	RetentionDays      int                `json:"retention_days"`
	TeamSeatsLimit     int                `json:"team_seats_limit"`
	RateLimitPerSecond int                `json:"rate_limit_per_second"`
	ProjectsLimit      int                `json:"projects_limit"`
	CreatedAt          time.Time          `json:"created_at"`
}

// ToPublic converts Organization to PublicOrganization
func (o *Organization) ToPublic() *PublicOrganization {
	return &PublicOrganization{
		ID:                 o.ID,
		Name:               o.Name,
		Slug:               o.Slug,
		OwnerID:            o.OwnerID,
		PlanTier:           o.PlanTier,
		BillingPeriod:      o.BillingPeriod,
		SubscriptionStatus: o.SubscriptionStatus,
		CurrentPeriodEnd:   o.CurrentPeriodEnd,
		MonthlyEventsLimit: o.MonthlyEventsLimit,
		MonthlyEventsUsed:  o.MonthlyEventsUsed,
		RetentionDays:      o.RetentionDays,
		TeamSeatsLimit:     o.TeamSeatsLimit,
		RateLimitPerSecond: o.RateLimitPerSecond,
		ProjectsLimit:      o.ProjectsLimit,
		CreatedAt:          o.CreatedAt,
	}
}