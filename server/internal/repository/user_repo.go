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
	// UPDATED: Added company_id, department, designation, employee_id, avatar fields, permissions
	userSelectColumns = `
		id, company_id, email, password_hash, 
		email_verified, email_verified_at,
		two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
		full_name, phone, timezone, language, date_of_birth, gender,
		role, department, designation, employee_id,
		avatar_url, avatar_thumbnail_url,
		permissions,
		status, is_active,
		last_login_at, last_login_ip, last_active_at,
		failed_login_attempts, locked_until,
		password_changed_at, must_change_password,
		invited_by, invited_at, joined_at,
		created_at, updated_at, deleted_at
	`

	baseUserQuery = `SELECT ` + userSelectColumns + ` FROM users WHERE deleted_at IS NULL`

	// UPDATED: Added company_id and new fields
	createUserQuery = `
		INSERT INTO users (
			id, company_id, email, password_hash, full_name, phone,
			role, department, designation, employee_id,
			status, timezone, language,
			email_verified, two_factor_enabled,
			password_changed_at, invited_by, invited_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17, $18,
			$19, $20
		)
		RETURNING id, created_at, updated_at
	`

	// UPDATED: Added new fields
	updateUserQuery = `
		UPDATE users SET
			email = $2,
			full_name = $3,
			phone = $4,
			timezone = $5,
			language = $6,
			date_of_birth = $7,
			gender = $8,
			role = $9,
			department = $10,
			designation = $11,
			employee_id = $12,
			avatar_url = $13,
			avatar_thumbnail_url = $14,
			status = $15,
			email_verified = $16,
			email_verified_at = $17,
			two_factor_enabled = $18,
			two_factor_secret = $19,
			two_factor_recovery_codes = $20,
			two_factor_verified_at = $21,
			must_change_password = $22,
			updated_at = $23
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`
)

var (
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrEmailExists        = fmt.Errorf("email already exists")
	ErrUserAlreadyDeleted = fmt.Errorf("user not found or already deleted")
	ErrUserNotDeleted     = fmt.Errorf("user not found or not deleted")
)

// ============================================================================
// Create Operations
// ============================================================================

// Create creates a new user
// UPDATED: Now requires company_id
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	
	// Validation: CompanyID is required
	if user.CompanyID == uuid.Nil {
		return fmt.Errorf("company_id is required")
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, createUserQuery,
		user.ID, user.CompanyID, user.Email, user.PasswordHash, user.FullName, user.Phone,
		user.Role, user.Department, user.Designation, user.EmployeeID,
		user.Status, user.Timezone, user.Language,
		user.EmailVerified, user.TwoFactorEnabled,
		user.PasswordChangedAt, user.InvitedBy, user.InvitedAt,
		user.CreatedAt, user.UpdatedAt,
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
// UPDATED: Added company_id for multi-tenant isolation
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "id = $1", id)
}

// GetByIDAndCompany retrieves user by ID with company verification
// NEW: Multi-tenant safe query
func (r *UserRepository) GetByIDAndCompany(ctx context.Context, id, companyID uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "id = $1 AND company_id = $2", id, companyID)
}

// GetByEmail retrieves a user by email
// NOTE: Email alone is not unique - need company_id for proper lookup
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.getUser(ctx, "email = $1", strings.ToLower(email))
}

// GetByEmailAndCompany retrieves user by email within company
// NEW: Multi-tenant safe email lookup
func (r *UserRepository) GetByEmailAndCompany(ctx context.Context, email string, companyID uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "email = $1 AND company_id = $2", strings.ToLower(email), companyID)
}

// GetByEmployeeID retrieves user by employee ID within company
// NEW: Internal employee ID lookup
func (r *UserRepository) GetByEmployeeID(ctx context.Context, employeeID string, companyID uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "employee_id = $1 AND company_id = $2", employeeID, companyID)
}

