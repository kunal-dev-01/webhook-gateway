package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
)

type RegisterHandler struct {
	registerService *services.RegisterService
	log             *logger.Logger
}

func NewRegisterHandler(registerService *services.RegisterService, log *logger.Logger) *RegisterHandler {
	return &RegisterHandler{
		registerService: registerService,
		log:         log,
	}
}

// ============================================================================
// REGISTER
// ============================================================================

func (h *RegisterHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.registerService.Register(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

func (h *RegisterHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.registerService.VerifyEmail(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *RegisterHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.registerService.ResendVerificationEmail(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
