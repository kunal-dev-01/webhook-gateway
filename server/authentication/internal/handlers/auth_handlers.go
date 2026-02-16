package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LOGIN
// ============================================================================

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		return err
	}

	if resp.TwoFactorRequired {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    resp,
		})
	}

	setRefreshTokenCookie(c, resp.RefreshToken)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{
			"two_factor_required": resp.TwoFactorRequired,
			"access_token"        : resp.AccessToken,
			"expires_in"          : resp.ExpiresIn,
			"User":            resp.User,
		},
	})
}

func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	var req dto.TwoFactorVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.Verify2FA(c.Context(), &req)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, resp.RefreshToken)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{
			"access_token"        : resp.AccessToken,
			"expires_in"          : resp.ExpiresIn,
			"User":            resp.User,
		},
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return errors.Unauthorized("Refresh token not found")
	}

	req := &dto.RefreshTokenRequest{RefreshToken: refreshToken}

	resp, err := h.authService.RefreshToken(c.Context(), req)
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
// EMAIL VERIFICATION
// ============================================================================

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.VerifyEmail(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.ResendVerificationEmail(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// PASSWORD MANAGEMENT
// ============================================================================

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.ForgotPassword(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.ResetPassword(c.Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	resp, err := h.authService.ChangePassword(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// ============================================================================
// LOGOUT & SESSIONS
// ============================================================================

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
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

	resp, err := h.authService.Logout(c.Context(), userID, sessionID, &req)
	if err != nil {
		return err
	}

	clearRefreshTokenCookie(c)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) GetActiveSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessions, err := h.authService.GetActiveSessions(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sessions,
	})
}

func (h *AuthHandler) RevokeSession(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	sessionID := c.Params("session_id")
	if sessionID == "" {
		return errors.BadRequest("Session ID is required")
	}

	if err := h.authService.RevokeSessionByID(c.Context(), userID, sessionID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session revoked successfully",
	})
}

// ============================================================================
// HELPERS
// ============================================================================

func setRefreshTokenCookie(c *fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}