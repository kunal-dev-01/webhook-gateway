package middleware

import "github.com/gofiber/fiber/v2"

// ============================================================================
// Security Headers Middleware
// ============================================================================

// SecurityHeaders returns a middleware that sets security-related HTTP headers.
// HSTS is only applied in production (requires real HTTPS).
func SecurityHeaders(isProduction bool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Clickjacking protection
		c.Set("X-Frame-Options", "DENY")

		// Disable MIME sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enforce HTTPS (production only)
		if isProduction {
			c.Set("Strict-Transport-Security",
				"max-age=63072000; includeSubDomains; preload")
		}

		// Strong Content Security Policy
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self';")

		// Control referrer info
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable dangerous browser APIs
		c.Set("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=(), usb=()")

		// Hide backend info
		c.Set("X-Powered-By", "")

		return c.Next()
	}
}

// ============================================================================
// Cookie Settings (Environment-Aware)
// ============================================================================

// CookieSettings holds environment-aware cookie security attributes.
type CookieSettings struct {
	Secure   bool
	SameSite string
	HTTPOnly bool
}

// NewCookieSettings returns cookie settings appropriate for the environment.
// Production: Secure=true, SameSite=Strict, HTTPOnly=true
// Development: Secure=false, SameSite=Lax, HTTPOnly=true
func NewCookieSettings(isProduction bool) CookieSettings {
	if isProduction {
		return CookieSettings{
			Secure:   true,
			SameSite: "Strict",
			HTTPOnly: true,
		}
	}
	return CookieSettings{
		Secure:   false,
		SameSite: "Lax",
		HTTPOnly: true,
	}
}
