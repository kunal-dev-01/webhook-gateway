package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"authentication/internal/db"
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

// ============================================================================
// CREATE
// ============================================================================

// Create creates a new user with default preferences
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, avatar_url,
			preferences, email_verified,
			two_factor_enabled, two_factor_secret, two_factor_backup_codes, two_factor_confirmed_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.AvatarURL,
		user.Preferences, user.EmailVerified,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.TwoFactorBackupCodes, user.TwoFactorConfirmedAt,
		user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create user", err, "email", user.Email)
		return errors.Wrap(err, 500, "Failed to create user")
	}

	logger.Debug("User created successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// ============================================================================
// READ
// ============================================================================

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT 
			id, email, password_hash, full_name, avatar_url,
			preferences, email_verified,
			two_factor_enabled, two_factor_secret, two_factor_backup_codes, two_factor_confirmed_at,
			deleted_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.AvatarURL,
		&user.Preferences, &user.EmailVerified,
		&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorBackupCodes, &user.TwoFactorConfirmedAt,
		&user.DeletedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by ID", err, "user_id", id)
		return nil, errors.Internal(err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT 
			id, email, password_hash, full_name, avatar_url,
			preferences, email_verified,
			two_factor_enabled, two_factor_secret, two_factor_backup_codes, two_factor_confirmed_at,
			deleted_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.AvatarURL,
		&user.Preferences, &user.EmailVerified,
		&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorBackupCodes, &user.TwoFactorConfirmedAt,
		&user.DeletedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by email", err, "email", email)
		return nil, errors.Internal(err)
	}

	return user, nil
}

// ExistsByEmail checks if a user exists with given email
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
// UPDATE
// ============================================================================

// Update updates user profile information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	
	query := `
		UPDATE users SET
			email = $2,
			full_name = $3,
			avatar_url = $4,
			preferences = $5,
			updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.FullName, user.AvatarURL,
		user.Preferences, user.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update user", err, "user_id", user.ID)
		return errors.Wrap(err, 500, "Failed to update user")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Debug("User updated successfully", "user_id", user.ID)
	return nil
}

// UpdatePassword updates user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $2, updated_at = $3 
		WHERE id = $1 AND deleted_at IS NULL`
	
	now := time.Now()
	result, err := r.db.Exec(ctx, query, userID, passwordHash, now)
	
	if err != nil {
		logger.Error("Failed to update password", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Debug("Password updated successfully", "user_id", userID)
	return nil
}

// MarkEmailVerified marks user's email as verified
func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET email_verified = true, updated_at = $2 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now)
	
	if err != nil {
		logger.Error("Failed to mark email verified", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Debug("Email marked as verified", "user_id", userID)
	return nil
}

// UpdatePreferences updates user preferences
func (r *UserRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences models.UserPreferences) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET preferences = $2, updated_at = $3 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, preferences, now)
	
	if err != nil {
		logger.Error("Failed to update preferences", err, "user_id", userID)
		return errors.Wrap(err, 500, "Failed to update preferences")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Debug("User preferences updated", "user_id", userID)
	return nil
}

// ============================================================================
// TWO-FACTOR AUTHENTICATION (2FA)
// ============================================================================

// Enable2FA enables two-factor authentication for user
func (r *UserRepository) Enable2FA(ctx context.Context, userID uuid.UUID, secret string, backupCodes []string) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET 
			two_factor_enabled = true,
			two_factor_secret = $2,
			two_factor_backup_codes = $3,
			two_factor_confirmed_at = $4,
			updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, secret, backupCodes, now, now)
	
	if err != nil {
		logger.Error("Failed to enable 2FA", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Info("2FA enabled for user", "user_id", userID)
	return nil
}

// Disable2FA disables two-factor authentication for user
func (r *UserRepository) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET 
			two_factor_enabled = false,
			two_factor_secret = NULL,
			two_factor_backup_codes = NULL,
			two_factor_confirmed_at = NULL,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now)
	
	if err != nil {
		logger.Error("Failed to disable 2FA", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	logger.Info("2FA disabled for user", "user_id", userID)
	return nil
}

// ============================================================================
// SOFT DELETE
// ============================================================================

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	
	// Append timestamp to email to allow recreation with same email
	query := `
		UPDATE users 
		SET 
			deleted_at = $2,
			email = CONCAT(email, '.deleted.', $3),
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now, now.Unix(), now)
	
	if err != nil {
		logger.Error("Failed to soft delete user", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found or already deleted")
	}

	logger.Info("User soft deleted", "user_id", userID)
	return nil
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	
	// Remove .deleted.timestamp suffix from email
	query := `
		UPDATE users 
		SET 
			deleted_at = NULL,
			email = REGEXP_REPLACE(email, '\.deleted\.\d+$', ''),
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := r.db.Exec(ctx, query, userID, now)
	
	if err != nil {
		logger.Error("Failed to restore user", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found or not deleted")
	}

	logger.Info("User restored", "user_id", userID)
	return nil
}