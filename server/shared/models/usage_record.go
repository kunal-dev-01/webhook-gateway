package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// USAGE RECORD MODEL
// ============================================================================

// UsageRecord tracks monthly event usage for billing and plan limit enforcement
// Organization-level (project_id = nil) = billing
// Project-level (project_id set)        = analytics breakdown
type UsageRecord struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organization_id"`
	ProjectID      *uuid.UUID `db:"project_id" json:"project_id,omitempty"`

	// Billing period
	PeriodYear  int `db:"period_year" json:"period_year"`
	PeriodMonth int `db:"period_month" json:"period_month"`

	// Event counters
	EventsReceived        int64 `db:"events_received" json:"events_received"`
	EventsDelivered       int64 `db:"events_delivered" json:"events_delivered"`
	EventsFailed          int64 `db:"events_failed" json:"events_failed"`
	EventsDeadLetter      int64 `db:"events_dead_letter" json:"events_dead_letter"`
	DeliveryAttemptsTotal int64 `db:"delivery_attempts_total" json:"delivery_attempts_total"`

	// Bandwidth
	BytesReceived  int64 `db:"bytes_received" json:"bytes_received"`
	BytesDelivered int64 `db:"bytes_delivered" json:"bytes_delivered"`

	// Billing context
	PlanEventsLimit    int64      `db:"plan_events_limit" json:"plan_events_limit"`
	LimitExceededAt    *time.Time `db:"limit_exceeded_at" json:"limit_exceeded_at,omitempty"`
	LimitWarningSentAt *time.Time `db:"limit_warning_sent_at" json:"limit_warning_sent_at,omitempty"`

	// Finalization
	IsFinalized bool       `db:"is_finalized" json:"is_finalized"`
	FinalizedAt *time.Time `db:"finalized_at" json:"finalized_at,omitempty"`

	LastSyncedAt time.Time `db:"last_synced_at" json:"last_synced_at"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// HELPERS
// ============================================================================

// IsOrgLevel checks if this is an org-level aggregate record (for billing)
func (u *UsageRecord) IsOrgLevel() bool {
	return u.ProjectID == nil
}

// GetUsagePercentage returns usage as percentage of plan limit (0.0 - 100.0)
func (u *UsageRecord) GetUsagePercentage() float64 {
	if u.PlanEventsLimit == 0 {
		return 0
	}
	pct := (float64(u.EventsReceived) / float64(u.PlanEventsLimit)) * 100
	if pct > 100 {
		return 100
	}
	return pct
}

// IsWithinLimit checks if current usage is within plan limit
func (u *UsageRecord) IsWithinLimit() bool {
	return u.EventsReceived < u.PlanEventsLimit
}

// GetRemainingEvents returns remaining events for this period
func (u *UsageRecord) GetRemainingEvents() int64 {
	remaining := u.PlanEventsLimit - u.EventsReceived
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ShouldWarn checks if usage warning should be sent (>= 80%)
func (u *UsageRecord) ShouldWarn() bool {
	return u.GetUsagePercentage() >= 80 && u.LimitWarningSentAt == nil
}

// ShouldBlock checks if requests should be blocked (>= 100%)
func (u *UsageRecord) ShouldBlock() bool {
	return u.EventsReceived >= u.PlanEventsLimit
}

// ============================================================================
// COUNTER METHODS
// ============================================================================

// IncrementReceived adds to received events counter
func (u *UsageRecord) IncrementReceived(count int64, bytes int64) {
	u.EventsReceived += count
	u.BytesReceived += bytes
	u.LastSyncedAt = time.Now()

	// Record when limit was first exceeded
	if u.ShouldBlock() && u.LimitExceededAt == nil {
		now := time.Now()
		u.LimitExceededAt = &now
	}
}

// IncrementDelivered adds to delivered events counter
func (u *UsageRecord) IncrementDelivered(count int64, bytes int64) {
	u.EventsDelivered += count
	u.BytesDelivered += bytes
	u.DeliveryAttemptsTotal += count
	u.LastSyncedAt = time.Now()
}

// IncrementFailed adds to failed events counter
func (u *UsageRecord) IncrementFailed(count int64) {
	u.EventsFailed += count
	u.DeliveryAttemptsTotal += count
	u.LastSyncedAt = time.Now()
}

// IncrementDeadLetter adds to dead letter counter
func (u *UsageRecord) IncrementDeadLetter(count int64) {
	u.EventsDeadLetter += count
	u.LastSyncedAt = time.Now()
}

// MarkWarningSent records that the usage warning was sent
func (u *UsageRecord) MarkWarningSent() {
	now := time.Now()
	u.LimitWarningSentAt = &now
}

// Finalize marks this period as finalized (no more updates)
func (u *UsageRecord) Finalize() error {
	if u.IsFinalized {
		return fmt.Errorf("usage record is already finalized")
	}
	now := time.Now()
	u.IsFinalized = true
	u.FinalizedAt = &now
	return nil
}

// GetDeliverySuccessRate returns percentage of successful deliveries
func (u *UsageRecord) GetDeliverySuccessRate() float64 {
	total := u.EventsDelivered + u.EventsFailed
	if total == 0 {
		return 0
	}
	return (float64(u.EventsDelivered) / float64(total)) * 100
}

// ============================================================================
// VALIDATION
// ============================================================================

func (u *UsageRecord) Validate() error {
	if u.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if u.PeriodYear < 2024 || u.PeriodYear > 2100 {
		return fmt.Errorf("invalid period year")
	}
	if u.PeriodMonth < 1 || u.PeriodMonth > 12 {
		return fmt.Errorf("period month must be between 1 and 12")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicUsageRecord struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organization_id"`
	ProjectID             *uuid.UUID `json:"project_id,omitempty"`
	PeriodYear            int        `json:"period_year"`
	PeriodMonth           int        `json:"period_month"`
	EventsReceived        int64      `json:"events_received"`
	EventsDelivered       int64      `json:"events_delivered"`
	EventsFailed          int64      `json:"events_failed"`
	DeliveryAttemptsTotal int64      `json:"delivery_attempts_total"`
	BytesReceived         int64      `json:"bytes_received"`
	BytesDelivered        int64      `json:"bytes_delivered"`
	PlanEventsLimit       int64      `json:"plan_events_limit"`
	RemainingEvents       int64      `json:"remaining_events"`
	UsagePercentage       float64    `json:"usage_percentage"`
	DeliverySuccessRate   float64    `json:"delivery_success_rate"`
	IsFinalized           bool       `json:"is_finalized"`
}

func (u *UsageRecord) ToPublic() *PublicUsageRecord {
	return &PublicUsageRecord{
		ID:                    u.ID,
		OrganizationID:        u.OrganizationID,
		ProjectID:             u.ProjectID,
		PeriodYear:            u.PeriodYear,
		PeriodMonth:           u.PeriodMonth,
		EventsReceived:        u.EventsReceived,
		EventsDelivered:       u.EventsDelivered,
		EventsFailed:          u.EventsFailed,
		DeliveryAttemptsTotal: u.DeliveryAttemptsTotal,
		BytesReceived:         u.BytesReceived,
		BytesDelivered:        u.BytesDelivered,
		PlanEventsLimit:       u.PlanEventsLimit,
		RemainingEvents:       u.GetRemainingEvents(),
		UsagePercentage:       u.GetUsagePercentage(),
		DeliverySuccessRate:   u.GetDeliverySuccessRate(),
		IsFinalized:           u.IsFinalized,
	}
}