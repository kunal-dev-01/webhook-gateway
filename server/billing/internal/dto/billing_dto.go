package dto

import (
	"time"

	"github.com/google/uuid"

	"shared/models"
)

// ============================================================================
// PLAN DTOs
// ============================================================================

// PublicPlanWithSavings — enriched plan for pricing page
type PublicPlanWithSavings struct {
	*models.PublicSubscriptionPlan
	FormattedPrice         string `json:"formatted_price"`
	FormattedPricePerMonth string `json:"formatted_price_per_month"`
	SavingsText            string `json:"savings_text,omitempty"`
}

// ListPlansResponse — pricing page
type ListPlansResponse struct {
	Plans []*PublicPlanWithSavings `json:"plans"`
}

// ============================================================================
// CHECKOUT DTOs
// ============================================================================

// CreateCheckoutRequest — user wants to subscribe to a plan
type CreateCheckoutRequest struct {
	PlanTier      models.PlanTier      `json:"plan_tier"`
	BillingPeriod models.BillingPeriod `json:"billing_period"`
}

// CreateCheckoutResponse — returns the checkout URL
type CreateCheckoutResponse struct {
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ============================================================================
// SUBSCRIPTION DTOs
// ============================================================================

// SubscriptionResponse — full subscription with plan details
type SubscriptionResponse struct {
	Subscription *models.PublicSubscription     `json:"subscription"`
	Plan         *models.PublicSubscriptionPlan `json:"plan"`
	// Customer portal URL for payment method updates
	CustomerPortalURL *string `json:"customer_portal_url,omitempty"`
}

// CancelSubscriptionRequest — optional reason
type CancelSubscriptionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ChangePlanRequest — upgrade / downgrade
type ChangePlanRequest struct {
	PlanTier      models.PlanTier      `json:"plan_tier"`
	BillingPeriod models.BillingPeriod `json:"billing_period"`
}

// ============================================================================
// PAYMENT DTOs
// ============================================================================

// ListPaymentsRequest — pagination
type ListPaymentsRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ListPaymentsResponse — paginated payment history
type ListPaymentsResponse struct {
	Payments []*models.PublicPayment `json:"payments"`
	Total    int                     `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}


// ============================================================================
// USAGE DTOs
// ============================================================================

// UsageResponse — current org usage
type UsageResponse struct {
	OrganizationID     uuid.UUID `json:"organization_id"`
	MonthlyEventsLimit int64     `json:"monthly_events_limit"`
	MonthlyEventsUsed  int64     `json:"monthly_events_used"`
	RemainingEvents    int64     `json:"remaining_events"`
	UsagePercentage    float64   `json:"usage_percentage"`
	RetentionDays      int       `json:"retention_days"`
	TeamSeatsLimit     int       `json:"team_seats_limit"`
	ProjectsLimit      int       `json:"projects_limit"`
	RateLimitPerSecond int       `json:"rate_limit_per_second"`
}

// ToUsageResponse builds UsageResponse from Organization
func ToUsageResponse(org *models.Organization) *UsageResponse {
	return &UsageResponse{
		OrganizationID:     org.ID,
		MonthlyEventsLimit: org.MonthlyEventsLimit,
		MonthlyEventsUsed:  org.MonthlyEventsUsed,
		RemainingEvents:    org.GetRemainingEvents(),
		UsagePercentage:    org.GetUsagePercentage(),
		RetentionDays:      org.RetentionDays,
		TeamSeatsLimit:     org.TeamSeatsLimit,
		ProjectsLimit:      org.ProjectsLimit,
		RateLimitPerSecond: org.RateLimitPerSecond,
	}
}