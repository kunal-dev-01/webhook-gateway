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

	// ── User routes ───────────────────────────────────────────────────────────
	users := protected.Group("/user")
	users.Get("/me", deps.UserHandler.GetMe)
	users.Patch("/me", deps.UserHandler.UpdateUser)
	users.Put("/me/avatar", deps.UserHandler.UpdateAvatar)
	users.Patch("/me/preferences", deps.UserHandler.UpdatePreferences)

	// ── Organization routes ───────────────────────────────────────────────────
	orgs := protected.Group("/user/organizations")
	orgs.Get("/", deps.OrgHandler.ListMyOrganizations)
	orgs.Get("/:org_id", deps.OrgHandler.GetOrganization)
	orgs.Patch("/:org_id", deps.OrgHandler.UpdateName)

	// Members
	orgs.Get("/:org_id/members", deps.OrgHandler.ListMembers)
	orgs.Patch("/:org_id/members/:user_id/role", deps.OrgHandler.UpdateMemberRole)
	orgs.Delete("/:org_id/members/:user_id", deps.OrgHandler.RemoveMember)

	// Invitations (org-scoped)
	orgs.Post("/:org_id/invitations", deps.InvitationHandler.InviteMember)
	orgs.Get("/:org_id/invitations", deps.InvitationHandler.ListInvitations)
	orgs.Delete("/:org_id/invitations/:invitation_id", deps.InvitationHandler.RevokeInvitation)

	// ── Invitation accept (not org-scoped — user accepts via token) ───────────
	invitations := protected.Group("/invitations")
	invitations.Post("/accept", deps.InvitationHandler.AcceptInvitation)

	// ── Notification routes ───────────────────────────────────────────────────
	notifications := protected.Group("/notifications")
	notifications.Get("/", deps.NotificationHandler.ListNotifications)
	notifications.Get("/unread-count", deps.NotificationHandler.GetUnreadCount)
	notifications.Patch("/read-all", deps.NotificationHandler.MarkAllRead)
	notifications.Patch("/:notification_id/read", deps.NotificationHandler.MarkRead)
	notifications.Delete("/:notification_id", deps.NotificationHandler.DeleteNotification)
}