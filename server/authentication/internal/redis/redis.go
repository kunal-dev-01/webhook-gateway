package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"shared/config" 
	"shared/logger"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client
type Client struct {
	*redis.Client
	cfg *config.RedisConfig
}

var instance *Client

// Connect establishes Redis connection
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

	// TLS Configuration
	if cfg.TLSEnabled {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	var client *redis.Client
	var err error

	// Retry Logic
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		logger.Info("Connecting to Redis...", "attempt", attempt)

		client = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
		err = client.Ping(ctx).Err()
		cancel()

		if err == nil {
			break
		}

		logger.Error("Failed to ping Redis", err, "attempt", attempt)
		client.Close()
		time.Sleep(cfg.RetryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("redis connection failed after %d attempts: %w", cfg.RetryAttempts, err)
	}

	instance = &Client{
		Client: client,
		cfg:    cfg,
	}

	logger.Info("Redis connected successfully", "pool_size", cfg.PoolSize)
	return instance, nil
}

// Get returns the singleton instance
func Get() *Client {
	if instance == nil {
		panic("Redis not initialized. Call Connect() first")
	}
	return instance
}

// Close closes the connection
func (r *Client) Close() error {
	logger.Info("Closing Redis connection...")
	return r.Client.Close()
}

// ───────────────── Helper Functions ─────────────────

const incrExpiryLua = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`

// IncrWithExpiry executes atomic increment with expiration
func (r *Client) IncrWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	seconds := int(expiry.Seconds())
	if seconds == 0 {
		seconds = 1
	}
	return r.Eval(ctx, incrExpiryLua, []string{key}, seconds).Int64()
}

func buildTLSConfig(cfg *config.RedisConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.TLSCA != "" {
		caCert, err := os.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}