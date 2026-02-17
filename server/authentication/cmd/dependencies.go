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
	"shared/logger"
)

// ============================================================================
// Dependencies Definition
// ============================================================================

type Dependencies struct {
	Config *config.Config
	DB     *db.DB
	Redis  *redis.Client
	Logger *logger.Logger

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
	RegisterService  *services.RegisterService
	LoginService  *services.LoginService
	TwoFactorService  *services.TwoFactorService
	PasswordService  *services.PasswordService
	SessionService  *services.SessionService
	EmailService *utils.EmailService

	// Handlers
	RegisterHandler *handlers.RegisterHandler
	LoginHandler *handlers.LoginHandler
	TwoFactorHandler *handlers.TwoFactorHandler
	PasswordHandler *handlers.PasswordHandler
	SessionHandler *handlers.SessionHandler
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
	loggerCfg := logger.Config{
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
	
	log := logger.New(loggerCfg)
	logger.Init(loggerCfg)
	
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

	// 9. Initialize Services
	registerService := services.NewRegisterService(
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
	log.Info("✅ Services initialized")

	loginService := services.NewLoginService(
		userRepo,
		organizationRepo,
		tokenRepo,
		sessionManager,
		emailService,
		jwtManager,
		totpManager,
		cfg,
		log,
	)
	log.Info("✅ Services initialized")

	twoFactorService := services.NewTwoFactorService(
		userRepo,
		tokenRepo,
		totpManager,
		log,
	)
	log.Info("✅ 2FA Service initialized")

	passwordService := services.NewPasswordService(
		userRepo,
		tokenRepo,
		emailService,
		sessionManager,
		cfg,
		log,
	)
	log.Info("✅ Password Service initialized")

	sessionService := services.NewSessionService(
		userRepo,
		sessionManager,
		emailService,
		jwtManager,
		cfg,
		log,
	)
	log.Info("✅ Session Service initialized")

	// 10. Initialize Handlers
	registerHandler := handlers.NewRegisterHandler(registerService, log)
	loginHandler := handlers.NewLoginHandler(loginService, log)
	twoFactorHandler := handlers.NewTwoFactorHandler(twoFactorService, log)
	passwordHandler := handlers.NewPasswordHandler(passwordService, log)
	sessionHandler := handlers.NewSessionHandler(sessionService, log)
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
		RegisterService:  registerService,
		LoginService:      loginService,
		TwoFactorService: twoFactorService,
		PasswordService:   passwordService,
		SessionService:    sessionService,

		RegisterHandler:  registerHandler,
		LoginHandler:      loginHandler,
		TwoFactorHandler: twoFactorHandler,
		PasswordHandler:   passwordHandler,
		SessionHandler:    sessionHandler,
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

	if d.Logger != nil {
		if err := d.Logger.Close(); err != nil {
			lastErr = err
		}
	}

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

	// Cleanup on exit
	defer func() {
		if err := deps.Cleanup(); err != nil {
			deps.Logger.Error("Cleanup finished with errors", err)
		}
	}()

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
	app.Use(deps.Logger.HTTPMiddleware())
	app.Use(deps.Logger.RecoveryMiddleware())
	app.Use(cors.New(getCORSConfig(deps.Config)))

	// Setup Routes
	SetupRoutes(app, deps)

	// Health Check Endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Quick DB ping
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		dbHealthy := true
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
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

	// Detailed Health Check
	app.Get("/health/detailed", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Database check
		dbHealthy := true
		dbLatency := time.Duration(0)
		start := time.Now()
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
		} else {
			dbLatency = time.Since(start)
		}
		
		// Get pool stats
		stats := deps.DB.Stats()
		
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "authentication",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"checks": fiber.Map{
				"database": fiber.Map{
					"healthy":  dbHealthy,
					"latency":  dbLatency.Milliseconds(),
					"pool": fiber.Map{
						"open":    stats.OpenConnections,
						"in_use":  stats.InUse,
						"idle":    stats.Idle,
					},
				},
			},
		})
	})

	// Start Server
	port := deps.Config.App.Port
	deps.Logger.Info("🚀 Starting server...", "port", port, "env", deps.Config.App.Env)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal("Server startup failed", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	deps.Logger.Info("🛑 Shutting down server...")
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		deps.Logger.Error("Server forced to shutdown", err)
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
			origins = "http://localhost:3000,http://localhost:4321,http://localhost:5173"
		}
		return cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
			AllowCredentials: true,
			MaxAge:           3600,
		}
	}
	
	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
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
	
	var result strings.Builder
	for i, origin := range origins {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(origin)
	}
	return result.String()
}