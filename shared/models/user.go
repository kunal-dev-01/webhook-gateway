package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleHR         UserRole = "hr"
	RoleAccountant UserRole = "accountant"
	RoleViewer     UserRole = "viewer"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

// User represents a user in the system
type User struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Authentication
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"` // Never expose in JSON

	// Email verification
	EmailVerified   bool       `db:"email_verified" json:"email_verified"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`

	// Two-Factor Authentication
	TwoFactorEnabled       bool           `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret        *string        `db:"two_factor_secret" json:"-"` // Never expose
	TwoFactorRecoveryCodes pq.StringArray `db:"two_factor_recovery_codes" json:"-"` // Never expose
	TwoFactorVerifiedAt    *time.Time     `db:"two_factor_verified_at" json:"two_factor_verified_at,omitempty"`

	// Profile information
	FullName string  `db:"full_name" json:"full_name"`
	Phone    *string `db:"phone" json:"phone,omitempty"`
	Timezone string  `db:"timezone" json:"timezone"`
	Language string  `db:"language" json:"language"`

	// User role and status
	Role   UserRole   `db:"role" json:"role"`
	Status UserStatus `db:"status" json:"status"`

	// Account security
	LastLoginAt         *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP         *string    `db:"last_login_ip" json:"last_login_ip,omitempty"`
	FailedLoginAttempts int        `db:"failed_login_attempts" json:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until" json:"locked_until,omitempty"`
	PasswordChangedAt   time.Time  `db:"password_changed_at" json:"password_changed_at"`
	MustChangePassword  bool       `db:"must_change_password" json:"must_change_password"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// Password Methods
// ============================================================================

const bcryptCost = 12 // Recommended cost for production

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.PasswordChangedAt = time.Now()
	return nil
}

// CheckPassword verifies if the provided password matches the hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ============================================================================
// Account Status Methods
// ============================================================================

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// CanLogin checks if the user can login (not locked, not deleted, active)
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsLocked() && !u.IsDeleted()
}

// ============================================================================
// Email Verification Methods
// ============================================================================

// MarkEmailAsVerified marks the email as verified
func (u *User) MarkEmailAsVerified() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
}

// ============================================================================
// Two-Factor Authentication Methods
// ============================================================================

// Enable2FA enables two-factor authentication with the provided secret
func (u *User) Enable2FA(secret string, recoveryCodes []string) {
	now := time.Now()
	u.TwoFactorEnabled = true
	u.TwoFactorSecret = &secret
	u.TwoFactorRecoveryCodes = recoveryCodes
	u.TwoFactorVerifiedAt = &now
}

// Disable2FA disables two-factor authentication
func (u *User) Disable2FA() {
	u.TwoFactorEnabled = false
	u.TwoFactorSecret = nil
	u.TwoFactorRecoveryCodes = nil
	u.TwoFactorVerifiedAt = nil
}

// ============================================================================
// Login Attempt Methods
// ============================================================================

const (
	maxFailedAttempts = 5
	lockoutDuration   = 30 * time.Minute
)

// IncrementFailedAttempts increments failed login attempts and locks if needed
func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxFailedAttempts {
		lockUntil := time.Now().Add(lockoutDuration)
		u.LockedUntil = &lockUntil
		u.Status = StatusSuspended
	}
}

// ResetFailedAttempts resets failed login attempts
func (u *User) ResetFailedAttempts(loginIP string) {
	now := time.Now()
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
	u.LastLoginIP = &loginIP
	
	// Unlock if was suspended due to failed attempts
	if u.Status == StatusSuspended && (u.LockedUntil == nil || u.LockedUntil.Before(now)) {
		u.Status = StatusActive
	}
}

// ============================================================================
// Soft Delete Methods
// ============================================================================

// SoftDelete soft deletes the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.Status = StatusDeleted
	// Append timestamp to email to allow future signups with same email
	u.Email = u.Email + ".deleted." + time.Now().Format("20060102150405")
}

// Restore restores a soft-deleted user
func (u *User) Restore() error {
	if u.DeletedAt == nil {
		return fmt.Errorf("user is not deleted")
	}
	
	u.DeletedAt = nil
	u.Status = StatusActive
	
	// Remove .deleted.timestamp suffix from email
	re := regexp.MustCompile(`\.deleted\.\d+$`)
	u.Email = re.ReplaceAllString(u.Email, "")
	
	return nil
}

// ============================================================================
// Role Permission Methods
// ============================================================================

// HasRole checks if user has a specific role
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin || u.Role == RoleHR
}

// CanRunPayroll checks if user can run payroll
func (u *User) CanRunPayroll() bool {
	return u.Role == RoleAdmin || u.Role == RoleHR || u.Role == RoleAccountant
}

// CanViewReports checks if user can view reports
func (u *User) CanViewReports() bool {
	// All roles can view reports
	return true
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the user struct
func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	
	if !isValidEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	
	if u.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	
	if u.Phone != nil && *u.Phone != "" && !isValidPhone(*u.Phone) {
		return fmt.Errorf("invalid phone number format")
	}
	
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidPhone validates phone number format
func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// ============================================================================
// JSON Marshaling (Custom for security)
// ============================================================================

// MarshalJSON custom JSON marshaling to exclude sensitive fields
func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		PasswordHash           string         `json:"-"`
		TwoFactorSecret        *string        `json:"-"`
		TwoFactorRecoveryCodes pq.StringArray `json:"-"`
	}{
		Alias: (*Alias)(u),
	})
}

// ============================================================================
// Public Response Methods (Safe for API)
// ============================================================================

// PublicUser represents user data safe for public API responses
type PublicUser struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Phone           *string    `json:"phone,omitempty"`
	Role            UserRole   `json:"role"`
	Status          UserStatus `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ToPublic converts User to PublicUser (safe for API response)
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:               u.ID,
		Email:            u.Email,
		FullName:         u.FullName,
		Phone:            u.Phone,
		Role:             u.Role,
		Status:           u.Status,
		EmailVerified:    u.EmailVerified,
		TwoFactorEnabled: u.TwoFactorEnabled,
		LastLoginAt:      u.LastLoginAt,
		CreatedAt:        u.CreatedAt,
	}
}