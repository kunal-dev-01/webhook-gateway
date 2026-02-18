package main

import (
	"github.com/gofiber/fiber/v2"

	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
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
	users := protected.Group("/user")
	users.Get("/me", deps.UserHandler.GetMe)
	users.Patch("/me", deps.UserHandler.UpdateUser)
	users.Put("/me/avatar", deps.UserHandler.UpdateAvatar)
	users.Patch("/me/preferences", deps.UserHandler.UpdatePreferences)
}