package handlers

import (
	"billing/internal/dto"
	"billing/internal/services"
	"shared/errors"
	"shared/logger"
	"shared/middleware"
	"shared/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ============================================================================
// BILLING HANDLER
// ============================================================================

type BillingHandler struct {
	billingService *services.BillingService
	log            *logger.Logger
}

func NewBillingHandler(billingService *services.BillingService, log *logger.Logger) *BillingHandler {
	return &BillingHandler{billingService: billingService, log: log}
}

// ── Create Checkout ───────────────────────────────────────────────────────────

// CreateCheckout — POST /api/v1/billing/:org_id/checkout
func (h *BillingHandler) CreateCheckout(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	var req dto.CreateCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if req.PlanTier == "" {
		return errors.BadRequest("plan_tier is required")
	}
	if req.BillingPeriod == "" {
		return errors.BadRequest("billing_period is required")
	}

	resp, err := h.billingService.CreateCheckout(
		c.Context(),
		orgID,
		authCtx.UserID,
		authCtx.Email,
		"", // Name not in JWT — can add if needed
		&req,
	)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ── Subscription ─────────────────────────────────────────────────────────────

// GetSubscription — GET /api/v1/billing/:org_id/subscription
func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.billingService.GetSubscription(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// CancelSubscription — DELETE /api/v1/billing/:org_id/subscription
func (h *BillingHandler) CancelSubscription(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.billingService.CancelSubscription(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription will be cancelled at the end of the billing period",
		"data":    resp,
	})
}

// ResumeSubscription — POST /api/v1/billing/:org_id/subscription/resume
func (h *BillingHandler) ResumeSubscription(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.billingService.ResumeSubscription(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription resumed successfully",
		"data":    resp,
	})
}

// ChangePlan — PATCH /api/v1/billing/:org_id/subscription/plan
func (h *BillingHandler) ChangePlan(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	var req dto.ChangePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if req.PlanTier == "" {
		return errors.BadRequest("plan_tier is required")
	}
	if req.BillingPeriod == "" {
		return errors.BadRequest("billing_period is required")
	}

	// Validate values
	validTiers := map[models.PlanTier]bool{
		models.PlanDeveloper: true, models.PlanStarter: true,
		models.PlanGrowth: true, models.PlanScale: true,
	}
	validPeriods := map[models.BillingPeriod]bool{
		models.BillingMonthly: true, models.BillingAnnual: true,
	}

	if !validTiers[req.PlanTier] {
		return errors.BadRequest("Invalid plan_tier")
	}
	if !validPeriods[req.BillingPeriod] {
		return errors.BadRequest("Invalid billing_period. Must be 'monthly' or 'annual'")
	}

	resp, err := h.billingService.ChangePlan(c.Context(), orgID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Plan changed successfully",
		"data":    resp,
	})
}

// ── Payments ─────────────────────────────────────────────────────────────────

// ListPayments — GET /api/v1/billing/:org_id/payments
func (h *BillingHandler) ListPayments(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	req := &dto.ListPaymentsRequest{}
	if err := c.QueryParser(req); err != nil {
		return errors.BadRequest("Invalid query parameters")
	}

	resp, err := h.billingService.ListPayments(c.Context(), orgID, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ── Usage ─────────────────────────────────────────────────────────────────────

// GetUsage — GET /api/v1/billing/:org_id/usage
func (h *BillingHandler) GetUsage(c *fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.billingService.GetUsage(c.Context(), orgID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}