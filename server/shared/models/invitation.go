package models

import (
	"slices"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// InvitationStatus represents invitation status
type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationExpired  InvitationStatus = "expired"
	InvitationRevoked  InvitationStatus = "revoked"
)

// ============================================================================
// INVITATION METADATA
// ============================================================================

// InvitationMetadata represents invitation metadata stored as JSONB
type InvitationMetadata struct {
	CustomMessage string `json:"custom_message,omitempty"`
	InviterName   string `json:"inviter_name,omitempty"`
	OrgName       string `json:"org_name,omitempty"`
}

// Value implements driver.Valuer for InvitationMetadata
func (m InvitationMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for InvitationMetadata
func (m *InvitationMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal InvitationMetadata value")
	}
	return json.Unmarshal(bytes, m)
}

// ============================================================================
// INVITATION MODEL
// ============================================================================

// Invitation represents a team member invitation
type Invitation struct {
	// Primary identification
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Invitee details
	Email string `db:"email" json:"email"` // Email of person being invited

	// Invitation details
	Role      OrganizationMemberRole `db:"role" json:"role"`
	InvitedBy uuid.UUID              `db:"invited_by" json:"invited_by"`

	// Token for accepting invitation (also stored in Redis with TTL)
	Token string `db:"token" json:"token"`

	// Status
	Status InvitationStatus `db:"status" json:"status"`

	// Acceptance tracking
	AcceptedBy *uuid.UUID `db:"accepted_by" json:"accepted_by,omitempty"`
	AcceptedAt *time.Time `db:"accepted_at" json:"accepted_at,omitempty"`

	// Expiration
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`

	// Metadata
	Metadata InvitationMetadata `db:"metadata" json:"metadata"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

// IsPending checks if invitation is pending
func (i *Invitation) IsPending() bool {
	return i.Status == InvitationPending
}

// IsExpired checks if invitation is expired
func (i *Invitation) IsExpired() bool {
	return i.Status == InvitationExpired || time.Now().After(i.ExpiresAt)
}

// IsAccepted checks if invitation is accepted
func (i *Invitation) IsAccepted() bool {
	return i.Status == InvitationAccepted
}

// IsRevoked checks if invitation is revoked
func (i *Invitation) IsRevoked() bool {
	return i.Status == InvitationRevoked
}

// CanBeAccepted checks if invitation can be accepted
func (i *Invitation) CanBeAccepted() bool {
	return i.IsPending() && !i.IsExpired()
}

// ============================================================================
// ACTIONS
// ============================================================================

// Accept marks invitation as accepted
func (i *Invitation) Accept(userID uuid.UUID) error {
	if !i.CanBeAccepted() {
		return fmt.Errorf("invitation cannot be accepted (status: %s, expired: %v)", i.Status, i.IsExpired())
	}

	now := time.Now()
	i.Status = InvitationAccepted
	i.AcceptedBy = &userID
	i.AcceptedAt = &now

	return nil
}

// Revoke marks invitation as revoked
func (i *Invitation) Revoke() error {
	if !i.IsPending() {
		return fmt.Errorf("only pending invitations can be revoked")
	}

	i.Status = InvitationRevoked
	return nil
}

// MarkExpired marks invitation as expired
func (i *Invitation) MarkExpired() {
	i.Status = InvitationExpired
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the invitation struct
func (i *Invitation) Validate() error {
	if i.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if err := utils.ValidateEmail(i.Email); err != nil {
		return err
	}
	if i.InvitedBy == uuid.Nil {
		return fmt.Errorf("inviter ID is required")
	}
	if i.Token == "" {
		return fmt.Errorf("invitation token is required")
	}
	if i.ExpiresAt.IsZero() {
		return fmt.Errorf("expiration time is required")
	}
	// Validate role (owner cannot be invited)
	validRoles := []OrganizationMemberRole{RoleAdmin, RoleMember, RoleViewer}
	validRole := slices.Contains(validRoles, i.Role)
	if !validRole {
		return fmt.Errorf("invalid invitation role (owner cannot be invited)")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicInvitation represents safe invitation data for API responses
type PublicInvitation struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	Email          string                 `json:"email"`
	Role           OrganizationMemberRole `json:"role"`
	InvitedBy      uuid.UUID              `json:"invited_by"`
	Status         InvitationStatus       `json:"status"`
	ExpiresAt      time.Time              `json:"expires_at"`
	Metadata       InvitationMetadata     `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ToPublic converts Invitation to PublicInvitation
func (i *Invitation) ToPublic() *PublicInvitation {
	return &PublicInvitation{
		ID:             i.ID,
		OrganizationID: i.OrganizationID,
		Email:          i.Email,
		Role:           i.Role,
		InvitedBy:      i.InvitedBy,
		Status:         i.Status,
		ExpiresAt:      i.ExpiresAt,
		Metadata:       i.Metadata,
		CreatedAt:      i.CreatedAt,
	}
}