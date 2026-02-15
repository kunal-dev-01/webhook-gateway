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

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentPending           PaymentStatus = "pending"
	PaymentPaid              PaymentStatus = "paid"
	PaymentFailed            PaymentStatus = "failed"
	PaymentRefunded          PaymentStatus = "refunded"
	PaymentPartiallyRefunded PaymentStatus = "partially_refunded"
)

// ============================================================================
// WEBHOOK PAYLOAD
// ============================================================================

// WebhookPayload represents raw webhook data from LemonSqueezy
type WebhookPayload map[string]interface{}

// Value implements driver.Valuer for WebhookPayload
func (p WebhookPayload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for WebhookPayload
func (p *WebhookPayload) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal WebhookPayload value")
	}
	return json.Unmarshal(bytes, p)
}

// ============================================================================
// PAYMENT MODEL
// ============================================================================

// Payment represents a payment transaction
type Payment struct {
	// Primary identification
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organization_id"`
	SubscriptionID *uuid.UUID `db:"subscription_id" json:"subscription_id,omitempty"`

	// LemonSqueezy payment data
	LemonSqueezyOrderID    string `db:"lemon_squeezy_order_id" json:"lemon_squeezy_order_id"`
	LemonSqueezyCustomerID string `db:"lemon_squeezy_customer_id" json:"lemon_squeezy_customer_id"`

	// Payment details
	Amount   float64       `db:"amount" json:"amount"`
	Currency string        `db:"currency" json:"currency"`
	Status   PaymentStatus `db:"status" json:"status"`

	// Payment method
	PaymentMethod *string `db:"payment_method" json:"payment_method,omitempty"` // 'card', 'paypal', etc.

	// Invoice/Receipt
	InvoiceURL *string `db:"invoice_url" json:"invoice_url,omitempty"`
	ReceiptURL *string `db:"receipt_url" json:"receipt_url,omitempty"`

	// Refund tracking
	RefundedAmount float64    `db:"refunded_amount" json:"refunded_amount"`
	RefundedAt     *time.Time `db:"refunded_at" json:"refunded_at,omitempty"`
	RefundReason   *string    `db:"refund_reason" json:"refund_reason,omitempty"`

	// Timestamps
	PaidAt   *time.Time `db:"paid_at" json:"paid_at,omitempty"`
	FailedAt *time.Time `db:"failed_at" json:"failed_at,omitempty"`

	// Webhook data (for debugging)
	WebhookPayload WebhookPayload `db:"webhook_payload" json:"webhook_payload"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

// IsPaid checks if payment is successful
func (p *Payment) IsPaid() bool {
	return p.Status == PaymentPaid
}

// IsFailed checks if payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentFailed
}

// IsRefunded checks if payment is fully refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentRefunded
}

// IsPartiallyRefunded checks if payment is partially refunded
func (p *Payment) IsPartiallyRefunded() bool {
	return p.Status == PaymentPartiallyRefunded
}

// GetNetAmount returns amount after refunds
func (p *Payment) GetNetAmount() float64 {
	return p.Amount - p.RefundedAmount
}

// ============================================================================
// ACTIONS
// ============================================================================

// MarkAsPaid marks payment as paid
func (p *Payment) MarkAsPaid() {
	now := time.Now()
	p.Status = PaymentPaid
	p.PaidAt = &now
}

// MarkAsFailed marks payment as failed
func (p *Payment) MarkAsFailed() {
	now := time.Now()
	p.Status = PaymentFailed
	p.FailedAt = &now
}

// Refund processes a refund
func (p *Payment) Refund(amount float64, reason string) error {
	if !p.IsPaid() {
		return fmt.Errorf("only paid payments can be refunded")
	}
	if amount <= 0 {
		return fmt.Errorf("refund amount must be positive")
	}
	if p.RefundedAmount+amount > p.Amount {
		return fmt.Errorf("refund amount exceeds payment amount")
	}

	now := time.Now()
	p.RefundedAmount += amount
	p.RefundedAt = &now
	p.RefundReason = &reason

	if p.RefundedAmount >= p.Amount {
		p.Status = PaymentRefunded
	} else {
		p.Status = PaymentPartiallyRefunded
	}

	return nil
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the payment struct
func (p *Payment) Validate() error {
	if p.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if p.LemonSqueezyOrderID == "" {
		return fmt.Errorf("LemonSqueezy order ID is required")
	}
	if p.LemonSqueezyCustomerID == "" {
		return fmt.Errorf("LemonSqueezy customer ID is required")
	}
	if p.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	if p.RefundedAmount > p.Amount {
		return fmt.Errorf("refunded amount cannot exceed payment amount")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicPayment represents safe payment data for API responses
type PublicPayment struct {
	ID             uuid.UUID     `json:"id"`
	OrganizationID uuid.UUID     `json:"organization_id"`
	SubscriptionID *uuid.UUID    `json:"subscription_id,omitempty"`
	Amount         float64       `json:"amount"`
	Currency       string        `json:"currency"`
	Status         PaymentStatus `json:"status"`
	PaymentMethod  *string       `json:"payment_method,omitempty"`
	InvoiceURL     *string       `json:"invoice_url,omitempty"`
	ReceiptURL     *string       `json:"receipt_url,omitempty"`
	RefundedAmount float64       `json:"refunded_amount"`
	RefundedAt     *time.Time    `json:"refunded_at,omitempty"`
	PaidAt         *time.Time    `json:"paid_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

// ToPublic converts Payment to PublicPayment
func (p *Payment) ToPublic() *PublicPayment {
	return &PublicPayment{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		SubscriptionID: p.SubscriptionID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		Status:         p.Status,
		PaymentMethod:  p.PaymentMethod,
		InvoiceURL:     p.InvoiceURL,
		ReceiptURL:     p.ReceiptURL,
		RefundedAmount: p.RefundedAmount,
		RefundedAt:     p.RefundedAt,
		PaidAt:         p.PaidAt,
		CreatedAt:      p.CreatedAt,
	}
}