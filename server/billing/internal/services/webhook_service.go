package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"billing/internal/lemonsqueezy"
	"billing/internal/repository"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// WEBHOOK SERVICE
// ============================================================================

type WebhookService struct {
	orgRepo          *repository.OrgRepository
	subRepo          *repository.SubscriptionRepository
	paymentRepo      *repository.PaymentRepository
	planRepo         *repository.PlanRepository
	webhookEventRepo *repository.WebhookEventRepository
	lsClient         *lemonsqueezy.Client
	log              *logger.Logger
}

func NewWebhookService(
	orgRepo *repository.OrgRepository,
	subRepo *repository.SubscriptionRepository,
	paymentRepo *repository.PaymentRepository,
	planRepo *repository.PlanRepository,
	webhookEventRepo *repository.WebhookEventRepository,
	lsClient *lemonsqueezy.Client,
	log *logger.Logger,
) *WebhookService {
	return &WebhookService{
		orgRepo:          orgRepo,
		subRepo:          subRepo,
		paymentRepo:      paymentRepo,
		planRepo:         planRepo,
		webhookEventRepo: webhookEventRepo,
		lsClient:         lsClient,
		log:              log,
	}
}

// ============================================================================
// PROCESS WEBHOOK
// ============================================================================

// ProcessWebhook is the main entry point — verify → log → dispatch
func (s *WebhookService) ProcessWebhook(ctx context.Context, rawBody []byte, signature string) error {
	// 1. Verify HMAC signature
	if !s.lsClient.VerifyWebhookSignature(rawBody, signature) {
		s.log.Security("webhook_signature_invalid", "high", "signature", signature[:min(len(signature), 20)])
		return errors.Unauthorized("Invalid webhook signature")
	}

	// 2. Parse top-level payload
	var payload lemonsqueezy.WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		s.log.Error("Failed to parse webhook payload", err)
		return errors.BadRequest("Invalid webhook payload")
	}

	eventType := payload.Meta.EventName
	webhookID := payload.Meta.WebhookID

	s.log.Info("Webhook received", "event_type", eventType, "webhook_id", webhookID)

	// 3. Idempotency check — skip if already successfully processed
	if webhookID != "" {
		alreadyProcessed, err := s.webhookEventRepo.ExistsByLemonSqueezyEventID(ctx, webhookID)
		if err != nil {
			s.log.Error("Idempotency check failed", err, "webhook_id", webhookID)
		} else if alreadyProcessed {
			s.log.Info("Webhook already processed — skipping", "webhook_id", webhookID, "event_type", eventType)
			return nil
		}
	}

	// 4. Create audit record
	rawPayload := models.WebhookEventPayload{}
	_ = json.Unmarshal(rawBody, &rawPayload)

	event := &models.WebhookEvent{
		ID:                  uuid.New(),
		EventType:           eventType,
		Status:              models.WebhookPending,
		Payload:             rawPayload,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if webhookID != "" {
		event.LemonSqueezyEventID = &webhookID
	}

	if err := s.webhookEventRepo.Create(ctx, event); err != nil {
		s.log.Error("Failed to save webhook event", err, "event_type", eventType)
		// Don't fail — still try to process
	}

	// 5. Mark processing
	event.MarkAsProcessing()
	_ = s.webhookEventRepo.Update(ctx, event)

	// 6. Dispatch to handler
	var processErr error
	switch eventType {
	case models.EventSubscriptionCreated:
		processErr = s.handleSubscriptionCreated(ctx, event, rawBody)
	case models.EventSubscriptionUpdated:
		processErr = s.handleSubscriptionUpdated(ctx, event, rawBody)
	case models.EventSubscriptionCanceled:
		processErr = s.handleSubscriptionCanceled(ctx, event, rawBody)
	case models.EventSubscriptionResumed:
		processErr = s.handleSubscriptionResumed(ctx, event, rawBody)
	case models.EventSubscriptionExpired:
		processErr = s.handleSubscriptionExpired(ctx, event, rawBody)
	case models.EventSubscriptionPaused:
		processErr = s.handleSubscriptionPaused(ctx, event, rawBody)
	case models.EventSubscriptionUnpaused:
		processErr = s.handleSubscriptionResumed(ctx, event, rawBody)
	case models.EventSubscriptionPaymentSuccess:
		processErr = s.handlePaymentSuccess(ctx, event, rawBody)
	case models.EventSubscriptionPaymentFailed:
		processErr = s.handlePaymentFailed(ctx, event, rawBody)
	case models.EventOrderCreated:
		processErr = s.handleOrderCreated(ctx, event, rawBody)
	case models.EventOrderRefunded:
		processErr = s.handleOrderRefunded(ctx, event, rawBody)
	default:
		event.MarkAsSkipped(fmt.Sprintf("unhandled event type: %s", eventType))
		_ = s.webhookEventRepo.Update(ctx, event)
		s.log.Info("Webhook event skipped — no handler", "event_type", eventType)
		return nil
	}

	// 7. Update audit record with result
	if processErr != nil {
		event.MarkAsFailed(processErr.Error())
		_ = s.webhookEventRepo.Update(ctx, event)
		s.log.Error("Webhook processing failed", processErr, "event_type", eventType, "event_id", event.ID)
		return processErr
	}

	event.MarkAsProcessed()
	_ = s.webhookEventRepo.Update(ctx, event)
	s.log.Info("Webhook processed successfully", "event_type", eventType, "event_id", event.ID)

	return nil
}

