package models

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ============================================================================
// ENUMS
// ============================================================================

// AddonType represents what resource the add-on provides
type AddonType string

const (
	AddonTypeEvents    AddonType = "events"
	AddonTypeRetention AddonType = "retention"
	AddonTypeSeats     AddonType = "seats"
	AddonTypeProjects  AddonType = "projects"
)

func (a AddonType) IsValid() bool {
	switch a {
	case AddonTypeEvents, AddonTypeRetention, AddonTypeSeats, AddonTypeProjects:
		return true
	}
	return false
}

// AddonStatus represents the status of an org's add-on subscription
type AddonStatus string

const (
	AddonStatusActive   AddonStatus = "active"
	AddonStatusCanceled AddonStatus = "canceled"
	AddonStatusExpired  AddonStatus = "expired"
	AddonStatusPaused   AddonStatus = "paused"
)

func (s AddonStatus) IsValid() bool {
	switch s {
	case AddonStatusActive, AddonStatusCanceled, AddonStatusExpired, AddonStatusPaused:
		return true
	}
	return false
}

// ============================================================================
// ADDON PLAN MODEL
// ============================================================================

// AddonPlan represents an available add-on that customers can purchase
// Add-ons sit on top of subscription plans to extend specific limits
type AddonPlan struct {
	ID uuid.UUID `db:"id" json:"id"`

	// Identification
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description,omitempty"`

	// What this add-on provides
	AddonType AddonType `db:"addon_type" json:"addon_type"`
	Quantity  int64     `db:"quantity" json:"quantity"`
	// events:    100000 (100k extra events/month)
	// retention: 7      (7 extra retention days)
	// seats:     5      (5 extra team seats)
	// projects:  10     (10 extra projects)

	// Pricing
	Price    float64 `db:"price" json:"price"`
	Currency string  `db:"currency" json:"currency"`

	// LemonSqueezy
	LemonSqueezyVariantID string `db:"lemon_squeezy_variant_id" json:"lemon_squeezy_variant_id"`

	// Which plans can purchase this add-on (Developer cannot)
	CompatiblePlans pq.StringArray `db:"compatible_plans" json:"compatible_plans"`

	// Display
	DisplayOrder int  `db:"display_order" json:"display_order"`
	IsActive     bool `db:"is_active" json:"is_active"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ADDON PLAN HELPERS
// ============================================================================

// IsCompatibleWith checks if this add-on can be purchased by a given plan tier
func (a *AddonPlan) IsCompatibleWith(tier PlanTier) bool {
	return slices.Contains(a.CompatiblePlans, string(tier))
}

// GetFormattedPrice returns human-readable price
func (a *AddonPlan) GetFormattedPrice() string {
	return fmt.Sprintf("$%.2f/month", a.Price)
}

// GetQuantityLabel returns human-readable quantity
func (a *AddonPlan) GetQuantityLabel() string {
	switch a.AddonType {
	case AddonTypeEvents:
		if a.Quantity >= 1000000 {
			return fmt.Sprintf("%dM events", a.Quantity/1000000)
		}
		return fmt.Sprintf("%dK events", a.Quantity/1000)
	case AddonTypeRetention:
		return fmt.Sprintf("+%d days", a.Quantity)
	case AddonTypeSeats:
		return fmt.Sprintf("+%d seats", a.Quantity)
	case AddonTypeProjects:
		return fmt.Sprintf("+%d projects", a.Quantity)
	}
	return fmt.Sprintf("+%d", a.Quantity)
}

// ============================================================================
// VALIDATION
// ============================================================================

func (a *AddonPlan) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !a.AddonType.IsValid() {
		return fmt.Errorf("invalid addon type")
	}
	if a.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if a.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if a.LemonSqueezyVariantID == "" {
		return fmt.Errorf("LemonSqueezy variant ID is required")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicAddonPlan struct {
	ID              uuid.UUID      `json:"id"`
	Name            string         `json:"name"`
	Description     *string        `json:"description,omitempty"`
	AddonType       AddonType      `json:"addon_type"`
	Quantity        int64          `json:"quantity"`
	QuantityLabel   string         `json:"quantity_label"`
	Price           float64        `json:"price"`
	Currency        string         `json:"currency"`
	CompatiblePlans pq.StringArray `json:"compatible_plans"`
	DisplayOrder    int            `json:"display_order"`
}

func (a *AddonPlan) ToPublic() *PublicAddonPlan {
	return &PublicAddonPlan{
		ID:              a.ID,
		Name:            a.Name,
		Description:     a.Description,
		AddonType:       a.AddonType,
		Quantity:        a.Quantity,
		QuantityLabel:   a.GetQuantityLabel(),
		Price:           a.Price,
		Currency:        a.Currency,
		CompatiblePlans: a.CompatiblePlans,
		DisplayOrder:    a.DisplayOrder,
	}
}

// ============================================================================
// ORGANIZATION ADDON MODEL
// ============================================================================

// OrganizationAddon represents an active add-on subscription for an organization
// When active, this add-on's quantity is added to the org's base plan limits
type OrganizationAddon struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	AddonPlanID    uuid.UUID `db:"addon_plan_id" json:"addon_plan_id"`

	// LemonSqueezy subscription for this add-on
	LemonSqueezySubscriptionID string  `db:"lemon_squeezy_subscription_id" json:"lemon_squeezy_subscription_id"`
	LemonSqueezyOrderID        *string `db:"lemon_squeezy_order_id" json:"lemon_squeezy_order_id,omitempty"`

	// Denormalized from addon_plans for quick limit calculations
	AddonType AddonType `db:"addon_type" json:"addon_type"`
	Quantity  int64     `db:"quantity" json:"quantity"`

	// Pricing snapshot at purchase time
	PriceAmount float64 `db:"price_amount" json:"price_amount"`
	Currency    string  `db:"currency" json:"currency"`

	// Status
	Status AddonStatus `db:"status" json:"status"`

	// Billing period
	CurrentPeriodStart time.Time `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time `db:"current_period_end" json:"current_period_end"`

	// Cancellation
	CancelAtPeriodEnd bool       `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	CanceledAt        *time.Time `db:"canceled_at" json:"canceled_at,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ORG ADDON HELPERS
// ============================================================================

func (o *OrganizationAddon) IsActive() bool {
	return o.Status == AddonStatusActive
}

func (o *OrganizationAddon) IsCanceled() bool {
	return o.Status == AddonStatusCanceled
}

func (o *OrganizationAddon) IsExpired() bool {
	return o.Status == AddonStatusExpired
}

func (o *OrganizationAddon) WillCancelAtPeriodEnd() bool {
	return o.CancelAtPeriodEnd
}

func (o *OrganizationAddon) IsInCurrentPeriod() bool {
	now := time.Now()
	return now.After(o.CurrentPeriodStart) && now.Before(o.CurrentPeriodEnd)
}

func (o *OrganizationAddon) DaysUntilRenewal() int {
	return int(time.Until(o.CurrentPeriodEnd).Hours() / 24)
}

// ============================================================================
// ACTIONS
// ============================================================================

// Cancel marks add-on to cancel at period end
func (o *OrganizationAddon) Cancel() {
	now := time.Now()
	o.CancelAtPeriodEnd = true
	o.CanceledAt = &now
	o.Status = AddonStatusCanceled
}

// Expire marks add-on as expired
func (o *OrganizationAddon) Expire() {
	o.Status = AddonStatusExpired
}

// Renew renews for next billing period
func (o *OrganizationAddon) Renew(nextPeriodEnd time.Time) {
	o.CurrentPeriodStart = o.CurrentPeriodEnd
	o.CurrentPeriodEnd = nextPeriodEnd
	o.Status = AddonStatusActive
	o.CancelAtPeriodEnd = false
}

// ============================================================================
// VALIDATION
// ============================================================================

func (o *OrganizationAddon) Validate() error {
	if o.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if o.AddonPlanID == uuid.Nil {
		return fmt.Errorf("addon plan ID is required")
	}
	if o.LemonSqueezySubscriptionID == "" {
		return fmt.Errorf("LemonSqueezy subscription ID is required")
	}
	if !o.AddonType.IsValid() {
		return fmt.Errorf("invalid addon type")
	}
	if o.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if !o.Status.IsValid() {
		return fmt.Errorf("invalid addon status")
	}
	if o.CurrentPeriodEnd.Before(o.CurrentPeriodStart) {
		return fmt.Errorf("period end must be after period start")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicOrganizationAddon struct {
	ID                         uuid.UUID   `json:"id"`
	OrganizationID             uuid.UUID   `json:"organization_id"`
	AddonPlanID                uuid.UUID   `json:"addon_plan_id"`
	AddonType                  AddonType   `json:"addon_type"`
	Quantity                   int64       `json:"quantity"`
	PriceAmount                float64     `json:"price_amount"`
	Currency                   string      `json:"currency"`
	Status                     AddonStatus `json:"status"`
	CurrentPeriodStart         time.Time   `json:"current_period_start"`
	CurrentPeriodEnd           time.Time   `json:"current_period_end"`
	CancelAtPeriodEnd          bool        `json:"cancel_at_period_end"`
	CanceledAt                 *time.Time  `json:"canceled_at,omitempty"`
	CreatedAt                  time.Time   `json:"created_at"`
}

func (o *OrganizationAddon) ToPublic() *PublicOrganizationAddon {
	return &PublicOrganizationAddon{
		ID:                 o.ID,
		OrganizationID:     o.OrganizationID,
		AddonPlanID:        o.AddonPlanID,
		AddonType:          o.AddonType,
		Quantity:           o.Quantity,
		PriceAmount:        o.PriceAmount,
		Currency:           o.Currency,
		Status:             o.Status,
		CurrentPeriodStart: o.CurrentPeriodStart,
		CurrentPeriodEnd:   o.CurrentPeriodEnd,
		CancelAtPeriodEnd:  o.CancelAtPeriodEnd,
		CanceledAt:         o.CanceledAt,
		CreatedAt:          o.CreatedAt,
	}
}