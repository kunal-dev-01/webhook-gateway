package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusInactive            UserStatus = "inactive"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusDeleted             UserStatus = "deleted"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

// UserRole represents user roles in the system
type UserRole string

const (
	UserRoleOwner      UserRole = "owner"
	UserRoleAdmin      UserRole = "admin"
	UserRoleManager    UserRole = "manager"
	UserRoleHR         UserRole = "hr"
	UserRoleAccountant UserRole = "accountant"
	UserRoleViewer     UserRole = "viewer"
)

// Permissions represents granular user permissions
type Permissions struct {
	Employees PermissionSet `json:"employees"`
	Payroll   PayrollPerms  `json:"payroll"`
	Attendance PermissionSet `json:"attendance"`
	Leaves    LeavePerms    `json:"leaves"`
	Reports   ReportPerms   `json:"reports"`
	Settings  PermissionSet `json:"settings"`
}

type PermissionSet struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

type PayrollPerms struct {
	View    bool `json:"view"`
	Create  bool `json:"create"`
	Edit    bool `json:"edit"`
	Delete  bool `json:"delete"`
	Process bool `json:"process"`
}

type LeavePerms struct {
	View    bool `json:"view"`
	Create  bool `json:"create"`
	Edit    bool `json:"edit"`
	Delete  bool `json:"delete"`
	Approve bool `json:"approve"`
}

type ReportPerms struct {
	View   bool `json:"view"`
	Export bool `json:"export"`
}

// Value implements driver.Valuer for Permissions
func (p Permissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for Permissions
func (p *Permissions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal Permissions value")
	}
	return json.Unmarshal(bytes, p)
}

// ============================================================================
// USER MODEL
// ============================================================================