// ============================================================================
// SUBSCRIPTION HANDLERS
// ============================================================================

func (s *WebhookService) handleSubscriptionCreated(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_created payload: %w", err)
	}

	attrs := payload.Data.Attributes
	lsSubID := payload.Data.ID
	customData := payload.Meta.CustomData

	// Extract org_id from custom data
	orgID, err := extractOrgID(customData)
	if err != nil {
		return fmt.Errorf("missing org_id in custom data: %w", err)
	}

	// Find plan by variant ID
	variantID := strconv.Itoa(attrs.VariantID)
	plan, err := s.planRepo.GetByVariantID(ctx, variantID)
	if err != nil {
		return fmt.Errorf("plan not found for variant %s: %w", variantID, err)
	}

	// Build subscription
	now := time.Now()
	lsOrderID := strconv.Itoa(attrs.OrderID)
	lsCustomerID := strconv.Itoa(attrs.CustomerID)

	sub := &models.Subscription{
		ID:                         uuid.New(),
		OrganizationID:             orgID,
		PlanID:                     plan.ID,
		LemonSqueezySubscriptionID: lsSubID,
		LemonSqueezyCustomerID:     lsCustomerID,
		LemonSqueezyOrderID:        &lsOrderID,
		Status:                     mapLSStatus(attrs.Status),
		CurrentPeriodStart:         now,
		CancelAtPeriodEnd:          attrs.Cancelled,
		PriceAmount:                plan.Price,
		Currency:                   "USD",
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}

	// Set period end from renews_at
	if attrs.RenewsAt != nil {
		sub.CurrentPeriodEnd = *attrs.RenewsAt
	} else {
		// Fallback: monthly = +30 days, annual = +365 days
		if plan.BillingPeriod == models.BillingAnnual {
			sub.CurrentPeriodEnd = now.AddDate(1, 0, 0)
		} else {
			sub.CurrentPeriodEnd = now.AddDate(0, 1, 0)
		}
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Sync org
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}

	org.PlanTier = plan.Tier
	org.BillingPeriod = plan.BillingPeriod
	org.SubscriptionStatus = models.SubscriptionActive
	org.LemonSqueezyCustomerID = &lsCustomerID
	org.LemonSqueezySubscriptionID = &lsSubID
	org.CurrentPeriodStart = &sub.CurrentPeriodStart
	org.CurrentPeriodEnd = &sub.CurrentPeriodEnd
	org.MonthlyEventsLimit = plan.MonthlyEventsLimit
	org.RetentionDays = plan.RetentionDays
	org.TeamSeatsLimit = plan.TeamSeatsLimit
	org.RateLimitPerSecond = plan.RateLimitPerSecond
	org.ProjectsLimit = plan.ProjectsLimit

	if err := s.orgRepo.SyncSubscription(ctx, org); err != nil {
		return fmt.Errorf("failed to sync org subscription: %w", err)
	}

	// Set event references
	event.OrganizationID = &orgID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription created",
		"org_id", orgID, "sub_id", sub.ID,
		"plan", plan.Tier, "billing", plan.BillingPeriod,
	)
	return nil
}

