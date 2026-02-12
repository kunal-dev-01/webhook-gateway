package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// Registration & Login
// ============================================================================

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
	FullName    string `json:"full_name" validate:"required,min=2,max=255"`
	CompanyName string `json:"company_name" validate:"required,min=2,max=255"` // NEW: For new company creation
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents successful login
type LoginResponse struct {
	AccessToken      string             `json:"access_token"`
	RefreshToken     string             `json:"-"` // Hidden from JSON, used for cookie
	ExpiresIn        int64              `json:"expires_in"`
	TokenType        string             `json:"token_type"`
	User             *models.PublicUser `json:"user"` // Already has CompanyID
	Require2FA       bool               `json:"require_2fa,omitempty"`
	TempToken        string             `json:"temp_token,omitempty"` // For 2FA verification
}

// ============================================================================
// Email Verification
// ============================================================================

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ResendVerificationRequest represents resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ============================================================================
// Password Reset
// ============================================================================

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ============================================================================
// Two-Factor Authentication
// ============================================================================

// Setup2FAResponse represents 2FA setup response
type Setup2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// Enable2FARequest: Client sirf Code bhejega
type Enable2FARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// Enable2FAResponse: Success hone par Recovery Codes milenge
type Enable2FAResponse struct {
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// Verify2FARequest represents 2FA code verification during login
type Verify2FARequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required,min=6,max=8"`
}

// Disable2FARequest represents disable 2FA request
type Disable2FARequest struct {
	Password string `json:"password" validate:"required"`
}

// ============================================================================
// Session Management
// ============================================================================

// SessionResponse represents a user session
// UPDATED: Added CompanyID for context
type SessionResponse struct {
	SessionID    string    `json:"session_id"`
	CompanyID    uuid.UUID `json:"company_id"`    // NEW: Multi-tenant context
	DeviceInfo   string    `json:"device_info,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// ListSessionsResponse represents list of user sessions
type ListSessionsResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
	Total    int                `json:"total"`
}

// ============================================================================
// User Management (Admin)
// ============================================================================

// CreateUserRequest represents admin creating a user
// NOTE: CompanyID comes from auth context, not from request
type CreateUserRequest struct {
	Email       string          `json:"email" validate:"required,email"`
	Password    string          `json:"password" validate:"required,min=8,max=72"`
	FullName    string          `json:"full_name" validate:"required,min=2,max=255"`
	Phone       *string         `json:"phone" validate:"omitempty,e164"`
	Role        models.UserRole `json:"role" validate:"required,oneof=admin manager hr accountant viewer"`
	Department  *string         `json:"department" validate:"omitempty,max=100"`
	Designation *string         `json:"designation" validate:"omitempty,max=100"`
	EmployeeID  *string         `json:"employee_id" validate:"omitempty,max=50"`
	Timezone    string          `json:"timezone" validate:"omitempty"`
	Language    string          `json:"language" validate:"omitempty,len=2"`
}

// InviteUserRequest represents user invitation
// NEW: For inviting users to company
type InviteUserRequest struct {
	Email       string          `json:"email" validate:"required,email"`
	FullName    string          `json:"full_name" validate:"required,min=2,max=255"`
	Role        models.UserRole `json:"role" validate:"required,oneof=admin manager hr accountant viewer"`
	Department  *string         `json:"department" validate:"omitempty,max=100"`
	Designation *string         `json:"designation" validate:"omitempty,max=100"`
}

// ============================================================================
// Current User (/auth/me)
// ============================================================================

// MeResponse represents current logged-in user
// UPDATED: Added CompanyID and new fields
type MeResponse struct {
	ID                uuid.UUID           `json:"id"`
	CompanyID         uuid.UUID           `json:"company_id"` // NEW: Multi-tenant context
	Email             string              `json:"email"`
	FullName          string              `json:"full_name"`
	Phone             *string             `json:"phone,omitempty"`
	Role              models.UserRole     `json:"role"`
	Department        *string             `json:"department,omitempty"`
	Designation       *string             `json:"designation,omitempty"`
	EmployeeID        *string             `json:"employee_id,omitempty"`
	AvatarURL         *string             `json:"avatar_url,omitempty"`
	Status            models.UserStatus   `json:"status"`
	EmailVerified     bool                `json:"email_verified"`
	TwoFactorEnabled  bool                `json:"two_factor_enabled"`
	Permissions       models.Permissions  `json:"permissions"`        // NEW: Granular permissions
	LastLoginAt       *time.Time          `json:"last_login_at,omitempty"`
	LastActiveAt      *time.Time          `json:"last_active_at,omitempty"`
	PasswordChangedAt time.Time           `json:"password_changed_at"`
	CreatedAt         time.Time           `json:"created_at"`
}

// UpdateMeRequest represents current user profile update
// NEW: Self-service profile updates
type UpdateMeRequest struct {
	FullName    *string    `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone       *string    `json:"phone" validate:"omitempty,e164"`
	Timezone    *string    `json:"timezone" validate:"omitempty"`
	Language    *string    `json:"language" validate:"omitempty,len=2"`
	DateOfBirth *time.Time `json:"date_of_birth" validate:"omitempty"`
	Gender      *string    `json:"gender" validate:"omitempty,oneof=male female other prefer_not_to_say"`
}

