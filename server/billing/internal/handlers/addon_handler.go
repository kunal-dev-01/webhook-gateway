// billing/internal/handlers/addon_handler.go
package handlers

import (
    "billing/internal/dto"
    "billing/internal/services"
    "shared/errors"
    "shared/logger"
    "shared/middleware"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

type AddonHandler struct {
    addonService *services.AddonService
    log          *logger.Logger
}

func NewAddonHandler(addonService *services.AddonService, log *logger.Logger) *AddonHandler {
    return &AddonHandler{addonService: addonService, log: log}
}

// GET /api/v1/billing/:org_id/addons/available
func (h *AddonHandler) ListAvailableAddons(c *fiber.Ctx) error {
    orgID, err := uuid.Parse(c.Params("org_id"))
    if err != nil {
        return errors.BadRequest("Invalid organization ID")
    }

    resp, err := h.addonService.ListAvailableAddons(c.Context(), orgID)
    if err != nil {
        return err
    }
    return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/billing/:org_id/addons/checkout
func (h *AddonHandler) CreateAddonCheckout(c *fiber.Ctx) error {
    authCtx, err := middleware.GetAuthContext(c)
    if err != nil {
        return errors.Unauthorized("Unauthorized")
    }

    orgID, err := uuid.Parse(c.Params("org_id"))
    if err != nil {
        return errors.BadRequest("Invalid organization ID")
    }

    var req dto.CreateAddonCheckoutRequest
    if err := c.BodyParser(&req); err != nil {
        return errors.BadRequest("Invalid request body")
    }
    if req.AddonPlanID == uuid.Nil {
        return errors.BadRequest("addon_plan_id is required")
    }

    resp, err := h.addonService.CreateAddonCheckout(
        c.Context(), orgID, authCtx.UserID,
        authCtx.Email, "", &req,
    )
    if err != nil {
        return err
    }
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/billing/:org_id/addons
func (h *AddonHandler) ListOrgAddons(c *fiber.Ctx) error {
    orgID, err := uuid.Parse(c.Params("org_id"))
    if err != nil {
        return errors.BadRequest("Invalid organization ID")
    }

    resp, err := h.addonService.ListOrgAddons(c.Context(), orgID)
    if err != nil {
        return err
    }
    return c.JSON(fiber.Map{"success": true, "data": resp})
}

// DELETE /api/v1/billing/:org_id/addons/:addon_id
func (h *AddonHandler) CancelAddon(c *fiber.Ctx) error {
    orgID, err := uuid.Parse(c.Params("org_id"))
    if err != nil {
        return errors.BadRequest("Invalid organization ID")
    }

    addonID, err := uuid.Parse(c.Params("addon_id"))
    if err != nil {
        return errors.BadRequest("Invalid addon ID")
    }

    public, err := h.addonService.CancelAddon(c.Context(), orgID, addonID)
    if err != nil {
        return err
    }
    return c.JSON(fiber.Map{
        "success": true,
        "message": "Add-on will be cancelled at the end of the billing period",
        "data":    public,
    })
}

// GET /api/v1/billing/:org_id/addons/limits
func (h *AddonHandler) GetEffectiveLimits(c *fiber.Ctx) error {
    orgID, err := uuid.Parse(c.Params("org_id"))
    if err != nil {
        return errors.BadRequest("Invalid organization ID")
    }

    resp, err := h.addonService.GetEffectiveLimits(c.Context(), orgID)
    if err != nil {
        return err
    }
    return c.JSON(fiber.Map{"success": true, "data": resp})
}