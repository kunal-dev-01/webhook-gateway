package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"billing/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// PAYMENT REPOSITORY
// ============================================================================

type PaymentRepository struct {
	db *db.DB
}

func NewPaymentRepository(database *db.DB) *PaymentRepository {
	return &PaymentRepository{db: database}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (r *PaymentRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

const paymentCols = `
	id, organization_id, subscription_id,
	lemon_squeezy_order_id, lemon_squeezy_customer_id,
	amount, currency, status, payment_method,
	invoice_url, receipt_url,
	refunded_amount, refunded_at, refund_reason,
	paid_at, failed_at, webhook_payload,
	created_at, updated_at`

func (r *PaymentRepository) scanPayment(scanner interface{ Scan(...any) error }) (*models.Payment, error) {
	p := &models.Payment{}
	err := scanner.Scan(
		&p.ID, &p.OrganizationID, &p.SubscriptionID,
		&p.LemonSqueezyOrderID, &p.LemonSqueezyCustomerID,
		&p.Amount, &p.Currency, &p.Status, &p.PaymentMethod,
		&p.InvoiceURL, &p.ReceiptURL,
		&p.RefundedAmount, &p.RefundedAt, &p.RefundReason,
		&p.PaidAt, &p.FailedAt, &p.WebhookPayload,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO payments (
			id, organization_id, subscription_id,
			lemon_squeezy_order_id, lemon_squeezy_customer_id,
			amount, currency, status, payment_method,
			invoice_url, receipt_url,
			refunded_amount, refunded_at, refund_reason,
			paid_at, failed_at, webhook_payload,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	_, err := r.db.Exec(writeCtx, query,
		payment.ID, payment.OrganizationID, payment.SubscriptionID,
		payment.LemonSqueezyOrderID, payment.LemonSqueezyCustomerID,
		payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod,
		payment.InvoiceURL, payment.ReceiptURL,
		payment.RefundedAmount, payment.RefundedAt, payment.RefundReason,
		payment.PaidAt, payment.FailedAt, payment.WebhookPayload,
		payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create payment", err, "org_id", payment.OrganizationID)
		return errors.Wrap(err, 500, "Failed to create payment record")
	}

	logger.Debug("Payment created", "payment_id", payment.ID, "org_id", payment.OrganizationID)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	query := `SELECT ` + paymentCols + ` FROM payments WHERE id = $1`

	p, err := r.scanPayment(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Payment not found")
	}
	if err != nil {
		logger.Error("Failed to get payment", err, "payment_id", id)
		return nil, errors.Internal(err)
	}
	return p, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	query := `SELECT ` + paymentCols + ` FROM payments WHERE lemon_squeezy_order_id = $1`

	p, err := r.scanPayment(r.db.QueryRow(ctx, query, orderID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Payment not found")
	}
	if err != nil {
		logger.Error("Failed to get payment by order ID", err, "order_id", orderID)
		return nil, errors.Internal(err)
	}
	return p, nil
}

func (r *PaymentRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Payment, int, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM payments WHERE organization_id = $1`,
		orgID,
	).Scan(&total); err != nil {
		logger.Error("Failed to count payments", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}

	query := `
		SELECT ` + paymentCols + `
		FROM payments
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		logger.Error("Failed to list payments", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		p, err := r.scanPayment(rows)
		if err != nil {
			logger.Error("Failed to scan payment", err)
			return nil, 0, errors.Internal(err)
		}
		payments = append(payments, p)
	}

	return payments, total, rows.Err()
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *PaymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	payment.UpdatedAt = time.Now()

	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE payments SET
			status = $2, refunded_amount = $3, refunded_at = $4,
			refund_reason = $5, paid_at = $6, failed_at = $7,
			updated_at = $8
		WHERE id = $1`

	_, err := r.db.Exec(writeCtx, query,
		payment.ID, payment.Status, payment.RefundedAmount, payment.RefundedAt,
		payment.RefundReason, payment.PaidAt, payment.FailedAt, payment.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to update payment", err, "payment_id", payment.ID)
		return errors.Wrap(err, 500, "Failed to update payment")
	}

	return nil
}

// ExistsByOrderID — fast duplicate check before creating payment
func (r *PaymentRepository) ExistsByOrderID(ctx context.Context, orderID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM payments WHERE lemon_squeezy_order_id = $1)`,
		orderID,
	).Scan(&exists); err != nil {
		return false, errors.Internal(err)
	}
	return exists, nil
}