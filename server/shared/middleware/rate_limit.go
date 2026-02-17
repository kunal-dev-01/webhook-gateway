package middleware

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"shared/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Lua Script — Atomic INCR + EXPIRE (only on first request in window)
// ============================================================================

// EXPIRE sirf tab set hoti hai jab key nai banti (count == 1)
// Isse window kabhi reset nahi hoti mid-way
const incrWithExpiryLua = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`

// ============================================================================
// Preset Configs
// ============================================================================

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowDuration    time.Duration
}

// General — authenticated routes ke liye (100 req/min per user)
func GeneralRateLimit() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 100,
		WindowDuration:    60 * time.Second,
	}
}

// Strict — sensitive endpoints ke liye (5 req/min per IP)
func StrictRateLimit() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 5,
		WindowDuration:    60 * time.Second,
	}
}

// ============================================================================
// Metrics
// ============================================================================

type rateLimiterMetrics struct {
	totalRequests atomic.Int64
	rateLimited   atomic.Int64
	redisErrors   atomic.Int64
}

// ============================================================================
// Rate Limiter
// ============================================================================

type RateLimiter struct {
	redis   *redis.Client
	metrics rateLimiterMetrics
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redisClient}
}

// ============================================================================
// Core: Atomic Increment via Lua
// ============================================================================

func (rl *RateLimiter) increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	seconds := int(window.Seconds())
	if seconds <= 0 {
		seconds = 60
	}
	return rl.redis.Eval(ctx, incrWithExpiryLua, []string{key}, seconds).Int64()
}

// ============================================================================
// IP-Based Middleware — public routes ke liye
// ============================================================================

func (rl *RateLimiter) ByIP(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rl.metrics.totalRequests.Add(1)

		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		key := "rl:ip:" + ip + ":" + c.Path()

		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		count, err := rl.increment(ctx, key, cfg.WindowDuration)
		if err != nil {
			rl.metrics.redisErrors.Add(1)
			logger.Warn("Rate limiter Redis error, allowing request", "ip", ip, "error", err.Error())
			return c.Next() // Fail open — Redis down hone pe block mat karo
		}

		limit := cfg.RequestsPerWindow
		remaining := max(limit - int(count), 0)

		ttl, _ := rl.redis.TTL(ctx, key).Result()
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if count > int64(limit) {
			rl.metrics.rateLimited.Add(1)

			retryAfter := int(ttl.Seconds())
			if retryAfter <= 0 {
				retryAfter = int(cfg.WindowDuration.Seconds())
			}

			c.Set("Retry-After", strconv.Itoa(retryAfter))

			logger.Warn("Rate limit exceeded", "ip", ip, "path", c.Path(), "count", count, "limit", limit)

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"error":       "Too many requests. Please try again later.",
				"retry_after": retryAfter,
			})
		}

		return c.Next()
	}
}

// ============================================================================
// User-Based Middleware — authenticated routes ke liye
// ============================================================================

func (rl *RateLimiter) ByUser(cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rl.metrics.totalRequests.Add(1)

		userID, err := GetUserID(c)
		if err != nil {
			// Auth nahi hua — skip karo, RequireAuth handle karega
			return c.Next()
		}

		key := "rl:user:" + userID.String() + ":" + c.Path()

		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		count, err := rl.increment(ctx, key, cfg.WindowDuration)
		if err != nil {
			rl.metrics.redisErrors.Add(1)
			logger.Warn("User rate limiter Redis error, allowing request", "user_id", userID.String(), "error", err.Error())
			return c.Next()
		}

		limit := cfg.RequestsPerWindow
		remaining := limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		ttl, _ := rl.redis.TTL(ctx, key).Result()
		resetAt := time.Now().Add(ttl).Unix()

		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if count > int64(limit) {
			rl.metrics.rateLimited.Add(1)

			retryAfter := int(ttl.Seconds())
			if retryAfter <= 0 {
				retryAfter = int(cfg.WindowDuration.Seconds())
			}

			c.Set("Retry-After", strconv.Itoa(retryAfter))

			logger.Warn("User rate limit exceeded", "user_id", userID.String(), "path", c.Path(), "count", count, "limit", limit)

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"error":       "Too many requests. Please try again later.",
				"retry_after": retryAfter,
			})
		}

		return c.Next()
	}
}

// ============================================================================
// Metrics
// ============================================================================

func (rl *RateLimiter) GetMetrics() map[string]int64 {
	return map[string]int64{
		"total_requests": rl.metrics.totalRequests.Load(),
		"rate_limited":   rl.metrics.rateLimited.Load(),
		"redis_errors":   rl.metrics.redisErrors.Load(),
	}
}