package middleware

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

/* ───────────────── Config ───────────────── */

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowDuration    time.Duration
	
	// Bypass Options
	WhitelistIPs  []string
	BypassPaths   []string
	CustomKeyFunc func(*fiber.Ctx) string
	
	// Algorithm
	UseSlidingWindow bool // false = fixed window, true = sliding window
}

/* ───────────────── Metrics ───────────────── */

type RateLimiterMetrics struct {
	TotalRequests   atomic.Int64
	RateLimited     atomic.Int64
	RedisErrors     atomic.Int64
	WhitelistBypass atomic.Int64
	PathBypass      atomic.Int64
}

/* ───────────────── Limiter ───────────────── */

type RateLimiter struct {
	redis   *redis.Client
	metrics RateLimiterMetrics
}

// NewRateLimiter creates a new rate limiter with Redis client
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
	}
}

/* ───────────────── Helper Functions ───────────────── */

// isWhitelisted checks if IP is in whitelist
func isWhitelisted(ip string, whitelist []string) bool {
	for _, whitelistedIP := range whitelist {
		if ip == whitelistedIP || strings.HasPrefix(ip, whitelistedIP) {
			return true
		}
	}
	return false
}

// isBypassPath checks if path should bypass rate limiting
func isBypassPath(path string, bypassPaths []string) bool {
	for _, bypassPath := range bypassPaths {
		if path == bypassPath || strings.HasPrefix(path, bypassPath) {
			return true
		}
	}
	return false
}

/* ───────────────── Redis Helper Methods ───────────────── */

// IncrWithExpiry atomically increments a key and sets expiry if it's new
func (rl *RateLimiter) incrWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	pipe := rl.redis.Pipeline()
	
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incrCmd.Val(), nil
}

/* ───────────────── Middleware ───────────────── */

// Middleware applies rate limiting
func (rl *RateLimiter) Middleware(cfg RateLimitConfig) fiber.Handler {
	// Use sliding window if enabled
	if cfg.UseSlidingWindow {
		return rl.slidingWindowMiddleware(cfg)
	}
	return rl.fixedWindowMiddleware(cfg)
}

// fixedWindowMiddleware implements fixed window rate limiting
func (rl *RateLimiter) fixedWindowMiddleware(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rl.metrics.TotalRequests.Add(1)
		
		/* ── Identify Client ── */
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		/* ── Check Whitelist ── */
		if isWhitelisted(ip, cfg.WhitelistIPs) {
			rl.metrics.WhitelistBypass.Add(1)
			return c.Next()
		}

		/* ── Check Bypass Paths ── */
		if isBypassPath(c.Path(), cfg.BypassPaths) {
			rl.metrics.PathBypass.Add(1)
			return c.Next()
		}

		/* ── Generate Key ── */
		var key string
		if cfg.CustomKeyFunc != nil {
			key = cfg.CustomKeyFunc(c)
		} else {
			key = "ratelimit:" + ip + ":" + c.Path()
		}

		/* ── Timeout Context ── */
		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		/* ── Atomic INCR + EXPIRE ── */
		current, err := rl.incrWithExpiry(ctx, key, cfg.WindowDuration)
		if err != nil {
			rl.metrics.RedisErrors.Add(1)
			
			// Graceful degradation: If Redis fails, allow the request
			log.Error().Err(err).Str("ip", ip).Msg("Rate limit Redis error, allowing request")
			return c.Next()
		}

		/* ── Calculate Remaining & TTL ── */
		ttl, err := rl.redis.TTL(ctx, key).Result()
		if err != nil {
			ttl = cfg.WindowDuration
		}

		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		resetAt := time.Now().Add(ttl).Unix()

		/* ── Set Headers ── */
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if current > int64(cfg.RequestsPerWindow) {
			rl.metrics.RateLimited.Add(1)
			
			retry := int(ttl.Seconds())
			if retry <= 0 {
				retry = 1
			}

			c.Set("Retry-After", strconv.Itoa(retry))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"error":       "Too many requests",
				"retry_after": retry,
			})
		}

		return c.Next()
	}
}

