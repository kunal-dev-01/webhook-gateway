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
// API Key Handler
// ============================================================================

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
	log           *logger.Logger
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService, log *logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService, log: log}
}

// POST /api/v1/user/organizations/:org_id/api-keys
func (h *APIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, authCtx.OrganizationID); err != nil {
		return err
	}

	var req dto.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, svcErr := h.apiKeyService.CreateAPIKey(c.Context(), authCtx.OrganizationID, authCtx.UserID, &req)
	if svcErr != nil {
		return svcErr
	}

	// 201 + plaintext key — this is the ONLY time the key is returned
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// GET /api/v1/user/organizations/:org_id/api-keys
func (h *APIKeyHandler) ListAPIKeys(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	resp, svcErr := h.apiKeyService.ListAPIKeys(c.Context(), orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/user/organizations/:org_id/api-keys/:key_id
func (h *APIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	keyID, err := parseUUID(c, "key_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.apiKeyService.GetAPIKey(c.Context(), keyID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/organizations/:org_id/api-keys/:key_id/revoke
func (h *APIKeyHandler) RevokeAPIKey(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, authCtx.OrganizationID); err != nil {
		return err
	}

	keyID, err := parseUUID(c, "key_id")
	if err != nil {
		return err
	}

	var req dto.RevokeAPIKeyRequest
	// Reason is optional — parse but don't fail if empty body
	_ = c.BodyParser(&req)

	resp, svcErr := h.apiKeyService.RevokeAPIKey(c.Context(), keyID, authCtx.OrganizationID, authCtx.UserID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// DELETE /api/v1/user/organizations/:org_id/api-keys/:key_id
func (h *APIKeyHandler) DeleteAPIKey(c *fiber.Ctx) error {
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}
	if err := requireOrgMatch(c, orgID); err != nil {
		return err
	}

	keyID, err := parseUUID(c, "key_id")
	if err != nil {
		return err
	}

	resp, svcErr := h.apiKeyService.DeleteAPIKey(c.Context(), keyID, orgID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}