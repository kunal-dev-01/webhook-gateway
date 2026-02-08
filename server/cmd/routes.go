package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"shared/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, deps *Dependencies) {
	// ========================================================================
	// Global Middleware
	// ========================================================================

	// Recover from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// Security headers
	app.Use(middleware.SecurityHeaders())

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.Frontend.FrontendUrl,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Request logging
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))

	// ========================================================================
	// Health Check
	// ========================================================================

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "auth-server",
			"time":    time.Now().UTC(),
		})
	})

	// ========================================================================
	// API v1 Routes
	// ========================================================================

	api := app.Group("/api/v1")

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(deps.Redis)

	// ========================================================================
	// Public Auth Routes (no auth required)
	// ========================================================================

	auth := api.Group("/auth")

	// Rate limit for sensitive endpoints
	authRateLimit := rateLimiter.Middleware(middleware.RateLimitConfig{
		RequestsPerWindow: 10,
		WindowDuration:    15 * time.Minute,
	})

	// Registration & Login
	auth.Post("/register", authRateLimit, deps.AuthHandler.Register)
	auth.Post("/login", authRateLimit, deps.AuthHandler.Login)
	auth.Post("/verify-2fa", authRateLimit, deps.AuthHandler.Verify2FA)

	// Email Verification
	auth.Post("/verify-email", deps.AuthHandler.VerifyEmail)
	auth.Post("/resend-verification", authRateLimit, deps.AuthHandler.ResendVerification)

	// Password Reset
	auth.Post("/forgot-password", authRateLimit, deps.AuthHandler.ForgotPassword)
	auth.Post("/reset-password", authRateLimit, deps.AuthHandler.ResetPassword)

	// Token Refresh (no auth, but requires refresh token in cookie)
	auth.Post("/refresh", deps.AuthHandler.RefreshToken)

	// ========================================================================
	// Protected Auth Routes (auth required)
	// ========================================================================

	authProtected := auth.Group("", authMiddleware.RequireAuth())

	// Get current user info
	authProtected.Get("/me", deps.AuthHandler.GetMe)

	// Password Management
	authProtected.Patch("/change-password", deps.AuthHandler.ChangePassword)

	// Logout
	authProtected.Post("/logout", deps.AuthHandler.Logout)
	authProtected.Post("/logout-all", deps.AuthHandler.LogoutAll)

	// Two-Factor Authentication
	authProtected.Post("/2fa/setup", deps.AuthHandler.Setup2FA)
	authProtected.Post("/2fa/enable", deps.AuthHandler.Enable2FA)
	authProtected.Post("/2fa/disable", deps.AuthHandler.Disable2FA)

	// Session Management
	authProtected.Get("/sessions", deps.AuthHandler.ListSessions)
	authProtected.Delete("/sessions/:sessionId", deps.AuthHandler.DeleteSession)

	// ========================================================================
	// 404 Handler
	// ========================================================================

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Route not found",
		})
	})
}