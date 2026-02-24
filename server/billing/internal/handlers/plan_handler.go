package handlers

import (
	"billing/internal/services"
	"shared/errors"
	"shared/logger"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// PLAN HANDLER
// ============================================================================

type PlanHandler struct {
	planService *services.PlanService
	log         *logger.Logger
}

func NewPlanHandler(planService *services.PlanService, log *logger.Logger) *PlanHandler {
	return &PlanHandler{planService: planService, log: log}
}

// ListPlans — GET /api/v1/plans (public)
func (h *PlanHandler) ListPlans(c *fiber.Ctx) error {
	resp, err := h.planService.ListPlans(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GetPlan — GET /api/v1/plans/:tier/:billing_period (public)
func (h *PlanHandler) GetPlan(c *fiber.Ctx) error {
	tier := c.Params("tier")
	billingPeriod := c.Params("billing_period")

	if tier == "" || billingPeriod == "" {
		return errors.BadRequest("tier and billing_period are required")
	}

	resp, err := h.planService.ListPlans(c.Context())
	if err != nil {
		return err
	}

	// Filter from list
	for _, plan := range resp.Plans {
		if string(plan.Tier) == tier && string(plan.BillingPeriod) == billingPeriod {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    plan,
			})
		}
	}

	return errors.NotFound("Plan not found")
}