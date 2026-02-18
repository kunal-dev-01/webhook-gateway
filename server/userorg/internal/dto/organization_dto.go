package dto

import (
	"time"

	"github.com/google/uuid"

	"shared/models"
)

// ============================================================================
// RESPONSE MODELS
// ============================================================================

type OrganizationResponse struct {
	ID                 uuid.UUID                 `json:"id"`
	Name               string                    `json:"name"`
	Slug               string                    `json:"slug"`
	OwnerID            uuid.UUID                 `json:"owner_id"`
	Role               string                    `json:"role"` 
	PlanTier           models.PlanTier           `json:"plan_tier"`
	BillingPeriod      models.BillingPeriod      `json:"billing_period"`
	SubscriptionStatus models.SubscriptionStatus `json:"subscription_status"`
	MonthlyEventsLimit int64                     `json:"monthly_events_limit"`
	MonthlyEventsUsed  int64                     `json:"monthly_events_used"`
	RetentionDays      int                       `json:"retention_days"`
	TeamSeatsLimit     int                       `json:"team_seats_limit"`
	RateLimitPerSecond int                       `json:"rate_limit_per_second"`
	ProjectsLimit      int                       `json:"projects_limit"`
	CreatedAt          time.Time                 `json:"created_at"`
}

type MemberResponse struct {
	ID                   uuid.UUID                      `json:"id"`
	OrganizationID       uuid.UUID                      `json:"organization_id"`
	UserID               uuid.UUID                      `json:"user_id"`
	Role                 models.OrganizationMemberRole  `json:"role"`
	InvitedBy            *uuid.UUID                     `json:"invited_by,omitempty"`
	InvitationAcceptedAt *time.Time                     `json:"invitation_accepted_at,omitempty"`
	JoinedAt             time.Time                      `json:"joined_at"`
}

// ============================================================================
// REQUEST MODELS
// ============================================================================

type ListMembersRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ============================================================================
// RESPONSE WRAPPERS
// ============================================================================

type ListOrganizationsResponse struct {
	Organizations []*OrganizationResponse `json:"organizations"`
	Total         int                     `json:"total"`
}

type GetOrganizationResponse struct {
	Organization *OrganizationResponse `json:"organization"`
}

type ListMembersResponse struct {
	Members []*MemberResponse `json:"members"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func ToOrganizationResponse(org *models.Organization, userRole string) *OrganizationResponse {
	return &OrganizationResponse{
		ID:                 org.ID,
		Name:               org.Name,
		Slug:               org.Slug,
		OwnerID:            org.OwnerID,
		Role:               userRole,
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

func ToMemberResponse(m *models.OrganizationMember) *MemberResponse {
	return &MemberResponse{
		ID:                   m.ID,
		OrganizationID:       m.OrganizationID,
		UserID:               m.UserID,
		Role:                 m.Role,
		InvitedBy:            m.InvitedBy,
		InvitationAcceptedAt: m.InvitationAcceptedAt,
		JoinedAt:             m.CreatedAt,
	}
}