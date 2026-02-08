package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"shared/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client with additional functionality
type Client struct {
	*redis.Client
	Config *config.RedisConfig
}

var instance *Client

// Connect establishes Redis connection with retry logic
func Connect(cfg *config.RedisConfig) (*Client, error) {
	var client *redis.Client
	var err error

	// Build options
	opts := &redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,

		MaxRetries:      cfg.MaxRetries,
		MinIdleConns:    cfg.MinIdleConns,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		MaxIdleConns:    cfg.MinIdleConns * 2,
		ConnMaxIdleTime: cfg.IdleTimeout,
		ConnMaxLifetime: cfg.MaxConnAge,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

	}

	// TLS Configuration
	if cfg.TLSEnabled {
		tlsConfig, err := buildRedisTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	// Retry logic
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		log.Printf("🔄 Connecting to Redis (attempt %d/%d)...", attempt, cfg.RetryAttempts)

		client = redis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
		defer cancel()

		if err = client.Ping(ctx).Err(); err != nil {
			log.Printf("❌ Failed to ping Redis: %v", err)
			client.Close()
			if attempt < cfg.RetryAttempts {
				log.Printf("⏳ Retrying in %v...", cfg.RetryDelay)
				time.Sleep(cfg.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", cfg.RetryAttempts, err)
		}

		break
	}

	instance = &Client{
		Client: client,
		Config: cfg,
	}

	log.Println("✅ Redis connected successfully!")
	logRedisStats(instance)

	return instance, nil
}

// GetInstance returns the singleton Redis instance
func GetInstance() *Client {
	if instance == nil {
		log.Fatal("❌ Redis not initialized. Call Connect() first.")
	}
	return instance
}

// Close gracefully closes Redis connection
func (r *Client) Close() error {
	if r.Client != nil {
		log.Println("🔌 Closing Redis connection...")
		return r.Client.Close()
	}
	return nil
}

// ============================================================================
// Helper: Increment with Expiry (for rate limiting)
// ============================================================================

// IncrWithExpiry atomically increments and sets expiry if new key
func (r *Client) IncrWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	count, err := r.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	
	// Set expiry only on first increment
	if count == 1 {
		r.Expire(ctx, key, expiry)
	}
	
	return count, nil
}

// TTL gets time-to-live for a key
func (r *Client) TTL(ctx context.Context, key string) time.Duration {
	ttl, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		return 0
	}
	return ttl
}

// ============================================================================
// Helper Functions
// ============================================================================

func buildRedisTLSConfig(cfg *config.RedisConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.TLSCA != "" {
		caCert, err := os.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

func logRedisStats(r *Client) {
	stats := r.PoolStats()
	log.Printf("📊 Redis Connection Pool Stats:")
	log.Printf("   Pool Size: %d", r.Config.PoolSize)
	log.Printf("   Min Idle Connections: %d", r.Config.MinIdleConns)
	log.Printf("   Hits: %d", stats.Hits)
	log.Printf("   Misses: %d", stats.Misses)
	log.Printf("   Timeouts: %d", stats.Timeouts)
	log.Printf("   Total Conns: %d", stats.TotalConns)
	log.Printf("   Idle Conns: %d", stats.IdleConns)
	log.Printf("   Stale Conns: %d", stats.StaleConns)
}