// UpdateUserRequest represents user update request (by admin)
// UPDATED: Added new fields
type UpdateUserRequest struct {
	FullName    *string            `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone       *string            `json:"phone" validate:"omitempty,e164"`
	Role        *models.UserRole   `json:"role" validate:"omitempty,oneof=admin manager hr accountant viewer"`
	Department  *string            `json:"department" validate:"omitempty,max=100"`
	Designation *string            `json:"designation" validate:"omitempty,max=100"`
	EmployeeID  *string            `json:"employee_id" validate:"omitempty,max=50"`
	Status      *models.UserStatus `json:"status" validate:"omitempty,oneof=active inactive suspended"`
	Timezone    *string            `json:"timezone" validate:"omitempty"`
	Language    *string            `json:"language" validate:"omitempty,len=2"`
}

// UserResponse represents user details
// UPDATED: Added CompanyID and new fields
type UserResponse struct {
	ID               uuid.UUID          `json:"id"`
	CompanyID        uuid.UUID          `json:"company_id"` // NEW: Multi-tenant context
	Email            string             `json:"email"`
	FullName         string             `json:"full_name"`
	Phone            *string            `json:"phone,omitempty"`
	Role             models.UserRole    `json:"role"`
	Department       *string            `json:"department,omitempty"`
	Designation      *string            `json:"designation,omitempty"`
	EmployeeID       *string            `json:"employee_id,omitempty"`
	AvatarURL        *string            `json:"avatar_url,omitempty"`
	Status           models.UserStatus  `json:"status"`
	EmailVerified    bool               `json:"email_verified"`
	TwoFactorEnabled bool               `json:"two_factor_enabled"`
	Permissions      models.Permissions `json:"permissions"` // NEW: Show permissions
	LastLoginAt      *time.Time         `json:"last_login_at,omitempty"`
	LastActiveAt     *time.Time         `json:"last_active_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

// ListUsersResponse represents paginated users list
type ListUsersResponse struct {
	Users []*UserResponse `json:"users"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

// ============================================================================
// User Statistics (Admin Dashboard)
// ============================================================================

// UserStatsResponse represents user statistics
// NEW: For admin dashboard
type UserStatsResponse struct {
	TotalUsers      int                        `json:"total_users"`
	ActiveUsers     int                        `json:"active_users"`
	InactiveUsers   int                        `json:"inactive_users"`
	PendingUsers    int                        `json:"pending_users"`
	UsersByRole     map[models.UserRole]int    `json:"users_by_role"`
	UsersByDept     map[string]int             `json:"users_by_department"`
	RecentLogins    int                        `json:"recent_logins_24h"`
	With2FA         int                        `json:"users_with_2fa"`
}

// ============================================================================
// Generic Responses
// ============================================================================

// MessageResponse represents a simple message response
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// ============================================================================
// Bulk Operations (Admin)
// ============================================================================

// BulkUpdateStatusRequest for bulk status updates
// NEW: Batch operations
type BulkUpdateStatusRequest struct {
	UserIDs []uuid.UUID        `json:"user_ids" validate:"required,min=1,max=100"`
	Status  models.UserStatus  `json:"status" validate:"required,oneof=active inactive suspended"`
}

// BulkDeleteRequest for bulk soft delete
type BulkDeleteRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" validate:"required,min=1,max=100"`
}

// BulkOperationResponse for bulk operation results
type BulkOperationResponse struct {
	Success      bool        `json:"success"`
	TotalCount   int         `json:"total_count"`
	SuccessCount int         `json:"success_count"`
	FailedCount  int         `json:"failed_count"`
	FailedIDs    []uuid.UUID `json:"failed_ids,omitempty"`
	Message      string      `json:"message"`
}