package main

import (
	"fmt"
	"log"

	"server/internal/handlers"
	"server/internal/repository"
	"server/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"shared/auth"
	"shared/config"
	"shared/models"
	redisClient "shared/redis"
	"shared/utils"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	Config *config.Config
	DB     *sqlx.DB
	Redis  *redisClient.Client
	Logger zerolog.Logger

	// Repositories
	UserRepo  *repository.UserRepository
	TokenRepo *repository.TokenRepository

	// Services
	JWTManager   *auth.Manager
	TOTPManager  *auth.TOTPManager
	SessionMgr   *auth.SessionManager
	EmailService *utils.EmailService
	AuthService  *services.AuthService
	UserService  *services.UserService

	// Handlers
	AuthHandler *handlers.AuthHandler
	UserHandler *handlers.UserHandler
}

// InitDependencies initializes all application dependencies
func InitDependencies() (*Dependencies, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()}).
		With().
		Timestamp().
		Logger()

	logger.Info().Msg("🚀 Initializing dependencies...")

	// Connect to database
	db, err := models.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db.DB, "postgres")

	logger.Info().Msg("✅ Database connected")

	// Connect to Redis
	redis, err := redisClient.Connect(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info().Msg("✅ Redis connected")

	// Initialize JWT Manager
	jwtManager := auth.NewManager(&cfg.JWT)
	logger.Info().Msg("✅ JWT Manager initialized")

	// Initialize TOTP Manager
	totpManager, err := auth.NewTOTPManager(&cfg.TOTP)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TOTP manager: %w", err)
	}
	logger.Info().Msg("✅ TOTP Manager initialized")

	// Initialize Session Manager
	sessionMgr := auth.NewSessionManager(redis.Client)
	logger.Info().Msg("✅ Session Manager initialized")

	// Initialize Email Service
	emailService, err := utils.NewEmailService(&cfg.AWS, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}
	logger.Info().Msg("✅ Email Service initialized")

	// Initialize Repositories
	userRepo := repository.NewUserRepository(sqlxDB, redis)
	tokenRepo := repository.NewTokenRepository(redis)
	logger.Info().Msg("✅ Repositories initialized")

	// Initialize Services
	authService := services.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
		totpManager,
		sessionMgr,
		emailService,
		cfg,
	)
	logger.Info().Msg("✅ Services initialized")

	userService := services.NewUserService(userRepo)
	logger.Info().Msg("✅ Services initialized")

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	logger.Info().Msg("✅ Handlers initialized")

	userHandler := handlers.NewUserHandler(userService)
	logger.Info().Msg("✅ Handlers initialized")

	logger.Info().Msg("✅ All dependencies initialized successfully")

	return &Dependencies{
		Config:       cfg,
		DB:           sqlxDB,
		Redis:        redis,
		Logger:       logger,
		UserRepo:     userRepo,
		TokenRepo:    tokenRepo,
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

// Cleanup gracefully closes all connections
func (d *Dependencies) Cleanup() error {
	d.Logger.Info().Msg("🔌 Cleaning up resources...")

	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			d.Logger.Error().Err(err).Msg("Failed to close database")
			return err
		}
		d.Logger.Info().Msg("✅ Database closed")
	}

	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			d.Logger.Error().Err(err).Msg("Failed to close Redis")
			return err
		}
		d.Logger.Info().Msg("✅ Redis closed")
	}

	d.Logger.Info().Msg("✅ Cleanup completed")
	return nil
}