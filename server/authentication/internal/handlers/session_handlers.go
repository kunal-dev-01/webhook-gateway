package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"
	"authentication/internal/utils"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type SessionHandler struct {
	sessionService *services.SessionService
	sessionMgr     *utils.Manager
	log            *logger.Logger
}

func NewSessionHandler(sessionService *services.SessionService, log *logger.Logger, sessionMgr *utils.Manager) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		sessionMgr:     sessionMgr,
		log:            log,
	}
}

// ============================================================================
// REFRESH
// ============================================================================

func (h *SessionHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return errors.Unauthorized("Refresh token not found")
	}

	req := &dto.RefreshTokenRequest{RefreshToken: refreshToken}

	resp, err := h.sessionService.RefreshToken(c.Context(), req)
	if err != nil {
		clearRefreshTokenCookie(c)
		return err
	}

	setRefreshTokenCookie(c, resp.RefreshToken)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LOGOUT & SESSIONS
// ============================================================================

func (h *SessionHandler) Logout(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessionID, err := middleware.GetSessionID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		req.AllDevices = false
	}

	resp, err := h.sessionService.Logout(c.Context(), userID, sessionID, &req)
	if err != nil {
		return err
	}

	clearRefreshTokenCookie(c)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *SessionHandler) GetActiveSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	// 1. Current session ID nikal lo
	currentSessionID, err := middleware.GetSessionID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	// 2. Database se saare sessions le aao
	allSessions, err := h.sessionService.GetActiveSessions(c.Context(), userID)
	if err != nil {
		return err
	}

	// 3. Current session ko filter out kar do
	var filteredSessions []utils.SessionData // ✅ Package name 'utils' use karein
	for _, session := range allSessions {
		// Sirf wo session add karo jo current nahi hain
		if session.SessionID != currentSessionID {
			filteredSessions = append(filteredSessions, session)
		}
	}

	// Agar nil slice ban jaye (sirf current session hi tha), toh empty array return karo
	if filteredSessions == nil {
		filteredSessions = []utils.SessionData{} // ✅ Yahan bhi 'utils' aayega
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"sessions": filteredSessions,
			"total":    len(filteredSessions), // Ab count bhi filtered ka jayega
		},
	})
}

func (h *SessionHandler) RevokeSession(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessionID := c.Params("session_id")
	if sessionID == "" {
		return errors.BadRequest("Session ID is required")
	}

	if err := h.sessionService.RevokeSessionByID(c.Context(), userID, sessionID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session revoked successfully",
	})
}
