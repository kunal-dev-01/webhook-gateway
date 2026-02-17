package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	v1 := app.Group("/api/v1")

	// ============================================================================
	// PUBLIC AUTH ROUTES
	// ============================================================================
	auth := v1.Group("/auth")

	// Registration & Email (10 req/min per IP)
	auth.Post("/register", rateLimiter.ByIP(middleware.RateLimitConfig{
		RequestsPerWindow: 10,
		WindowDuration:    60 * time.Second,
	}), deps.RegisterHandler.Register)

	auth.Post("/verify-email", rateLimiter.ByIP(middleware.RateLimitConfig{
		RequestsPerWindow: 10,
		WindowDuration:    60 * time.Second,
	}), deps.RegisterHandler.VerifyEmail)

	auth.Post("/resend-verification", rateLimiter.ByIP(middleware.StrictRateLimit()), deps.RegisterHandler.ResendVerificationEmail)

	// Login — strict (5 req/min per IP)
	auth.Post("/login", rateLimiter.ByIP(middleware.StrictRateLimit()), deps.LoginHandler.Login)

	// 2FA verify during login — strict (5 req/min per IP)
	auth.Post("/2fa/verify", rateLimiter.ByIP(middleware.StrictRateLimit()), deps.LoginHandler.Verify2FA)

	// Token refresh — moderate (20 req/min per IP)
	auth.Post("/refresh", rateLimiter.ByIP(middleware.RateLimitConfig{
		RequestsPerWindow: 20,
		WindowDuration:    60 * time.Second,
	}), deps.SessionHandler.RefreshToken)

	// Password reset — very strict (3 req/min per IP)
	auth.Post("/forgot-password", rateLimiter.ByIP(middleware.RateLimitConfig{
		RequestsPerWindow: 3,
		WindowDuration:    60 * time.Second,
	}), deps.PasswordHandler.ForgotPassword)

	auth.Post("/reset-password", rateLimiter.ByIP(middleware.RateLimitConfig{
		RequestsPerWindow: 3,
		WindowDuration:    60 * time.Second,
	}), deps.PasswordHandler.ResetPassword)

	// ============================================================================
	// PROTECTED ROUTES
	// ============================================================================
	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.ByUser(middleware.GeneralRateLimit())) // 100 req/min per user

	// Account management
	protected.Post("/auth/change-password", deps.PasswordHandler.ChangePassword)
	protected.Post("/auth/logout", deps.SessionHandler.Logout)

	// Sessions — moderate (20 req/min per user)
	sessions := protected.Group("/auth/sessions")
	sessions.Use(rateLimiter.ByUser(middleware.RateLimitConfig{
		RequestsPerWindow: 20,
		WindowDuration:    60 * time.Second,
	}))
	sessions.Get("/", deps.SessionHandler.GetActiveSessions)
	sessions.Delete("/:session_id", deps.SessionHandler.RevokeSession)

	// 2FA management — strict (10 req/min per user)
	twoFactor := protected.Group("/auth/2fa")
	twoFactor.Use(rateLimiter.ByUser(middleware.RateLimitConfig{
		RequestsPerWindow: 10,
		WindowDuration:    60 * time.Second,
	}))
	twoFactor.Get("/status", deps.TwoFactorHandler.Get2FAStatus)
	twoFactor.Post("/enable", deps.TwoFactorHandler.Enable2FA)
	twoFactor.Post("/confirm", deps.TwoFactorHandler.Confirm2FA)
	twoFactor.Post("/disable", deps.TwoFactorHandler.Disable2FA)
	twoFactor.Post("/regenerate-codes", deps.TwoFactorHandler.RegenerateBackupCodes)
}