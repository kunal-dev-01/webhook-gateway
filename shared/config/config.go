package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	App      AppConfig
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

type AppConfig struct {
	Env  string
	Port string
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
			DBName:          getEnv("DB_NAME", "postgres"),
			SSLMode:         getEnv("DB_SSLMODE", "require"),
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
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
	}

	// Validate required fields
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

// Helper functions
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

	// Add SSL certificates if provided
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