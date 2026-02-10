package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Common Errors
var (
	ErrTokenNotFound = errors.New("token not found or expired")
)

// Key Prefixes (Centralized Management)
const (
	PrefixVerifyEmail   = "verify:email"
	PrefixResetPass     = "reset:pass"
	Prefix2FATemp       = "2fa:temp"
	Prefix2FASetup      = "2fa:setup"
	PrefixRecoveryUsed  = "2fa:recovery:used"
	PrefixRateLimit     = "ratelimit"
)

// TokenRepository handles token storage in Redis
type TokenRepository struct {
	redis *redis.Client
}

// NewTokenRepository creates new token repository
func NewTokenRepository(redisClient *redis.Client) *TokenRepository {
	return &TokenRepository{
		redis: redisClient,
	}
}

// ============================================================================
// Internal Key Helpers (Prevents Typos)
// ============================================================================

func keyVerifyEmail(token string) string {
	return fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
}

func keyResetPass(token string) string {
	return fmt.Sprintf("%s:%s", PrefixResetPass, token)
}

func key2FATemp(token string) string {
	return fmt.Sprintf("%s:%s", Prefix2FATemp, token)
}

func key2FASetup(userID string) string {
	return fmt.Sprintf("%s:%s", Prefix2FASetup, userID)
}

func keyRecoveryUsed(userID string, idx int) string {
	return fmt.Sprintf("%s:%s:%d", PrefixRecoveryUsed, userID, idx)
}

// ============================================================================
// Email Verification Tokens
// ============================================================================

func (r *TokenRepository) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	return r.redis.Set(ctx, keyVerifyEmail(token), userID.String(), ttl).Err()
}

func (r *TokenRepository) GetVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := r.redis.Get(ctx, keyVerifyEmail(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, ErrTokenNotFound
		}
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

func (r *TokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	// Unlink is non-blocking (faster than Del for large keys, mostly same for small)
	return r.redis.Unlink(ctx, keyVerifyEmail(token)).Err()
}

// ============================================================================
// Password Reset Tokens
// ============================================================================

func (r *TokenRepository) StoreResetToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	return r.redis.Set(ctx, keyResetPass(token), userID.String(), ttl).Err()
}

func (r *TokenRepository) GetResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := r.redis.Get(ctx, keyResetPass(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, ErrTokenNotFound
		}
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

func (r *TokenRepository) DeleteResetToken(ctx context.Context, token string) error {
	return r.redis.Unlink(ctx, keyResetPass(token)).Err()
}

// ============================================================================
// 2FA Temporary Tokens (Login Flow)
// ============================================================================

func (r *TokenRepository) Store2FATempToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	return r.redis.Set(ctx, key2FATemp(token), userID.String(), ttl).Err()
}

func (r *TokenRepository) Get2FATempToken(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := r.redis.Get(ctx, key2FATemp(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, ErrTokenNotFound
		}
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

func (r *TokenRepository) Delete2FATempToken(ctx context.Context, token string) error {
	return r.redis.Unlink(ctx, key2FATemp(token)).Err()
}

// ============================================================================
// 2FA Setup State (Secret Storage)
// ============================================================================

func (r *TokenRepository) Store2FASetupSecret(ctx context.Context, userID uuid.UUID, secret string, ttl time.Duration) error {
	return r.redis.Set(ctx, key2FASetup(userID.String()), secret, ttl).Err()
}

func (r *TokenRepository) Get2FASetupSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	val, err := r.redis.Get(ctx, key2FASetup(userID.String())).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenNotFound
		}
		return "", err
	}
	return val, nil
}

func (r *TokenRepository) Delete2FASetupSecret(ctx context.Context, userID uuid.UUID) error {
	return r.redis.Unlink(ctx, key2FASetup(userID.String())).Err()
}

// ============================================================================
// Recovery Code Tracking
// ============================================================================

func (r *TokenRepository) MarkRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int, ttl time.Duration) error {
	// Set "1" to mark as used
	return r.redis.Set(ctx, keyRecoveryUsed(userID.String(), codeIndex), "1", ttl).Err()
}

func (r *TokenRepository) IsRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int) (bool, error) {
	exists, err := r.redis.Exists(ctx, keyRecoveryUsed(userID.String(), codeIndex)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// ============================================================================
// Rate Limiting (Optimized with Pipeline)
// ============================================================================

// CheckEmailRateLimit checks limit by Email
func (r *TokenRepository) CheckEmailRateLimit(ctx context.Context, email string, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("%s:email:%s", PrefixRateLimit, email)
	return r.checkRateLimit(ctx, key, maxAttempts, window)
}

// CheckLoginRateLimit checks limit by IP
func (r *TokenRepository) CheckLoginRateLimit(ctx context.Context, ip string, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("%s:login:%s", PrefixRateLimit, ip)
	return r.checkRateLimit(ctx, key, maxAttempts, window)
}

// Check2FARateLimit checks limit by UserID
func (r *TokenRepository) Check2FARateLimit(ctx context.Context, userID uuid.UUID, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("%s:2fa:%s", PrefixRateLimit, userID.String())
	return r.checkRateLimit(ctx, key, maxAttempts, window)
}

// Reset2FARateLimit clears the counter
func (r *TokenRepository) Reset2FARateLimit(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("%s:2fa:%s", PrefixRateLimit, userID.String())
	return r.redis.Unlink(ctx, key).Err()
}

// Internal Helper: Atomic Rate Limiting
func (r *TokenRepository) checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Optimized: Use Pipeline to save network RTT
	pipe := r.redis.Pipeline()
	
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis rate limit error: %w", err)
	}

	// Logic: If current count > limit, block it
	return incr.Val() > int64(limit), nil
}