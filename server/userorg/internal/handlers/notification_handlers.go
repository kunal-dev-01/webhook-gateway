package handlers

import (
	"userorg/internal/dto"
	"userorg/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"shared/errors"
	"shared/logger"
	"shared/middleware"
)

// ============================================================================
// Notification Handler
// ============================================================================

type NotificationHandler struct {
	notifService *services.NotificationService
	log          *logger.Logger
}

func NewNotificationHandler(notifService *services.NotificationService, log *logger.Logger) *NotificationHandler {
	return &NotificationHandler{notifService: notifService, log: log}
}

// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	var req dto.ListNotificationsRequest
	if err := c.QueryParser(&req); err != nil {
		return errors.BadRequest("Invalid query parameters")
	}

	resp, err := h.notifService.ListNotifications(c.Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

// PATCH /api/v1/notifications/:notification_id/read
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	notifID, err := uuid.Parse(c.Params("notification_id"))
	if err != nil {
		return errors.BadRequest("Invalid notification ID")
	}

	if err := h.notifService.MarkRead(c.Context(), notifID, userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification marked as read",
	})
}

// PATCH /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	if err := h.notifService.MarkAllRead(c.Context(), userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All notifications marked as read",
	})
}

// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	count, err := h.notifService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"unread_count": count},
	})
}

// DELETE /api/v1/notifications/:notification_id
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errors.Unauthorized("Unauthorized")
	}

	notifID, err := uuid.Parse(c.Params("notification_id"))
	if err != nil {
		return errors.BadRequest("Invalid notification ID")
	}

	if err := h.notifService.DeleteNotification(c.Context(), notifID, userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notification deleted successfully",
	})
}