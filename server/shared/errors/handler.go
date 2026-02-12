package errors

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"shared/logger" // Tera logger package yahan import kar
)

// FiberHandler converts errors into a standard JSON response
func FiberHandler(c *fiber.Ctx, err error) error {
	// Default Response
	code := http.StatusInternalServerError
	message := "Internal Server Error"
	var fields map[string]string

	// 1. Check if it's OUR custom AppError
	var e *AppError
	if errors.As(err, &e) {
		code = e.Code
		message = e.Message
		fields = e.Fields

		// Agar 500 error hai, toh hi log karo (DB fail, Logic fail)
		if code >= 500 {
			logger.Error("Internal App Error", e.Err,
				"path", c.Path(),
				"method", c.Method(),
				"request_id", c.Locals("requestid"), // Agar middleware use kar rahe ho
			)
		}
	} else if fiberErr, ok := err.(*fiber.Error); ok {
		// 2. Check if it's a Fiber Error (e.g. Entity Too Large, Method Not Allowed)
		code = fiberErr.Code
		message = fiberErr.Message
	} else {
		// 3. Unknown Panic or Raw Error
		logger.Error("Unhandled Server Error", err,
			"path", c.Path(),
			"method", c.Method(),
		)
	}

	// Final JSON Response
	response := fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	}

	// Add fields if validation error
	if len(fields) > 0 {
		response["error"].(fiber.Map)["fields"] = fields
	}

	return c.Status(code).JSON(response)
}