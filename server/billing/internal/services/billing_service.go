package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"billing/internal/dto"
	"billing/internal/lemonsqueezy"
	"billing/internal/repository"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// BILLING SERVICE
// ============================================================================

type BillingService struct {
	orgRepo     *repository.OrgRepository
	subRepo     *repository.SubscriptionRepository
	paymentRepo *repository.PaymentRepository
	planRepo    *repository.PlanRepository
	lsClient    *lemonsqueezy.Client
	log         *logger.Logger
}

func NewBillingService(
	orgRepo *repository.OrgRepository,
	subRepo *repository.SubscriptionRepository,
	paymentRepo *repository.PaymentRepository,
	planRepo *repository.PlanRepository,
	lsClient *lemonsqueezy.Client,
	log *logger.Logger,
) *BillingService {
	return &BillingService{
		orgRepo:     orgRepo,
		subRepo:     subRepo,
		paymentRepo: paymentRepo,
		planRepo:    planRepo,
		lsClient:    lsClient,
		log:         log,
	}
}

// ============================================================================
// CREATE CHECKOUT
// ============================================================================

// CreateCheckout creates a LemonSqueezy checkout session for a plan upgrade
func (s *BillingService) CreateCheckout(
	ctx context.Context,
	orgID, userID uuid.UUID,
	userEmail, userName string,
	req *dto.CreateCheckoutRequest,
) (*dto.CreateCheckoutResponse, error) {
	// Validate plan exists
	plan, err := s.planRepo.GetByTierAndBilling(ctx, req.PlanTier, req.BillingPeriod)
	if err != nil {
		return nil, err
	}

	// Free plan has no checkout
	if plan.IsFreePlan() {
		return nil, errors.BadRequest("Developer plan is free — no checkout required")
	}

	// Parse variant ID to int
	variantID, err := strconv.Atoi(plan.LemonSqueezyVariantID)
	if err != nil {
		s.log.Error("Invalid variant ID in plan", err, "plan_id", plan.ID, "variant_id", plan.LemonSqueezyVariantID)
		return nil, errors.Internal(fmt.Errorf("invalid variant ID configuration"))
	}

	// Checkout expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Build checkout request
	checkoutReq := &lemonsqueezy.CreateCheckoutRequest{}
	checkoutReq.Data.Type = "checkouts"
	checkoutReq.Data.Attributes.StoreID = s.lsClient.GetStoreIDInt()
	checkoutReq.Data.Attributes.VariantID = variantID
	checkoutReq.Data.Attributes.ExpiresAt = &expiresAt

	// Pre-fill customer data
	checkoutReq.Data.Attributes.CheckoutData.Email = userEmail
	checkoutReq.Data.Attributes.CheckoutData.Name = userName
	checkoutReq.Data.Attributes.CheckoutData.Custom = map[string]string{
		"org_id":  orgID.String(),
		"user_id": userID.String(),
	}

	// Options
	checkoutReq.Data.Attributes.CheckoutOptions.SubscriptionPreview = true

	// Create checkout via LemonSqueezy
	checkout, err := s.lsClient.CreateCheckout(ctx, checkoutReq)
	if err != nil {
		s.log.Error("Failed to create LemonSqueezy checkout", err, "org_id", orgID, "plan_tier", req.PlanTier)
		return nil, errors.Internal(err)
	}

	s.log.Info("Checkout created", "org_id", orgID, "plan_tier", req.PlanTier, "billing", req.BillingPeriod)

	return &dto.CreateCheckoutResponse{
		CheckoutURL: checkout.Attributes.URL,
		ExpiresAt:   checkout.Attributes.ExpiresAt,
	}, nil
}

// ============================================================================
// GET SUBSCRIPTION
// ============================================================================

func (s *BillingService) GetSubscription(ctx context.Context, orgID uuid.UUID) (*dto.SubscriptionResponse, error) {
	// Get org to verify it exists
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Get subscription
	sub, err := s.subRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		// No subscription = free plan, return org plan info
		if isNotFound(err) {
			plan, planErr := s.planRepo.GetByTierAndBilling(ctx, org.PlanTier, org.BillingPeriod)
			if planErr != nil {
				return nil, planErr
			}
			return &dto.SubscriptionResponse{
				Plan: plan.ToPublic(),
			}, nil
		}
		return nil, err
	}

	// Get plan details
	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	// Optionally fetch customer portal URL from LemonSqueezy
	var portalURL *string
	if sub.LemonSqueezySubscriptionID != "" {
		lsSub, lsErr := s.lsClient.GetSubscription(ctx, sub.LemonSqueezySubscriptionID)
		if lsErr == nil {
			portalURL = &lsSub.Attributes.Urls.CustomerPortal
		}
	}

	return &dto.SubscriptionResponse{
		Subscription:      sub.ToPublic(),
		Plan:              plan.ToPublic(),
		CustomerPortalURL: portalURL,
	}, nil
}

// ============================================================================
// CANCEL SUBSCRIPTION
// ============================================================================

