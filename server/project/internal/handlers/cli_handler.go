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
// CLI Session Handler
// ============================================================================

type CLISessionHandler struct {
	cliSessionService *services.CLISessionService
	log               *logger.Logger
}

func NewCLISessionHandler(cliSessionService *services.CLISessionService, log *logger.Logger) *CLISessionHandler {
	return &CLISessionHandler{cliSessionService: cliSessionService, log: log}
}

// GET /api/v1/user/cli-sessions
func (h *CLISessionHandler) ListSessions(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	// Mark current session in list (if using CLI session token)
	currentSessionID := getCurrentSessionID(c)

	resp, svcErr := h.cliSessionService.ListSessions(c.Context(), authCtx.UserID, currentSessionID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// GET /api/v1/user/cli-sessions/:session_id
func (h *CLISessionHandler) GetSession(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessionID, err := parseUUID(c, "session_id")
	if err != nil {
		return err
	}

	currentSessionID := getCurrentSessionID(c)

	resp, svcErr := h.cliSessionService.GetSession(c.Context(), sessionID, authCtx.UserID, currentSessionID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/cli-sessions/:session_id/revoke
func (h *CLISessionHandler) RevokeSession(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessionID, err := parseUUID(c, "session_id")
	if err != nil {
		return err
	}

	var req dto.RevokeCLISessionRequest
	_ = c.BodyParser(&req)

	resp, svcErr := h.cliSessionService.RevokeSession(c.Context(), sessionID, authCtx.UserID, &req)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// POST /api/v1/user/cli-sessions/revoke-all
func (h *CLISessionHandler) RevokeAllSessions(c *fiber.Ctx) error {
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	resp, svcErr := h.cliSessionService.RevokeAllSessions(c.Context(), authCtx.UserID)
	if svcErr != nil {
		return svcErr
	}

	return c.JSON(fiber.Map{"success": true, "data": resp})
}

// ============================================================================
// HELPER
// ============================================================================

// getCurrentSessionID extracts the current CLI session ID from JWT claims if present.
// Returns nil if not available (e.g., browser JWT session, not CLI).
func getCurrentSessionID(c *fiber.Ctx) *uuid.UUID {
	if id, ok := c.Locals("cli_session_id").(string); ok && id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	return nil
}