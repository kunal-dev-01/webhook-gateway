package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// WebhookEventStatus represents webhook processing status
type WebhookEventStatus string

const (
	WebhookPending    WebhookEventStatus = "pending"
	WebhookProcessing WebhookEventStatus = "processing"
	WebhookProcessed  WebhookEventStatus = "processed"
	WebhookFailed     WebhookEventStatus = "failed"
	WebhookSkipped    WebhookEventStatus = "skipped"
)

// ============================================================================
// WEBHOOK EVENT PAYLOAD
// ============================================================================

// WebhookEventPayload represents raw webhook payload from LemonSqueezy
type WebhookEventPayload map[string]interface{}

// Value implements driver.Valuer for WebhookEventPayload
func (p WebhookEventPayload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for WebhookEventPayload
func (p *WebhookEventPayload) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal WebhookEventPayload value")
	}
	return json.Unmarshal(bytes, p)
}

// ============================================================================
// WEBHOOK EVENT MODEL
// ============================================================================

// WebhookEvent represents a LemonSqueezy webhook event for audit trail
type WebhookEvent struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// LemonSqueezy webhook data
	LemonSqueezyEventID *string `db:"lemon_squeezy_event_id" json:"lemon_squeezy_event_id,omitempty"`
	EventType           string  `db:"event_type" json:"event_type"` // 'subscription_created', 'order_created', etc.

	// Processing status
	Status WebhookEventStatus `db:"status" json:"status"`

	// Related entities (nullable - not all events have these)
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	SubscriptionID *uuid.UUID `db:"subscription_id" json:"subscription_id,omitempty"`
	PaymentID      *uuid.UUID `db:"payment_id" json:"payment_id,omitempty"`

	// Raw webhook payload (for debugging and replay)
	Payload WebhookEventPayload `db:"payload" json:"payload"`

	// Processing details
	ProcessedAt  *time.Time `db:"processed_at" json:"processed_at,omitempty"`
	FailedAt     *time.Time `db:"failed_at" json:"failed_at,omitempty"`
	ErrorMessage *string    `db:"error_message" json:"error_message,omitempty"`
	RetryCount   int        `db:"retry_count" json:"retry_count"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

// IsPending checks if event is pending processing
func (w *WebhookEvent) IsPending() bool {
	return w.Status == WebhookPending
}

// IsProcessing checks if event is currently being processed
func (w *WebhookEvent) IsProcessing() bool {
	return w.Status == WebhookProcessing
}

// IsProcessed checks if event was successfully processed
func (w *WebhookEvent) IsProcessed() bool {
	return w.Status == WebhookProcessed
}

// IsFailed checks if event processing failed
func (w *WebhookEvent) IsFailed() bool {
	return w.Status == WebhookFailed
}

// IsSkipped checks if event was skipped
func (w *WebhookEvent) IsSkipped() bool {
	return w.Status == WebhookSkipped
}

// CanRetry checks if event can be retried
func (w *WebhookEvent) CanRetry() bool {
	return w.IsFailed() && w.RetryCount < 5 // Max 5 retries
}

// ============================================================================
// ACTIONS
// ============================================================================

// MarkAsProcessing marks event as currently being processed
func (w *WebhookEvent) MarkAsProcessing() {
	w.Status = WebhookProcessing
}

// MarkAsProcessed marks event as successfully processed
func (w *WebhookEvent) MarkAsProcessed() {
	now := time.Now()
	w.Status = WebhookProcessed
	w.ProcessedAt = &now
	w.ErrorMessage = nil
}

// MarkAsFailed marks event as failed with error message
func (w *WebhookEvent) MarkAsFailed(errorMsg string) {
	now := time.Now()
	w.Status = WebhookFailed
	w.FailedAt = &now
	w.ErrorMessage = &errorMsg
	w.RetryCount++
}

// MarkAsSkipped marks event as skipped
func (w *WebhookEvent) MarkAsSkipped(reason string) {
	w.Status = WebhookSkipped
	w.ErrorMessage = &reason
}

// ============================================================================
// WEBHOOK EVENT TYPES
// ============================================================================

// Common LemonSqueezy webhook event types
const (
	EventSubscriptionCreated     = "subscription_created"
	EventSubscriptionUpdated     = "subscription_updated"
	EventSubscriptionCanceled    = "subscription_canceled"
	EventSubscriptionResumed     = "subscription_resumed"
	EventSubscriptionExpired     = "subscription_expired"
	EventSubscriptionPaused      = "subscription_paused"
	EventSubscriptionUnpaused    = "subscription_unpaused"
	EventSubscriptionPaymentSuccess = "subscription_payment_success"
	EventSubscriptionPaymentFailed  = "subscription_payment_failed"
	EventOrderCreated            = "order_created"
	EventOrderRefunded           = "order_refunded"
	EventLicenseKeyCreated       = "license_key_created"
)

// IsSubscriptionEvent checks if event is subscription-related
func (w *WebhookEvent) IsSubscriptionEvent() bool {
	subscriptionEvents := []string{
		EventSubscriptionCreated,
		EventSubscriptionUpdated,
		EventSubscriptionCanceled,
		EventSubscriptionResumed,
		EventSubscriptionExpired,
		EventSubscriptionPaused,
		EventSubscriptionUnpaused,
		EventSubscriptionPaymentSuccess,
		EventSubscriptionPaymentFailed,
	}
	for _, evt := range subscriptionEvents {
		if w.EventType == evt {
			return true
		}
	}
	return false
}

// IsPaymentEvent checks if event is payment-related
func (w *WebhookEvent) IsPaymentEvent() bool {
	paymentEvents := []string{
		EventOrderCreated,
		EventOrderRefunded,
		EventSubscriptionPaymentSuccess,
		EventSubscriptionPaymentFailed,
	}
	for _, evt := range paymentEvents {
		if w.EventType == evt {
			return true
		}
	}
	return false
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the webhook event struct
func (w *WebhookEvent) Validate() error {
	if w.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if w.Payload == nil {
		return fmt.Errorf("webhook payload is required")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicWebhookEvent represents safe webhook event data for API responses (admin only)
type PublicWebhookEvent struct {
	ID                  uuid.UUID          `json:"id"`
	LemonSqueezyEventID *string            `json:"lemon_squeezy_event_id,omitempty"`
	EventType           string             `json:"event_type"`
	Status              WebhookEventStatus `json:"status"`
	OrganizationID      *uuid.UUID         `json:"organization_id,omitempty"`
	SubscriptionID      *uuid.UUID         `json:"subscription_id,omitempty"`
	PaymentID           *uuid.UUID         `json:"payment_id,omitempty"`
	ProcessedAt         *time.Time         `json:"processed_at,omitempty"`
	FailedAt            *time.Time         `json:"failed_at,omitempty"`
	ErrorMessage        *string            `json:"error_message,omitempty"`
	RetryCount          int                `json:"retry_count"`
	CreatedAt           time.Time          `json:"created_at"`
}

// ToPublic converts WebhookEvent to PublicWebhookEvent
func (w *WebhookEvent) ToPublic() *PublicWebhookEvent {
	return &PublicWebhookEvent{
		ID:                  w.ID,
		LemonSqueezyEventID: w.LemonSqueezyEventID,
		EventType:           w.EventType,
		Status:              w.Status,
		OrganizationID:      w.OrganizationID,
		SubscriptionID:      w.SubscriptionID,
		PaymentID:           w.PaymentID,
		ProcessedAt:         w.ProcessedAt,
		FailedAt:            w.FailedAt,
		ErrorMessage:        w.ErrorMessage,
		RetryCount:          w.RetryCount,
		CreatedAt:           w.CreatedAt,
	}
}