func (s *WebhookService) handleSubscriptionUpdated(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_updated payload: %w", err)
	}

	attrs := payload.Data.Attributes
	lsSubID := payload.Data.ID

	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found for ls_id %s: %w", lsSubID, err)
	}

	// Find plan by variant ID
	variantID := strconv.Itoa(attrs.VariantID)
	plan, err := s.planRepo.GetByVariantID(ctx, variantID)
	if err != nil {
		return fmt.Errorf("plan not found for variant %s: %w", variantID, err)
	}

	// Update subscription
	sub.PlanID = plan.ID
	sub.Status = mapLSStatus(attrs.Status)
	sub.CancelAtPeriodEnd = attrs.Cancelled
	sub.PriceAmount = plan.Price

	if attrs.RenewsAt != nil {
		sub.CurrentPeriodEnd = *attrs.RenewsAt
	}
	if attrs.EndsAt != nil && attrs.Cancelled {
		canceledAt := *attrs.EndsAt
		sub.CanceledAt = &canceledAt
	}

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Sync org
	org, err := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}

	org.PlanTier = plan.Tier
	org.BillingPeriod = plan.BillingPeriod
	org.SubscriptionStatus = sub.Status
	org.CurrentPeriodEnd = &sub.CurrentPeriodEnd
	org.MonthlyEventsLimit = plan.MonthlyEventsLimit
	org.RetentionDays = plan.RetentionDays
	org.TeamSeatsLimit = plan.TeamSeatsLimit
	org.RateLimitPerSecond = plan.RateLimitPerSecond
	org.ProjectsLimit = plan.ProjectsLimit

	if err := s.orgRepo.SyncSubscription(ctx, org); err != nil {
		return fmt.Errorf("failed to sync org: %w", err)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription updated", "org_id", sub.OrganizationID, "sub_id", sub.ID, "status", sub.Status)
	return nil
}

func (s *WebhookService) handleSubscriptionCanceled(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_canceled payload: %w", err)
	}

	lsSubID := payload.Data.ID
	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	now := time.Now()
	sub.CancelAtPeriodEnd = true
	sub.CanceledAt = &now

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update canceled subscription: %w", err)
	}

	// Sync org status
	org, _ := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if org != nil {
		org.SubscriptionStatus = models.SubscriptionCanceled
		_ = s.orgRepo.SyncSubscription(ctx, org)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription canceled", "org_id", sub.OrganizationID, "sub_id", sub.ID)
	return nil
}

func (s *WebhookService) handleSubscriptionResumed(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_resumed payload: %w", err)
	}

	lsSubID := payload.Data.ID
	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.Status = models.SubscriptionActive
	now := time.Now()
	sub.ResumedAt = &now

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update resumed subscription: %w", err)
	}

	org, _ := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if org != nil {
		org.SubscriptionStatus = models.SubscriptionActive
		_ = s.orgRepo.SyncSubscription(ctx, org)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription resumed", "org_id", sub.OrganizationID, "sub_id", sub.ID)
	return nil
}

