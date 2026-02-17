package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"shared/errors"
)

const (
	PrefixVerifyEmail  = "verify:email"
	PrefixResetPass    = "reset:pass"
	Prefix2FATemp      = "2fa:temp"
	Prefix2FASetup     = "2fa:setup"
	PrefixRecoveryUsed = "2fa:recovery:used"
)

type TokenRepository struct {
	redis *redis.Client
}

func NewTokenRepository(redisClient *redis.Client) *TokenRepository {
	return &TokenRepository{redis: redisClient}
}

// ============================================================================
// PRIVATE HELPERS (To reduce duplication)
// ============================================================================

func (r *TokenRepository) set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if err := r.redis.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) getUUID(ctx context.Context, key, notFoundMsg string) (uuid.UUID, error) {
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, errors.NotFound(notFoundMsg)
		}
		return uuid.Nil, errors.Internal(err)
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.Internal(err)
	}
	return id, nil
}

func (r *TokenRepository) getString(ctx context.Context, key, notFoundMsg string) (string, error) {
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.NotFound(notFoundMsg)
		}
		return "", errors.Internal(err)
	}
	return val, nil
}

func (r *TokenRepository) delete(ctx context.Context, key string) error {
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

func (r *TokenRepository) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	return r.set(ctx, key, userID.String(), ttl)
}

func (r *TokenRepository) GetVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	return r.getUUID(ctx, key, "Verification token not found or expired")
}

func (r *TokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	return r.delete(ctx, key)
}

// ============================================================================
// PASSWORD RESET
// ============================================================================

func (r *TokenRepository) StoreResetToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	return r.set(ctx, key, userID.String(), ttl)
}

func (r *TokenRepository) GetResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	return r.getUUID(ctx, key, "Reset token not found or expired")
}

func (r *TokenRepository) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	return r.delete(ctx, key)
}

// ============================================================================
// 2FA TEMPORARY TOKENS (Login Flow)
// ============================================================================

func (r *TokenRepository) Store2FATempToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	return r.set(ctx, key, userID.String(), ttl)
}

func (r *TokenRepository) Get2FATempToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	return r.getUUID(ctx, key, "2FA token not found or expired")
}

func (r *TokenRepository) Delete2FATempToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	return r.delete(ctx, key)
}

// ============================================================================
// 2FA SETUP STATE
// ============================================================================

func (r *TokenRepository) Store2FASetupSecret(ctx context.Context, userID uuid.UUID, secret string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	return r.set(ctx, key, secret, ttl)
}

func (r *TokenRepository) Get2FASetupSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	return r.getString(ctx, key, "2FA setup secret not found or expired")
}

func (r *TokenRepository) Delete2FASetupSecret(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	return r.delete(ctx, key)
}

// ============================================================================
// RECOVERY CODE TRACKING
// ============================================================================

func (r *TokenRepository) MarkRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s:%d", PrefixRecoveryUsed, userID.String(), codeIndex)
	return r.set(ctx, key, "1", ttl)
}

func (r *TokenRepository) IsRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int) (bool, error) {
	key := fmt.Sprintf("%s:%s:%d", PrefixRecoveryUsed, userID.String(), codeIndex)
	exists, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Internal(err)
	}
	return exists > 0, nil
}