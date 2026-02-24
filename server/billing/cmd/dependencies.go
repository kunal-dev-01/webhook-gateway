package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"shared/config"
	"billing/internal/db"
	"billing/internal/handlers"
	"billing/internal/lemonsqueezy"
	"billing/internal/redis"
	"billing/internal/repository"
	"billing/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"

	"shared/auth"
	sharedErrors "shared/errors"
	"shared/logger"
)

// ============================================================================
// DEPENDENCIES
// ============================================================================

type Dependencies struct {
	Config *config.Config
	DB     *db.DB
	Redis  *redis.Client
	Logger *logger.Logger

	// Repositories
	OrgRepo          *repository.OrgRepository
	SubRepo          *repository.SubscriptionRepository
	PaymentRepo      *repository.PaymentRepository
	PlanRepo         *repository.PlanRepository
	WebhookEventRepo *repository.WebhookEventRepository

	// Managers
	JWTManager *auth.TokenManager
	LSClient   *lemonsqueezy.Client

	// Services
	PlanService    *services.PlanService
	BillingService *services.BillingService
	WebhookService *services.WebhookService

	// Handlers
	PlanHandler    *handlers.PlanHandler
	BillingHandler *handlers.BillingHandler
	WebhookHandler *handlers.WebhookHandler
}

// ============================================================================
// INIT
// ============================================================================

func InitDependencies() (*Dependencies, error) {
	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	// 2. Logger
	loggerCfg := logger.Config{
		Environment:   cfg.App.Env,
		ServiceName:   "billing-service",
		Level:         parseLogLevel(cfg.Logger.Level),
		EnableConsole: cfg.Logger.EnableConsole,
		EnableFile:    cfg.Logger.EnableFile,
		FilePath:      cfg.Logger.FilePath,
		MaxSize:       cfg.Logger.MaxSize,
		MaxBackups:    cfg.Logger.MaxBackups,
		MaxAge:        cfg.Logger.MaxAge,
		EnableCaller:  cfg.Logger.EnableCaller,
	}

	log := logger.New(loggerCfg)
	logger.Init(loggerCfg)

	log.Info("🚀 Initializing Billing Service...")

	// 3. Database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	log.Info("✅ Database connected")

	// 4. Redis
	redisClient, err := redis.Connect(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}
	log.Info("✅ Redis connected")

	// 5. JWT Manager (verify-only in billing — no token creation)
	jwtManager := auth.NewTokenManager(&cfg.JWT)
	log.Info("✅ JWT Manager initialized")

	// 6. LemonSqueezy Client
	lsClient := lemonsqueezy.NewClient(&cfg.LemonSqueezy)
	log.Info("✅ LemonSqueezy client initialized")

	// 7. Repositories
	orgRepo          := repository.NewOrgRepository(database)
	subRepo          := repository.NewSubscriptionRepository(database)
	paymentRepo      := repository.NewPaymentRepository(database)
	planRepo         := repository.NewPlanRepository(database)
	webhookEventRepo := repository.NewWebhookEventRepository(database)
	log.Info("✅ Repositories initialized")

	// 8. Services
	planService    := services.NewPlanService(planRepo, log)
	billingService := services.NewBillingService(orgRepo, subRepo, paymentRepo, planRepo, lsClient, log)
	webhookService := services.NewWebhookService(orgRepo, subRepo, paymentRepo, planRepo, webhookEventRepo, lsClient, log)
	log.Info("✅ Services initialized")

	// 9. Handlers
	planHandler    := handlers.NewPlanHandler(planService, log)
	billingHandler := handlers.NewBillingHandler(billingService, log)
	webhookHandler := handlers.NewWebhookHandler(webhookService, log)
	log.Info("✅ Handlers initialized")

	log.Info("🎉 All dependencies initialized successfully")

	return &Dependencies{
		Config: cfg,
		DB:     database,
		Redis:  redisClient,
		Logger: log,

		OrgRepo:          orgRepo,
		SubRepo:          subRepo,
		PaymentRepo:      paymentRepo,
		PlanRepo:         planRepo,
		WebhookEventRepo: webhookEventRepo,

		JWTManager: jwtManager,
		LSClient:   lsClient,

		PlanService:    planService,
		BillingService: billingService,
		WebhookService: webhookService,

		PlanHandler:    planHandler,
		BillingHandler: billingHandler,
		WebhookHandler: webhookHandler,
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

	return lastErr
}

// ============================================================================
// MAIN
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
		AppName:               "Upnixo Billing Service",
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

	// ── Health Checks ─────────────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbHealthy := deps.DB.PingContext(ctx) == nil

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "billing",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": dbHealthy,
			},
		})
	})

	app.Get("/health/detailed", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbHealthy := true
		start := time.Now()
		dbLatency := time.Duration(0)
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
		} else {
			dbLatency = time.Since(start)
		}

		stats := deps.DB.Stats()

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "billing",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": fiber.Map{
					"healthy": dbHealthy,
					"latency": dbLatency.Milliseconds(),
					"pool": fiber.Map{
						"open":   stats.OpenConnections,
						"in_use": stats.InUse,
						"idle":   stats.Idle,
					},
				},
			},
		})
	})

	port := deps.Config.App.Port
	deps.Logger.Info("🚀 Starting Billing Service...", "port", port, "env", deps.Config.App.Env)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal("Server startup failed", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	deps.Logger.Info("🛑 Shutting down billing service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		deps.Logger.Error("Server forced to shutdown", err)
	}

	deps.Logger.Info("✅ Billing service stopped gracefully")
}

// ============================================================================
// HELPERS
// ============================================================================

func getCORSConfig(cfg *config.Config) cors.Config {
	origins := strings.Join(cfg.App.AllowedOrigins, ",")
	if cfg.App.IsDevelopment() && (len(cfg.App.AllowedOrigins) == 0 || cfg.App.AllowedOrigins[0] == "*") {
		origins = "http://localhost:3000,http://localhost:4321,http://localhost:5173"
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
	default:
		return zerolog.InfoLevel
	}
}