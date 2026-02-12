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
	CompanyName string `json:"company_name" validate:"required,min=2,max=100"`
	FullName    string `json:"full_name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
}

type RegisterResponse struct {
	Message string `json:"message"`
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
	User         *UserResponse        `json:"user,omitempty"`
	Preferences  *PreferencesResponse `json:"preferences,omitempty"`
	AccessToken  string               `json:"access_token,omitempty"`
	RefreshToken string               `json:"refresh_token,omitempty"`
	ExpiresIn    int64                `json:"expires_in,omitempty"`

	// If 2FA is ENABLED
	TwoFactorRequired bool   `json:"two_factor_required"`
	TempToken         string `json:"temp_token,omitempty"` // Valid for 5 minutes
}

// ============================================================================
// 2FA VERIFICATION
// ============================================================================

type TwoFactorVerifyRequest struct {
	TempToken    string `json:"temp_token" validate:"required"`
	Code         string `json:"code" validate:"required"`         // TOTP code or recovery code
	IsRecovery   bool   `json:"is_recovery"`                      // true if using recovery code
}

type TwoFactorVerifyResponse struct {
	User         *UserResponse        `json:"user"`
	Preferences  *PreferencesResponse `json:"preferences"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int64                `json:"expires_in"`
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

type Enable2FAResponse struct {
	QRCodeURL     string   `json:"qr_code_url"`
	Secret        string   `json:"secret"`         // For manual entry
}

type Confirm2FARequest struct {
	Code string `json:"code" validate:"required,len=6"` // TOTP code to confirm setup
}

type Confirm2FAResponse struct {
	Message       string   `json:"message"`
	RecoveryCodes []string `json:"recovery_codes"`
}

type Disable2FARequest struct {
	Password string `json:"password" validate:"required"` // Require password for security
}

type Disable2FAResponse struct {
	Message string `json:"message"`
}

type Regenerate2FARecoveryCodesRequest struct {
	Password string `json:"password" validate:"required"`
}

type Regenerate2FARecoveryCodesResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// ============================================================================
// LOGOUT
// ============================================================================

type LogoutRequest struct {
	SessionID string `json:"session_id,omitempty"` // Optional: logout specific session
	AllSessions bool `json:"all_sessions"`         // true = logout all devices
}

type LogoutResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// USER RESPONSE (for auth responses)
// ============================================================================

type UserResponse struct {
	ID               uuid.UUID              `json:"id"`
	CompanyID        uuid.UUID              `json:"company_id"`
	Email            string                 `json:"email"`
	FullName         string                 `json:"full_name"`
	Phone            *string                `json:"phone,omitempty"`
	Role             models.UserRole        `json:"role"`
	Department       *string                `json:"department,omitempty"`
	Designation      *string                `json:"designation,omitempty"`
	AvatarURL        *string                `json:"avatar_url,omitempty"`
	Status           models.UserStatus      `json:"status"`
	EmailVerified    bool                   `json:"email_verified"`
	TwoFactorEnabled bool                   `json:"two_factor_enabled"`
	LastLoginAt      *time.Time             `json:"last_login_at,omitempty"`
	Permissions      models.Permissions     `json:"permissions"`
	Timezone         string                 `json:"timezone"`
	Language         string                 `json:"language"`
	CreatedAt        time.Time              `json:"created_at"`
}

type PreferencesResponse struct {
	Theme                ThemeResponse              `json:"theme"`
	Notifications        NotificationSettingsResponse `json:"notifications"`
	Dashboard            DashboardSettingsResponse   `json:"dashboard"`
	Privacy              PrivacySettingsResponse     `json:"privacy"`
	Accessibility        AccessibilitySettingsResponse `json:"accessibility"`
	Language             string                      `json:"language"`
	Timezone             string                      `json:"timezone"`
	DateFormat           string                      `json:"date_format"`
	TimeFormat           string                      `json:"time_format"`
	Currency             string                      `json:"currency"`
	CompactMode          bool                        `json:"compact_mode"`
	SidebarCollapsed     bool                        `json:"sidebar_collapsed"`
	EmailDigestFrequency string                      `json:"email_digest_frequency"`
}

type ThemeResponse struct {
	Mode string `json:"mode"` // light, dark, auto
}

type NotificationSettingsResponse struct {
	Email NotificationChannelResponse `json:"email"`
	Push  NotificationChannelResponse `json:"push"`
	InApp InAppNotificationsResponse  `json:"in_app"`
	SMS   NotificationChannelResponse `json:"sms"`
}

type NotificationChannelResponse struct {
	Enabled bool            `json:"enabled"`
	Digest  string          `json:"digest,omitempty"`
	Events  map[string]bool `json:"events"`
}