func (s *WebhookService) handleSubscriptionExpired(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_expired payload: %w", err)
	}

	lsSubID := payload.Data.ID
	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	sub.Status = models.SubscriptionExpired
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to expire subscription: %w", err)
	}

	// Downgrade org to free plan
	org, err := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}

	freePlan, _ := s.planRepo.GetByTierAndBilling(ctx, models.PlanDeveloper, models.BillingMonthly)

	org.SubscriptionStatus = models.SubscriptionExpired
	org.PlanTier = models.PlanDeveloper
	org.BillingPeriod = models.BillingMonthly
	org.LemonSqueezySubscriptionID = nil

	if freePlan != nil {
		org.MonthlyEventsLimit = freePlan.MonthlyEventsLimit
		org.RetentionDays = freePlan.RetentionDays
		org.TeamSeatsLimit = freePlan.TeamSeatsLimit
		org.RateLimitPerSecond = freePlan.RateLimitPerSecond
		org.ProjectsLimit = freePlan.ProjectsLimit
	}

	if err := s.orgRepo.SyncSubscription(ctx, org); err != nil {
		return fmt.Errorf("failed to sync org after expiry: %w", err)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription expired — org downgraded to free", "org_id", sub.OrganizationID)
	return nil
}

func (s *WebhookService) handleSubscriptionPaused(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription_paused payload: %w", err)
	}

	lsSubID := payload.Data.ID
	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	sub.Pause()
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	org, _ := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if org != nil {
		org.SubscriptionStatus = models.SubscriptionPaused
		_ = s.orgRepo.SyncSubscription(ctx, org)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Subscription paused", "org_id", sub.OrganizationID, "sub_id", sub.ID)
	return nil
}

// ============================================================================
// PAYMENT HANDLERS
// ============================================================================

func (s *WebhookService) handlePaymentSuccess(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse payment_success payload: %w", err)
	}

	lsSubID := payload.Data.ID
	attrs := payload.Data.Attributes

	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Renew subscription period
	if attrs.RenewsAt != nil {
		sub.Renew(*attrs.RenewsAt)
		if err := s.subRepo.Update(ctx, sub); err != nil {
			return fmt.Errorf("failed to renew subscription: %w", err)
		}
	}

	// Update org period
	org, _ := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if org != nil {
		org.SubscriptionStatus = models.SubscriptionActive
		org.CurrentPeriodEnd = &sub.CurrentPeriodEnd
		_ = s.orgRepo.SyncSubscription(ctx, org)
	}

	// Create payment record
	orderID := strconv.Itoa(attrs.OrderID)
	alreadyExists, _ := s.paymentRepo.ExistsByOrderID(ctx, orderID)
	if !alreadyExists {
		now := time.Now()
		customerID := strconv.Itoa(attrs.CustomerID)
		plan, _ := s.planRepo.GetByID(ctx, sub.PlanID)
		amount := 0.0
		if plan != nil {
			amount = plan.Price
		}

		payment := &models.Payment{
			ID:                     uuid.New(),
			OrganizationID:         sub.OrganizationID,
			SubscriptionID:         &sub.ID,
			LemonSqueezyOrderID:    orderID,
			LemonSqueezyCustomerID: customerID,
			Amount:                 amount,
			Currency:               "USD",
			Status:                 models.PaymentPaid,
			PaidAt:                 &now,
			WebhookPayload:         models.WebhookPayload(event.Payload),
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		payment.MarkAsPaid()

		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			s.log.Error("Failed to create payment record", err, "org_id", sub.OrganizationID)
		} else {
			event.PaymentID = &payment.ID
		}
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Info("Payment success processed", "org_id", sub.OrganizationID, "sub_id", sub.ID)
	return nil
}

func (s *WebhookService) handlePaymentFailed(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.SubscriptionWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse payment_failed payload: %w", err)
	}

	lsSubID := payload.Data.ID
	sub, err := s.subRepo.GetByLemonSqueezyID(ctx, lsSubID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Mark subscription as past due
	sub.Status = models.SubscriptionPastDue
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription to past_due: %w", err)
	}

	org, _ := s.orgRepo.GetByID(ctx, sub.OrganizationID)
	if org != nil {
		org.SubscriptionStatus = models.SubscriptionPastDue
		_ = s.orgRepo.SyncSubscription(ctx, org)
	}

	event.OrganizationID = &sub.OrganizationID
	event.SubscriptionID = &sub.ID

	s.log.Warn("Payment failed — subscription marked past_due", "org_id", sub.OrganizationID, "sub_id", sub.ID)
	return nil
}

