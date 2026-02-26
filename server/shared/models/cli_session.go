package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// CLI SESSION MODEL
// ============================================================================

// CLISession represents an authenticated CLI tool session
// Used for interactive login (whk login)
// Non-interactive CI/CD uses api_keys table instead
type CLISession struct {
	ID     uuid.UUID `db:"id" json:"id"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`

	// Currently selected organization (switchable via `whk switch-org`)
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`

	// Device info
	DeviceName        *string `db:"device_name" json:"device_name,omitempty"`
	DeviceFingerprint *string `db:"device_fingerprint" json:"device_fingerprint,omitempty"`
	CLIVersion        *string `db:"cli_version" json:"cli_version,omitempty"`
	OSInfo            *string `db:"os_info" json:"os_info,omitempty"`

	// Token hashes (never store plaintext)
	SessionTokenHash *string `db:"session_token_hash" json:"-"` // Short-lived
	RefreshTokenHash *string `db:"refresh_token_hash" json:"-"` // Long-lived

	// Expiration
	ExpiresAt        time.Time  `db:"expires_at" json:"expires_at"`
	RefreshExpiresAt *time.Time `db:"refresh_expires_at" json:"refresh_expires_at,omitempty"`

	// Usage tracking
	LastUsedAt      *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	LastUsedIP      *string    `db:"last_used_ip" json:"last_used_ip,omitempty"`
	LastUsedCommand *string    `db:"last_used_command" json:"last_used_command,omitempty"`
	UsageCount      int        `db:"usage_count" json:"usage_count"`

	// Revocation
	RevokedAt     *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	RevokedReason *string    `db:"revoked_reason" json:"revoked_reason,omitempty"`

	// Login IP (security audit)
	CreatedIP *string `db:"created_ip" json:"created_ip,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

func (s *CLISession) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *CLISession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *CLISession) IsRefreshExpired() bool {
	if s.RefreshExpiresAt == nil {
		return s.IsExpired()
	}
	return time.Now().After(*s.RefreshExpiresAt)
}

// IsValid checks if session can be used for authentication
func (s *CLISession) IsValid() bool {
	return !s.IsRevoked() && !s.IsExpired()
}

// CanRefresh checks if this session can be refreshed
func (s *CLISession) CanRefresh() bool {
	return !s.IsRevoked() && !s.IsRefreshExpired() && s.RefreshTokenHash != nil
}

// TimeUntilExpiry returns duration until session expires
func (s *CLISession) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}

// ============================================================================
// ACTIONS
// ============================================================================

// Revoke invalidates this session
func (s *CLISession) Revoke(reason string) {
	now := time.Now()
	s.RevokedAt = &now
	s.RevokedReason = &reason
}

// RecordUsage updates usage tracking
func (s *CLISession) RecordUsage(ip, command string) {
	now := time.Now()
	s.LastUsedAt = &now
	s.LastUsedIP = &ip
	s.LastUsedCommand = &command
	s.UsageCount++
}

// SwitchOrganization changes the currently selected org
func (s *CLISession) SwitchOrganization(orgID uuid.UUID) {
	s.OrganizationID = &orgID
}

// ClearOrganization removes org selection
func (s *CLISession) ClearOrganization() {
	s.OrganizationID = nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (s *CLISession) Validate() error {
	if s.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if s.ExpiresAt.IsZero() {
		return fmt.Errorf("expiration time is required")
	}
	if s.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration time must be in the future")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicCLISession struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	OrganizationID    *uuid.UUID `json:"organization_id,omitempty"`
	DeviceName        *string    `json:"device_name,omitempty"`
	CLIVersion        *string    `json:"cli_version,omitempty"`
	OSInfo            *string    `json:"os_info,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	LastUsedCommand   *string    `json:"last_used_command,omitempty"`
	UsageCount        int        `json:"usage_count"`
	IsExpired         bool       `json:"is_expired"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (s *CLISession) ToPublic() *PublicCLISession {
	return &PublicCLISession{
		ID:              s.ID,
		UserID:          s.UserID,
		OrganizationID:  s.OrganizationID,
		DeviceName:      s.DeviceName,
		CLIVersion:      s.CLIVersion,
		OSInfo:          s.OSInfo,
		ExpiresAt:       s.ExpiresAt,
		LastUsedAt:      s.LastUsedAt,
		LastUsedCommand: s.LastUsedCommand,
		UsageCount:      s.UsageCount,
		IsExpired:       s.IsExpired(),
		RevokedAt:       s.RevokedAt,
		CreatedAt:       s.CreatedAt,
	}
}