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
// EMAIL VERIFICATION
// ============================================================================

func (r *TokenRepository) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) GetVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, errors.NotFound("Verification token not found or expired")
		}
		return uuid.Nil, errors.Internal(err)
	}
	
	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.Internal(err)
	}
	return userID, nil
}

func (r *TokenRepository) DeleteVerificationToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", PrefixVerifyEmail, token)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// PASSWORD RESET
// ============================================================================

func (r *TokenRepository) StoreResetToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) GetResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, errors.NotFound("Reset token not found or expired")
		}
		return uuid.Nil, errors.Internal(err)
	}
	
	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.Internal(err)
	}
	return userID, nil
}

func (r *TokenRepository) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", PrefixResetPass, token)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// 2FA TEMPORARY TOKENS (Login Flow)
// ============================================================================

func (r *TokenRepository) Store2FATempToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	if err := r.redis.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) Get2FATempToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, errors.NotFound("2FA token not found or expired")
		}
		return uuid.Nil, errors.Internal(err)
	}
	
	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, errors.Internal(err)
	}
	return userID, nil
}

func (r *TokenRepository) Delete2FATempToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s:%s", Prefix2FATemp, token)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// 2FA SETUP STATE
// ============================================================================

func (r *TokenRepository) Store2FASetupSecret(ctx context.Context, userID uuid.UUID, secret string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	if err := r.redis.Set(ctx, key, secret, ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) Get2FASetupSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.NotFound("2FA setup secret not found or expired")
		}
		return "", errors.Internal(err)
	}
	return val, nil
}

func (r *TokenRepository) Delete2FASetupSecret(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("%s:%s", Prefix2FASetup, userID.String())
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// RECOVERY CODE TRACKING
// ============================================================================

func (r *TokenRepository) MarkRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s:%d", PrefixRecoveryUsed, userID.String(), codeIndex)
	if err := r.redis.Set(ctx, key, "1", ttl).Err(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

func (r *TokenRepository) IsRecoveryCodeUsed(ctx context.Context, userID uuid.UUID, codeIndex int) (bool, error) {
	key := fmt.Sprintf("%s:%s:%d", PrefixRecoveryUsed, userID.String(), codeIndex)
	exists, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Internal(err)
	}
	return exists > 0, nil
}