// List retrieves users with filters and pagination
// UPDATED: Now requires companyID for multi-tenant isolation
func (r *UserRepository) List(ctx context.Context, companyID uuid.UUID, filters ListUsersFilter) ([]*models.User, int, error) {
	whereClauses, args := r.buildFilters(companyID, filters)
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

// GetCompanyOwner retrieves the owner of a company
// NEW: Fetch company owner
func (r *UserRepository) GetCompanyOwner(ctx context.Context, companyID uuid.UUID) (*models.User, error) {
	return r.getUser(ctx, "company_id = $1 AND role = 'owner' AND is_active = true", companyID)
}

// CountByCompany returns user count for a company
// NEW: Company-specific count
func (r *UserRepository) CountByCompany(ctx context.Context, companyID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.GetContext(ctx, &count, query, companyID)
	if err != nil {
		return 0, fmt.Errorf("failed to count company users: %w", err)
	}
	return count, nil
}

// ============================================================================
// Update Operations
// ============================================================================

// Update updates a user
// UPDATED: Now handles all new fields
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, updateUserQuery,
		user.ID, user.Email, user.FullName, user.Phone,
		user.Timezone, user.Language, user.DateOfBirth, user.Gender,
		user.Role, user.Department, user.Designation, user.EmployeeID,
		user.AvatarURL, user.AvatarThumbnailURL,
		user.Status,
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
			last_active_at = $2,
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

// UpdateActivity updates last active timestamp
// NEW: Track user activity
func (r *UserRepository) UpdateActivity(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users SET
			last_active_at = $2,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, userID, now, "update activity")
}

// IncrementFailedAttempts increments failed login attempts
// UPDATED: Now works with email + company_id for multi-tenant
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, email string, companyID uuid.UUID) error {
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
		WHERE email = $1 AND company_id = $2 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, strings.ToLower(email), companyID, "increment failed attempts")
}

// MarkEmailAsVerified marks email as verified
func (r *UserRepository) MarkEmailAsVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users SET
			email_verified = true,
			email_verified_at = $2,
			status = 'active',
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, userID, now, "mark email as verified")
}

// UpdatePermissions updates user permissions
// NEW: Granular permission updates
func (r *UserRepository) UpdatePermissions(ctx context.Context, userID uuid.UUID, permissions models.Permissions) error {
	query := `
		UPDATE users SET
			permissions = $2,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	permJSON, err := permissions.Value()
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	return r.execAndCheck(ctx, query, userID, permJSON, "update permissions")
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
			is_active = false,
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
			is_active = true,
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
// DEPRECATED: Use ExistsInCompany for multi-tenant safety
func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, strings.ToLower(email))
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// ExistsInCompany checks if email exists within company
// NEW: Multi-tenant safe existence check
func (r *UserRepository) ExistsInCompany(ctx context.Context, email string, companyID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND company_id = $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, strings.ToLower(email), companyID)
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

// UPDATED: Now includes company_id filter by default
func (r *UserRepository) buildFilters(companyID uuid.UUID, filters ListUsersFilter) ([]string, []any) {
	// CRITICAL: Always filter by company_id for multi-tenant isolation
	whereClauses := []string{"deleted_at IS NULL", "company_id = $1"}
	args := []any{companyID}
	argCounter := 2

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
	if filters.Department != nil {
		addFilter("department", *filters.Department)
	}
	if filters.EmailVerified != nil {
		addFilter("email_verified", *filters.EmailVerified)
	}
	if filters.IsActive != nil {
		addFilter("is_active", *filters.IsActive)
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(full_name ILIKE $%d OR email ILIKE $%d OR employee_id ILIKE $%d)",
			argCounter, argCounter, argCounter,
		))
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
		// Unique constraint violation
		if strings.Contains(pqErr.Constraint, "email") {
			return ErrEmailExists
		}
		return fmt.Errorf("duplicate entry: %s", pqErr.Constraint)
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// ============================================================================
// Filter Struct
// ============================================================================

// ListUsersFilter represents filters for listing users
// UPDATED: Added new filter fields
type ListUsersFilter struct {
	Status        *models.UserStatus
	Role          *models.UserRole
	Department    *string
	EmailVerified *bool
	IsActive      *bool
	Search        string // Searches: name, email, employee_id
	SortBy        string
	SortDesc      bool
	Limit         int
	Offset        int
}