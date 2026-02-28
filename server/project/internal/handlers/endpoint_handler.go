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
// Endpoint Handler
// ============================================================================

type EndpointHandler struct {
	endpointService *services.EndpointService
	log             *logger.Logger
}

func NewEndpointHandler(endpointService *services.EndpointService, log *logger.Logger) *EndpointHandler {
	return &EndpointHandler{endpointService: endpointService, log: log}
}

// POST /api/v1/user/organizations/:org_id/projects/:project_id/endpoints
func (h *EndpointHandler) CreateEndpoint(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}

	var req dto.CreateEndpointRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.endpointService.CreateEndpoint(c.Context(), projectID, orgID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GET /api/v1/user/organizations/:org_id/projects/:project_id/endpoints
func (h *EndpointHandler) ListEndpoints(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.endpointService.ListEndpoints(c.Context(), projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id
func (h *EndpointHandler) GetEndpoint(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}
	endpointID, err := parseUUID(c, "endpoint_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.endpointService.GetEndpoint(c.Context(), endpointID, projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// PATCH /api/v1/user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id
func (h *EndpointHandler) UpdateEndpoint(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}
	endpointID, err := parseUUID(c, "endpoint_id")
	if err != nil {
		return err
	}

	var req dto.UpdateEndpointRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.endpointService.UpdateEndpoint(c.Context(), endpointID, projectID, orgID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// DELETE /api/v1/user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id
func (h *EndpointHandler) DeleteEndpoint(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}
	endpointID, err := parseUUID(c, "endpoint_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.endpointService.DeleteEndpoint(c.Context(), endpointID, projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/organizations/:org_id/projects/:project_id/endpoints/:endpoint_id/toggle
func (h *EndpointHandler) ToggleEndpoint(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	projectID, err := parseUUID(c, "project_id")
	if err != nil {
		return err
	}
	endpointID, err := parseUUID(c, "endpoint_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.endpointService.ToggleEndpoint(c.Context(), endpointID, projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}