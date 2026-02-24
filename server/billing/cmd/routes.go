package main

import (
	"github.com/gofiber/fiber/v2"

	"shared/middleware"
)

func SetupRoutes(app *fiber.App, deps *Dependencies) {
	authMiddleware := middleware.NewAuthMiddleware(deps.JWTManager, deps.Redis.Client)
	rateLimiter := middleware.NewRateLimiter(deps.Redis.Client)

	// ============================================================================
	// PUBLIC ROUTES
	// ============================================================================

	// Pricing page — no auth needed
	v1 := app.Group("/api/v1")
	plans := v1.Group("/billing/plans")
	plans.Get("/", deps.PlanHandler.ListPlans)
	plans.Get("/:tier/:billing_period", deps.PlanHandler.GetPlan)

	// ============================================================================
	// WEBHOOK ROUTES — No auth, raw body needed
	// ============================================================================

	// IMPORTANT: BodyLimit here must be large enough for webhooks
	webhooks := app.Group("/webhooks")
	webhooks.Use(rateLimiter.ByIP(middleware.StrictRateLimit()))
	webhooks.Post("/lemonsqueezy", deps.WebhookHandler.HandleLemonSqueezy)

	// ============================================================================
	// PROTECTED ROUTES
	// ============================================================================

	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.ByUser(middleware.GeneralRateLimit()))

	// ── Billing (per-org, owner only for mutations) ─────────────────────────────
	billing := protected.Group("/billing/:org_id")
	billing.Use(authMiddleware.RequireCanManageBilling()) // owner/admin only

	// Checkout
	billing.Post("/checkout", deps.BillingHandler.CreateCheckout)

	// Subscription
	billing.Get("/subscription", deps.BillingHandler.GetSubscription)
	billing.Delete("/subscription", deps.BillingHandler.CancelSubscription)
	billing.Post("/subscription/resume", deps.BillingHandler.ResumeSubscription)
	billing.Patch("/subscription/plan", deps.BillingHandler.ChangePlan)

	// Payments
	billing.Get("/payments", deps.BillingHandler.ListPayments)

	// Usage (member+ can see usage)
	usage := protected.Group("/billing/:org_id")
	usage.Get("/usage", deps.BillingHandler.GetUsage)
}