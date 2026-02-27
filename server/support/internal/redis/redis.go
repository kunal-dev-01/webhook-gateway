package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	"support/internal/config"
	"shared/logger"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Client
// ============================================================================

// Client wraps redis.Client with config
type Client struct {
	*redis.Client
	cfg *config.RedisConfig
}

// ============================================================================
// Singleton — thread-safe
// ============================================================================

var (
	instance *Client
	once     sync.Once
	mu       sync.RWMutex
)

// ============================================================================
// Connect
// ============================================================================

// Connect establishes Redis connection with retry logic.
// Safe to call from multiple goroutines — only connects once.
func Connect(cfg *config.RedisConfig) (*Client, error) {
	opts := &redis.Options{
		Addr:            cfg.GetRedisAddr(),
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		MinIdleConns:    cfg.MinIdleConns,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
		ConnMaxLifetime: cfg.MaxConnAge,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	}

	// TLS
	if cfg.TLSEnabled {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	// Minimum 1 attempt
	attempts := cfg.RetryAttempts
	if attempts <= 0 {
		attempts = 1
	}

	var (
		client *redis.Client
		err    error
	)

	for attempt := 1; attempt <= attempts; attempt++ {
		logger.Info("Connecting to Redis...", "attempt", attempt, "addr", cfg.GetRedisAddr())

		c := redis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
		err = c.Ping(ctx).Err()
		cancel()

		if err == nil {
			client = c
			break
		}

		logger.Error("Failed to ping Redis", err, "attempt", attempt)
		_ = c.Close()

		if attempt < attempts {
			time.Sleep(cfg.RetryDelay)
		}
	}

	if client == nil {
		return nil, fmt.Errorf("redis connection failed after %d attempt(s): %w", attempts, err)
	}

	mu.Lock()
	instance = &Client{
		Client: client,
		cfg:    cfg,
	}
	mu.Unlock()

	logger.Info("Redis connected successfully",
		"addr", cfg.GetRedisAddr(),
		"pool_size", cfg.PoolSize,
		"db", cfg.DB,
	)

	return instance, nil
}

// ============================================================================
// Get — thread-safe singleton access
// ============================================================================

// Get returns the singleton instance. Panics if Connect() was not called first.
func Get() *Client {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		panic("Redis not initialized — call Connect() first")
	}
	return instance
}

// ============================================================================
// Close
// ============================================================================

func (r *Client) Close() error {
	logger.Info("Closing Redis connection...")

	mu.Lock()
	instance = nil
	mu.Unlock()

	return r.Client.Close()
}

// ============================================================================
// Health Check
// ============================================================================

// Ping checks if Redis is reachable
func (r *Client) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// ============================================================================
// TLS Helper
// ============================================================================

func buildTLSConfig(cfg *config.RedisConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS keypair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.TLSCA != "" {
		caCert, err := os.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}