package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type PasswordHandler struct {
	passwordService *services.PasswordService
	log         *logger.Logger
}

func NewPasswordHandler(passwordService *services.PasswordService, log *logger.Logger) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
		log:         log,
	}
}


// ============================================================================
// PASSWORD MANAGEMENT
// ============================================================================

func (h *PasswordHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.passwordService.ForgotPassword(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *PasswordHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.passwordService.ResetPassword(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *PasswordHandler) ChangePassword(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.passwordService.ChangePassword(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
