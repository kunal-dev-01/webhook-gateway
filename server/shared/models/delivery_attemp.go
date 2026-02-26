package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// DeliveryStatus represents the status of a delivery attempt
type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "pending"
	DeliveryInFlight   DeliveryStatus = "in_flight"
	DeliverySucceeded  DeliveryStatus = "succeeded"
	DeliveryFailed     DeliveryStatus = "failed"
	DeliveryDeadLetter DeliveryStatus = "dead_letter"
	DeliveryCanceled   DeliveryStatus = "canceled"
)

func (d DeliveryStatus) IsValid() bool {
	switch d {
	case DeliveryPending, DeliveryInFlight, DeliverySucceeded,
		DeliveryFailed, DeliveryDeadLetter, DeliveryCanceled:
		return true
	}
	return false
}

// IsTerminal returns true if status won't change anymore
func (d DeliveryStatus) IsTerminal() bool {
	switch d {
	case DeliverySucceeded, DeliveryDeadLetter, DeliveryCanceled:
		return true
	}
	return false
}

// ============================================================================
// DELIVERY ATTEMPT MODEL
// ============================================================================

// DeliveryAttempt represents an active webhook delivery in the retry queue
// NOTE: This is only the QUEUE. Full history is in ClickHouse delivery_logs.
// Rows are deleted/archived after terminal status is reached.
type DeliveryAttempt struct {
	ID             uuid.UUID `db:"id" json:"id"`
	EndpointID     uuid.UUID `db:"endpoint_id" json:"endpoint_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	ProjectID      uuid.UUID `db:"project_id" json:"project_id"`

	// Reference to ClickHouse webhook_events table
	CHEventID string `db:"ch_event_id" json:"ch_event_id"`

	// Attempt tracking
	AttemptNumber int `db:"attempt_number" json:"attempt_number"`
	MaxAttempts   int `db:"max_attempts" json:"max_attempts"`

	// Status
	Status DeliveryStatus `db:"status" json:"status"`

	// Scheduling
	ScheduledAt *time.Time `db:"scheduled_at" json:"scheduled_at"`
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`

	// Retry scheduling
	NextRetryAt       *time.Time `db:"next_retry_at" json:"next_retry_at,omitempty"`
	RetryDelaySeconds *int       `db:"retry_delay_seconds" json:"retry_delay_seconds,omitempty"`

	// Quick response summary (full details in ClickHouse)
	HTTPStatusCode *int    `db:"http_status_code" json:"http_status_code,omitempty"`
	ErrorType      *string `db:"error_type" json:"error_type,omitempty"`
	ErrorMessage   *string `db:"error_message" json:"error_message,omitempty"`
	DurationMS     *int    `db:"duration_ms" json:"duration_ms,omitempty"`

	// Worker lock (prevents duplicate processing)
	LockedBy      *string    `db:"locked_by" json:"locked_by,omitempty"`
	LockedAt      *time.Time `db:"locked_at" json:"locked_at,omitempty"`
	LockExpiresAt *time.Time `db:"lock_expires_at" json:"lock_expires_at,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

func (d *DeliveryAttempt) IsPending() bool    { return d.Status == DeliveryPending }
func (d *DeliveryAttempt) IsInFlight() bool   { return d.Status == DeliveryInFlight }
func (d *DeliveryAttempt) IsSucceeded() bool  { return d.Status == DeliverySucceeded }
func (d *DeliveryAttempt) IsFailed() bool     { return d.Status == DeliveryFailed }
func (d *DeliveryAttempt) IsDeadLetter() bool { return d.Status == DeliveryDeadLetter }
func (d *DeliveryAttempt) IsCanceled() bool   { return d.Status == DeliveryCanceled }
func (d *DeliveryAttempt) IsTerminal() bool   { return d.Status.IsTerminal() }

// HasMoreRetries checks if more retry attempts are allowed
func (d *DeliveryAttempt) HasMoreRetries() bool {
	return d.AttemptNumber < d.MaxAttempts
}

// IsLockExpired checks if the worker lock has expired
func (d *DeliveryAttempt) IsLockExpired() bool {
	if d.LockExpiresAt == nil {
		return false
	}
	return time.Now().After(*d.LockExpiresAt)
}

// IsReadyForRetry checks if this failed attempt is ready to be retried
func (d *DeliveryAttempt) IsReadyForRetry() bool {
	if !d.IsFailed() || d.NextRetryAt == nil {
		return false
	}
	return time.Now().After(*d.NextRetryAt)
}

// ============================================================================
// WORKER ACTIONS
// ============================================================================

// Lock claims this attempt for a specific worker
func (d *DeliveryAttempt) Lock(workerID string, ttl time.Duration) {
	now := time.Now()
	expires := now.Add(ttl)
	d.Status = DeliveryInFlight
	d.StartedAt = &now
	d.LockedBy = &workerID
	d.LockedAt = &now
	d.LockExpiresAt = &expires
}

// MarkSucceeded records a successful delivery
func (d *DeliveryAttempt) MarkSucceeded(statusCode int, durationMS int) {
	now := time.Now()
	d.Status = DeliverySucceeded
	d.CompletedAt = &now
	d.HTTPStatusCode = &statusCode
	d.DurationMS = &durationMS
	d.LockedBy = nil
	d.LockedAt = nil
	d.LockExpiresAt = nil
}

// MarkFailed records a failed attempt and schedules retry
func (d *DeliveryAttempt) MarkFailed(errType, errMsg string, statusCode *int, durationMS int, nextRetryAt *time.Time) {
	now := time.Now()
	d.Status = DeliveryFailed
	d.CompletedAt = &now
	d.ErrorType = &errType
	d.ErrorMessage = &errMsg
	d.HTTPStatusCode = statusCode
	d.DurationMS = &durationMS
	d.NextRetryAt = nextRetryAt
	d.LockedBy = nil
	d.LockedAt = nil
	d.LockExpiresAt = nil
}

// MoveToDeadLetter moves attempt to dead letter queue (no more retries)
func (d *DeliveryAttempt) MoveToDeadLetter() {
	d.Status = DeliveryDeadLetter
	d.NextRetryAt = nil
}

// PrepareForRetry increments attempt number and resets for next try
func (d *DeliveryAttempt) PrepareForRetry(delaySecs int) {
	d.AttemptNumber++
	d.Status = DeliveryPending
	d.StartedAt = nil
	d.CompletedAt = nil
	retryDelay := delaySecs
	d.RetryDelaySeconds = &retryDelay
}

// Cancel cancels this delivery attempt
func (d *DeliveryAttempt) Cancel() {
	d.Status = DeliveryCanceled
	d.NextRetryAt = nil
	d.LockedBy = nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (d *DeliveryAttempt) Validate() error {
	if d.EndpointID == uuid.Nil {
		return fmt.Errorf("endpoint ID is required")
	}
	if d.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if d.CHEventID == "" {
		return fmt.Errorf("ClickHouse event ID is required")
	}
	if d.AttemptNumber < 1 {
		return fmt.Errorf("attempt number must be >= 1")
	}
	if d.MaxAttempts < 1 || d.MaxAttempts > 10 {
		return fmt.Errorf("max attempts must be between 1 and 10")
	}
	if !d.Status.IsValid() {
		return fmt.Errorf("invalid delivery status")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicDeliveryAttempt struct {
	ID             uuid.UUID      `json:"id"`
	EndpointID     uuid.UUID      `json:"endpoint_id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	CHEventID      string         `json:"ch_event_id"`
	AttemptNumber  int            `json:"attempt_number"`
	MaxAttempts    int            `json:"max_attempts"`
	Status         DeliveryStatus `json:"status"`
	ScheduledAt    *time.Time     `json:"scheduled_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	NextRetryAt    *time.Time     `json:"next_retry_at,omitempty"`
	HTTPStatusCode *int           `json:"http_status_code,omitempty"`
	ErrorType      *string        `json:"error_type,omitempty"`
	ErrorMessage   *string        `json:"error_message,omitempty"`
	DurationMS     *int           `json:"duration_ms,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

func (d *DeliveryAttempt) ToPublic() *PublicDeliveryAttempt {
	return &PublicDeliveryAttempt{
		ID:             d.ID,
		EndpointID:     d.EndpointID,
		OrganizationID: d.OrganizationID,
		CHEventID:      d.CHEventID,
		AttemptNumber:  d.AttemptNumber,
		MaxAttempts:    d.MaxAttempts,
		Status:         d.Status,
		ScheduledAt:    d.ScheduledAt,
		CompletedAt:    d.CompletedAt,
		NextRetryAt:    d.NextRetryAt,
		HTTPStatusCode: d.HTTPStatusCode,
		ErrorType:      d.ErrorType,
		ErrorMessage:   d.ErrorMessage,
		DurationMS:     d.DurationMS,
		CreatedAt:      d.CreatedAt,
	}
}