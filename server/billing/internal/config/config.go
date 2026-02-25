package config

import (
	"slices"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	LemonSqueezy LemonSqueezyConfig
	TOTP     TOTPConfig
	Security SecurityConfig
	Frontend FrontendConfig
	AWS      AWSConfig
	R2       R2Config 
	App      AppConfig
	Logger   LoggerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	
	// SSL Configuration
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
	
	// Connection Pool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	
	// Timeouts
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
	
	// Advanced
	LogLevel      string
	AutoMigrate   bool
	RetryAttempts int
	RetryDelay    time.Duration
}

type RedisConfig struct {
	// Connection
	Host     string
	Port     int
	Password string
	DB       int
	
	// Connection Pool
	MaxRetries      int
	MinIdleConns    int
	PoolSize        int
	PoolTimeout     time.Duration
	MaxConnAge      time.Duration
	IdleTimeout     time.Duration
	IdleCheckFreq   time.Duration
	
	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	
	// TLS/SSL
	TLSEnabled bool
	TLSCert    string
	TLSKey     string
	TLSCA      string
	
	// Advanced
	EnableTracing      bool
	RetryAttempts      int
	RetryDelay         time.Duration
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type LemonSqueezyConfig struct {
	APIKey        string        // LEMONSQUEEZY_API_KEY
	StoreID       string        // LEMONSQUEEZY_STORE_ID
	WebhookSecret string        // LEMONSQUEEZY_WEBHOOK_SECRET
	BaseURL       string        // https://api.lemonsqueezy.com/v1
	Timeout       time.Duration // HTTP client timeout
}

type TOTPConfig struct {
	Issuer string
	AESKey string // 32 bytes for AES-256
}

type SecurityConfig struct {
	MaxLoginAttempts int
	LockoutDuration  time.Duration
	RateLimitMax     int
}

type FrontendConfig struct {
	FrontendUrl     string
	VerificationUrl string
	ResetUrl        string
}

type AWSConfig struct {
	AWSAccessKey   string
	AWSSecretKey   string
	AWSRegion      string
	AWSSQSQueueURL string
	AWSFromEmail   string
}

type AppConfig struct {
	Env  string
	Port string
	
	// Server Timeouts
	RequestBodyLimit int           // Max request body size in bytes
	ReadTimeout      time.Duration // Request read timeout
	WriteTimeout     time.Duration // Response write timeout
	IdleTimeout      time.Duration // Keep-alive timeout
	
	// Security
	TrustedProxies []string // Trusted proxy IPs (for X-Forwarded-For)
	AllowedOrigins []string // CORS allowed origins
	
	// Features
	EnableMetrics bool
	EnableHealth  bool
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string // e.g. "https://cdn.yourdomain.com"
}

type LoggerConfig struct {
	// Log Level
	Level string // "trace", "debug", "info", "warn", "error", "fatal", "panic"
	
	// Output
	EnableConsole bool
	EnableFile    bool
	FilePath      string
	
	// File Rotation
	MaxSize    int  // Maximum size in MB before rotation
	MaxBackups int  // Maximum number of old log files to retain
	MaxAge     int  // Maximum number of days to retain old log files
	Compress   bool // Compress old log files
	
	// Performance
	EnableSampling  bool
	SamplingInitial int // Log first N of each level
	SamplingThereafter int // Then 1 every N
	
	// Features
	EnableCaller     bool // Add file:line to logs
	EnableStackTrace bool // Add stack trace on error/fatal
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "upnixo"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			SSLCert:         getEnv("DB_SSLCERT", ""),
			SSLKey:          getEnv("DB_SSLKEY", ""),
			SSLRootCert:     getEnv("DB_SSLROOTCERT", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
			ConnectTimeout:  getEnvAsDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
			QueryTimeout:    getEnvAsDuration("DB_QUERY_TIMEOUT", 30*time.Second),
			LogLevel:        getEnv("DB_LOG_LEVEL", "info"),
			AutoMigrate:     getEnvAsBool("DB_AUTO_MIGRATE", false),
			RetryAttempts:   getEnvAsInt("DB_RETRY_ATTEMPTS", 3),
			RetryDelay:      getEnvAsDuration("DB_RETRY_DELAY", 5*time.Second),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			
			MaxRetries:    getEnvAsInt("REDIS_MAX_RETRIES", 3),
			MinIdleConns:  getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			PoolSize:      getEnvAsInt("REDIS_POOL_SIZE", 50),
			PoolTimeout:   getEnvAsDuration("REDIS_POOL_TIMEOUT", 5*time.Second),
			MaxConnAge:    getEnvAsDuration("REDIS_MAX_CONN_AGE", 30*time.Minute),
			IdleTimeout:   getEnvAsDuration("REDIS_IDLE_TIMEOUT", 10*time.Minute),
			IdleCheckFreq: getEnvAsDuration("REDIS_IDLE_CHECK_FREQ", 1*time.Minute),
			
			DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			
			TLSEnabled: getEnvAsBool("REDIS_TLS_ENABLED", false),
			TLSCert:    getEnv("REDIS_TLS_CERT", ""),
			TLSKey:     getEnv("REDIS_TLS_KEY", ""),
			TLSCA:      getEnv("REDIS_TLS_CA", ""),
			
			EnableTracing: getEnvAsBool("REDIS_ENABLE_TRACING", false),
			RetryAttempts: getEnvAsInt("REDIS_RETRY_ATTEMPTS", 3),
			RetryDelay:    getEnvAsDuration("REDIS_RETRY_DELAY", 1*time.Second),
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""),
			AccessTokenDuration:  getEnvAsDuration("ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			Issuer:               getEnv("JWT_ISSUER", "upnixo.com"),
		},
		LemonSqueezy: LemonSqueezyConfig{
			APIKey:        getEnv("LEMONSQUEEZY_API_KEY", ""),
			StoreID:       getEnv("LEMONSQUEEZY_STORE_ID", ""),
			WebhookSecret: getEnv("LEMONSQUEEZY_WEBHOOK_SECRET", ""),
			BaseURL:       getEnv("LEMONSQUEEZY_BASE_URL", "https://api.lemonsqueezy.com/v1"),
			Timeout:       getEnvAsDuration("LEMONSQUEEZY_TIMEOUT", 30*time.Second),
		},
		TOTP: TOTPConfig{
			Issuer: getEnv("TOTP_ISSUER", "Upnixo"),
			AESKey: getEnv("TOTP_AES_KEY", ""),
		},
		Security: SecurityConfig{
			MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:  getEnvAsDuration("LOCKOUT_DURATION", 30*time.Minute),
			RateLimitMax:     getEnvAsInt("RATE_LIMIT_MAX", 100),
		},
		Frontend: FrontendConfig{
			FrontendUrl:     getEnv("FRONTEND_URL", "http://localhost:4321"),
			VerificationUrl: getEnv("VERIFICATION_URL", "http://localhost:4321/auth/verify"),
			ResetUrl:        getEnv("RESET_URL", "http://localhost:4321/auth/reset"),
		},
		AWS: AWSConfig{
			AWSAccessKey:   getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretKey:   getEnv("AWS_SECRET_ACCESS_KEY", ""),
			AWSRegion:      getEnv("AWS_REGION", "eu-central-1"),
			AWSSQSQueueURL: getEnv("AWS_SQS_QUEUE_URL", ""),
			AWSFromEmail:   getEnv("AWS_FROM_EMAIL", ""),
		},
		R2: R2Config{
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			Bucket:          getEnv("R2_BUCKET", ""),
			PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		},
		App: AppConfig{
			Env:              getEnv("APP_ENV", "development"),
			Port:             getEnv("APP_PORT", "8080"),
			RequestBodyLimit: getEnvAsInt("REQUEST_BODY_LIMIT", 10*1024*1024), // 10MB default
			ReadTimeout:      getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:     getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:      getEnvAsDuration("IDLE_TIMEOUT", 120*time.Second),
			TrustedProxies:   getEnvAsSlice("TRUSTED_PROXIES", []string{"127.0.0.1"}),
			AllowedOrigins:   getEnvAsSlice("ALLOWED_ORIGINS", []string{"*"}),
			EnableMetrics:    getEnvAsBool("ENABLE_METRICS", true),
			EnableHealth:     getEnvAsBool("ENABLE_HEALTH", true),
		},
		Logger: LoggerConfig{
			Level:              getEnv("LOG_LEVEL", "info"),
			EnableConsole:      getEnvAsBool("LOG_ENABLE_CONSOLE", true),
			EnableFile:         getEnvAsBool("LOG_ENABLE_FILE", false),
			FilePath:           getEnv("LOG_FILE_PATH", "/var/log/app/app.log"),
			MaxSize:            getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxBackups:         getEnvAsInt("LOG_MAX_BACKUPS", 10),
			MaxAge:             getEnvAsInt("LOG_MAX_AGE", 30),
			Compress:           getEnvAsBool("LOG_COMPRESS", true),
			EnableSampling:     getEnvAsBool("LOG_ENABLE_SAMPLING", false),
			SamplingInitial:    getEnvAsInt("LOG_SAMPLING_INITIAL", 10),
			SamplingThereafter: getEnvAsInt("LOG_SAMPLING_THEREAFTER", 100),
			EnableCaller:       getEnvAsBool("LOG_ENABLE_CALLER", false),
			EnableStackTrace:   getEnvAsBool("LOG_ENABLE_STACKTRACE", true),
		},
	}

	// Validate required fields
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if cfg.App.Env == "production" && cfg.Redis.Password == "" {
		return nil, fmt.Errorf("REDIS_PASSWORD is required in production")
	}

	if cfg.Redis.Password != "" && len(cfg.Redis.Password) < 32 {
		return nil, fmt.Errorf("REDIS_PASSWORD must be at least 32 characters")
	}

	if cfg.LemonSqueezy.APIKey == "" {
		return nil, fmt.Errorf("LEMON_SQUEEZY_API_KEY is required")
	}

	if cfg.LemonSqueezy.StoreID == "" {
		return nil, fmt.Errorf("LEMON_SQUEEZY_STORE_ID is required")
	}

	if cfg.LemonSqueezy.WebhookSecret == "" {
		return nil, fmt.Errorf("LEMON_SQUEEZY_WEBHOOK_SECRET is required")
	}

	if cfg.LemonSqueezy.BaseURL == "" {
		return nil, fmt.Errorf("LEMON_SQUEEZY_BASE_URL is required")
	}

	if cfg.LemonSqueezy.Timeout == 0 {
		return nil, fmt.Errorf("LEMON_SQUEEZY_TIMEOUT is required")
	}

	// Validate log level
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, cfg.Logger.Level) {
		return nil, fmt.Errorf("LOG_LEVEL must be one of: %v", validLevels)
	}

	return cfg, nil
}

/* ───────────────── Helper Functions ───────────────── */

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func getEnvAsSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return fallback
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

/* ───────────────── DSN Builders ───────────────── */

// GetDSN builds PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
		int(c.ConnectTimeout.Seconds()),
	)

	if c.SSLCert != "" {
		dsn += fmt.Sprintf(" sslcert=%s", c.SSLCert)
	}
	if c.SSLKey != "" {
		dsn += fmt.Sprintf(" sslkey=%s", c.SSLKey)
	}
	if c.SSLRootCert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", c.SSLRootCert)
	}

	return dsn
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

/* ───────────────── Environment-Specific Helpers ───────────────── */

// IsDevelopment checks if running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction checks if running in production mode
func (c *AppConfig) IsProduction() bool {
	return c.Env == "production"
}

// IsStaging checks if running in staging mode
func (c *AppConfig) IsStaging() bool {
	return c.Env == "staging"
}
