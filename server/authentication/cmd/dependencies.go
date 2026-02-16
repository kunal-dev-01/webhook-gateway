package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"authentication/internal/db"
	"authentication/internal/handlers"
	"authentication/internal/redis"
	"authentication/internal/repository"
	"authentication/internal/services"
	"authentication/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"

	"shared/auth"
	"shared/config"
	sharedErrors "shared/errors"
	sharedLogger "shared/logger" // Renamed to avoid collision
)

// ============================================================================
// Dependencies Definition
// ============================================================================

type Dependencies struct {
	Config *config.Config
	DB     *db.DB
	Redis  *redis.Client
	Logger *sharedLogger.Logger

	// Repositories
	UserRepo         *repository.UserRepository
	OrganizationRepo *repository.OrganizationRepository
	PlanRepo         *repository.SubscriptionPlanRepository
	TokenRepo        *repository.TokenRepository

	// Managers
	JWTManager     *auth.TokenManager
	TOTPManager    *utils.TOTPManager
	SessionManager *utils.Manager

	// Services
	AuthService      *services.AuthService
	TwoFactorService *services.TwoFactorService
	EmailService     *utils.EmailService

	// Handlers
	AuthHandler      *handlers.AuthHandler
	TwoFactorHandler *handlers.TwoFactorHandler
}

// ============================================================================
// Initialization
// ============================================================================

func InitDependencies() (*Dependencies, error) {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	// 2. Initialize Logger
	loggerCfg := sharedLogger.Config{
		Environment:      cfg.App.Env,
		ServiceName:      "authentication-service",
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
		LogQueryParams:   false,
	}

	log := sharedLogger.New(loggerCfg)
	sharedLogger.Init(loggerCfg)

	log.Info("🚀 Initializing Authentication Service dependencies...")

	// 3. Database Connection
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	log.Info("✅ Database connected")

	// 4. Redis Connection
	redisClient, err := redis.Connect(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}
	log.Info("✅ Redis connected")

	// 5. Initialize Managers
	jwtManager := auth.NewTokenManager(&cfg.JWT)
	log.Info("✅ JWT Manager initialized")

	totpManager, err := utils.NewTOTPManager(&cfg.TOTP)
	if err != nil {
		return nil, fmt.Errorf("TOTP manager initialization error: %w", err)
	}
	log.Info("✅ TOTP Manager initialized")

	sessionManager := utils.NewManager(
		redisClient.Client,
		5,
		7*24*time.Hour,
	)
	log.Info("✅ Session Manager initialized")

	// 6. Email Service
	emailService, err := utils.NewEmailService(&cfg.AWS, *log.Logger())
	if err != nil {
		return nil, fmt.Errorf("email service initialization error: %w", err)
	}
	log.Info("✅ Email Service initialized")

	// 7. Initialize Repositories
	userRepo := repository.NewUserRepository(database)
	organizationRepo := repository.NewOrganizationRepository(database)
	planRepo := repository.NewSubscriptionPlanRepository(database)
	tokenRepo := repository.NewTokenRepository(redisClient.Client)
	log.Info("✅ Repositories initialized")

	// 8. Initialize Services
	authService := services.NewAuthService(
		userRepo,
		organizationRepo,
		planRepo,
		tokenRepo,
		sessionManager,
		emailService,
		jwtManager,
		totpManager,
		cfg,
		log,
	)
	log.Info("✅ Auth Service initialized")

	twoFactorService := services.NewTwoFactorService(
		userRepo,
		tokenRepo,
		totpManager,
		log,
	)
	log.Info("✅ 2FA Service initialized")

	// 9. Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	twoFactorHandler := handlers.NewTwoFactorHandler(twoFactorService, log)
	log.Info("✅ Handlers initialized")

	log.Info("🎉 All dependencies initialized successfully")

	return &Dependencies{
		Config:           cfg,
		DB:               database,
		Redis:            redisClient,
		Logger:           log,
		UserRepo:         userRepo,
		OrganizationRepo: organizationRepo,
		PlanRepo:         planRepo,
		TokenRepo:        tokenRepo,
		JWTManager:       jwtManager,
		TOTPManager:      totpManager,
		SessionManager:   sessionManager,
		EmailService:     emailService,
		AuthService:      authService,
		TwoFactorService: twoFactorService,
		AuthHandler:      authHandler,
		TwoFactorHandler: twoFactorHandler,
	}, nil
}