// User represents a user in the system (can login and manage)
type User struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Company relationship
	CompanyID uuid.UUID `db:"company_id" json:"company_id"`

	// Role & Department
	Role        UserRole  `db:"role" json:"role"`
	Department  *string   `db:"department" json:"department,omitempty"`
	Designation *string   `db:"designation" json:"designation,omitempty"`
	EmployeeID  *string   `db:"employee_id" json:"employee_id,omitempty"`

	// Authentication
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"` // Never expose

	// Email verification (tokens in Redis)
	EmailVerified   bool       `db:"email_verified" json:"email_verified"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`

	// Two-Factor Authentication
	TwoFactorEnabled       bool           `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret        *string        `db:"two_factor_secret" json:"-"`
	TwoFactorRecoveryCodes pq.StringArray `db:"two_factor_recovery_codes" json:"-"`
	TwoFactorVerifiedAt    *time.Time     `db:"two_factor_verified_at" json:"two_factor_verified_at,omitempty"`

	// Profile information
	FullName    string     `db:"full_name" json:"full_name"`
	Phone       *string    `db:"phone" json:"phone,omitempty"`
	Timezone    string     `db:"timezone" json:"timezone"`
	Language    string     `db:"language" json:"language"`
	DateOfBirth *time.Time `db:"date_of_birth" json:"date_of_birth,omitempty"`
	Gender      *string    `db:"gender" json:"gender,omitempty"`

	// Avatar/Profile Picture
	AvatarURL          *string `db:"avatar_url" json:"avatar_url,omitempty"`
	AvatarThumbnailURL *string `db:"avatar_thumbnail_url" json:"avatar_thumbnail_url,omitempty"`

	// Permissions (JSONB)
	Permissions Permissions `db:"permissions" json:"permissions"`

	// Account status
	Status   UserStatus `db:"status" json:"status"`
	IsActive bool       `db:"is_active" json:"is_active"`

	// Account security
	LastLoginAt         *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP         *string    `db:"last_login_ip" json:"last_login_ip,omitempty"`
	LastActiveAt        *time.Time `db:"last_active_at" json:"last_active_at,omitempty"`
	FailedLoginAttempts int        `db:"failed_login_attempts" json:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until" json:"locked_until,omitempty"`
	PasswordChangedAt   time.Time  `db:"password_changed_at" json:"password_changed_at"`
	MustChangePassword  bool       `db:"must_change_password" json:"must_change_password"`

	// Invitation tracking
	InvitedBy *uuid.UUID `db:"invited_by" json:"invited_by,omitempty"`
	InvitedAt *time.Time `db:"invited_at" json:"invited_at,omitempty"`
	JoinedAt  *time.Time `db:"joined_at" json:"joined_at,omitempty"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// PASSWORD METHODS
// ============================================================================

const bcryptCost = 12 // Production recommended

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
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

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 72 { // bcrypt limit
		return fmt.Errorf("password must be less than 72 characters")
	}
	
	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Check for at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	
	if !hasNumber || !hasLetter {
		return fmt.Errorf("password must contain both letters and numbers")
	}
	
	return nil
}

// ============================================================================
// ACCOUNT STATUS METHODS
// ============================================================================

// IsAccountActive checks if user account is active
func (u *User) IsAccountActive() bool {
	return u.Status == UserStatusActive && u.IsActive && u.DeletedAt == nil
}

// IsLocked checks if account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// CanLogin checks if user can login
func (u *User) CanLogin() bool {
	return u.IsAccountActive() && !u.IsLocked() && u.EmailVerified
}

// ============================================================================
// ROLE & PERMISSION METHODS
// ============================================================================

// IsOwner checks if user is company owner
func (u *User) IsOwner() bool {
	return u.Role == UserRoleOwner
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOwner
}

// HasPermission checks if user has specific permission
func (u *User) HasPermission(module, action string) bool {
	switch module {
	case "employees":
		return u.checkPermission(u.Permissions.Employees, action)
	case "attendance":
		return u.checkPermission(u.Permissions.Attendance, action)
	case "settings":
		return u.checkPermission(u.Permissions.Settings, action)
	case "payroll":
		return u.checkPayrollPermission(action)
	case "leaves":
		return u.checkLeavePermission(action)
	case "reports":
		return u.checkReportPermission(action)
	default:
		return false
	}
}

func (u *User) checkPermission(perms PermissionSet, action string) bool {
	switch action {
	case "view":
		return perms.View
	case "create":
		return perms.Create
	case "edit":
		return perms.Edit
	case "delete":
		return perms.Delete
	default:
		return false
	}
}

func (u *User) checkPayrollPermission(action string) bool {
	switch action {
	case "view":
		return u.Permissions.Payroll.View
	case "create":
		return u.Permissions.Payroll.Create
	case "edit":
		return u.Permissions.Payroll.Edit
	case "delete":
		return u.Permissions.Payroll.Delete
	case "process":
		return u.Permissions.Payroll.Process
	default:
		return false
	}
}

func (u *User) checkLeavePermission(action string) bool {
	switch action {
	case "view":
		return u.Permissions.Leaves.View
	case "create":
		return u.Permissions.Leaves.Create
	case "edit":
		return u.Permissions.Leaves.Edit
	case "delete":
		return u.Permissions.Leaves.Delete
	case "approve":
		return u.Permissions.Leaves.Approve
	default:
		return false
	}
}

func (u *User) checkReportPermission(action string) bool {
	switch action {
	case "view":
		return u.Permissions.Reports.View
	case "export":
		return u.Permissions.Reports.Export
	default:
		return false
	}
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

// MarkEmailAsVerified marks email as verified
func (u *User) MarkEmailAsVerified() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
	u.Status = UserStatusActive
}

// ============================================================================
// TWO-FACTOR AUTHENTICATION
// ============================================================================

// Enable2FA enables two-factor authentication
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
// LOGIN ATTEMPTS
// ============================================================================

const (
	maxFailedAttempts = 5
	lockoutDuration   = 30 * time.Minute
)

// IncrementFailedAttempts increments failed login attempts
func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= maxFailedAttempts {
		lockUntil := time.Now().Add(lockoutDuration)
		u.LockedUntil = &lockUntil
		u.Status = UserStatusSuspended
	}
}

// ResetFailedAttempts resets failed attempts on successful login
func (u *User) ResetFailedAttempts(loginIP string) {
	now := time.Now()
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
	u.LastLoginIP = &loginIP
	u.LastActiveAt = &now

	// Unlock if was suspended
	if u.Status == UserStatusSuspended {
		u.Status = UserStatusActive
	}
}

// UpdateActivity updates last active timestamp
func (u *User) UpdateActivity() {
	now := time.Now()
	u.LastActiveAt = &now
}

// ============================================================================
// SOFT DELETE
// ============================================================================

// SoftDelete soft deletes the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.Status = UserStatusDeleted
	u.IsActive = false
	// Append timestamp to email
	u.Email = fmt.Sprintf("%s.deleted.%d", u.Email, now.Unix())
}

// Restore restores a soft-deleted user
func (u *User) Restore() error {
	if u.DeletedAt == nil {
		return fmt.Errorf("user is not deleted")
	}

	u.DeletedAt = nil
	u.Status = UserStatusActive
	u.IsActive = true

	// Remove .deleted.timestamp suffix
	re := regexp.MustCompile(`\.deleted\.\d+$`)
	u.Email = re.ReplaceAllString(u.Email, "")

	return nil
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the user struct
func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !utils.IsValidEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	if u.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if u.Phone != nil && *u.Phone != "" && !utils.IsValidIndianPhone(*u.Phone) {
		return fmt.Errorf("invalid phone number format")
	}
	if u.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicUser represents safe user data for API responses
type PublicUser struct {
	ID                 uuid.UUID   `json:"id"`
	CompanyID          uuid.UUID   `json:"company_id"`
	Email              string      `json:"email"`
	FullName           string      `json:"full_name"`
	Phone              *string     `json:"phone,omitempty"`
	Role               UserRole    `json:"role"`
	Department         *string     `json:"department,omitempty"`
	Designation        *string     `json:"designation,omitempty"`
	AvatarURL          *string     `json:"avatar_url,omitempty"`
	Status             UserStatus  `json:"status"`
	EmailVerified      bool        `json:"email_verified"`
	TwoFactorEnabled   bool        `json:"two_factor_enabled"`
	LastLoginAt        *time.Time  `json:"last_login_at,omitempty"`
	Permissions        Permissions `json:"permissions"`
	CreatedAt          time.Time   `json:"created_at"`
}

// ToPublic converts User to PublicUser
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:               u.ID,
		CompanyID:        u.CompanyID,
		Email:            u.Email,
		FullName:         u.FullName,
		Phone:            u.Phone,
		Role:             u.Role,
		Department:       u.Department,
		Designation:      u.Designation,
		AvatarURL:        u.AvatarURL,
		Status:           u.Status,
		EmailVerified:    u.EmailVerified,
		TwoFactorEnabled: u.TwoFactorEnabled,
		LastLoginAt:      u.LastLoginAt,
		Permissions:      u.Permissions,
		CreatedAt:        u.CreatedAt,
	}
}