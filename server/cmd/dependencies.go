package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/internal/handlers"
	"server/internal/repository"
	"server/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"shared/auth"
	"shared/config"
	"shared/models"
	redisClient "shared/redis"
	"shared/utils"
)

// ============================================================================
// Dependencies Definition
// ============================================================================

type Dependencies struct {
	Config *config.Config
	DB     *sqlx.DB
	Redis  *redisClient.Client
	Logger zerolog.Logger

	// Repositories
	UserRepo  *repository.UserRepository
	TokenRepo *repository.TokenRepository
	CompanyRepo *repository.CompanyRepository

	// Managers
	JWTManager  *auth.TokenManager // ✅ Field Name is JWTManager
	TOTPManager *auth.TOTPManager
	SessionMgr  *auth.SessionManager

	// Services
	AuthService  *services.AuthService
	UserService  *services.UserService
	EmailService *utils.EmailService

	// Handlers
	AuthHandler *handlers.AuthHandler
	UserHandler *handlers.UserHandler
}

// ============================================================================
// Initialization
// ============================================================================

func InitDependencies() (*Dependencies, error) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()

	logger.Info().Msg("🚀 Initializing dependencies...")

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	// 2. Database
	rawDB, err := models.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("db connection error: %w", err)
	}
	db := sqlx.NewDb(rawDB.DB, "postgres")
	logger.Info().Msg("✅ Database connected")

	// 3. Redis
	redis, err := redisClient.Connect(&cfg.Redis, logger)
	if err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}

	// 4. Managers
	jwtManager := auth.NewTokenManager(&cfg.JWT)
	totpManager, err := auth.NewTOTPManager(&cfg.TOTP)
	if err != nil {
		return nil, err
	}
	// ✅ FIXED: Pass redis client and config
	sessionMgr := auth.NewSessionManager(
		redis.Client, // *redis.Client
		auth.SessionConfig{
			MaxSessions: 5,              // Max 5 concurrent sessions per user
			DefaultTTL:  24 * time.Hour, // Session expires in 24 hours
		},
	) // Pass raw client

	// 5. External Services
	emailService, err := utils.NewEmailService(&cfg.AWS, logger)
	if err != nil {
		return nil, err
	}

	// 6. Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redis.Client) 
	companyRepo := repository.NewCompanyRepository(db)


	// 7. Services
	authService := services.NewAuthService(
		userRepo, companyRepo, tokenRepo, jwtManager, totpManager, sessionMgr, emailService, cfg, logger,
	)
	userService := services.NewUserService(userRepo, logger)

	// 8. Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	return &Dependencies{
		Config:       cfg,
		DB:           db,
		Redis:        redis,
		Logger:       logger,
		UserRepo:     userRepo,
		TokenRepo:    tokenRepo,
		CompanyRepo:  companyRepo,
		JWTManager:   jwtManager,
		TOTPManager:  totpManager,
		SessionMgr:   sessionMgr,
		EmailService: emailService,
		AuthService:  authService,
		UserService:  userService,
		AuthHandler:  authHandler,
		UserHandler:  userHandler,
	}, nil
}

// ✅ FIXED: Cleanup now returns error to match main.go usage
func (d *Dependencies) Cleanup() error {
	d.Logger.Info().Msg("🔌 Cleaning up resources...")
	var err error

	if d.Redis != nil {
		if e := d.Redis.Close(); e != nil {
			d.Logger.Error().Err(e).Msg("Error closing Redis")
			err = e
		} else {
			d.Logger.Info().Msg("✅ Redis closed")
		}
	}

	if d.DB != nil {
		if e := d.DB.Close(); e != nil {
			d.Logger.Error().Err(e).Msg("Error closing Database")
			err = e
		} else {
			d.Logger.Info().Msg("✅ Database closed")
		}
	}

	return err
}

// ============================================================================
// Main Entry Point
// ============================================================================

func main() {
	// Initialize
	deps, err := InitDependencies()
	if err != nil {
		// Use standard log here because deps might be nil
		fmt.Printf("❌ Failed to initialize dependencies: %v\n", err)
		os.Exit(1)
	}

	// Cleanup on exit
	defer func() {
		if err := deps.Cleanup(); err != nil {
			deps.Logger.Error().Err(err).Msg("Cleanup finished with errors")
		}
	}()

	// Fiber App
	app := fiber.New(fiber.Config{
		AppName:               "Upnixo Auth Server",
		DisableStartupMessage: true,
		ErrorHandler:          customErrorHandler,
		BodyLimit:             4 * 1024 * 1024, // 4MB
	})

	// Setup Routes
	SetupRoutes(app, deps)

	// Start Server
	port := deps.Config.App.Port
	deps.Logger.Info().Str("port", port).Msg("🚀 Starting server...")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			deps.Logger.Fatal().Err(err).Msg("Server startup failed")
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	deps.Logger.Info().Msg("🛑 Shutting down server...")
	if err := app.Shutdown(); err != nil {
		deps.Logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	deps.Logger.Info().Msg("✅ Server stopped gracefully")
}

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
