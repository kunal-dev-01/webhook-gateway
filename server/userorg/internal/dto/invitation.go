package dto

import (
	"time"

	"github.com/google/uuid"

	"shared/models"
)

// ============================================================================
// REQUEST MODELS
// ============================================================================

type InviteMemberRequest struct {
	Email         string                        `json:"email"`
	Role          models.OrganizationMemberRole `json:"role"`
	CustomMessage string                        `json:"custom_message,omitempty"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

type InvitationResponse struct {
	ID             uuid.UUID                     `json:"id"`
	OrganizationID uuid.UUID                     `json:"organization_id"`
	Email          string                        `json:"email"`
	Role           models.OrganizationMemberRole `json:"role"`
	InvitedBy      uuid.UUID                     `json:"invited_by"`
	Status         models.InvitationStatus       `json:"status"`
	ExpiresAt      time.Time                     `json:"expires_at"`
	Metadata       models.InvitationMetadata     `json:"metadata"`
	CreatedAt      time.Time                     `json:"created_at"`
}

type ListInvitationsResponse struct {
	Invitations []*InvitationResponse `json:"invitations"`
	Total       int                   `json:"total"`
}

type AcceptInvitationResponse struct {
	Organization *OrganizationResponse `json:"organization"`
	Member       *MemberResponse       `json:"member"`
}

// ============================================================================
// HELPERS
// ============================================================================

func ToInvitationResponse(inv *models.Invitation) *InvitationResponse {
	return &InvitationResponse{
		ID:             inv.ID,
		OrganizationID: inv.OrganizationID,
		Email:          inv.Email,
		Role:           inv.Role,
		InvitedBy:      inv.InvitedBy,
		Status:         inv.Status,
		ExpiresAt:      inv.ExpiresAt,
		Metadata:       inv.Metadata,
		CreatedAt:      inv.CreatedAt,
	}
}