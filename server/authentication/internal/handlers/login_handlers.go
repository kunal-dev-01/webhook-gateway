package handlers

import (
	"authentication/internal/dto"
	"authentication/internal/services"

	"github.com/gofiber/fiber/v2"

	"shared/errors"
	"shared/logger"
)

type LoginHandler struct {
	loginService *services.LoginService
	log         *logger.Logger
}

func NewLoginHandler(loginService *services.LoginService, log *logger.Logger) *LoginHandler {
	return &LoginHandler{
		loginService: loginService,
		log:         log,
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
                "temp_token":         resp.TempToken,
            },
        })
    }

    // Normal login - set cookie and return tokens
    setRefreshTokenCookie(c, resp.RefreshToken)
    
    return c.JSON(fiber.Map{
        "success": true,
        "data": fiber.Map{
            "two_factor_required": false,
            "access_token":        resp.AccessToken,
            "expires_in":          resp.ExpiresIn,
            "user":                resp.User, // ✅ lowercase 'user'
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

    setRefreshTokenCookie(c, resp.RefreshToken)
    
    return c.JSON(fiber.Map{
        "success": true,
        "data": fiber.Map{
            "access_token": resp.AccessToken,
            "expires_in":   resp.ExpiresIn,
            "user":         resp.User, // ✅ lowercase 'user'
        },
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