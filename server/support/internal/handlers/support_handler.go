package handlers

import (
	"support/internal/dto"
	"support/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
	sharedUtils "shared/utils"
)

// ============================================================================
// Support Handler
// ============================================================================

type SupportHandler struct {
	supportService *services.SupportService
	log            *logger.Logger
}

func NewSupportHandler(supportService *services.SupportService, log *logger.Logger) *SupportHandler {
	return &SupportHandler{
		supportService: supportService,
		log:            log,
	}
}

// ============================================================================
// CREATE TICKET    POST /api/v1/support/tickets
// ============================================================================

func (h *SupportHandler) CreateTicket(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.CreateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.supportService.CreateTicket(c.Context(), authCtx.UserID, authCtx.OrganizationID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LIST TICKETS     GET /api/v1/support/tickets
// ============================================================================

func (h *SupportHandler) ListTickets(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	query := &dto.ListTicketsQuery{}
	if err := c.QueryParser(query); err != nil {
		return errors.BadRequest("Invalid query parameters")
	}

	resp, svcErr := h.supportService.ListTickets(c.Context(), orgID, query)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// GET TICKET       GET /api/v1/support/tickets/:ticket_id
// ============================================================================

func (h *SupportHandler) GetTicket(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.supportService.GetTicket(c.Context(), ticketID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// ADD MESSAGE      POST /api/v1/support/tickets/:ticket_id/messages
// ============================================================================

func (h *SupportHandler) AddMessage(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	var req dto.AddMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.supportService.AddMessage(c.Context(), ticketID, authCtx.OrganizationID, authCtx.UserID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LIST MESSAGES    GET /api/v1/support/tickets/:ticket_id/messages
// ============================================================================

func (h *SupportHandler) ListMessages(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.supportService.ListMessages(c.Context(), ticketID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// UPLOAD ATTACHMENT   POST /api/v1/support/tickets/:ticket_id/messages/:message_id/attachments
// ============================================================================

func (h *SupportHandler) UploadAttachment(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	messageID, err := parseUUID(c, "message_id")
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errors.BadRequest("file is required (field name: 'file')")
	}

	// Validate MIME type — allow images and common doc types
	contentType, err := sharedUtils.AttachmentContentType(fileHeader)
	if err != nil {
		return errors.BadRequest(err.Error())
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errors.Internal(err)
	}
	defer file.Close()

	resp, svcErr := h.supportService.UploadAttachment(
		c.Context(),
		ticketID, authCtx.OrganizationID, messageID, authCtx.UserID,
		file, fileHeader.Filename, contentType, fileHeader.Size,
	)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// GET DOWNLOAD URL   GET /api/v1/support/tickets/:ticket_id/attachments/:attachment_id/download
// ============================================================================

func (h *SupportHandler) GetAttachmentDownloadURL(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	attachmentID, err := parseUUID(c, "attachment_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.supportService.GetAttachmentDownloadURL(c.Context(), ticketID, orgID, attachmentID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// SUBMIT SATISFACTION   POST /api/v1/support/tickets/:ticket_id/satisfaction
// ============================================================================

func (h *SupportHandler) SubmitSatisfaction(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	var req dto.SubmitSatisfactionRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.supportService.SubmitSatisfaction(c.Context(), ticketID, orgID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// REOPEN TICKET    POST /api/v1/support/tickets/:ticket_id/reopen
// ============================================================================

func (h *SupportHandler) ReopenTicket(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	ticketID, err := parseUUID(c, "ticket_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.supportService.ReopenTicket(c.Context(), ticketID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// HELPER
// ============================================================================

func parseUUID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	raw := c.Params(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.BadRequest("Invalid " + param + " format")
	}
	return id, nil
}