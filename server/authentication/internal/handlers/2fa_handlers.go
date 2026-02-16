package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shared/logger"
)

type TwoFactorHandler struct {
	twoFactorService *services.TwoFactorService
	log              *logger.Logger
}

func NewTwoFactorHandler(twoFactorService *services.TwoFactorService, log *logger.Logger) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorService: twoFactorService,
		log:              log,
	}
}

// ============================================================================
// GET 2FA STATUS
// ============================================================================

// Get2FAStatus godoc
// @Summary Get two-factor authentication status
// @Description Check if 2FA is enabled for the current user
// @Tags 2FA
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TwoFactorStatusResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/2fa/status [get]
func (h *TwoFactorHandler) Get2FAStatus(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	resp, err := h.twoFactorService.Get2FAStatus(c.Context(), userID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// ENABLE 2FA (Step 1: Generate QR Code)
// ============================================================================

// Enable2FA godoc
// @Summary Enable two-factor authentication (Step 1)
// @Description Generate QR code and secret for 2FA setup
// @Tags 2FA
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Enable2FAResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/2fa/enable [post]
func (h *TwoFactorHandler) Enable2FA(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	resp, err := h.twoFactorService.Enable2FA(c.Context(), userID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// CONFIRM 2FA (Step 2: Verify TOTP Code)
// ============================================================================

// Confirm2FA godoc
// @Summary Confirm two-factor authentication setup (Step 2)
// @Description Verify TOTP code and complete 2FA setup
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.Confirm2FARequest true "TOTP code"
// @Success 200 {object} dto.Confirm2FAResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/2fa/confirm [post]
func (h *TwoFactorHandler) Confirm2FA(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req dto.Confirm2FARequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.twoFactorService.Confirm2FA(c.Context(), userID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// DISABLE 2FA
// ============================================================================

// Disable2FA godoc
// @Summary Disable two-factor authentication
// @Description Disable 2FA with password and TOTP code or recovery code
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.Disable2FARequest true "Password and code"
// @Success 200 {object} dto.Disable2FAResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/2fa/disable [post]
func (h *TwoFactorHandler) Disable2FA(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req dto.Disable2FARequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.twoFactorService.Disable2FA(c.Context(), userID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// REGENERATE BACKUP CODES
// ============================================================================

// RegenerateBackupCodes godoc
// @Summary Regenerate backup codes
// @Description Generate new backup codes for 2FA
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RegenerateBackupCodesRequest true "Password"
// @Success 200 {object} dto.RegenerateBackupCodesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/2fa/regenerate-codes [post]
func (h *TwoFactorHandler) RegenerateBackupCodes(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req dto.RegenerateBackupCodesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.twoFactorService.RegenerateBackupCodes(c.Context(), userID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}