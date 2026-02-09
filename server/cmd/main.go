package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize dependencies
	deps, err := InitDependencies()
	if err != nil {
		log.Fatalf("❌ Failed to initialize dependencies: %v", err)
	}
	defer func() {
		if err := deps.Cleanup(); err != nil {
			deps.Logger.Error().Err(err).Msg("Failed to cleanup resources")
		}
	}()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Upnixo Auth Server",
		DisableStartupMessage: true,
		ErrorHandler:          customErrorHandler,
		Prefork:               false, 
		CaseSensitive:         true,
		StrictRouting:         false,
		ServerHeader:          "", // Hide server info
		BodyLimit:             4 * 1024 * 1024, // 4MB
	})

	// Setup routes
	SetupRoutes(app, deps)

	// Start server
	port := deps.Config.App.Port
	deps.Logger.Info().
		Str("port", port).
		Str("env", deps.Config.App.Env).
		Msg("🚀 Starting server...")

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	deps.Logger.Info().
		Str("url", fmt.Sprintf("http://localhost:%s", port)).
		Msg("✅ Server is running")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	deps.Logger.Info().Msg("🛑 Shutting down server...")

	if err := app.Shutdown(); err != nil {
		deps.Logger.Error().Err(err).Msg("Server shutdown failed")
	}

	deps.Logger.Info().Msg("✅ Server stopped gracefully")
}

// customErrorHandler handles fiber errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}