// Cleanup cleans up all resources
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

	// Logger close logic removed as zerolog doesn't need explicit close usually, 
	// unless using a buffered writer that needs flush. 
	// Assuming sharedLogger handles this internally if needed.

	return lastErr
}

// ============================================================================
// Main Entry Point
// ============================================================================

func main() {
	// Initialize Dependencies
	deps, err := InitDependencies()
	if err != nil {
		fmt.Printf("❌ Failed to initialize dependencies: %v\n", err)
		os.Exit(1)
	}

	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName:               "Upnixo Authentication Service",
		DisableStartupMessage: true,
		ErrorHandler:          sharedErrors.FiberHandler,
		BodyLimit:             deps.Config.App.RequestBodyLimit,
		ReadTimeout:           deps.Config.App.ReadTimeout,
		WriteTimeout:          deps.Config.App.WriteTimeout,
		IdleTimeout:           deps.Config.App.IdleTimeout,
	})

	// Global Middleware
	// Note: Ensure Logger.HTTPMiddleware() is compatible with Fiber v2
	// If it's a custom middleware in shared/logger, use it.
	// app.Use(deps.Logger.HTTPMiddleware()) 
	
	app.Use(deps.Logger.RecoveryMiddleware())
	app.Use(cors.New(getCORSConfig(deps.Config)))

	// Setup Routes (Assuming this function exists in another file in main package)
	// SetupRoutes(app, deps) 

	// ───────────────── Health Check Endpoint ─────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		// Use PingContext for timeout safety
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		dbHealthy := true
		// FIX: Use PingContext instead of Ping(ctx)
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
			deps.Logger.Error("Health check DB ping failed", err)
		}

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "authentication",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": dbHealthy,
			},
		})
	})

	// ───────────────── Detailed Health Check ─────────────────
	app.Get("/health/detailed", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		// Database check
		dbHealthy := true
		dbLatency := time.Duration(0)
		start := time.Now()
		
		// FIX: Use PingContext here too
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
			deps.Logger.Error("Detailed health check DB ping failed", err)
		} else {
			dbLatency = time.Since(start)
		}

		// Get pool stats
		stats := deps.DB.Stats()

		// Plans check
		plansHealthy := true
		plansCount := 0
		// Assuming ListAll takes context
		if plans, err := deps.PlanRepo.ListAll(ctx); err != nil {
			plansHealthy = false
			// Don't log full error for plans check unless critical
		} else {
			plansCount = len(plans)
		}

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "authentication",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": fiber.Map{
					"healthy": dbHealthy,
					"latency_ms": dbLatency.Milliseconds(),
					"pool": fiber.Map{
						"open":   stats.OpenConnections,
						"in_use": stats.InUse,
						"idle":   stats.Idle,
					},
				},
				"subscription_plans": fiber.Map{
					"healthy": plansHealthy,
					"count":   plansCount,
				},
			},
		})
	})

	// Start Server
	port := deps.Config.App.Port
	deps.Logger.Info("🚀 Starting server...", "port", port, "env", deps.Config.App.Env)

	// Graceful Shutdown Channel
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal("Server startup failed", err)
		}
	}()

	<-quit
	deps.Logger.Info("🛑 Shutting down server...")

	// Shutdown context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		deps.Logger.Error("Server forced to shutdown", err)
	}

	// Cleanup dependencies
	if err := deps.Cleanup(); err != nil {
		deps.Logger.Error("Cleanup finished with errors", err)
	}

	deps.Logger.Info("✅ Server stopped gracefully")
}

// ============================================================================
// Helper Functions
// ============================================================================

func getCORSConfig(cfg *config.Config) cors.Config {
	origins := getAllowedOrigins(cfg.App.AllowedOrigins)

	if cfg.App.IsDevelopment() {
		if len(cfg.App.AllowedOrigins) == 0 || cfg.App.AllowedOrigins[0] == "*" {
			// Added common local dev ports
			origins = "http://localhost:3000,http://localhost:4321,http://localhost:5173,http://127.0.0.1:3000"
		}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS", // Added PATCH
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           3600,
	}
}

func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
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
	case "panic":
		return zerolog.PanicLevel
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