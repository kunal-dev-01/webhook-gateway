package middleware

import (
	"context"
	"strconv"
	"time"

	"shared/redis"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

/* ───────────────── Config ───────────────── */

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowDuration    time.Duration
}

/* ───────────────── Limiter ───────────────── */

type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
	}
}

// ✅ OPTIMIZED: Non-blocking with timeout
func (rl *RateLimiter) Middleware(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		/* ── Identify Client ── */
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		key := "ratelimit:" + ip + ":" + c.Path()

		// ✅ OPTIMIZED: Timeout context to prevent blocking
		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		/* ── Atomic INCR + EXPIRE ── */
		// Note: IncrWithExpiry returns (int64, error)
		current, err := rl.redis.IncrWithExpiry(ctx, key, cfg.WindowDuration)
		if err != nil {
			// Graceful degradation: If Redis fails, allow the request
			log.Error().Err(err).Str("ip", ip).Msg("Rate limit Redis error, allowing request")
			return c.Next()
		}

		/* ── Calculate Remaining & TTL ── */
		// FIX: Added .Result() to get the duration value
		ttl, err := rl.redis.TTL(ctx, key).Result()
		if err != nil {
			// Fallback if TTL fails
			ttl = cfg.WindowDuration
		}

		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if current > int64(cfg.RequestsPerWindow) {
			retry := int(ttl.Seconds())
			if retry <= 0 {
				retry = 1 // Prevent 0 or negative retry
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

// ✅ NEW: Per-user rate limiting (for authenticated endpoints)
func (rl *RateLimiter) UserRateLimit(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := GetUserID(c) // This comes from auth_middleware.go
		if err != nil {
			// Not authenticated, assume global limit handles it or skip
			return c.Next()
		}

		key := "ratelimit:user:" + userID.String() + ":" + c.Path()

		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		current, err := rl.redis.IncrWithExpiry(ctx, key, cfg.WindowDuration)
		if err != nil {
			log.Error().Err(err).Msg("User rate limit Redis error")
			return c.Next()
		}

		// FIX: Added .Result()
		ttl, err := rl.redis.TTL(ctx, key).Result()
		if err != nil {
			ttl = cfg.WindowDuration
		}

		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if current > int64(cfg.RequestsPerWindow) {
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

// Helper for older Go versions (Go 1.21+ has built-in max)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}