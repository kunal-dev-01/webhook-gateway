package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"support/internal/db"
	"support/internal/handlers"
	"support/internal/redis"
	"support/internal/repository"
	"support/internal/services"
	"support/internal/utils"
	"support/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"

	"shared/auth"
	sharedErrors "shared/errors"
	"shared/logger"
)

// ============================================================================
// Dependencies
// ============================================================================

type Dependencies struct {
	Config *config.Config
	DB     *db.DB
	Redis  *redis.Client
	Logger *logger.Logger

	// Repositories
	TicketRepo     *repository.TicketRepository
	MessageRepo    *repository.MessageRepository
	AttachmentRepo *repository.AttachmentRepository

	// Managers
	JWTManager *auth.TokenManager

	// Utils
	R2 *utils.R2Client

	// Services
	SupportService *services.SupportService

	// Handlers
	SupportHandler *handlers.SupportHandler
}

// ============================================================================
// Initialization
// ============================================================================

func InitDependencies() (*Dependencies, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	loggerCfg := logger.Config{
		Environment:      cfg.App.Env,
		ServiceName:      "support-service",
		Level:            parseLogLevel(cfg.Logger.Level),
		EnableConsole:    cfg.Logger.EnableConsole,
		EnableFile:       cfg.Logger.EnableFile,
		FilePath:         cfg.Logger.FilePath,
		MaxSize:          cfg.Logger.MaxSize,
		MaxBackups:       cfg.Logger.MaxBackups,
		MaxAge:           cfg.Logger.MaxAge,
		Compress:         cfg.Logger.Compress,
		EnableCaller:     cfg.Logger.EnableCaller,
		EnableStackTrace: cfg.Logger.EnableStackTrace,
	}

	log := logger.New(loggerCfg)
	logger.Init(loggerCfg)

	log.Info("🚀 Initializing Support Service dependencies...")

	database, err := db.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	log.Info("✅ Database connected")

	redisClient, err := redis.Connect(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}
	log.Info("✅ Redis connected")

	jwtManager, err := auth.NewTokenManager(auth.Config{
		Secret:               cfg.JWT.Secret,
		AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
		RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		Issuer:               cfg.JWT.Issuer,
	})
	if err != nil {
		return nil, fmt.Errorf("JWT manager init error: %w", err)
	}
	log.Info("✅ JWT Manager initialized")

	r2Client, err := utils.NewR2Client(&cfg.R2)
	if err != nil {
		return nil, fmt.Errorf("R2 client initialization error: %w", err)
	}
	log.Info("✅ R2 Client initialized")

	// Repositories
	ticketRepo     := repository.NewTicketRepository(database)
	messageRepo    := repository.NewMessageRepository(database)
	attachmentRepo := repository.NewAttachmentRepository(database)
	log.Info("✅ Repositories initialized")

	// Service
	supportService := services.NewSupportService(ticketRepo, messageRepo, attachmentRepo, r2Client, log)
	log.Info("✅ Support Service initialized")

	// Handler
	supportHandler := handlers.NewSupportHandler(supportService, log)
	log.Info("✅ Support Handler initialized")

	log.Info("🎉 All dependencies initialized successfully")

	return &Dependencies{
		Config:         cfg,
		DB:             database,
		Redis:          redisClient,
		Logger:         log,
		TicketRepo:     ticketRepo,
		MessageRepo:    messageRepo,
		AttachmentRepo: attachmentRepo,
		JWTManager:     jwtManager,
		R2:             r2Client,
		SupportService: supportService,
		SupportHandler: supportHandler,
	}, nil
}

func (d *Dependencies) Cleanup() error {
	d.Logger.Info("🔌 Cleaning up resources...")
	var lastErr error

	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			d.Logger.Error("Error closing Redis", err)
			lastErr = err
		} else {
			d.Logger.Info("✅ Redis closed")
		}
	}

	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			d.Logger.Error("Error closing Database", err)
			lastErr = err
		} else {
			d.Logger.Info("✅ Database closed")
		}
	}

	if d.Logger != nil {
		_ = d.Logger.Close()
	}

	return lastErr
}

// ============================================================================
// Main
// ============================================================================

func main() {
	deps, err := InitDependencies()
	if err != nil {
		fmt.Printf("❌ Failed to initialize dependencies: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := deps.Cleanup(); err != nil {
			deps.Logger.Error("Cleanup finished with errors", err)
		}
	}()

	app := fiber.New(fiber.Config{
		AppName:               "Support Service",
		DisableStartupMessage: true,
		ErrorHandler:          sharedErrors.FiberHandler,
		BodyLimit:             deps.Config.App.RequestBodyLimit,
		ReadTimeout:           deps.Config.App.ReadTimeout,
		WriteTimeout:          deps.Config.App.WriteTimeout,
		IdleTimeout:           deps.Config.App.IdleTimeout,
	})

	app.Use(deps.Logger.HTTPMiddleware())
	app.Use(deps.Logger.RecoveryMiddleware())
	app.Use(cors.New(getCORSConfig(deps.Config)))

	SetupRoutes(app, deps)

	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbHealthy := true
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
		}

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "support",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": dbHealthy,
			},
		})
	})

	port := deps.Config.App.Port
	deps.Logger.Info("🚀 Starting Support Service...", "port", port, "env", deps.Config.App.Env)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal("Server startup failed", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	deps.Logger.Info("🛑 Shutting down Support Service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		deps.Logger.Error("Server forced to shutdown", err)
	}

	deps.Logger.Info("✅ Support Service stopped gracefully")
}

// ============================================================================
// Helpers
// ============================================================================

func getCORSConfig(cfg *config.Config) cors.Config {
	origins := getAllowedOrigins(cfg.App.AllowedOrigins)

	if cfg.App.IsDevelopment() {
		if len(cfg.App.AllowedOrigins) == 0 || cfg.App.AllowedOrigins[0] == "*" {
			origins = "http://localhost:3000,http://localhost:4321,http://localhost:5173"
		}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func getAllowedOrigins(origins []string) string {
	if len(origins) == 0 {
		return "*"
	}
	return strings.Join(origins, ",")
}