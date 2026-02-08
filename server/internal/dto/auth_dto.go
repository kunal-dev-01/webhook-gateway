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
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
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
	User             *models.PublicUser `json:"user"`
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
	Secret        string   `json:"secret"`
	QRCodeURL     string   `json:"qr_code_url"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// Enable2FARequest represents enable 2FA request
type Enable2FARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// Verify2FARequest represents 2FA code verification during login
type Verify2FARequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required,len=6"`
}

// Disable2FARequest represents disable 2FA request
type Disable2FARequest struct {
	Password string `json:"password" validate:"required"`
}

// ============================================================================
// Session Management
// ============================================================================

// SessionResponse represents a user session
type SessionResponse struct {
	SessionID    string    `json:"session_id"`
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
type CreateUserRequest struct {
	Email    string            `json:"email" validate:"required,email"`
	Password string            `json:"password" validate:"required,min=8,max=72"`
	FullName string            `json:"full_name" validate:"required,min=2,max=255"`
	Phone    *string           `json:"phone" validate:"omitempty,e164"`
	Role     models.UserRole   `json:"role" validate:"required,oneof=admin hr accountant viewer"`
	Timezone string            `json:"timezone" validate:"omitempty"`
	Language string            `json:"language" validate:"omitempty,len=2"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FullName *string             `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone    *string             `json:"phone" validate:"omitempty,e164"`
	Role     *models.UserRole    `json:"role" validate:"omitempty,oneof=admin hr accountant viewer"`
	Status   *models.UserStatus  `json:"status" validate:"omitempty,oneof=active inactive suspended"`
	Timezone *string             `json:"timezone" validate:"omitempty"`
	Language *string             `json:"language" validate:"omitempty,len=2"`
}

// UserResponse represents user details
type UserResponse struct {
	ID               uuid.UUID          `json:"id"`
	Email            string             `json:"email"`
	FullName         string             `json:"full_name"`
	Phone            *string            `json:"phone,omitempty"`
	Role             models.UserRole    `json:"role"`
	Status           models.UserStatus  `json:"status"`
	EmailVerified    bool               `json:"email_verified"`
	TwoFactorEnabled bool               `json:"two_factor_enabled"`
	LastLoginAt      *time.Time         `json:"last_login_at,omitempty"`
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