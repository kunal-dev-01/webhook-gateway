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

func (rl *RateLimiter) Middleware(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {

		/* ── Identify Client ── */
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		key := "ratelimit:" + ip + ":" + c.Path()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		/* ── Atomic INCR + EXPIRE ── */
		current, err := rl.redis.IncrWithExpiry(
			ctx,
			key,
			cfg.WindowDuration,
		)

		if err != nil {
			log.Error().
				Err(err).
				Msg("Rate limit Redis error, allowing request")

			return c.Next()
		}

		/* ── Headers ── */
		remaining := max(cfg.RequestsPerWindow - int(current), 0)

		ttl := rl.redis.TTL(ctx, key)

		resetAt := time.Now().
			Add(ttl).
			Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		/* ── Enforce Limit ── */
		if current > int64(cfg.RequestsPerWindow) {

			retry := int(ttl.Seconds())

			c.Set("Retry-After", strconv.Itoa(retry))

			return c.Status(fiber.StatusTooManyRequests).
				JSON(fiber.Map{
					"success":     false,
					"error":       "Too many requests",
					"retry_after": retry,
				})
		}

		return c.Next()
	}
}