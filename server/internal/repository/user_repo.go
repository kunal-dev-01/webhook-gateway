package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"shared/models"
	"shared/redis"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db    *sqlx.DB
	redis *redis.Client
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB, redisClient *redis.Client) *UserRepository {
	return &UserRepository{
		db:    db,
		redis: redisClient,
	}
}

// Cache key constants
const (
	userCachePrefix      = "user:cache:"
	userEmailCachePrefix = "user:email:cache:"
	userCacheTTL         = 15 * time.Minute
)

// ============================================================================
// Cache Helpers - PRIVATE
// ============================================================================

func (r *UserRepository) getUserCacheKey(id uuid.UUID) string {
	return userCachePrefix + id.String()
}

func (r *UserRepository) getUserEmailCacheKey(email string) string {
	return userEmailCachePrefix + strings.ToLower(email)
}

func (r *UserRepository) cacheUser(ctx context.Context, user *models.User) {
	if user == nil {
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		return
	}

	// Cache by ID
	r.redis.Set(ctx, r.getUserCacheKey(user.ID), data, userCacheTTL)
	
	// Cache by email (store user ID for lookup)
	r.redis.Set(ctx, r.getUserEmailCacheKey(user.Email), user.ID.String(), userCacheTTL)
}

func (r *UserRepository) getUserFromCache(ctx context.Context, id uuid.UUID) (*models.User, bool) {
	data, err := r.redis.Get(ctx, r.getUserCacheKey(id)).Result()
	if err != nil {
		return nil, false
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, false
	}

	return &user, true
}

// ✅ FIXED: Invalidate BOTH ID and email cache
func (r *UserRepository) invalidateUserCache(ctx context.Context, user *models.User) {
	if user == nil {
		return
	}

	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(user.ID))
	pipe.Del(ctx, r.getUserEmailCacheKey(user.Email))
	pipe.Exec(ctx)
}

// ✅ NEW: Invalidate cache by ID only (when we don't have full user object)
func (r *UserRepository) invalidateUserCacheByID(ctx context.Context, userID uuid.UUID) {
	// We only have ID, so we can only invalidate ID cache
	// Email cache will expire naturally
	r.redis.Del(ctx, r.getUserCacheKey(userID))
}

// ✅ NEW: Invalidate cache by email only
func (r *UserRepository) invalidateUserCacheByEmail(ctx context.Context, email string) {
	email = strings.ToLower(email)
	
	// Try to get user ID from email cache first
	userIDStr, err := r.redis.Get(ctx, r.getUserEmailCacheKey(email)).Result()
	if err == nil {
		if userID, parseErr := uuid.Parse(userIDStr); parseErr == nil {
			// We have the ID, invalidate both caches
			pipe := r.redis.Pipeline()
			pipe.Del(ctx, r.getUserCacheKey(userID))
			pipe.Del(ctx, r.getUserEmailCacheKey(email))
			pipe.Exec(ctx)
			return
		}
	}
	
	// Couldn't get ID, just invalidate email cache
	r.redis.Del(ctx, r.getUserEmailCacheKey(email))
}

// ============================================================================
// Create Operations
// ============================================================================

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, phone,
			role, status, timezone, language,
			email_verified, two_factor_enabled,
			password_changed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14
		)
		RETURNING id, created_at, updated_at
	`

	// Generate UUID if not set
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.Role,
		user.Status,
		user.Timezone,
		user.Language,
		user.EmailVerified,
		user.TwoFactorEnabled,
		user.PasswordChangedAt,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Cache the new user
	r.cacheUser(ctx, user)

	return nil
}

// ============================================================================
// Read Operations - OPTIMIZED WITH CACHE
// ============================================================================

// GetByID retrieves a user by ID - OPTIMIZED
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// ✅ 1. Check cache first
	if user, found := r.getUserFromCache(ctx, id); found {
		return user, nil
	}

	// ✅ 2. Cache miss - query database
	query := `
		SELECT 
			id, email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language,
			role, status,
			last_login_at, last_login_ip, failed_login_attempts, locked_until,
			password_changed_at, must_change_password,
			created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// ✅ 3. Cache for next time
	r.cacheUser(ctx, &user)

	return &user, nil
}

