package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shared/redis"
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
// Email Verification Tokens
// ============================================================================

// StoreVerificationToken stores email verification token
func (r *TokenRepository) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("verify:email:%s", token)

	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	return nil
}

// GetVerificationToken retrieves user ID from verification token
func (r *TokenRepository) GetVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("verify:email:%s", token)

	userIDStr, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid or expired verification token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token")
	}

	return userID, nil
}

// DeleteVerificationToken removes verification token
func (r *TokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("verify:email:%s", token)
	return r.redis.Del(ctx, key).Err()
}

// ============================================================================
// Password Reset Tokens
// ============================================================================

// StoreResetToken stores password reset token
func (r *TokenRepository) StoreResetToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("reset:password:%s", token)

	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	return nil
}

// GetResetToken retrieves user ID from reset token
func (r *TokenRepository) GetResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("reset:password:%s", token)

	userIDStr, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid or expired reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token")
	}

	return userID, nil
}

// DeleteResetToken removes reset token
func (r *TokenRepository) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("reset:password:%s", token)
	return r.redis.Del(ctx, key).Err()
}

// ============================================================================
// 2FA Temporary Tokens (during login) - OPTIMIZED
// ============================================================================

// Store2FATempToken stores temporary token for 2FA verification during login
func (r *TokenRepository) Store2FATempToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("2fa:temp:%s", token)

	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to store 2FA temp token: %w", err)
	}

	return nil
}

// Get2FATempToken retrieves user ID from 2FA temp token
func (r *TokenRepository) Get2FATempToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("2fa:temp:%s", token)

	userIDStr, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid or expired 2FA token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token")
	}

	return userID, nil
}

// Delete2FATempToken removes 2FA temp token
func (r *TokenRepository) Delete2FATempToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("2fa:temp:%s", token)
	return r.redis.Del(ctx, key).Err()
}

// ============================================================================
// Recovery Code Tracking - NEW
// ============================================================================

// ✅ NEW: MarkRecoveryCodeUsed marks a recovery code as used (prevent reuse)
func (r *TokenRepository) MarkRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int, ttl time.Duration) error {
	key := fmt.Sprintf("2fa:recovery:used:%s:%d", userID.String(), codeIndex)

	if err := r.redis.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("failed to mark recovery code as used: %w", err)
	}

	return nil
}

// ✅ NEW: IsRecoveryCodeUsed checks if recovery code has been used
func (r *TokenRepository) IsRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int) (bool, error) {
	key := fmt.Sprintf("2fa:recovery:used:%s:%d", userID.String(), codeIndex)

	exists, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check recovery code usage: %w", err)
	}

	return exists > 0, nil
}

// ============================================================================
// Rate Limiting Helpers - OPTIMIZED
// ============================================================================

// CheckEmailRateLimit checks if email sending is rate limited
func (r *TokenRepository) CheckEmailRateLimit(ctx context.Context, email string, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:email:%s", email)

	count, err := r.redis.IncrWithExpiry(ctx, key, window)
	if err != nil {
		return false, err
	}

	return count > int64(maxAttempts), nil
}

// CheckLoginRateLimit checks if login attempts are rate limited
func (r *TokenRepository) CheckLoginRateLimit(ctx context.Context, ip string, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:login:%s", ip)

	count, err := r.redis.IncrWithExpiry(ctx, key, window)
	if err != nil {
		return false, err
	}

	return count > int64(maxAttempts), nil
}

// ✅ NEW: Check2FARateLimit checks if 2FA attempts are rate limited (prevent brute force)
func (r *TokenRepository) Check2FARateLimit(ctx context.Context, userID uuid.UUID, maxAttempts int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:2fa:%s", userID.String())

	count, err := r.redis.IncrWithExpiry(ctx, key, window)
	if err != nil {
		return false, err
	}

	return count > int64(maxAttempts), nil
}

// ✅ NEW: Reset2FARateLimit resets 2FA rate limit after successful auth
func (r *TokenRepository) Reset2FARateLimit(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("ratelimit:2fa:%s", userID.String())
	return r.redis.Del(ctx, key).Err()
}

// ============================================================================
// Cleanup Helpers - OPTIMIZED WITH PIPELINE
// ============================================================================

// ✅ NEW: CleanupExpiredTokens removes all expired tokens for a user
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context, userID uuid.UUID) error {
	// ✅ OPTIMIZED: Use pipeline for batch delete
	pipe := r.redis.Pipeline()

	// Pattern matching for user's tokens
	patterns := []string{
		fmt.Sprintf("verify:email:*:%s", userID.String()),
		fmt.Sprintf("reset:password:*:%s", userID.String()),
		fmt.Sprintf("2fa:temp:*:%s", userID.String()),
	}

	for _, pattern := range patterns {
		keys, err := r.redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			pipe.Del(ctx, key)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}