package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type TwoFactorHandler struct {
	twoFactorService *services.TwoFactorService
	log              *logger.Logger
}

func NewTwoFactorHandler(
	twoFactorService *services.TwoFactorService,
	log *logger.Logger,
) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorService: twoFactorService,
		log:              log,
	}
}

// ============================================================================
// GET 2FA STATUS
// ============================================================================

func (h *TwoFactorHandler) Get2FAStatus(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	resp, err := h.twoFactorService.Get2FAStatus(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// ENABLE 2FA (STEP 1)
// ============================================================================

func (h *TwoFactorHandler) Enable2FA(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	resp, err := h.twoFactorService.Enable2FA(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// CONFIRM 2FA (STEP 2)
// ============================================================================

func (h *TwoFactorHandler) Confirm2FA(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.Confirm2FARequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.twoFactorService.Confirm2FA(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// DISABLE 2FA
// ============================================================================

func (h *TwoFactorHandler) Disable2FA(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.Disable2FARequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.twoFactorService.Disable2FA(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// REGENERATE BACKUP CODES
// ============================================================================

func (h *TwoFactorHandler) RegenerateBackupCodes(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.RegenerateBackupCodesRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.twoFactorService.RegenerateBackupCodes(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}