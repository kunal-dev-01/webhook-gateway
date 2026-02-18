package dto

import (
	"time"
	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// UPDATE USER 
// ============================================================================
type UpdateUserRequest struct {
	FullName *string `json:"full_name"`
}

type UpdateUserResponse struct {
	FullName  string  `json:"full_name"`
}

// ============================================================================
// UPDATE AVATAR
// ============================================================================

type UpdateAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

// ============================================================================
// UPDATE PREFERENCES
// ============================================================================

type UpdateUserPreferencesRequest struct {
	Preferences models.UserPreferences `json:"preferences"`
}

type UpdateUserPreferencesResponse struct {
	Preferences models.UserPreferences `json:"preferences"`
}

// ============================================================================
// GET CURRENT USER
// ============================================================================

type GetMeResponse struct {
	User         *UserResponse         `json:"user"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

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

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

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