// slidingWindowMiddleware implements sliding window rate limiting
func (rl *RateLimiter) slidingWindowMiddleware(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rl.metrics.TotalRequests.Add(1)
		
		/* ── Identify Client ── */
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		/* ── Check Whitelist ── */
		if isWhitelisted(ip, cfg.WhitelistIPs) {
			rl.metrics.WhitelistBypass.Add(1)
			return c.Next()
		}

		/* ── Check Bypass Paths ── */
		if isBypassPath(c.Path(), cfg.BypassPaths) {
			rl.metrics.PathBypass.Add(1)
			return c.Next()
		}

		/* ── Generate Key ── */
		var key string
		if cfg.CustomKeyFunc != nil {
			key = cfg.CustomKeyFunc(c)
		} else {
			key = "ratelimit:sliding:" + ip + ":" + c.Path()
		}

		/* ── Timeout Context ── */
		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		/* ── Sliding Window Logic ── */
		now := time.Now().UnixNano()
		windowStart := now - cfg.WindowDuration.Nanoseconds()

		// Use Redis pipeline for atomic operations
		pipe := rl.redis.Pipeline()
		
		// Remove old entries outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
		
		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: strconv.FormatInt(now, 10),
		})
		
		// Count requests in window
		pipe.ZCard(ctx, key)
		
		// Set expiry
		pipe.Expire(ctx, key, cfg.WindowDuration*2)
		
		// Execute pipeline
		cmds, err := pipe.Exec(ctx)
		if err != nil {
			rl.metrics.RedisErrors.Add(1)
			
			log.Error().Err(err).Str("ip", ip).Msg("Sliding window Redis error, allowing request")
			return c.Next()
		}

		// Get count from pipeline results (3rd command)
		count := cmds[2].(*redis.IntCmd).Val()

		/* ── Calculate Remaining ── */
		remaining := max(cfg.RequestsPerWindow-int(count), 0)
		resetAt := time.Now().Add(cfg.WindowDuration).Unix()

		/* ── Set Headers ── */
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if count > int64(cfg.RequestsPerWindow) {
			rl.metrics.RateLimited.Add(1)
			
			retry := int(cfg.WindowDuration.Seconds())
			if retry <= 0 {
				retry = 1
			}

			c.Set("Retry-After", strconv.Itoa(retry))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"error":       "Too many requests",
				"retry_after": retry,
			})
		}

		return c.Next()
	}
}

/* ───────────────── User Rate Limiting ───────────────── */

// UserRateLimit applies rate limiting per authenticated user
func (rl *RateLimiter) UserRateLimit(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rl.metrics.TotalRequests.Add(1)
		
		userID, err := GetUserID(c)
		if err != nil {
			// Not authenticated, skip user rate limit
			return c.Next()
		}

		/* ── Check Bypass Paths ── */
		if isBypassPath(c.Path(), cfg.BypassPaths) {
			rl.metrics.PathBypass.Add(1)
			return c.Next()
		}

		/* ── Generate Key ── */
		var key string
		if cfg.CustomKeyFunc != nil {
			key = cfg.CustomKeyFunc(c)
		} else {
			key = "ratelimit:user:" + userID.String() + ":" + c.Path()
		}

		/* ── Timeout Context ── */
		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		/* ── Atomic INCR + EXPIRE ── */
		current, err := rl.incrWithExpiry(ctx, key, cfg.WindowDuration)
		if err != nil {
			rl.metrics.RedisErrors.Add(1)
			
			log.Error().Err(err).Msg("User rate limit Redis error")
			return c.Next()
		}

		/* ── Calculate Remaining & TTL ── */
		ttl, err := rl.redis.TTL(ctx, key).Result()
		if err != nil {
			ttl = cfg.WindowDuration
		}

		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		resetAt := time.Now().Add(ttl).Unix()

		/* ── Set Headers ── */
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if current > int64(cfg.RequestsPerWindow) {
			rl.metrics.RateLimited.Add(1)
			
			retry := int(ttl.Seconds())
			if retry <= 0 {
				retry = 1
			}

			c.Set("Retry-After", strconv.Itoa(retry))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"error":       "Too many requests",
				"retry_after": retry,
			})
		}

		return c.Next()
	}
}

/* ───────────────── Metrics ───────────────── */

// GetMetrics returns current metrics
func (rl *RateLimiter) GetMetrics() map[string]int64 {
	return map[string]int64{
		"total_requests":   rl.metrics.TotalRequests.Load(),
		"rate_limited":     rl.metrics.RateLimited.Load(),
		"redis_errors":     rl.metrics.RedisErrors.Load(),
		"whitelist_bypass": rl.metrics.WhitelistBypass.Load(),
		"path_bypass":      rl.metrics.PathBypass.Load(),
	}
}

// ResetMetrics resets all metrics
func (rl *RateLimiter) ResetMetrics() {
	rl.metrics.TotalRequests.Store(0)
	rl.metrics.RateLimited.Store(0)
	rl.metrics.RedisErrors.Store(0)
	rl.metrics.WhitelistBypass.Store(0)
	rl.metrics.PathBypass.Store(0)
}