package handlers

import (
	"project/internal/dto"
	"project/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

// ============================================================================
// Project Handler
// ============================================================================

type ProjectHandler struct {
	projectService *services.ProjectService
	log            *logger.Logger
}

func NewProjectHandler(projectService *services.ProjectService, log *logger.Logger) *ProjectHandler {
	return &ProjectHandler{projectService: projectService, log: log}
}

// POST /api/v1/user/organizations/:org_id/projects
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	if err := requireOrgMatch(c, authCtx.OrganizationID); err != nil {
		return err
	}

	// projectsLimit should come from org plan — default 5 for developer
	// In real app: fetch from org or inject via context
	projectsLimit := getPlanProjectsLimit(c)

	var req dto.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.projectService.CreateProject(c.Context(), authCtx.OrganizationID, projectsLimit, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GET /api/v1/user/organizations/:org_id/projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	resp, svcErr := h.projectService.ListProjects(c.Context(), orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/user/organizations/:org_id/projects/:project_id
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
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

	resp, svcErr := h.projectService.GetProject(c.Context(), projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// PATCH /api/v1/user/organizations/:org_id/projects/:project_id
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
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

	var req dto.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.projectService.UpdateProject(c.Context(), projectID, orgID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// DELETE /api/v1/user/organizations/:org_id/projects/:project_id
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
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

	resp, svcErr := h.projectService.DeleteProject(c.Context(), projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/organizations/:org_id/projects/:project_id/toggle
func (h *ProjectHandler) ToggleProject(c *fiber.Ctx) error {
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

	resp, svcErr := h.projectService.ToggleProject(c.Context(), projectID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// ============================================================================
// HELPERS — shared across all handlers in this package
// ============================================================================

func parseUUID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, errors.BadRequest("Invalid " + param + " format")
	}
	return id, nil
}

// requireOrgMatch ensures URL :org_id matches the JWT claim org
func requireOrgMatch(c *fiber.Ctx, claimsOrgID uuid.UUID) error {
	paramOrgID, err := parseUUID(c, "org_id")
	if err != nil {
		return err
	}
	if paramOrgID != claimsOrgID {
		return errors.Forbidden("Access denied: organization mismatch")
	}
	return nil
}

// getPlanProjectsLimit reads projects limit — inject from org context in real app
func getPlanProjectsLimit(c *fiber.Ctx) int {
	if limit, ok := c.Locals("projects_limit").(int); ok && limit > 0 {
		return limit
	}
	return 5 // developer plan default
}