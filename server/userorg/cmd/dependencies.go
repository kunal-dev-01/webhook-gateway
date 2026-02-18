package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"userorg/internal/db"
	"userorg/internal/handlers"
	"userorg/internal/redis"
	"userorg/internal/repository"
	"userorg/internal/services"
	"userorg/internal/utils"

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
	UserRepo *repository.UserRepository
	OrgRepo  *repository.OrganizationRepository

	// Managers
	JWTManager *auth.TokenManager

	// Utils
	R2          *utils.R2Client
	AvatarQueue *utils.AvatarUploadQueue

	// Services
	UserService *services.UserService
	OrgService  *services.OrganizationService

	// Handlers
	UserHandler *handlers.UserHandler
	OrgHandler  *handlers.OrganizationHandler

	// Worker lifecycle
	WorkerStop context.CancelFunc
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
		ServiceName:      "userorg-service",
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

	log.Info("🚀 Initializing UserOrg Service dependencies...")

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

	// 6. Initialize R2 Client
	r2Client, err := utils.NewR2Client(&cfg.R2)
	if err != nil {
		return nil, fmt.Errorf("R2 client initialization error: %w", err)
	}
	log.Info("✅ R2 Client initialized")

	// 7. Initialize Avatar Upload Queue
	avatarQueue := utils.NewAvatarUploadQueue(redisClient.Client)
	log.Info("✅ Avatar Upload Queue initialized")

	// 8. Initialize Repositories
	userRepo := repository.NewUserRepository(database)
	orgRepo  := repository.NewOrganizationRepository(database)
	log.Info("✅ Repositories initialized")

	// 9. Initialize Services
	userService := services.NewUserService(userRepo, r2Client, avatarQueue, log)
	log.Info("✅ User Service initialized")

	orgService := services.NewOrganizationService(orgRepo, log)
	log.Info("✅ Organization Service initialized")

	// 10. Initialize Handlers
	userHandler := handlers.NewUserHandler(userService, log)
	orgHandler  := handlers.NewOrganizationHandler(orgService, log)   
	log.Info("✅ Handlers initialized")



	log.Info("🎉 All dependencies initialized successfully")

	return &Dependencies{
		Config:      cfg,
		DB:          database,
		Redis:       redisClient,
		Logger:      log,

		UserRepo:    userRepo,
		OrgRepo:     orgRepo,

		JWTManager:  jwtManager,
		R2:          r2Client,
		AvatarQueue: avatarQueue,

		UserService: userService,
		OrgService:  orgService,

		UserHandler: userHandler,
		OrgHandler:  orgHandler,
	}, nil
}

// Cleanup cleans up all resources
func (d *Dependencies) Cleanup() error {
	d.Logger.Info("🔌 Cleaning up resources...")
	var lastErr error

	if d.WorkerStop != nil {
		d.WorkerStop()
		d.Logger.Info("✅ Avatar worker stopped")
	}

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

	// Start background avatar upload worker
	workerCtx, workerStop := context.WithCancel(context.Background())
	deps.WorkerStop = workerStop
	utils.StartAvatarWorker(workerCtx, deps.AvatarQueue, deps.R2, deps.UserRepo.UpdateAvatarURL)

	app := fiber.New(fiber.Config{
		AppName:               "Upnixo UserOrg Service",
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
			"service":   "userorg",
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
		dbLatency := time.Duration(0)
		start := time.Now()
		if err := deps.DB.PingContext(ctx); err != nil {
			dbHealthy = false
		} else {
			dbLatency = time.Since(start)
		}

		stats := deps.DB.Stats()

		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "userorg",
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
	deps.Logger.Info("🚀 Starting server...", "port", port, "env", deps.Config.App.Env)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal("Server startup failed", err)
		}
	}()

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
			AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
			AllowCredentials: true,
			MaxAge:           3600,
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