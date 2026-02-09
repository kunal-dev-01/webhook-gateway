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
		current, err := rl.redis.IncrWithExpiry(ctx, key, cfg.WindowDuration)
		if err != nil {
			// ✅ OPTIMIZED: Graceful degradation on Redis failure
			log.Error().
				Err(err).
				Str("ip", ip).
				Msg("Rate limit Redis error, allowing request")
			return c.Next()
		}

		/* ── Headers ── */
		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		ttl := rl.redis.TTL(ctx, key)
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if current > int64(cfg.RequestsPerWindow) {
			retry := int(ttl.Seconds())
			if retry <= 0 {
				retry = int(cfg.WindowDuration.Seconds())
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
		userID, err := GetUserID(c)
		if err != nil {
			// Not authenticated, skip
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

		remaining := max(cfg.RequestsPerWindow-int(current), 0)
		ttl := rl.redis.TTL(ctx, key)
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if current > int64(cfg.RequestsPerWindow) {
			retry := int(ttl.Seconds())
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