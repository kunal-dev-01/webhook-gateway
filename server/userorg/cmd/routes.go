package main

import (
	"github.com/gofiber/fiber/v2"

	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	// Auth middleware uses JWTManager + Redis from auth service
	// UserOrg service validates JWT only (stateless) — no Redis session check needed
	// Pass a nil redis client and handle via a lighter JWT-only middleware
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	v1 := app.Group("/api/v1")

	// ============================================================================
	// PROTECTED ROUTES
	// ============================================================================

	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.ByUser(middleware.GeneralRateLimit()))

	// User routes
	users := protected.Group("/users")
	users.Get("/me", deps.UserHandler.GetMe)
	users.Patch("/me", deps.UserHandler.UpdateUser)
	users.Patch("/me/preferences", deps.UserHandler.UpdatePreferences)
}