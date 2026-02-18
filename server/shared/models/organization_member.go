package models

import (
	"slices"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ORGANIZATION MEMBER MODEL
// ============================================================================

// OrganizationMember represents a team member in an organization
type OrganizationMember struct {
	// Primary identification
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`

	// Role in organization
	Role OrganizationMemberRole `db:"role" json:"role"`

	// Invitation tracking
	InvitedBy             *uuid.UUID `db:"invited_by" json:"invited_by,omitempty"`
	InvitationAcceptedAt  *time.Time `db:"invitation_accepted_at" json:"invitation_accepted_at,omitempty"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ROLE HELPERS
// ============================================================================

// IsOwner checks if member is owner
func (m *OrganizationMember) IsOwner() bool {
	return m.Role == RoleOwner
}

// IsAdmin checks if member is admin or owner
func (m *OrganizationMember) IsAdmin() bool {
	return m.Role == RoleAdmin || m.Role == RoleOwner
}

// CanManageMembers checks if member can manage other members
func (m *OrganizationMember) CanManageMembers() bool {
	return m.Role == RoleOwner || m.Role == RoleAdmin
}

// CanInviteMembers checks if member can invite new members
func (m *OrganizationMember) CanInviteMembers() bool {
	return m.Role == RoleOwner || m.Role == RoleAdmin
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the organization member struct
func (m *OrganizationMember) Validate() error {
	if m.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if m.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	// Validate role
	validRoles := []OrganizationMemberRole{RoleOwner, RoleAdmin, RoleMember, RoleViewer}
	validRole := slices.Contains(validRoles, m.Role)
	if !validRole {
		return fmt.Errorf("invalid role")
	}
	return nil
}

// ============================================================================
// SOFT DELETE
// ============================================================================

// SoftDelete soft deletes the member
func (m *OrganizationMember) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
}

// IsDeleted checks if member is soft deleted
func (m *OrganizationMember) IsDeleted() bool {
	return m.DeletedAt != nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicOrganizationMember represents safe member data for API responses
type PublicOrganizationMember struct {
	ID                   uuid.UUID              `json:"id"`
	OrganizationID       uuid.UUID              `json:"organization_id"`
	UserID               uuid.UUID              `json:"user_id"`
	Role                 OrganizationMemberRole `json:"role"`
	InvitedBy            *uuid.UUID             `json:"invited_by,omitempty"`
	InvitationAcceptedAt *time.Time             `json:"invitation_accepted_at,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
}

// ToPublic converts OrganizationMember to PublicOrganizationMember
func (m *OrganizationMember) ToPublic() *PublicOrganizationMember {
	return &PublicOrganizationMember{
		ID:                   m.ID,
		OrganizationID:       m.OrganizationID,
		UserID:               m.UserID,
		Role:                 m.Role,
		InvitedBy:            m.InvitedBy,
		InvitationAcceptedAt: m.InvitationAcceptedAt,
		CreatedAt:            m.CreatedAt,
	}
}