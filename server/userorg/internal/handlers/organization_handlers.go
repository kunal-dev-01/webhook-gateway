package handlers

import (
	"userorg/internal/dto"
	"userorg/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type OrganizationHandler struct {
	orgService *services.OrganizationService
	log        *logger.Logger
}

func NewOrganizationHandler(orgService *services.OrganizationService, log *logger.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		log:        log,
	}
}

// ============================================================================
// LIST MY ORGANIZATIONS
// ============================================================================

func (h *OrganizationHandler) ListMyOrganizations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	resp, err := h.orgService.ListMyOrganizations(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// GET ORGANIZATION DETAILS
// ============================================================================

func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.orgService.GetOrganization(c.Context(), orgID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// RENAME ORGANIZATION
// ============================================================================
func (h *OrganizationHandler) UpdateName(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if body.Name == "" {
		return errors.BadRequest("name is required")
	}

	resp, err := h.orgService.UpdateName(c.Context(), orgID, userID, body.Name)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LIST MEMBERS
// ============================================================================

func (h *OrganizationHandler) ListMembers(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	req := &dto.ListMembersRequest{}
	if err := c.QueryParser(req); err != nil {
		return errors.BadRequest("Invalid query parameters")
	}

	resp, err := h.orgService.ListMembers(c.Context(), orgID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// REMOVE MEMBER
// ============================================================================

func (h *OrganizationHandler) RemoveMember(c *fiber.Ctx) error {
	callerID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	targetUserID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return errors.BadRequest("Invalid user ID")
	}

	if err := h.orgService.RemoveMember(c.Context(), orgID, callerID, targetUserID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Member removed successfully",
	})
}