// GetByEmail retrieves a user by email - OPTIMIZED
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	email = strings.ToLower(email)

	// ✅ 1. Check if we have user ID in cache
	userIDStr, err := r.redis.Get(ctx, r.getUserEmailCacheKey(email)).Result()
	if err == nil {
		// Got user ID from email cache
		if userID, parseErr := uuid.Parse(userIDStr); parseErr == nil {
			// Try to get full user from cache
			if user, found := r.getUserFromCache(ctx, userID); found {
				return user, nil
			}
		}
	}

	// ✅ 2. Cache miss - query database
	query := `
		SELECT 
			id, email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language,
			role, status,
			last_login_at, last_login_ip, failed_login_attempts, locked_until,
			password_changed_at, must_change_password,
			created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	err = r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// ✅ 3. Cache for next time
	r.cacheUser(ctx, &user)

	return &user, nil
}

// List retrieves users with filters and pagination
func (r *UserRepository) List(ctx context.Context, filters ListUsersFilter) ([]*models.User, int, error) {
	// Build WHERE clause
	whereClauses := []string{"deleted_at IS NULL"}
	args := []any{}
	argCounter := 1

	if filters.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, *filters.Status)
		argCounter++
	}

	if filters.Role != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argCounter))
		args = append(args, *filters.Role)
		argCounter++
	}

	if filters.EmailVerified != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("email_verified = $%d", argCounter))
		args = append(args, *filters.EmailVerified)
		argCounter++
	}

	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(full_name ILIKE $%d OR email ILIKE $%d)",
			argCounter, argCounter,
		))
		args = append(args, "%"+filters.Search+"%")
		argCounter++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortDesc {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filters.SortBy, direction)
	}

	// Build pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := max(filters.Offset, 0)

	// Query users
	query := fmt.Sprintf(`
		SELECT 
			id, email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language,
			role, status,
			last_login_at, last_login_ip, failed_login_attempts, locked_until,
			password_changed_at, must_change_password,
			created_at, updated_at, deleted_at
		FROM users
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCounter, argCounter+1)

	args = append(args, limit, offset)

	var users []*models.User
	err = r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// ============================================================================
// Update Operations - FIXED CACHE INVALIDATION ✅
// ============================================================================

// Update updates a user - INVALIDATES CACHE
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			email = $2,
			full_name = $3,
			phone = $4,
			timezone = $5,
			language = $6,
			role = $7,
			status = $8,
			email_verified = $9,
			email_verified_at = $10,
			two_factor_enabled = $11,
			two_factor_secret = $12,
			two_factor_recovery_codes = $13,
			two_factor_verified_at = $14,
			must_change_password = $15,
			updated_at = $16
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	user.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.FullName,
		user.Phone,
		user.Timezone,
		user.Language,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.EmailVerifiedAt,
		user.TwoFactorEnabled,
		user.TwoFactorSecret,
		pq.Array(user.TwoFactorRecoveryCodes),
		user.TwoFactorVerifiedAt,
		user.MustChangePassword,
		user.UpdatedAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	// ✅ Invalidate cache after update
	r.invalidateUserCache(ctx, user)

	return nil
}

