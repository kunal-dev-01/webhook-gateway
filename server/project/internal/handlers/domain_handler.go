package handlers

import (
	"project/internal/dto"
	"project/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

// ============================================================================
// Domain Handler
// ============================================================================

type DomainHandler struct {
	domainService *services.DomainService
	log           *logger.Logger
}

func NewDomainHandler(domainService *services.DomainService, log *logger.Logger) *DomainHandler {
	return &DomainHandler{domainService: domainService, log: log}
}

// POST /api/v1/user/organizations/:org_id/domains
func (h *DomainHandler) AddDomain(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	var req dto.AddDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.domainService.AddDomain(c.Context(), orgID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GET /api/v1/user/organizations/:org_id/domains
func (h *DomainHandler) ListDomains(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	resp, svcErr := h.domainService.ListDomains(c.Context(), orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/user/organizations/:org_id/domains/:domain_id
func (h *DomainHandler) GetDomain(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	domainID, err := parseUUID(c, "domain_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.domainService.GetDomain(c.Context(), domainID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// DELETE /api/v1/user/organizations/:org_id/domains/:domain_id
func (h *DomainHandler) DeleteDomain(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	domainID, err := parseUUID(c, "domain_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.domainService.DeleteDomain(c.Context(), domainID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/organizations/:org_id/domains/:domain_id/verify
func (h *DomainHandler) TriggerVerification(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	domainID, err := parseUUID(c, "domain_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.domainService.TriggerVerification(c.Context(), domainID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}