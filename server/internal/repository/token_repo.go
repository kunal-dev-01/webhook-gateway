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
// 2FA Temporary Tokens (during login)
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
// Rate Limiting Helpers
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