func (s *BillingService) CancelSubscription(ctx context.Context, orgID uuid.UUID) (*dto.SubscriptionResponse, error) {
	sub, err := s.subRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !sub.IsActive() {
		return nil, errors.BadRequest("Only active subscriptions can be cancelled")
	}

	if sub.WillCancelAtPeriodEnd() {
		return nil, errors.BadRequest("Subscription is already scheduled for cancellation")
	}

	// Cancel via LemonSqueezy API (marks cancel_at_period_end)
	_, err = s.lsClient.CancelSubscription(ctx, sub.LemonSqueezySubscriptionID)
	if err != nil {
		s.log.Error("LemonSqueezy cancel failed", err, "org_id", orgID, "ls_sub_id", sub.LemonSqueezySubscriptionID)
		return nil, errors.Internal(err)
	}

	// Mark locally — webhook will also confirm
	sub.MarkForCancellation()
	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.log.Error("Failed to update subscription after cancel", err, "sub_id", sub.ID)
		return nil, err
	}

	s.log.Info("Subscription cancelled", "org_id", orgID, "sub_id", sub.ID, "ends_at", sub.CurrentPeriodEnd)

	plan, _ := s.planRepo.GetByID(ctx, sub.PlanID)
	return &dto.SubscriptionResponse{
		Subscription: sub.ToPublic(),
		Plan:         plan.ToPublic(),
	}, nil
}

// ============================================================================
// RESUME SUBSCRIPTION
// ============================================================================

func (s *BillingService) ResumeSubscription(ctx context.Context, orgID uuid.UUID) (*dto.SubscriptionResponse, error) {
	sub, err := s.subRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !sub.WillCancelAtPeriodEnd() && sub.IsActive() {
		return nil, errors.BadRequest("Subscription is already active and not scheduled for cancellation")
	}

	// Resume via LemonSqueezy API
	_, err = s.lsClient.ResumeSubscription(ctx, sub.LemonSqueezySubscriptionID)
	if err != nil {
		s.log.Error("LemonSqueezy resume failed", err, "org_id", orgID)
		return nil, errors.Internal(err)
	}

	// Remove cancellation flag locally
	sub.UndoCancellation()
	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.log.Error("Failed to update subscription after resume", err, "sub_id", sub.ID)
		return nil, err
	}

	s.log.Info("Subscription resumed", "org_id", orgID, "sub_id", sub.ID)

	plan, _ := s.planRepo.GetByID(ctx, sub.PlanID)
	return &dto.SubscriptionResponse{
		Subscription: sub.ToPublic(),
		Plan:         plan.ToPublic(),
	}, nil
}

// ============================================================================
// CHANGE PLAN (Upgrade / Downgrade)
// ============================================================================

func (s *BillingService) ChangePlan(ctx context.Context, orgID uuid.UUID, req *dto.ChangePlanRequest) (*dto.SubscriptionResponse, error) {
	sub, err := s.subRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !sub.IsActive() {
		return nil, errors.BadRequest("Only active subscriptions can be changed")
	}

	// Get target plan
	newPlan, err := s.planRepo.GetByTierAndBilling(ctx, req.PlanTier, req.BillingPeriod)
	if err != nil {
		return nil, err
	}

	if newPlan.IsFreePlan() {
		return nil, errors.BadRequest("Cannot downgrade to free plan directly — cancel your subscription instead")
	}

	// Same plan check
	currentPlan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}
	if currentPlan.ID == newPlan.ID {
		return nil, errors.BadRequest("You are already on this plan")
	}

	// Parse new variant ID
	variantID, err := strconv.Atoi(newPlan.LemonSqueezyVariantID)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("invalid variant ID configuration"))
	}

	// Change via LemonSqueezy API (immediate proration)
	_, err = s.lsClient.ChangeSubscriptionPlan(ctx, sub.LemonSqueezySubscriptionID, variantID)
	if err != nil {
		s.log.Error("LemonSqueezy plan change failed", err, "org_id", orgID)
		return nil, errors.Internal(err)
	}

	// Update locally — webhook will also confirm with full data
	sub.PlanID = newPlan.ID
	sub.PriceAmount = newPlan.Price
	if err := s.subRepo.Update(ctx, sub); err != nil {
		s.log.Error("Failed to update subscription after plan change", err, "sub_id", sub.ID)
		return nil, err
	}

	s.log.Info("Plan changed", "org_id", orgID, "from_plan", currentPlan.Tier, "to_plan", newPlan.Tier)

	return &dto.SubscriptionResponse{
		Subscription: sub.ToPublic(),
		Plan:         newPlan.ToPublic(),
	}, nil
}

// ============================================================================
// PAYMENT HISTORY
// ============================================================================

func (s *BillingService) ListPayments(ctx context.Context, orgID uuid.UUID, req *dto.ListPaymentsRequest) (*dto.ListPaymentsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	payments, total, err := s.paymentRepo.ListByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		s.log.Error("Failed to list payments", err, "org_id", orgID)
		return nil, err
	}

	publicPayments := make([]*models.PublicPayment, 0, len(payments))
	for _, p := range payments {
		publicPayments = append(publicPayments, p.ToPublic())
	}

	return &dto.ListPaymentsResponse{
		Payments: publicPayments,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// ============================================================================
// USAGE
// ============================================================================

func (s *BillingService) GetUsage(ctx context.Context, orgID uuid.UUID) (*dto.UsageResponse, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return dto.ToUsageResponse(org), nil
}

// ============================================================================
// HELPERS
// ============================================================================

func isNotFound(err error) bool {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Code == 404
	}
	return false
}

// GetByID on plan repo (add this method)
func (s *BillingService) getPlanByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	return s.planRepo.GetByID(ctx, id)
}