package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	// API v1 Group
	v1 := app.Group("/api/v1")

	// ============================================================================
	// PUBLIC AUTH ROUTES (No authentication required)
	// ============================================================================
	auth := v1.Group("/auth")
	{
		// Rate limit: 10 requests per minute for auth endpoints
		authRateLimit := rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 10,
			WindowDuration:    60 * time.Second,
			BypassPaths:       []string{"/api/v1/auth/verify-email"},
		})

		auth.Use(authRateLimit)

		// Registration & Email Verification
		auth.Post("/register", deps.AuthHandler.Register)
		auth.Post("/verify-email", deps.AuthHandler.VerifyEmail)
		auth.Post("/resend-verification", deps.AuthHandler.ResendVerificationEmail)

		// Login & Token Refresh
		auth.Post("/login", deps.AuthHandler.Login)
		auth.Post("/refresh", deps.AuthHandler.RefreshToken)

		// 2FA Verification (during login flow)
		auth.Post("/2fa/verify", deps.AuthHandler.Verify2FA)

		// Password Reset
		auth.Post("/forgot-password", deps.AuthHandler.ForgotPassword)
		auth.Post("/reset-password", deps.AuthHandler.ResetPassword)
	}

	// ============================================================================
	// PROTECTED ROUTES (Authentication required)
	// ============================================================================
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Password Management
		protected.Post("/auth/change-password", deps.AuthHandler.ChangePassword)

		// Logout
		protected.Post("/auth/logout", deps.AuthHandler.Logout)

		// Session Management
		sessions := protected.Group("/sessions")
		{
			sessions.Get("/", deps.AuthHandler.GetActiveSessions)
			sessions.Delete("/:session_id", deps.AuthHandler.RevokeSession)
		}

		// ============================================================================
		// TWO-FACTOR AUTHENTICATION (2FA) - Protected Routes
		// ============================================================================
		twoFactor := protected.Group("/auth/2fa")
		{
			// Get 2FA status
			twoFactor.Get("/status", deps.TwoFactorHandler.Get2FAStatus)

			// Enable 2FA (Step 1: Generate QR code)
			twoFactor.Post("/enable", deps.TwoFactorHandler.Enable2FA)

			// Confirm 2FA (Step 2: Verify TOTP code)
			twoFactor.Post("/confirm", deps.TwoFactorHandler.Confirm2FA)

			// Disable 2FA (Requires password + code/recovery)
			twoFactor.Post("/disable", deps.TwoFactorHandler.Disable2FA)

			// Regenerate backup codes
			twoFactor.Post("/regenerate-codes", deps.TwoFactorHandler.RegenerateBackupCodes)
		}
	}
}