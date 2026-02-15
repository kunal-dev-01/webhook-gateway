package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// OrganizationMemberRole represents user roles in an organization
type OrganizationMemberRole string

const (
	RoleOwner  OrganizationMemberRole = "owner"
	RoleAdmin  OrganizationMemberRole = "admin"
	RoleMember OrganizationMemberRole = "member"
	RoleViewer OrganizationMemberRole = "viewer"
)

// ============================================================================
// USER PREFERENCES
// ============================================================================

// UserPreferences represents user preferences stored as JSONB
type UserPreferences struct {
	Theme              string `json:"theme"`               // "light", "dark", "system"
	EmailNotifications bool   `json:"email_notifications"` // Email notifications enabled
	Timezone           string `json:"timezone"`            // User timezone (e.g., "UTC", "Asia/Kolkata")
}

// Value implements driver.Valuer for UserPreferences
func (p UserPreferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for UserPreferences
func (p *UserPreferences) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal UserPreferences value")
	}
	return json.Unmarshal(bytes, p)
}

// DefaultUserPreferences returns default user preferences
func DefaultUserPreferences() UserPreferences {
	return UserPreferences{
		Theme:              "system",
		EmailNotifications: true,
		Timezone:           "UTC",
	}
}

// ============================================================================
// USER MODEL
// ============================================================================

// User represents a user account in Upnixo
type User struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Authentication
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"` // Never expose in JSON

	// Profile
	FullName  string  `db:"full_name" json:"full_name"`
	AvatarURL *string `db:"avatar_url" json:"avatar_url,omitempty"`

	// User Preferences (JSONB)
	Preferences UserPreferences `db:"preferences" json:"preferences"`

	// Email verification (tokens stored in Redis)
	EmailVerified bool `db:"email_verified" json:"email_verified"`

	// Two-Factor Authentication (2FA)
	TwoFactorEnabled       bool           `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret        *string        `db:"two_factor_secret" json:"-"` // Encrypted TOTP secret
	TwoFactorBackupCodes   pq.StringArray `db:"two_factor_backup_codes" json:"-"` // Hashed backup codes
	TwoFactorConfirmedAt   *time.Time     `db:"two_factor_confirmed_at" json:"two_factor_confirmed_at,omitempty"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// PASSWORD METHODS
// ============================================================================

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if err := utils.ValidatePasswordSimple(password); err != nil {
		return err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	u.PasswordHash = hash
	return nil
}

// CheckPassword verifies if the provided password matches the hash
func (u *User) CheckPassword(password string) bool {
	return utils.CheckPassword(password, u.PasswordHash)
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

// MarkEmailAsVerified marks email as verified
func (u *User) MarkEmailAsVerified() {
	u.EmailVerified = true
}

// ============================================================================
// TWO-FACTOR AUTHENTICATION (2FA)
// ============================================================================

// Enable2FA enables two-factor authentication
func (u *User) Enable2FA(secret string, backupCodes []string) {
	now := time.Now()
	u.TwoFactorEnabled = true
	u.TwoFactorSecret = &secret
	u.TwoFactorBackupCodes = backupCodes
	u.TwoFactorConfirmedAt = &now
}

// Disable2FA disables two-factor authentication
func (u *User) Disable2FA() {
	u.TwoFactorEnabled = false
	u.TwoFactorSecret = nil
	u.TwoFactorBackupCodes = nil
	u.TwoFactorConfirmedAt = nil
}

// ============================================================================
// SOFT DELETE
// ============================================================================

// SoftDelete soft deletes the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	// Append timestamp to email to allow recreation
	u.Email = fmt.Sprintf("%s.deleted.%d", u.Email, now.Unix())
}

// Restore restores a soft-deleted user
func (u *User) Restore() error {
	if u.DeletedAt == nil {
		return fmt.Errorf("user is not deleted")
	}

	u.DeletedAt = nil

	// Remove .deleted.timestamp suffix from email
	re := regexp.MustCompile(`\.deleted\.\d+$`)
	u.Email = re.ReplaceAllString(u.Email, "")

	return nil
}

// IsDeleted checks if user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the user struct
func (u *User) Validate() error {
	if err := utils.ValidateRequired(u.Email, "email"); err != nil {
		return err
	}
	if err := utils.ValidateEmail(u.Email); err != nil {
		return err
	}
	if err := utils.ValidateRequired(u.FullName, "full name"); err != nil {
		return err
	}
	if err := utils.ValidateLength(u.FullName, 2, 255, "full name"); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE (SAFE FOR API)
// ============================================================================

// PublicUser represents safe user data for API responses
type PublicUser struct {
	ID                   uuid.UUID       `json:"id"`
	Email                string          `json:"email"`
	FullName             string          `json:"full_name"`
	AvatarURL            *string         `json:"avatar_url,omitempty"`
	EmailVerified        bool            `json:"email_verified"`
	TwoFactorEnabled     bool            `json:"two_factor_enabled"`
	TwoFactorConfirmedAt *time.Time      `json:"two_factor_confirmed_at,omitempty"`
	Preferences          UserPreferences `json:"preferences"`
	CreatedAt            time.Time       `json:"created_at"`
}

// ToPublic converts User to PublicUser (safe for API responses)
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:                   u.ID,
		Email:                u.Email,
		FullName:             u.FullName,
		AvatarURL:            u.AvatarURL,
		EmailVerified:        u.EmailVerified,
		TwoFactorEnabled:     u.TwoFactorEnabled,
		TwoFactorConfirmedAt: u.TwoFactorConfirmedAt,
		Preferences:          u.Preferences,
		CreatedAt:            u.CreatedAt,
	}
}