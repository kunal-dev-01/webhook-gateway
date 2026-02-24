package handlers

import (
	"billing/internal/services"
	"shared/errors"
	"shared/logger"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// WEBHOOK HANDLER
// ============================================================================

type WebhookHandler struct {
	webhookService *services.WebhookService
	log            *logger.Logger
}

func NewWebhookHandler(webhookService *services.WebhookService, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService, log: log}
}

// HandleLemonSqueezy — POST /webhooks/lemonsqueezy
// NOTE: This route must NOT go through auth middleware
// NOTE: Must read raw body BEFORE any body parser
func (h *WebhookHandler) HandleLemonSqueezy(c *fiber.Ctx) error {
	// Read raw body for HMAC verification
	rawBody := c.Body()
	if len(rawBody) == 0 {
		return errors.BadRequest("Empty webhook body")
	}

	// Get signature from header
	signature := c.Get("X-Signature")
	if signature == "" {
		h.log.Security("webhook_missing_signature", "high", "ip", c.IP())
		return errors.Unauthorized("Missing X-Signature header")
	}

	// Process webhook
	if err := h.webhookService.ProcessWebhook(c.Context(), rawBody, signature); err != nil {
		// Check if it's auth error (invalid signature) — return 401
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == 401 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid signature",
			})
		}
		// Other errors — return 500 but LemonSqueezy will retry
		h.log.Error("Webhook processing error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Webhook processing failed",
		})
	}

	// Always return 200 to LemonSqueezy to prevent retries on business logic errors
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Webhook received",
	})
}