package middleware

import "github.com/gofiber/fiber/v2"

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Clickjacking protection
		c.Set("X-Frame-Options", "DENY")

		// Disable MIME sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enforce HTTPS (prod only)
		c.Set("Strict-Transport-Security",
			"max-age=63072000; includeSubDomains; preload")

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