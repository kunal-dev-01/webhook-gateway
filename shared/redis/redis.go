package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	"shared/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Lua Script for Atomic Increment + Expiry
const incrExpiryLua = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`

// Client wraps redis.Client with additional functionality
type Client struct {
	*redis.Client
	Config *config.RedisConfig
	logger zerolog.Logger
}

var instance *Client

// Connect establishes Redis connection with retry logic
func Connect(cfg *config.RedisConfig, logger zerolog.Logger) (*Client, error) {
	log := logger.With().Str("component", "redis").Logger()

	opts := &redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,

		// Connection Pool Tuning
		MaxRetries:      cfg.MaxRetries,
		MinIdleConns:    cfg.MinIdleConns,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		MaxIdleConns:    cfg.MinIdleConns * 2,
		ConnMaxIdleTime: cfg.IdleTimeout,
		ConnMaxLifetime: cfg.MaxConnAge,

		// Timeouts
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	if cfg.TLSEnabled {
		tlsConfig, err := buildRedisTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	var client *redis.Client
	var err error

	// Retry logic with backoff
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		log.Info().Int("attempt", attempt).Int("max", cfg.RetryAttempts).Msg("connecting to redis")

		client = redis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
		err = client.Ping(ctx).Err()
		cancel()

		if err == nil {
			break
		}

		log.Error().Err(err).Msg("failed to ping redis")
		client.Close()

		if attempt < cfg.RetryAttempts {
			time.Sleep(cfg.RetryDelay)
			continue
		}
		return nil, fmt.Errorf("redis connection failed after %d attempts: %w", cfg.RetryAttempts, err)
	}

	instance = &Client{
		Client: client,
		Config: cfg,
		logger: log,
	}

	log.Info().Msg("✅ redis connected successfully")
	logRedisStats(instance)

	return instance, nil
}

// GetInstance returns the singleton Redis instance
func GetInstance() *Client {
	if instance == nil {
		panic("redis not initialized. call Connect() first")
	}
	return instance
}

// Close gracefully closes Redis connection
func (r *Client) Close() error {
	if r.Client != nil {
		r.logger.Info().Msg("🔌 closing redis connection")
		return r.Client.Close()
	}
	return nil
}

// IsHealthy checks if redis is reachable
func (r *Client) IsHealthy(ctx context.Context) bool {
	return r.Ping(ctx).Err() == nil
}

// ============================================================================
// Atomic Helpers (Optimized)
// ============================================================================

// IncrWithExpiry atomically increments and sets expiry using Lua.
// This is race-condition safe and prevents keys without TTL.
func (r *Client) IncrWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	// Convert duration to seconds for Redis
	seconds := int(expiry.Seconds())
	if seconds == 0 {
		seconds = 1 // Min 1s expiry
	}

	// Execution: Eval(script, keys, args)
	val, err := r.Eval(ctx, incrExpiryLua, []string{key}, seconds).Int64()
	if err != nil {
		return 0, fmt.Errorf("lua incr error: %w", err)
	}

	return val, nil
}

// ============================================================================
// Internal Helpers
// ============================================================================

func buildRedisTLSConfig(cfg *config.RedisConfig) (*tls.Config, error) {
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
			return nil, errors.New("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

func logRedisStats(r *Client) {
	stats := r.PoolStats()
	r.logger.Debug().
		Uint32("pool_size", stats.TotalConns).
		Uint32("idle_conns", stats.IdleConns).
		Uint32("stale_conns", stats.StaleConns).
		Uint32("hits", stats.Hits).
		Msg("📊 redis connection pool stats")
}