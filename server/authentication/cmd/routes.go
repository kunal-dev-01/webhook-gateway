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
	// 🟢 PUBLIC AUTH ROUTES
	// ============================================================================
	auth := v1.Group("/auth")
	{
		// ─────────────────────────────────────────────────────────────────
		// REGISTRATION & EMAIL VERIFICATION (Moderate: 10/min)
		// ─────────────────────────────────────────────────────────────────
		registration := auth.Group("")
		registration.Use(rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 10,
			WindowDuration:    60 * time.Second,
		}))
		
		registration.Post("/register", deps.AuthHandler.Register)
		registration.Post("/verify-email", deps.AuthHandler.VerifyEmail)
		registration.Post("/resend-verification", deps.AuthHandler.ResendVerificationEmail)

		// ─────────────────────────────────────────────────────────────────
		// LOGIN (Strict: 5/min to prevent brute force)
		// ─────────────────────────────────────────────────────────────────
		login := auth.Group("")
		login.Use(rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 5, // ✅ Stricter for login
			WindowDuration:    60 * time.Second,
		}))
		
		login.Post("/login", deps.AuthHandler.Login)

		// ─────────────────────────────────────────────────────────────────
		// 2FA VERIFICATION during login (Very Strict: 5/min)
		// ─────────────────────────────────────────────────────────────────
		twoFactorVerify := auth.Group("/2fa")
		twoFactorVerify.Use(rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 5, // ✅ Very strict - prevent TOTP brute force
			WindowDuration:    60 * time.Second,
		}))
		
		twoFactorVerify.Post("/verify", deps.AuthHandler.Verify2FA)

		// ─────────────────────────────────────────────────────────────────
		// TOKEN REFRESH (Moderate: 20/min)
		// ─────────────────────────────────────────────────────────────────
		refresh := auth.Group("")
		refresh.Use(rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 20, // ✅ More lenient for token refresh
			WindowDuration:    60 * time.Second,
		}))
		
		refresh.Post("/refresh", deps.AuthHandler.RefreshToken)

		// ─────────────────────────────────────────────────────────────────
		// PASSWORD RESET (Strict: 3/min per IP)
		// ─────────────────────────────────────────────────────────────────
		passwordReset := auth.Group("")
		passwordReset.Use(rateLimiter.Middleware(middleware.RateLimitConfig{
			RequestsPerWindow: 3, // ✅ Very strict - prevent abuse
			WindowDuration:    60 * time.Second,
		}))
		
		passwordReset.Post("/forgot-password", deps.AuthHandler.ForgotPassword)
		passwordReset.Post("/reset-password", deps.AuthHandler.ResetPassword)
	}

	// ============================================================================
	// 🔴 PROTECTED USER ROUTES
	// ============================================================================
	user := v1.Group("")
	user.Use(authMiddleware.RequireAuth())
	{
		// General authenticated rate limit: 100 requests per minute
		userRateLimit := rateLimiter.UserRateLimit(middleware.RateLimitConfig{
			RequestsPerWindow: 100,
			WindowDuration:    60 * time.Second,
		})
		user.Use(userRateLimit)

		// ─────────────────────────────────────────────────────────────────
		// PASSWORD & ACCOUNT MANAGEMENT
		// ─────────────────────────────────────────────────────────────────
		user.Post("/auth/change-password", deps.AuthHandler.ChangePassword)
		user.Post("/auth/logout", deps.AuthHandler.Logout)

		// ─────────────────────────────────────────────────────────────────
		// SESSION MANAGEMENT (Moderate limit: 20/min)
		// ─────────────────────────────────────────────────────────────────
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

		// ─────────────────────────────────────────────────────────────────
		// TWO-FACTOR AUTHENTICATION MANAGEMENT (Strict: 10/min)
		// ─────────────────────────────────────────────────────────────────
		twoFactor := user.Group("/auth/2fa")
		{
			// Strict rate limit for 2FA operations
			twoFactorRateLimit := rateLimiter.UserRateLimit(middleware.RateLimitConfig{
				RequestsPerWindow: 10, // ✅ Reasonable for management ops
				WindowDuration:    60 * time.Second,
			})
			twoFactor.Use(twoFactorRateLimit)

			// Get 2FA status
			twoFactor.Get("/status", deps.TwoFactorHandler.Get2FAStatus)

			// Enable 2FA (Step 1: Generate QR code)
			twoFactor.Post("/enable", deps.TwoFactorHandler.Enable2FA)

			// Confirm 2FA (Step 2: Verify TOTP code and get backup codes)
			twoFactor.Post("/confirm", deps.TwoFactorHandler.Confirm2FA)

			// Disable 2FA
			twoFactor.Post("/disable", deps.TwoFactorHandler.Disable2FA)

			// Regenerate backup codes
			twoFactor.Post("/regenerate-codes", deps.TwoFactorHandler.RegenerateBackupCodes)
		}
	}
}