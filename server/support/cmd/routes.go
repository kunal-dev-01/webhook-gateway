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
	// PROTECTED ROUTES — All support routes require authentication
	// ============================================================================

	protected := v1.Group("")
	protected.Use(authMiddleware.RequireAuth())
	protected.Use(rateLimiter.ByUser(middleware.GeneralRateLimit()))

	// ── Support Tickets ───────────────────────────────────────────────────────
	//
	// POST   /api/v1/support/tickets               → Create new ticket
	// GET    /api/v1/support/tickets               → List org's tickets (with filters)
	// GET    /api/v1/support/tickets/:ticket_id    → Get single ticket
	// POST   /api/v1/support/tickets/:ticket_id/reopen        → Reopen resolved ticket
	// POST   /api/v1/support/tickets/:ticket_id/satisfaction  → Submit rating after resolution
	//
	// ── Messages ─────────────────────────────────────────────────────────────
	//
	// POST   /api/v1/support/tickets/:ticket_id/messages              → Add message
	// GET    /api/v1/support/tickets/:ticket_id/messages              → List messages (user-visible only)
	//
	// ── Attachments ──────────────────────────────────────────────────────────
	//
	// POST   /api/v1/support/tickets/:ticket_id/messages/:message_id/attachments          → Upload file
	// GET    /api/v1/support/tickets/:ticket_id/attachments/:attachment_id/download       → Get presigned download URL

	tickets := protected.Group("/support/tickets")

	// Ticket CRUD
	tickets.Post("/", deps.SupportHandler.CreateTicket)
	tickets.Get("/", deps.SupportHandler.ListTickets)
	tickets.Get("/:ticket_id", deps.SupportHandler.GetTicket)

	// Ticket actions
	tickets.Post("/:ticket_id/reopen", deps.SupportHandler.ReopenTicket)
	tickets.Post("/:ticket_id/satisfaction", deps.SupportHandler.SubmitSatisfaction)

	// Messages
	tickets.Post("/:ticket_id/messages", deps.SupportHandler.AddMessage)
	tickets.Get("/:ticket_id/messages", deps.SupportHandler.ListMessages)

	// Attachments
	tickets.Post("/:ticket_id/messages/:message_id/attachments", deps.SupportHandler.UploadAttachment)
	tickets.Get("/:ticket_id/attachments/:attachment_id/download", deps.SupportHandler.GetAttachmentDownloadURL)
}