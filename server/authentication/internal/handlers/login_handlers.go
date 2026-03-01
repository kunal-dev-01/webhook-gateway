package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

type LoginHandler struct {
	loginService   *services.LoginService
	log            *logger.Logger
	cookieSettings middleware.CookieSettings
}

func NewLoginHandler(loginService *services.LoginService, log *logger.Logger, cookieSettings middleware.CookieSettings) *LoginHandler {
	return &LoginHandler{
		loginService:   loginService,
		log:            log,
		cookieSettings: cookieSettings,
	}
}

// ============================================================================
// LOGIN
// ============================================================================

func (h *LoginHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	req.IPAddress = c.IP()
	req.UserAgent = c.Get("User-Agent")

	resp, err := h.loginService.Login(c.Context(), &req)
	if err != nil {
		return err
	}

	// If 2FA required, return temp_token
	if resp.TwoFactorRequired {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"two_factor_required": true,
				"temp_token":          resp.TempToken,
			},
		})
	}

	// Normal login - set cookie and return tokens
	setRefreshTokenCookie(c, resp.RefreshToken, h.cookieSettings)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"two_factor_required": false,
			"access_token":        resp.AccessToken,
			"expires_in":          resp.ExpiresIn,
			"user":                resp.User,
		},
	})
}

func (h *LoginHandler) Verify2FA(c *fiber.Ctx) error {
	var req dto.TwoFactorVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	req.IPAddress = c.IP()
	req.UserAgent = c.Get("User-Agent")

	resp, err := h.loginService.Verify2FA(c.Context(), &req)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, resp.RefreshToken, h.cookieSettings)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token": resp.AccessToken,
			"expires_in":   resp.ExpiresIn,
			"user":         resp.User,
		},
	})
}

// ============================================================================
// COOKIE HELPERS (Environment-Aware)
// ============================================================================

func setRefreshTokenCookie(c *fiber.Ctx, refreshToken string, cs middleware.CookieSettings) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   cs.Secure,
		HTTPOnly: cs.HTTPOnly,
		SameSite: cs.SameSite,
	})
}

func clearRefreshTokenCookie(c *fiber.Ctx, cs middleware.CookieSettings) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   cs.Secure,
		HTTPOnly: cs.HTTPOnly,
		SameSite: cs.SameSite,
	})
}
