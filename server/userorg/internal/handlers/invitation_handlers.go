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

// ============================================================================
// Invitation Handler
// ============================================================================

type InvitationHandler struct {
	invService *services.InvitationService
	log        *logger.Logger
}

func NewInvitationHandler(invService *services.InvitationService, log *logger.Logger) *InvitationHandler {
	return &InvitationHandler{invService: invService, log: log}
}

// POST /api/v1/user/organizations/:org_id/invitations
func (h *InvitationHandler) InviteMember(c *fiber.Ctx) error {
	callerID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	var req dto.InviteMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if req.Email == "" {
		return errors.BadRequest("email is required")
	}
	if req.Role == "" {
		return errors.BadRequest("role is required")
	}

	resp, err := h.invService.InviteMember(c.Context(), orgID, callerID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// POST /api/v1/invitations/accept
func (h *InvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	callerID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.AcceptInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if req.Token == "" {
		return errors.BadRequest("token is required")
	}

	resp, err := h.invService.AcceptInvitation(c.Context(), callerID, req.Token)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// DELETE /api/v1/user/organizations/:org_id/invitations/:invitation_id
func (h *InvitationHandler) RevokeInvitation(c *fiber.Ctx) error {
	callerID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	invitationID, err := uuid.Parse(c.Params("invitation_id"))
	if err != nil {
		return errors.BadRequest("Invalid invitation ID")
	}

	if err := h.invService.RevokeInvitation(c.Context(), orgID, callerID, invitationID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invitation revoked successfully",
	})
}

// GET /api/v1/user/organizations/:org_id/invitations
func (h *InvitationHandler) ListInvitations(c *fiber.Ctx) error {
	callerID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	orgID, err := uuid.Parse(c.Params("org_id"))
	if err != nil {
		return errors.BadRequest("Invalid organization ID")
	}

	resp, err := h.invService.ListInvitations(c.Context(), orgID, callerID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
