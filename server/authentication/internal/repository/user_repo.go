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

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{db: database}
}

// ============================================================================
// CREATE
// ============================================================================

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, company_id, role, department, designation, employee_id,
			email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language, date_of_birth, gender,
			avatar_url, avatar_thumbnail_url, permissions, status, is_active,
			last_login_at, last_login_ip, last_active_at, failed_login_attempts, locked_until,
			password_changed_at, must_change_password, invited_by, invited_at, joined_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37
		)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.CompanyID, user.Role, user.Department, user.Designation, user.EmployeeID,
		user.Email, user.PasswordHash, user.EmailVerified, user.EmailVerifiedAt,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.TwoFactorRecoveryCodes, user.TwoFactorVerifiedAt,
		user.FullName, user.Phone, user.Timezone, user.Language, user.DateOfBirth, user.Gender,
		user.AvatarURL, user.AvatarThumbnailURL, user.Permissions, user.Status, user.IsActive,
		user.LastLoginAt, user.LastLoginIP, user.LastActiveAt, user.FailedLoginAttempts, user.LockedUntil,
		user.PasswordChangedAt, user.MustChangePassword, user.InvitedBy, user.InvitedAt, user.JoinedAt,
		user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create user", err, "email", user.Email)
		return errors.Wrap(err, 500, "Failed to create user")
	}

	if err := r.createDefaultPreferences(ctx, user.ID, user.Timezone, user.Language); err != nil {
		logger.Error("Failed to create default preferences", err, "user_id", user.ID)
		return err
	}

	return nil
}

