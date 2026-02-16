package dto

import (
	"time"

	"github.com/google/uuid"

	"shared/models"
)

// ============================================================================
// REGISTER
// ============================================================================

type RegisterRequest struct {
	FullName      string  `json:"full_name" validate:"required,min=2,max=255"`
	Email         string  `json:"email" validate:"required,email"`
	Password      string  `json:"password" validate:"required,min=8,max=72"`
	PlanTier      *string `json:"plan_tier,omitempty"`
	BillingPeriod *string `json:"billing_period,omitempty"`
}

type RegisterResponse struct {
	Message      string                `json:"message"`
}

// ============================================================================
// LOGIN
// ============================================================================

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	// If 2FA is DISABLED
	User         *UserResponse `json:"user,omitempty"`
	AccessToken  string        `json:"access_token,omitempty"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in,omitempty"`

	// If 2FA is ENABLED
	TwoFactorRequired bool   `json:"two_factor_required"`
	TempToken         string `json:"temp_token,omitempty"` // Valid for 5 minutes
}

// ============================================================================
// ORGANIZATION SELECTION (After Login)
// ============================================================================

type SelectOrganizationRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
}

type SelectOrganizationResponse struct {
	User         *UserResponse         `json:"user"`
	Organization *OrganizationResponse `json:"organization"`
	AccessToken  string                `json:"access_token"`
	RefreshToken string                `json:"refresh_token"`
	ExpiresIn    int64                 `json:"expires_in"`
}

// ============================================================================
// 2FA VERIFICATION
// ============================================================================

type TwoFactorVerifyRequest struct {
	TempToken  string `json:"temp_token" validate:"required"`
	Code       string `json:"code" validate:"required,len=6"` // TOTP code
	IsRecovery bool   `json:"is_recovery"`                    // true if using backup code
}

type TwoFactorVerifyResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}

// ============================================================================
// REFRESH TOKEN
// ============================================================================

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ============================================================================
// EMAIL VERIFICATION
// ============================================================================

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendVerificationResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// PASSWORD RESET
// ============================================================================

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// 2FA MANAGEMENT
// ============================================================================

// Get 2FA Status
type TwoFactorStatusResponse struct {
	Enabled     bool       `json:"enabled"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// Enable 2FA (Step 1)
type Enable2FAResponse struct {
	QRCodeURL string `json:"qr_code_url"` // Base64 data URL
	Secret    string `json:"secret"`      // For manual entry
}

// Confirm 2FA (Step 2)
type Confirm2FARequest struct {
	Code string `json:"code" validate:"required,len=6"` // TOTP code to confirm setup
}

type Confirm2FAResponse struct {
	Message     string   `json:"message"`
	BackupCodes []string `json:"backup_codes"` // Show once, user must save
}

// Disable 2FA
type Disable2FARequest struct {
	Password     string `json:"password" validate:"required"` // Require password for security
	Code         string `json:"code,omitempty"`               // TOTP code
	RecoveryCode string `json:"recovery_code,omitempty"`      // OR recovery code
}

type Disable2FAResponse struct {
	Message string `json:"message"`
}

// Regenerate Backup Codes
type RegenerateBackupCodesRequest struct {
	Password string `json:"password" validate:"required"`
}

type RegenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// ============================================================================
// LOGOUT
// ============================================================================

type LogoutRequest struct {
	AllDevices bool `json:"all_devices"` // true = logout all sessions
}

type LogoutResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// SWITCH ORGANIZATION
// ============================================================================

type SwitchOrganizationRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
}

type SwitchOrganizationResponse struct {
	Organization *OrganizationResponse `json:"organization"`
	AccessToken  string                `json:"access_token"`
	RefreshToken string                `json:"refresh_token"`
	ExpiresIn    int64                 `json:"expires_in"`
}

// ============================================================================
// GET CURRENT USER
// ============================================================================

type GetMeResponse struct {
	User         *UserResponse         `json:"user"`
	Organization *OrganizationResponse `json:"organization"` // Current active organization
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

// UserResponse represents user data in auth responses
type UserResponse struct {
	ID                   uuid.UUID              `json:"id"`
	Email                string                 `json:"email"`
	FullName             string                 `json:"full_name"`
	AvatarURL            *string                `json:"avatar_url,omitempty"`
	EmailVerified        bool                   `json:"email_verified"`
	TwoFactorEnabled     bool                   `json:"two_factor_enabled"`
	TwoFactorConfirmedAt *time.Time             `json:"two_factor_confirmed_at,omitempty"`
	Preferences          models.UserPreferences `json:"preferences"`
	CreatedAt            time.Time              `json:"created_at"`
}

// OrganizationResponse represents organization data in auth responses
type OrganizationResponse struct {
	ID                 uuid.UUID                     `json:"id"`
	Name               string                        `json:"name"`
	Slug               string                        `json:"slug"`
	OwnerID            uuid.UUID                     `json:"owner_id"`
	Role               string                        `json:"role"` // User's role in this org
	PlanTier           models.PlanTier               `json:"plan_tier"`
	BillingPeriod      models.BillingPeriod          `json:"billing_period"`
	SubscriptionStatus models.SubscriptionStatus     `json:"subscription_status"`
	MonthlyEventsLimit int64                         `json:"monthly_events_limit"`
	MonthlyEventsUsed  int64                         `json:"monthly_events_used"`
	RetentionDays      int                           `json:"retention_days"`
	TeamSeatsLimit     int                           `json:"team_seats_limit"`
	RateLimitPerSecond int                           `json:"rate_limit_per_second"`
	ProjectsLimit      int                           `json:"projects_limit"`
	CreatedAt          time.Time                     `json:"created_at"`
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// ToUserResponse converts User model to UserResponse DTO
func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:                   user.ID,
		Email:                user.Email,
		FullName:             user.FullName,
		AvatarURL:            user.AvatarURL,
		EmailVerified:        user.EmailVerified,
		TwoFactorEnabled:     user.TwoFactorEnabled,
		TwoFactorConfirmedAt: user.TwoFactorConfirmedAt,
		Preferences:          user.Preferences,
		CreatedAt:            user.CreatedAt,
	}
}

// ToOrganizationResponse converts Organization model to OrganizationResponse DTO
func ToOrganizationResponse(org *models.Organization, userRole string) *OrganizationResponse {
	return &OrganizationResponse{
		ID:                 org.ID,
		Name:               org.Name,
		Slug:               org.Slug,
		OwnerID:            org.OwnerID,
		Role:               userRole, // User's role in this organization
		PlanTier:           org.PlanTier,
		BillingPeriod:      org.BillingPeriod,
		SubscriptionStatus: org.SubscriptionStatus,
		MonthlyEventsLimit: org.MonthlyEventsLimit,
		MonthlyEventsUsed:  org.MonthlyEventsUsed,
		RetentionDays:      org.RetentionDays,
		TeamSeatsLimit:     org.TeamSeatsLimit,
		RateLimitPerSecond: org.RateLimitPerSecond,
		ProjectsLimit:      org.ProjectsLimit,
		CreatedAt:          org.CreatedAt,
	}
}