func (s *WebhookService) handleOrderCreated(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.OrderWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse order_created payload: %w", err)
	}

	attrs := payload.Data.Attributes
	orderID := payload.Data.ID

	// Idempotency
	alreadyExists, _ := s.paymentRepo.ExistsByOrderID(ctx, orderID)
	if alreadyExists {
		return nil
	}

	// Extract org from custom data
	orgID, err := extractOrgID(payload.Meta.CustomData)
	if err != nil {
		s.log.Warn("Order without org_id — skipping payment record", "order_id", orderID)
		return nil
	}

	amount := float64(attrs.Total) / 100.0 // LemonSqueezy returns cents
	customerID := strconv.Itoa(attrs.CustomerID)

	now := time.Now()
	payment := &models.Payment{
		ID:                     uuid.New(),
		OrganizationID:         orgID,
		LemonSqueezyOrderID:    orderID,
		LemonSqueezyCustomerID: customerID,
		Amount:                 amount,
		Currency:               attrs.Currency,
		Status:                 mapOrderStatus(attrs.Status),
		WebhookPayload:         models.WebhookPayload(event.Payload),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	if attrs.Status == "paid" {
		payment.MarkAsPaid()
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return fmt.Errorf("failed to create order payment: %w", err)
	}

	event.OrganizationID = &orgID
	event.PaymentID = &payment.ID

	s.log.Info("Order payment recorded", "org_id", orgID, "order_id", orderID, "amount", amount)
	return nil
}

func (s *WebhookService) handleOrderRefunded(ctx context.Context, event *models.WebhookEvent, rawBody []byte) error {
	var payload lemonsqueezy.OrderWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("failed to parse order_refunded payload: %w", err)
	}

	orderID := payload.Data.ID
	attrs := payload.Data.Attributes

	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("payment not found for order %s: %w", orderID, err)
	}

	refundAmount := payment.Amount // Full refund
	if err := payment.Refund(refundAmount, "order_refunded"); err != nil {
		return fmt.Errorf("failed to process refund: %w", err)
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update refunded payment: %w", err)
	}

	event.OrganizationID = &payment.OrganizationID
	event.PaymentID = &payment.ID

	s.log.Info("Order refunded", "org_id", payment.OrganizationID, "order_id", orderID, "amount", attrs.Total)
	return nil
}

// ============================================================================
// HELPERS
// ============================================================================

func extractOrgID(customData map[string]string) (uuid.UUID, error) {
	if customData == nil {
		return uuid.Nil, fmt.Errorf("no custom data")
	}
	orgIDStr, ok := customData["org_id"]
	if !ok || orgIDStr == "" {
		return uuid.Nil, fmt.Errorf("org_id missing from custom data")
	}
	return uuid.Parse(orgIDStr)
}

// mapLSStatus maps LemonSqueezy status strings to our SubscriptionStatus
func mapLSStatus(lsStatus string) models.SubscriptionStatus {
	switch lsStatus {
	case "active", "on_trial":
		return models.SubscriptionActive
	case "past_due", "unpaid":
		return models.SubscriptionPastDue
	case "cancelled":
		return models.SubscriptionCanceled
	case "paused":
		return models.SubscriptionPaused
	case "expired":
		return models.SubscriptionExpired
	default:
		return models.SubscriptionActive
	}
}

// mapOrderStatus maps LemonSqueezy order status to PaymentStatus
func mapOrderStatus(status string) models.PaymentStatus {
	switch status {
	case "paid":
		return models.PaymentPaid
	case "failed":
		return models.PaymentFailed
	case "refunded":
		return models.PaymentRefunded
	default:
		return models.PaymentPending
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}