func (r *UserRepository) createDefaultPreferences(ctx context.Context, userID uuid.UUID, timezone, language string) error {
	prefs := &models.UserPreferences{
		ID:                   uuid.New(),
		UserID:               userID,
		Theme:                models.ThemeModeAuto,
		CompactMode:          false,
		SidebarCollapsed:     false,
		Language:             language,
		Timezone:             timezone,
		DateFormat:           "DD/MM/YYYY",
		TimeFormat:           "24h",
		Currency:             "INR",
		EmailDigestFrequency: "daily",
		EmailDigestTime:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	prefs.Notifications = models.NotificationSettings{
		Email: models.NotificationChannel{Enabled: true, Digest: "instant", Events: make(map[string]bool)},
		Push:  models.NotificationChannel{Enabled: true, Digest: "instant", Events: make(map[string]bool)},
		InApp: models.InAppNotifications{Enabled: true, Sound: true, DesktopNotifications: true},
		SMS:   models.NotificationChannel{Enabled: false, Events: make(map[string]bool)},
	}

	prefs.Dashboard = models.DashboardSettings{
		Widgets:     []models.DashboardWidget{},
		DefaultView: "overview",
		ShowTips:    true,
	}

	prefs.Privacy = models.PrivacySettings{
		ShowProfilePicture: true,
		ShowPhone:          true,
		ShowEmail:          true,
		ShowBirthday:       false,
		AllowMentions:      true,
	}

	prefs.Accessibility = models.AccessibilitySettings{
		HighContrast:      false,
		ReduceMotion:      false,
		ScreenReader:      false,
		FontSize:          "medium",
		KeyboardShortcuts: true,
	}

	prefs.MobileSettings = models.MobileSettings{
		BiometricAuth: false,
		OfflineMode:   false,
		DataSaver:     false,
		AutoSync:      true,
	}

	query := `
		INSERT INTO user_preferences (
			id, user_id, theme, compact_mode, sidebar_collapsed,
			language, timezone, date_format, time_format, currency,
			notifications, dashboard, privacy, accessibility,
			email_digest_frequency, email_digest_time, mobile_settings,
			quick_actions, keyboard_shortcuts, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.Exec(ctx, query,
		prefs.ID, prefs.UserID, prefs.Theme, prefs.CompactMode, prefs.SidebarCollapsed,
		prefs.Language, prefs.Timezone, prefs.DateFormat, prefs.TimeFormat, prefs.Currency,
		prefs.Notifications, prefs.Dashboard, prefs.Privacy, prefs.Accessibility,
		prefs.EmailDigestFrequency, prefs.EmailDigestTime, prefs.MobileSettings,
		prefs.QuickActions, prefs.KeyboardShortcuts, prefs.CreatedAt, prefs.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, 500, "Failed to create user preferences")
	}

	return nil
}

// ============================================================================
// READ
// ============================================================================

type UserWithPreferences struct {
	User        *models.User
	Preferences *models.UserPreferences
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*UserWithPreferences, error) {
	user := &models.User{}
	query := `
		SELECT id, company_id, role, department, designation, employee_id,
			email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language, date_of_birth, gender,
			avatar_url, avatar_thumbnail_url, permissions, status, is_active,
			last_login_at, last_login_ip, last_active_at, failed_login_attempts, locked_until,
			password_changed_at, must_change_password, invited_by, invited_at, joined_at,
			created_at, updated_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.CompanyID, &user.Role, &user.Department, &user.Designation, &user.EmployeeID,
		&user.Email, &user.PasswordHash, &user.EmailVerified, &user.EmailVerifiedAt,
		&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorRecoveryCodes, &user.TwoFactorVerifiedAt,
		&user.FullName, &user.Phone, &user.Timezone, &user.Language, &user.DateOfBirth, &user.Gender,
		&user.AvatarURL, &user.AvatarThumbnailURL, &user.Permissions, &user.Status, &user.IsActive,
		&user.LastLoginAt, &user.LastLoginIP, &user.LastActiveAt, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.PasswordChangedAt, &user.MustChangePassword, &user.InvitedBy, &user.InvitedAt, &user.JoinedAt,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by ID", err, "user_id", id)
		return nil, errors.Internal(err)
	}

	prefs, err := r.getPreferences(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserWithPreferences{User: user, Preferences: prefs}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*UserWithPreferences, error) {
	user := &models.User{}
	query := `
		SELECT id, company_id, role, department, designation, employee_id,
			email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language, date_of_birth, gender,
			avatar_url, avatar_thumbnail_url, permissions, status, is_active,
			last_login_at, last_login_ip, last_active_at, failed_login_attempts, locked_until,
			password_changed_at, must_change_password, invited_by, invited_at, joined_at,
			created_at, updated_at, deleted_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.CompanyID, &user.Role, &user.Department, &user.Designation, &user.EmployeeID,
		&user.Email, &user.PasswordHash, &user.EmailVerified, &user.EmailVerifiedAt,
		&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorRecoveryCodes, &user.TwoFactorVerifiedAt,
		&user.FullName, &user.Phone, &user.Timezone, &user.Language, &user.DateOfBirth, &user.Gender,
		&user.AvatarURL, &user.AvatarThumbnailURL, &user.Permissions, &user.Status, &user.IsActive,
		&user.LastLoginAt, &user.LastLoginIP, &user.LastActiveAt, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.PasswordChangedAt, &user.MustChangePassword, &user.InvitedBy, &user.InvitedAt, &user.JoinedAt,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User not found")
	}
	if err != nil {
		logger.Error("Failed to get user by email", err, "email", email)
		return nil, errors.Internal(err)
	}

	prefs, err := r.getPreferences(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &UserWithPreferences{User: user, Preferences: prefs}, nil
}

func (r *UserRepository) getPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreferences, error) {
	prefs := &models.UserPreferences{}
	query := `
		SELECT id, user_id, theme, compact_mode, sidebar_collapsed,
			language, timezone, date_format, time_format, currency,
			notifications, dashboard, privacy, accessibility,
			email_digest_frequency, email_digest_time, mobile_settings,
			quick_actions, keyboard_shortcuts, created_at, updated_at
		FROM user_preferences 
		WHERE user_id = $1`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.Theme, &prefs.CompactMode, &prefs.SidebarCollapsed,
		&prefs.Language, &prefs.Timezone, &prefs.DateFormat, &prefs.TimeFormat, &prefs.Currency,
		&prefs.Notifications, &prefs.Dashboard, &prefs.Privacy, &prefs.Accessibility,
		&prefs.EmailDigestFrequency, &prefs.EmailDigestTime, &prefs.MobileSettings,
		&prefs.QuickActions, &prefs.KeyboardShortcuts, &prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("User preferences not found")
	}
	if err != nil {
		logger.Error("Failed to get user preferences", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	return prefs, nil
}

func (r *UserRepository) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*UserWithPreferences, int, error) {
	countQuery := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, companyID).Scan(&total); err != nil {
		logger.Error("Failed to count users", err, "company_id", companyID)
		return nil, 0, errors.Internal(err)
	}

	query := `
		SELECT id, company_id, role, department, designation, employee_id,
			email, password_hash, email_verified, email_verified_at,
			two_factor_enabled, two_factor_secret, two_factor_recovery_codes, two_factor_verified_at,
			full_name, phone, timezone, language, date_of_birth, gender,
			avatar_url, avatar_thumbnail_url, permissions, status, is_active,
			last_login_at, last_login_ip, last_active_at, failed_login_attempts, locked_until,
			password_changed_at, must_change_password, invited_by, invited_at, joined_at,
			created_at, updated_at, deleted_at
		FROM users 
		WHERE company_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, companyID, limit, offset)
	if err != nil {
		logger.Error("Failed to list users", err, "company_id", companyID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var users []*UserWithPreferences
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.CompanyID, &user.Role, &user.Department, &user.Designation, &user.EmployeeID,
			&user.Email, &user.PasswordHash, &user.EmailVerified, &user.EmailVerifiedAt,
			&user.TwoFactorEnabled, &user.TwoFactorSecret, &user.TwoFactorRecoveryCodes, &user.TwoFactorVerifiedAt,
			&user.FullName, &user.Phone, &user.Timezone, &user.Language, &user.DateOfBirth, &user.Gender,
			&user.AvatarURL, &user.AvatarThumbnailURL, &user.Permissions, &user.Status, &user.IsActive,
			&user.LastLoginAt, &user.LastLoginIP, &user.LastActiveAt, &user.FailedLoginAttempts, &user.LockedUntil,
			&user.PasswordChangedAt, &user.MustChangePassword, &user.InvitedBy, &user.InvitedAt, &user.JoinedAt,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		); err != nil {
			logger.Error("Failed to scan user", err)
			return nil, 0, errors.Internal(err)
		}

		prefs, _ := r.getPreferences(ctx, user.ID)
		users = append(users, &UserWithPreferences{User: user, Preferences: prefs})
	}

	return users, total, nil
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
// UPDATE
// ============================================================================

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users SET
			role = $2, department = $3, designation = $4, employee_id = $5,
			email = $6, full_name = $7, phone = $8, timezone = $9, language = $10,
			date_of_birth = $11, gender = $12, avatar_url = $13, avatar_thumbnail_url = $14,
			permissions = $15, status = $16, is_active = $17, must_change_password = $18,
			updated_at = $19
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		user.ID, user.Role, user.Department, user.Designation, user.EmployeeID,
		user.Email, user.FullName, user.Phone, user.Timezone, user.Language,
		user.DateOfBirth, user.Gender, user.AvatarURL, user.AvatarThumbnailURL,
		user.Permissions, user.Status, user.IsActive, user.MustChangePassword,
		user.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update user", err, "user_id", user.ID)
		return errors.Wrap(err, 500, "Failed to update user")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, password_changed_at = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	now := time.Now()

	result, err := r.db.Exec(ctx, query, userID, passwordHash, now, now)
	if err != nil {
		logger.Error("Failed to update password", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE users SET email_verified = true, email_verified_at = $2, status = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now, models.UserStatusActive, now)
	if err != nil {
		logger.Error("Failed to mark email verified", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) UpdateLoginAttempts(ctx context.Context, userID uuid.UUID, failedAttempts int, lockedUntil *time.Time, status models.UserStatus) error {
	query := `UPDATE users SET failed_login_attempts = $2, locked_until = $3, status = $4, updated_at = $5 WHERE id = $1 AND deleted_at IS NULL`
	now := time.Now()

	result, err := r.db.Exec(ctx, query, userID, failedAttempts, lockedUntil, status, now)
	if err != nil {
		logger.Error("Failed to update login attempts", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	now := time.Now()
	query := `UPDATE users SET last_login_at = $2, last_login_ip = $3, last_active_at = $4, failed_login_attempts = 0, locked_until = NULL, updated_at = $5 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now, ip, now, now)
	if err != nil {
		logger.Error("Failed to update last login", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) UpdateActivity(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE users SET last_active_at = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.Exec(ctx, query, userID, now, now)
	if err != nil {
		logger.Error("Failed to update activity", err, "user_id", userID)
		return errors.Internal(err)
	}

	return nil
}

func (r *UserRepository) Enable2FA(ctx context.Context, userID uuid.UUID, secret string, recoveryCodes []string) error {
	now := time.Now()
	query := `UPDATE users SET two_factor_enabled = true, two_factor_secret = $2, two_factor_recovery_codes = $3, two_factor_verified_at = $4, updated_at = $5 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, secret, recoveryCodes, now, now)
	if err != nil {
		logger.Error("Failed to enable 2FA", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

func (r *UserRepository) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE users SET two_factor_enabled = false, two_factor_secret = NULL, two_factor_recovery_codes = NULL, two_factor_verified_at = NULL, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now)
	if err != nil {
		logger.Error("Failed to disable 2FA", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found")
	}

	return nil
}

// ============================================================================
// PREFERENCES UPDATE
// ============================================================================

func (r *UserRepository) UpdatePreferences(ctx context.Context, prefs *models.UserPreferences) error {
	prefs.UpdatedAt = time.Now()
	query := `
		UPDATE user_preferences SET
			theme = $2, compact_mode = $3, sidebar_collapsed = $4,
			language = $5, timezone = $6, date_format = $7, time_format = $8, currency = $9,
			notifications = $10, dashboard = $11, privacy = $12, accessibility = $13,
			email_digest_frequency = $14, email_digest_time = $15, mobile_settings = $16,
			quick_actions = $17, keyboard_shortcuts = $18, updated_at = $19
		WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query,
		prefs.UserID, prefs.Theme, prefs.CompactMode, prefs.SidebarCollapsed,
		prefs.Language, prefs.Timezone, prefs.DateFormat, prefs.TimeFormat, prefs.Currency,
		prefs.Notifications, prefs.Dashboard, prefs.Privacy, prefs.Accessibility,
		prefs.EmailDigestFrequency, prefs.EmailDigestTime, prefs.MobileSettings,
		prefs.QuickActions, prefs.KeyboardShortcuts, prefs.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update preferences", err, "user_id", prefs.UserID)
		return errors.Wrap(err, 500, "Failed to update preferences")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User preferences not found")
	}

	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *UserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET deleted_at = $2, status = $3, is_active = false, email = CONCAT(email, '.deleted.', $4), updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, userID, now, models.UserStatusDeleted, now.Unix(), now)
	if err != nil {
		logger.Error("Failed to soft delete user", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found or already deleted")
	}

	return nil
}

func (r *UserRepository) Restore(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET deleted_at = NULL, status = $2, is_active = true, 
		    email = REGEXP_REPLACE(email, '\.deleted\.\d+$', ''), updated_at = $3
		WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := r.db.Exec(ctx, query, userID, models.UserStatusActive, now)
	if err != nil {
		logger.Error("Failed to restore user", err, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("User not found or not deleted")
	}

	return nil
}