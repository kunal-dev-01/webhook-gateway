package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"shared/models"
)

// UserRepository handles database operations for users - NO CACHE
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ============================================================================
// SQL Queries
// ============================================================================

const (
	userSelectColumns = `
		id, email, password_hash, email_verified, email_verified_at,
		two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
		full_name, phone, timezone, language,
		role, status,
		last_login_at, last_login_ip, failed_login_attempts, locked_until,
		password_changed_at, must_change_password,
		created_at, updated_at, deleted_at
	`

	baseUserQuery = `SELECT ` + userSelectColumns + ` FROM users WHERE deleted_at IS NULL`

	createUserQuery = `
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

	updateUserQuery = `
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
)

var (
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrEmailExists       = fmt.Errorf("email already exists")
	ErrUserAlreadyDeleted = fmt.Errorf("user not found or already deleted")
	ErrUserNotDeleted    = fmt.Errorf("user not found or not deleted")
)

// ============================================================================
// Create Operations
// ============================================================================

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, createUserQuery,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Phone,
		user.Role, user.Status, user.Timezone, user.Language,
		user.EmailVerified, user.TwoFactorEnabled,
		user.PasswordChangedAt, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return r.handleDBError(err, "create user")
	}
	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "id = $1", id)
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.getUser(ctx, "email = $1", strings.ToLower(email))
}

// List retrieves users with filters and pagination
func (r *UserRepository) List(ctx context.Context, filters ListUsersFilter) ([]*models.User, int, error) {
	whereClauses, args := r.buildFilters(filters)
	whereClause := strings.Join(whereClauses, " AND ")

	total, err := r.countUsers(ctx, whereClause, args)
	if err != nil {
		return nil, 0, err
	}

	users, err := r.fetchUsers(ctx, whereClause, args, filters)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ============================================================================
// Update Operations
// ============================================================================

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, updateUserQuery,
		user.ID, user.Email, user.FullName, user.Phone,
		user.Timezone, user.Language, user.Role, user.Status,
		user.EmailVerified, user.EmailVerifiedAt,
		user.TwoFactorEnabled, user.TwoFactorSecret,
		pq.Array(user.TwoFactorRecoveryCodes), user.TwoFactorVerifiedAt,
		user.MustChangePassword, user.UpdatedAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return r.handleDBError(err, "update user")
	}
	return nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	now := time.Now()
	query := `
		UPDATE users SET
			password_hash = $2,
			password_changed_at = $3,
			must_change_password = false,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, userID, passwordHash, now, "update password")
}

// UpdateLastLogin updates last login info
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	now := time.Now()
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
	`
	return r.execAndCheck(ctx, query, userID, now, ip, "update last login")
}

// IncrementFailedAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, email string) error {
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
	`
	return r.execAndCheck(ctx, query, strings.ToLower(email), "increment failed attempts")
}

// MarkEmailAsVerified marks email as verified
func (r *UserRepository) MarkEmailAsVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users SET
			email_verified = true,
			email_verified_at = $2,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, userID, now, "mark email as verified")
}

// ============================================================================
// Delete Operations
// ============================================================================

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users SET
			deleted_at = $2,
			status = 'deleted',
			email = email || '.deleted.' || EXTRACT(EPOCH FROM $2)::TEXT,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`
	if err := r.execAndCheck(ctx, query, userID, now, "soft delete user"); err != nil {
		if err == ErrUserNotFound {
			return ErrUserAlreadyDeleted
		}
		return err
	}
	return nil
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users SET
			deleted_at = NULL,
			status = 'active',
			email = REGEXP_REPLACE(email, '\.deleted\.\d+$', ''),
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NOT NULL
	`
	if err := r.execAndCheck(ctx, query, userID, now, "restore user"); err != nil {
		if err == ErrUserNotFound {
			return ErrUserNotDeleted
		}
		return err
	}
	return nil
}

// HardDelete permanently deletes a user (use with caution)
func (r *UserRepository) HardDelete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	return r.execAndCheck(ctx, query, userID, "hard delete user")
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
// Helper Methods
// ============================================================================

func (r *UserRepository) getUser(ctx context.Context, condition string, args ...any) (*models.User, error) {
	query := baseUserQuery + " AND " + condition
	var user models.User
	err := r.db.GetContext(ctx, &user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) buildFilters(filters ListUsersFilter) ([]string, []any) {
	whereClauses := []string{"deleted_at IS NULL"}
	args := []any{}
	argCounter := 1

	addFilter := func(column string, value any) {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", column, argCounter))
		args = append(args, value)
		argCounter++
	}

	if filters.Status != nil {
		addFilter("status", *filters.Status)
	}
	if filters.Role != nil {
		addFilter("role", *filters.Role)
	}
	if filters.EmailVerified != nil {
		addFilter("email_verified", *filters.EmailVerified)
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d)", argCounter, argCounter))
		args = append(args, "%"+filters.Search+"%")
		argCounter++
	}

	return whereClauses, args
}

func (r *UserRepository) countUsers(ctx context.Context, whereClause string, args []any) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return total, nil
}

func (r *UserRepository) fetchUsers(ctx context.Context, whereClause string, args []any, filters ListUsersFilter) ([]*models.User, error) {
	orderBy := "created_at DESC"
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortDesc {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filters.SortBy, direction)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := max(filters.Offset, 0)

	argCounter := len(args) + 1
	query := fmt.Sprintf(`
		SELECT %s FROM users 
		WHERE %s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d
	`, userSelectColumns, whereClause, orderBy, argCounter, argCounter+1)

	args = append(args, limit, offset)

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) execAndCheck(ctx context.Context, query string, args ...any) error {
	operation := args[len(args)-1].(string)
	execArgs := args[:len(args)-1]

	result, err := r.db.ExecContext(ctx, query, execArgs...)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", operation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for %s: %w", operation, err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) handleDBError(err error, operation string) error {
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return ErrEmailExists
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// ============================================================================
// Filter Struct
// ============================================================================

// ListUsersFilter represents filters for listing users
type ListUsersFilter struct {
	Status        *models.UserStatus
	Role          *models.UserRole
	EmailVerified *bool
	Search        string
	SortBy        string
	SortDesc      bool
	Limit         int
	Offset        int
}