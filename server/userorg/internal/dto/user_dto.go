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
	AvatarURL *string `json:"avatar_url"`
}

type UpdateUserResponse struct {
	FullName string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

// ===========================================================================
// UPDATE USER PREFERENCES
// ===========================================================================

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