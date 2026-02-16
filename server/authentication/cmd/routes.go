package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	// API v1 Group
	v1 := app.Group("/api/v1")

	// ============================================================================
	// 🟢 PUBLIC AUTH ROUTES (Strict rate limiting)
	// ============================================================================
	auth := v1.Group("/auth")
	{
		// Strict rate limit for public auth: 10 requests per minute
		authRateLimit := rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 10,
			WindowDuration:    60 * time.Second,
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
	// 🔴 PROTECTED USER ROUTES (Moderate rate limiting)
	// ============================================================================
	user := v1.Group("")
	user.Use(authMiddleware.RequireAuth())
	{
		// Moderate rate limit for authenticated users: 100 requests per minute
		userRateLimit := rateLimiter.UserRateLimit(middleware.RateLimitConfig{
			RequestsPerWindow: 100,
			WindowDuration:    60 * time.Second,
		})
		user.Use(userRateLimit)

		// Password & Account Management
		user.Post("/auth/change-password", deps.AuthHandler.ChangePassword)
		user.Post("/auth/logout", deps.AuthHandler.Logout)

		// Session Management (stricter limit)
		sessions := user.Group("/auth/sessions")
		{
			sessionRateLimit := rateLimiter.UserRateLimit(middleware.RateLimitConfig{
				RequestsPerWindow: 20,
				WindowDuration:    60 * time.Second,
			})
			sessions.Use(sessionRateLimit)

			sessions.Get("/", deps.AuthHandler.GetActiveSessions)
			sessions.Delete("/:session_id", deps.AuthHandler.RevokeSession)
		}

		// ============================================================================
		// TWO-FACTOR AUTHENTICATION (2FA) - Stricter limits
		// ============================================================================
		twoFactor := user.Group("/auth/2fa")
		{
			// Strict rate limit for 2FA operations: 15 requests per minute
			twoFactorRateLimit := rateLimiter.UserRateLimit(middleware.RateLimitConfig{
				RequestsPerWindow: 15,
				WindowDuration:    60 * time.Second,
			})
			twoFactor.Use(twoFactorRateLimit)

			// Get 2FA status
			twoFactor.Get("/status", deps.TwoFactorHandler.Get2FAStatus)

			// Enable 2FA (Step 1: Generate QR code)
			twoFactor.Post("/enable", deps.TwoFactorHandler.Enable2FA)

			// Confirm 2FA (Step 2: Verify TOTP code)
			twoFactor.Post("/confirm", deps.TwoFactorHandler.Confirm2FA)

			// Disable 2FA
			twoFactor.Post("/disable", deps.TwoFactorHandler.Disable2FA)

			// Regenerate backup codes
			twoFactor.Post("/regenerate-codes", deps.TwoFactorHandler.RegenerateBackupCodes)
		}
	}
}