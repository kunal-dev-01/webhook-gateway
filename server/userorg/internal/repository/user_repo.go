package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"userorg/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// User Repository
// ============================================================================

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{db: database}
}

// ───────────────── Helpers (For Optimization & Safety) ─────────────────

// withWriteContext creates a detached context for safe Write operations.
// Ensures data integrity even if the HTTP request is cancelled.
func (r *UserRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// scanUser handles scanning a single user row to avoid code duplication.
func (r *UserRepository) scanUser(row *sql.Row) (*models.User, error) {
	user := &models.User{}
	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.AvatarURL,
		&user.Preferences, &user.EmailVerified,
		&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorBackupCodes, &user.TwoFactorConfirmedAt,
		&user.DeletedAt, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

// performUpdate executes an update query safely using a background context.
// It handles execution, logging, and checking for RowsAffected.
func (r *UserRepository) performUpdate(query, opName string, userID uuid.UUID, args ...any) error {
	// Use Safe Write Context
	ctx, cancel := r.withWriteContext()
	defer cancel()

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to "+opName, err, "user_id", userID)
		return errors.Wrap(err, 500, "Failed to "+opName)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found or operation not applicable")
	}

	logger.Debug("Success: "+opName, "user_id", userID)
	return nil
}

// ============================================================================
// READ (Fast Reads using Request Context)
// ============================================================================

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, full_name, avatar_url,
			preferences, email_verified,
			two_factor_enabled, two_factor_secret, two_factor_backup_codes, two_factor_confirmed_at,
			deleted_at, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	user, err := r.scanUser(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by ID", err, "user_id", id)
		return nil, errors.Internal(err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, full_name, avatar_url,
			preferences, email_verified,
			two_factor_enabled, two_factor_secret, two_factor_backup_codes, two_factor_confirmed_at,
			deleted_at, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`

	user, err := r.scanUser(r.db.QueryRow(ctx, query, email))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by email", err, "email", email)
		return nil, errors.Internal(err)
	}
	return user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		logger.Error("Failed to check email existence", err, "email", email)
		return false, errors.Internal(err)
	}
	return exists, nil
}

// ============================================================================
// UPDATE (Uses performUpdate helper for consistency)
// ============================================================================

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users SET
			full_name = $2, avatar_url = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	return r.performUpdate(query, "update user", user.ID,
		user.ID, user.FullName, user.AvatarURL, user.UpdatedAt)
}

func (r *UserRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences models.UserPreferences) error {
	query := `UPDATE users SET preferences = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	return r.performUpdate(query, "update preferences", userID, userID, preferences, time.Now())
}