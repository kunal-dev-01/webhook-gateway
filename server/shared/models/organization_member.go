package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var roleLevel = map[OrganizationMemberRole]int{
	RoleViewer: 1,
	RoleMember: 2,
	RoleAdmin:  3,
	RoleOwner:  4,
}

func (r OrganizationMemberRole) Level() int {
	return roleLevel[r]
}

func (r OrganizationMemberRole) IsValid() bool {
	_, ok := roleLevel[r]
	return ok
}

// IsAtLeast returns true if this role has >= permissions than the given minimum.
func (r OrganizationMemberRole) IsAtLeast(min OrganizationMemberRole) bool {
	return r.Level() >= min.Level()
}

// CanManage returns true if caller can manage target.
// Rule: strictly higher level only.
func (r OrganizationMemberRole) CanManage(target OrganizationMemberRole) bool {
	return r.Level() > target.Level()
}

func (r OrganizationMemberRole) String() string {
	return string(r)
}

// ============================================================================
// ORGANIZATION MEMBER MODEL
// ============================================================================

type OrganizationMember struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	OrganizationID uuid.UUID              `db:"organization_id" json:"organization_id"`
	UserID         uuid.UUID              `db:"user_id" json:"user_id"`
	Role           OrganizationMemberRole `db:"role" json:"role"`

	// Invitation tracking
	InvitedBy            *uuid.UUID `db:"invited_by" json:"invited_by,omitempty"`
	InvitationAcceptedAt *time.Time `db:"invitation_accepted_at" json:"invitation_accepted_at,omitempty"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ROLE HELPERS
// ============================================================================

func (m *OrganizationMember) IsOwner() bool {
	return m.Role == RoleOwner
}

func (m *OrganizationMember) IsAdmin() bool {
	return m.Role == RoleAdmin || m.Role == RoleOwner
}

func (m *OrganizationMember) CanManageMembers() bool {
	return m.Role == RoleOwner || m.Role == RoleAdmin
}

func (m *OrganizationMember) CanInviteMembers() bool {
	return m.Role == RoleOwner || m.Role == RoleAdmin
}

// ============================================================================
// VALIDATION
// ============================================================================

func (m *OrganizationMember) Validate() error {
	if m.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if m.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if !m.Role.IsValid() {
		return fmt.Errorf("invalid role")
	}
	return nil
}

// ============================================================================
// SOFT DELETE
// ============================================================================

func (m *OrganizationMember) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
}

func (m *OrganizationMember) IsDeleted() bool {
	return m.DeletedAt != nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicOrganizationMember struct {
	ID                   uuid.UUID              `json:"id"`
	OrganizationID       uuid.UUID              `json:"organization_id"`
	UserID               uuid.UUID              `json:"user_id"`
	Role                 OrganizationMemberRole `json:"role"`
	InvitedBy            *uuid.UUID             `json:"invited_by,omitempty"`
	InvitationAcceptedAt *time.Time             `json:"invitation_accepted_at,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
}

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