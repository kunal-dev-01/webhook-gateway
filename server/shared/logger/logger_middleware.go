package logger

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

/* ───────────────── HTTP Middleware ───────────────── */

func (l *Logger) HTTPMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := getOrCreateRequestID(c)

		err := c.Next()

		l.logHTTPRequest(c, requestID, start, err)
		return err
	}
}

func getOrCreateRequestID(c *fiber.Ctx) string {
	if reqID := c.Locals("requestid"); reqID != nil {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	id := uuid.New().String()
	c.Locals("requestid", id)
	return id
}

func (l *Logger) logHTTPRequest(c *fiber.Ctx, requestID string, start time.Time, err error) {
	duration := time.Since(start)
	status := c.Response().StatusCode()

	level := l.selectLogLevel(status)
	event := level().
		Str("type", "http").
		Str("request_id", requestID).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("route", c.Route().Path).
		Int("status", status).
		Dur("duration_ms", duration).
		Str("ip", c.IP()).
		Str("user_agent", c.Get("User-Agent")).
		Int("bytes_sent", len(c.Response().Body()))

	// Add context
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(uuid.UUID); ok && id != uuid.Nil {
			event.Str("user_id", id.String())
		}
	}

	if err != nil {
		event.Err(err)
	}

	// Configurable query logging
	if l.config.LogQueryParams {
		if query := c.Request().URI().QueryString(); len(query) > 0 {
			event.Str("query", string(query))
		}
	}

	event.Msg("HTTP request completed")
}

func (l *Logger) selectLogLevel(status int) func() *zerolog.Event {
	switch {
	case status >= 500:
		return l.logger.Error
	case status >= 400:
		return l.logger.Warn
	default:
		return l.logger.Info
	}
}

/* ───────────────── Recovery Middleware ───────────────── */

func (l *Logger) RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				l.logPanic(c, r)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}

func (l *Logger) logPanic(c *fiber.Ctx, r any) {
	requestID := "unknown"
	if reqID := c.Locals("requestid"); reqID != nil {
		if id, ok := reqID.(string); ok {
			requestID = id
		}
	}

	event := l.logger.Error().
		Str("type", "panic").
		Str("request_id", requestID).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("ip", c.IP()).
		Interface("panic", r)

	if l.config.EnableStackTrace {
		event.Stack()
	}

	event.Msg("Panic recovered")
}