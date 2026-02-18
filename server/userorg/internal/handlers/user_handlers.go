package handlers

import (
	"userorg/internal/dto"
	"userorg/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type UserHandler struct {
	userService *services.UserService
	log         *logger.Logger
}

func NewUserHandler(userService *services.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// ============================================================================
// GET ME
// ============================================================================

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	resp, err := h.userService.GetMe(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// UPDATE USER
// ============================================================================

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.userService.UpdateUser(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// UPDATE PREFERENCES
// ============================================================================

func (h *UserHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.UpdateUserPreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.userService.UpdatePreferences(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}