// ✅ FIXED: UpdatePassword - Proper cache invalidation
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $2,
			password_changed_at = $3,
			must_change_password = false,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING email
	`

	now := time.Now()
	
	// ✅ Get email so we can invalidate both caches
	var email string
	err := r.db.QueryRowContext(ctx, query, userID, passwordHash, now).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update password: %w", err)
	}

	// ✅ Invalidate BOTH caches
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(email))
	pipe.Exec(ctx)

	return nil
}

// ✅ FIXED: UpdateLastLogin - Proper cache invalidation
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	query := `
		UPDATE users SET
			last_login_at = $2,
			last_login_ip = $3,
			failed_login_attempts = 0,
			locked_until = NULL,
			status = CASE 
				WHEN status = 'suspended' AND (locked_until IS NULL OR locked_until < $2) THEN 'active'
				ELSE status
			END,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING email
	`

	now := time.Now()
	
	// ✅ Get email so we can invalidate both caches
	var email string
	err := r.db.QueryRowContext(ctx, query, userID, now, ip).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update last login: %w", err)
	}

	// ✅ Invalidate BOTH caches
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(email))
	pipe.Exec(ctx)

	return nil
}

// ✅ FIXED: IncrementFailedAttempts - Proper cache invalidation
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, email string) error {
	email = strings.ToLower(email)
	
	query := `
		UPDATE users SET
			failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE 
				WHEN failed_login_attempts + 1 >= 5 THEN NOW() + INTERVAL '30 minutes'
				ELSE locked_until
			END,
			status = CASE 
				WHEN failed_login_attempts + 1 >= 5 THEN 'suspended'
				ELSE status
			END,
			updated_at = NOW()
		WHERE email = $1 AND deleted_at IS NULL
		RETURNING id
	`

	// ✅ Get user ID so we can invalidate both caches
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}

	// ✅ Invalidate BOTH caches (ID and email)
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(email))
	pipe.Exec(ctx)

	return nil
}

// ✅ FIXED: MarkEmailAsVerified - Proper cache invalidation
func (r *UserRepository) MarkEmailAsVerified(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users SET
			email_verified = true,
			email_verified_at = $2,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING email
	`

	now := time.Now()
	
	// ✅ Get email so we can invalidate both caches
	var email string
	err := r.db.QueryRowContext(ctx, query, userID, now).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	// ✅ Invalidate BOTH caches
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(email))
	pipe.Exec(ctx)

	return nil
}

// ============================================================================
// Delete Operations
// ============================================================================

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users SET
			deleted_at = $2,
			status = 'deleted',
			email = email || '.deleted.' || EXTRACT(EPOCH FROM $2)::TEXT,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING email
	`

	now := time.Now()
	
	var oldEmail string
	err := r.db.QueryRowContext(ctx, query, userID, now).Scan(&oldEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found or already deleted")
		}
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	// ✅ Invalidate both caches (use old email before it was modified)
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(oldEmail))
	pipe.Exec(ctx)

	return nil
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users SET
			deleted_at = NULL,
			status = 'active',
			email = REGEXP_REPLACE(email, '\.deleted\.\d+$', ''),
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NOT NULL
		RETURNING email
	`

	now := time.Now()
	
	var email string
	err := r.db.QueryRowContext(ctx, query, userID, now).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found or not deleted")
		}
		return fmt.Errorf("failed to restore user: %w", err)
	}

	// ✅ Invalidate both caches
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, r.getUserCacheKey(userID))
	pipe.Del(ctx, r.getUserEmailCacheKey(email))
	pipe.Exec(ctx)

	return nil
}

// HardDelete permanently deletes a user (use with caution)
func (r *UserRepository) HardDelete(ctx context.Context, userID uuid.UUID) error {
	// ✅ First get user to invalidate email cache
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to hard delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	// ✅ Invalidate both caches
	r.invalidateUserCache(ctx, user)

	return nil
}

// ============================================================================
// Utility Methods
// ============================================================================

// Exists checks if a user exists by email
func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, strings.ToLower(email))
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// Count returns total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// ============================================================================
// Filter Struct
// ============================================================================

// ListUsersFilter represents filters for listing users
type ListUsersFilter struct {
	Status        *models.UserStatus
	Role          *models.UserRole
	EmailVerified *bool
	Search        string // Search in name and email
	SortBy        string // Column to sort by
	SortDesc      bool   // Sort descending
	Limit         int    // Page size
	Offset        int    // Offset for pagination
}