package handlers

import (
	"server/internal/dto"
	"server/internal/services"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"shared/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
	validate    *validator.Validate
}

// NewAuthHandler creates new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// ============================================================================
// Registration
// ============================================================================

// Register handles user registration
// POST /auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if err := h.authService.Register(c.Context(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.MessageResponse{
		Success: true,
		Message: "Registration successful. Please check your email to verify your account.",
	})
}

// ============================================================================
// Login Handler
// ============================================================================

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	
	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Get IP and user agent for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	// Attempt login
	result, err := h.authService.Login(c.Context(), &req, ip, userAgent)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// ✅ If 2FA is required, return temp token (no cookies set yet)
	if result.Require2FA {
		return c.JSON(result)
	}

	// ✅ If login successful (no 2FA), set refresh token in HTTP-only cookie
	h.setRefreshTokenCookie(c, result.RefreshToken)

	return c.JSON(result)
}

// ============================================================================
// Verify 2FA Handler - UPDATED ✅
// ============================================================================

// ✅ Important: Used recovery codes are automatically deleted from database
func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	var req dto.Verify2FARequest
	
	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Get IP and user agent for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	// Verify 2FA code (TOTP or recovery code)
	result, err := h.authService.Verify2FA(c.Context(), &req, ip, userAgent)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// ✅ Set refresh token in HTTP-only cookie
	h.setRefreshTokenCookie(c, result.RefreshToken)

	return c.JSON(result)
}

// ============================================================================
// Email Verification
// ============================================================================

// VerifyEmail handles email verification
// POST /auth/verify-email
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if err := h.authService.VerifyEmail(c.Context(), req.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// ResendVerification handles resending verification email
// POST /auth/resend-verification
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if err := h.authService.ResendVerification(c.Context(), req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Verification email sent successfully",
	})
}

// ============================================================================
// Password Reset
// ============================================================================

// ForgotPassword handles forgot password request
// POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Always return success to prevent email enumeration
	_ = h.authService.ForgotPassword(c.Context(), req.Email)

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset with token
// POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if err := h.authService.ResetPassword(c.Context(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// ChangePassword handles password change for logged-in user
// POST /auth/change-password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Get user ID from context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	if err := h.authService.ChangePassword(c.Context(), userID, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshToken handles token refresh
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Refresh token not found",
		})
	}

	accessToken, err := h.authService.RefreshAccessToken(c.Context(), refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   15 * 60, // 15 minutes in seconds
		TokenType:   "Bearer",
	})
}

// ============================================================================
// Logout
// ============================================================================

// Logout handles user logout (current session)
// POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	sessionID, err := middleware.GetSessionID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Session not found",
		})
	}

	if err := h.authService.Logout(c.Context(), userID, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Failed to logout",
		})
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// LogoutAll handles logout from all devices
// POST /auth/logout-all
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	if err := h.authService.LogoutAll(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Failed to logout from all devices",
		})
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Logged out from all devices successfully",
	})
}

// ============================================================================
// Two-Factor Authentication
// ============================================================================

// Setup2FA handles 2FA setup
// POST /auth/2fa/setup
func (h *AuthHandler) Setup2FA(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	result, err := h.authService.Setup2FA(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(result)
}

// Enable2FA handles enabling 2FA
// POST /auth/2fa/enable
func (h *AuthHandler) Enable2FA(c *fiber.Ctx) error {
	var req dto.Enable2FARequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Get secret and recovery codes from request or session
	// For simplicity, we'll require them to be passed in the request
	// In production, you might want to store them temporarily in Redis
	secret := c.FormValue("secret")
	recoveryCodes := c.FormValue("recovery_codes") // Comma-separated

	if secret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Secret is required",
		})
	}

	// Parse recovery codes
	var codes []string
	if recoveryCodes != "" {
		codes = []string{} // Parse from comma-separated string
	}

	if err := h.authService.Enable2FA(c.Context(), userID, &req, secret, codes); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Two-factor authentication enabled successfully",
	})
}

// Disable2FA handles disabling 2FA
// POST /auth/2fa/disable
func (h *AuthHandler) Disable2FA(c *fiber.Ctx) error {
	var req dto.Disable2FARequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	if err := h.authService.Disable2FA(c.Context(), userID, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Two-factor authentication disabled successfully",
	})
}

// ============================================================================
// Session Management
// ============================================================================

// ListSessions handles listing user sessions
// GET /auth/sessions
func (h *AuthHandler) ListSessions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	result, err := h.authService.ListSessions(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(result)
}

// DeleteSession handles deleting a specific session
// DELETE /auth/sessions/:sessionId
func (h *AuthHandler) DeleteSession(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Session ID is required",
		})
	}

	if err := h.authService.DeleteSession(c.Context(), userID, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Session deleted successfully",
	})
}

// ============================================================================
// Helper Methods
// ============================================================================

// setRefreshTokenCookie sets refresh token in HTTP-only cookie
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "lax",
	})
}

// clearRefreshTokenCookie clears refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}