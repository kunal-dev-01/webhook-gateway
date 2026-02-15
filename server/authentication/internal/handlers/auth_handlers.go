package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"
	"shared/logger"
	"shared/errors"
	"shared/middleware"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewAuthHandler(authService *services.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

// ============================================================================
// REGISTER
// ============================================================================

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

// VerifyEmail godoc
// @Summary Verify user email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verification token"
// @Success 200 {object} dto.VerifyEmailResponse
// @Failure 400 {object} map[string]interface{}
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.VerifyEmail(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ResendVerificationEmail godoc
// @Summary Resend verification email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "User email"
// @Success 200 {object} dto.ResendVerificationResponse
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.ResendVerificationEmail(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// PASSWORD RESET
// ============================================================================

// ForgotPassword godoc
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "User email"
// @Success 200 {object} dto.ForgotPasswordResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.ForgotPassword(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ResetPassword godoc
// @Summary Reset password with token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} dto.ResetPasswordResponse
// @Failure 400 {object} map[string]interface{}
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.ResetPassword(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ChangePassword godoc
// @Summary Change user password (authenticated)
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} dto.ChangePasswordResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	resp, err := h.authService.ChangePassword(c.Context(), userID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LOGOUT
// ============================================================================

// Logout godoc
// @Summary Logout user
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Logout options"
// @Success 200 {object} dto.LogoutResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get user ID and session ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	sessionID, err := middleware.GetSessionID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req dto.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		// Default: logout current session only
		req.AllDevices = false
	}

	resp, err := h.authService.Logout(c.Context(), userID, sessionID, &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// SESSION MANAGEMENT
// ============================================================================

// GetActiveSessions godoc
// @Summary Get all active sessions
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {array} utils.SessionData
// @Failure 401 {object} map[string]interface{}
// @Router /auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	sessions, err := h.authService.GetActiveSessions(c.Context(), userID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sessions,
	})
}

// RevokeSession godoc
// @Summary Revoke a specific session
// @Tags Auth
// @Security BearerAuth
// @Param session_id path string true "Session ID to revoke"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/sessions/{session_id} [delete]
func (h *AuthHandler) RevokeSession(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	sessionID := c.Params("session_id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Session ID is required",
		})
	}

	if err := h.authService.RevokeSessionByID(c.Context(), userID, sessionID); err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// ============================================================================
// ERROR HANDLER
// ============================================================================

func handleError(c *fiber.Ctx, err error) error {
	// Type assertion for custom AppError
	if appErr, ok := err.(*errors.AppError); ok {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"error":   appErr.Message,
			"fields":  appErr.Fields,
		})
	}

	// Default internal server error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "Something went wrong",
	})
}