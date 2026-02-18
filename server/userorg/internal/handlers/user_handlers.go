package handlers

import (
	"userorg/internal/dto"
	"userorg/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
	sharedUtils "shared/utils"
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
// UPDATE USER (JSON — full_name only)
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
// UPDATE AVATAR (multipart — avatar file)
// ============================================================================

func (h *UserHandler) UpdateAvatar(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return errors.BadRequest("avatar file is required")
	}

	// Validate size and MIME type
	if err := sharedUtils.ValidateImage(fileHeader); err != nil {
		return errors.BadRequest(err.Error())
	}

	// Detect content type
	contentType, err := sharedUtils.ImageContentType(fileHeader)
	if err != nil {
		return errors.BadRequest(err.Error())
	}

	// Open file for reading
	file, err := fileHeader.Open()
	if err != nil {
		return errors.Internal(err)
	}
	defer file.Close()

	resp, err := h.userService.UpdateAvatar(c.Context(), userID, file, contentType)
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