type InAppNotificationsResponse struct {
	Enabled              bool `json:"enabled"`
	Sound                bool `json:"sound"`
	DesktopNotifications bool `json:"desktop_notifications"`
}

type DashboardSettingsResponse struct {
	Widgets     []DashboardWidgetResponse `json:"widgets"`
	DefaultView string                    `json:"default_view"`
	ShowTips    bool                      `json:"show_tips"`
}

type DashboardWidgetResponse struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Order   int    `json:"order"`
}

type PrivacySettingsResponse struct {
	ShowProfilePicture bool `json:"show_profile_picture"`
	ShowPhone          bool `json:"show_phone"`
	ShowEmail          bool `json:"show_email"`
	ShowBirthday       bool `json:"show_birthday"`
	AllowMentions      bool `json:"allow_mentions"`
}

type AccessibilitySettingsResponse struct {
	HighContrast      bool   `json:"high_contrast"`
	ReduceMotion      bool   `json:"reduce_motion"`
	ScreenReader      bool   `json:"screen_reader"`
	FontSize          string `json:"font_size"`
	KeyboardShortcuts bool   `json:"keyboard_shortcuts"`
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:               user.ID,
		CompanyID:        user.CompanyID,
		Email:            user.Email,
		FullName:         user.FullName,
		Phone:            user.Phone,
		Role:             user.Role,
		Department:       user.Department,
		Designation:      user.Designation,
		AvatarURL:        user.AvatarURL,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		LastLoginAt:      user.LastLoginAt,
		Permissions:      user.Permissions,
		Timezone:         user.Timezone,
		Language:         user.Language,
		CreatedAt:        user.CreatedAt,
	}
}

func ToPreferencesResponse(prefs *models.UserPreferences) *PreferencesResponse {
	return &PreferencesResponse{
		Theme: ThemeResponse{
			Mode: string(prefs.Theme),
		},
		Notifications: NotificationSettingsResponse{
			Email: NotificationChannelResponse{
				Enabled: prefs.Notifications.Email.Enabled,
				Digest:  prefs.Notifications.Email.Digest,
				Events:  prefs.Notifications.Email.Events,
			},
			Push: NotificationChannelResponse{
				Enabled: prefs.Notifications.Push.Enabled,
				Digest:  prefs.Notifications.Push.Digest,
				Events:  prefs.Notifications.Push.Events,
			},
			InApp: InAppNotificationsResponse{
				Enabled:              prefs.Notifications.InApp.Enabled,
				Sound:                prefs.Notifications.InApp.Sound,
				DesktopNotifications: prefs.Notifications.InApp.DesktopNotifications,
			},
			SMS: NotificationChannelResponse{
				Enabled: prefs.Notifications.SMS.Enabled,
				Digest:  prefs.Notifications.SMS.Digest,
				Events:  prefs.Notifications.SMS.Events,
			},
		},
		Dashboard: DashboardSettingsResponse{
			Widgets:     toDashboardWidgets(prefs.Dashboard.Widgets),
			DefaultView: prefs.Dashboard.DefaultView,
			ShowTips:    prefs.Dashboard.ShowTips,
		},
		Privacy: PrivacySettingsResponse{
			ShowProfilePicture: prefs.Privacy.ShowProfilePicture,
			ShowPhone:          prefs.Privacy.ShowPhone,
			ShowEmail:          prefs.Privacy.ShowEmail,
			ShowBirthday:       prefs.Privacy.ShowBirthday,
			AllowMentions:      prefs.Privacy.AllowMentions,
		},
		Accessibility: AccessibilitySettingsResponse{
			HighContrast:      prefs.Accessibility.HighContrast,
			ReduceMotion:      prefs.Accessibility.ReduceMotion,
			ScreenReader:      prefs.Accessibility.ScreenReader,
			FontSize:          prefs.Accessibility.FontSize,
			KeyboardShortcuts: prefs.Accessibility.KeyboardShortcuts,
		},
		Language:             prefs.Language,
		Timezone:             prefs.Timezone,
		DateFormat:           prefs.DateFormat,
		TimeFormat:           prefs.TimeFormat,
		Currency:             prefs.Currency,
		CompactMode:          prefs.CompactMode,
		SidebarCollapsed:     prefs.SidebarCollapsed,
		EmailDigestFrequency: prefs.EmailDigestFrequency,
	}
}

func toDashboardWidgets(widgets []models.DashboardWidget) []DashboardWidgetResponse {
	result := make([]DashboardWidgetResponse, len(widgets))
	for i, w := range widgets {
		result[i] = DashboardWidgetResponse{
			ID:      w.ID,
			Enabled: w.Enabled,
			Order:   w.Order,
		}
	}
	return result
}