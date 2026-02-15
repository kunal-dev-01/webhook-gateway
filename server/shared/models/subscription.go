package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// SUBSCRIPTION METADATA
// ============================================================================

// SubscriptionMetadata represents subscription metadata from LemonSqueezy
type SubscriptionMetadata struct {
	LemonSqueezyData map[string]interface{} `json:"lemon_squeezy_data,omitempty"`
	CustomFields     map[string]string      `json:"custom_fields,omitempty"`
}

// Value implements driver.Valuer for SubscriptionMetadata
func (m SubscriptionMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for SubscriptionMetadata
func (m *SubscriptionMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal SubscriptionMetadata value")
	}
	return json.Unmarshal(bytes, m)
}

// ============================================================================
// SUBSCRIPTION MODEL
// ============================================================================

// Subscription represents an active organization subscription
type Subscription struct {
	// Primary identification
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	PlanID         uuid.UUID `db:"plan_id" json:"plan_id"`

	// LemonSqueezy subscription data
	LemonSqueezySubscriptionID string  `db:"lemon_squeezy_subscription_id" json:"lemon_squeezy_subscription_id"`
	LemonSqueezyCustomerID     string  `db:"lemon_squeezy_customer_id" json:"lemon_squeezy_customer_id"`
	LemonSqueezyOrderID        *string `db:"lemon_squeezy_order_id" json:"lemon_squeezy_order_id,omitempty"`

	// Subscription status
	Status SubscriptionStatus `db:"status" json:"status"`

	// Billing cycle
	CurrentPeriodStart time.Time `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time `db:"current_period_end" json:"current_period_end"`

	// Cancellation
	CancelAtPeriodEnd bool       `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	CanceledAt        *time.Time `db:"canceled_at" json:"canceled_at,omitempty"`

	// Pause/Resume
	PausedAt  *time.Time `db:"paused_at" json:"paused_at,omitempty"`
	ResumedAt *time.Time `db:"resumed_at" json:"resumed_at,omitempty"`

	// Pricing snapshot
	PriceAmount float64 `db:"price_amount" json:"price_amount"`
	Currency    string  `db:"currency" json:"currency"`

	// Metadata
	Metadata SubscriptionMetadata `db:"metadata" json:"metadata"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

// IsActive checks if subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionActive
}

// IsPastDue checks if subscription is past due
func (s *Subscription) IsPastDue() bool {
	return s.Status == SubscriptionPastDue
}

// IsCanceled checks if subscription is canceled
func (s *Subscription) IsCanceled() bool {
	return s.Status == SubscriptionCanceled
}

// IsPaused checks if subscription is paused
func (s *Subscription) IsPaused() bool {
	return s.Status == SubscriptionPaused
}

// IsExpired checks if subscription is expired
func (s *Subscription) IsExpired() bool {
	return s.Status == SubscriptionExpired
}

// WillCancelAtPeriodEnd checks if subscription will cancel at end of period
func (s *Subscription) WillCancelAtPeriodEnd() bool {
	return s.CancelAtPeriodEnd
}

// IsInCurrentPeriod checks if we're in the current billing period
func (s *Subscription) IsInCurrentPeriod() bool {
	now := time.Now()
	return now.After(s.CurrentPeriodStart) && now.Before(s.CurrentPeriodEnd)
}

// DaysUntilRenewal returns days until next renewal
func (s *Subscription) DaysUntilRenewal() int {
	duration := time.Until(s.CurrentPeriodEnd)
	return int(duration.Hours() / 24)
}

// ============================================================================
// ACTIONS
// ============================================================================

// MarkForCancellation marks subscription to cancel at period end
func (s *Subscription) MarkForCancellation() {
	s.CancelAtPeriodEnd = true
}

// UndoCancellation removes cancellation flag
func (s *Subscription) UndoCancellation() {
	s.CancelAtPeriodEnd = false
}

// Cancel immediately cancels the subscription
func (s *Subscription) Cancel() {
	now := time.Now()
	s.Status = SubscriptionCanceled
	s.CanceledAt = &now
	s.CancelAtPeriodEnd = false
}

// Pause pauses the subscription
func (s *Subscription) Pause() {
	now := time.Now()
	s.Status = SubscriptionPaused
	s.PausedAt = &now
}

// Resume resumes a paused subscription
func (s *Subscription) Resume() error {
	if !s.IsPaused() {
		return fmt.Errorf("subscription is not paused")
	}
	now := time.Now()
	s.Status = SubscriptionActive
	s.ResumedAt = &now
	return nil
}

// Renew renews the subscription for next period
func (s *Subscription) Renew(nextPeriodEnd time.Time) {
	s.CurrentPeriodStart = s.CurrentPeriodEnd
	s.CurrentPeriodEnd = nextPeriodEnd
	s.Status = SubscriptionActive
	s.CancelAtPeriodEnd = false
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the subscription struct
func (s *Subscription) Validate() error {
	if s.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if s.PlanID == uuid.Nil {
		return fmt.Errorf("plan ID is required")
	}
	if s.LemonSqueezySubscriptionID == "" {
		return fmt.Errorf("LemonSqueezy subscription ID is required")
	}
	if s.LemonSqueezyCustomerID == "" {
		return fmt.Errorf("LemonSqueezy customer ID is required")
	}
	if s.CurrentPeriodEnd.Before(s.CurrentPeriodStart) {
		return fmt.Errorf("current period end must be after start")
	}
	if s.PriceAmount < 0 {
		return fmt.Errorf("price amount cannot be negative")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicSubscription represents safe subscription data for API responses
type PublicSubscription struct {
	ID                 uuid.UUID          `json:"id"`
	OrganizationID     uuid.UUID          `json:"organization_id"`
	PlanID             uuid.UUID          `json:"plan_id"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd  bool               `json:"cancel_at_period_end"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	PausedAt           *time.Time         `json:"paused_at,omitempty"`
	PriceAmount        float64            `json:"price_amount"`
	Currency           string             `json:"currency"`
	CreatedAt          time.Time          `json:"created_at"`
}

// ToPublic converts Subscription to PublicSubscription
func (s *Subscription) ToPublic() *PublicSubscription {
	return &PublicSubscription{
		ID:                 s.ID,
		OrganizationID:     s.OrganizationID,
		PlanID:             s.PlanID,
		Status:             s.Status,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CanceledAt:         s.CanceledAt,
		PausedAt:           s.PausedAt,
		PriceAmount:        s.PriceAmount,
		Currency:           s.Currency,
		CreatedAt:          s.CreatedAt,
	}
}