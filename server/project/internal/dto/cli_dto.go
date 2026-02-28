package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// LIST CLI SESSIONS
// ============================================================================

type ListCLISessionsResponse struct {
	Sessions []*CLISessionResponse `json:"sessions"`
	Total    int                   `json:"total"`
}

// ============================================================================
// GET CLI SESSION
// ============================================================================

type GetCLISessionResponse struct {
	Session *CLISessionResponse `json:"session"`
}

// ============================================================================
// REVOKE CLI SESSION
// ============================================================================

type RevokeCLISessionRequest struct {
	Reason string `json:"reason"`
}

type RevokeCLISessionResponse struct {
	Session *CLISessionResponse `json:"session"`
	Message string              `json:"message"`
}

// ============================================================================
// REVOKE ALL CLI SESSIONS
// ============================================================================

type RevokeAllCLISessionsResponse struct {
	RevokedCount int    `json:"revoked_count"`
	Message      string `json:"message"`
}

// ============================================================================
// RESPONSE MODEL
// ============================================================================

type CLISessionResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	OrganizationID  *uuid.UUID `json:"organization_id,omitempty"`
	DeviceName      *string    `json:"device_name,omitempty"`
	CLIVersion      *string    `json:"cli_version,omitempty"`
	OSInfo          *string    `json:"os_info,omitempty"`
	ExpiresAt       time.Time  `json:"expires_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	LastUsedCommand *string    `json:"last_used_command,omitempty"`
	UsageCount      int        `json:"usage_count"`
	IsExpired       bool       `json:"is_expired"`
	IsRevoked       bool       `json:"is_revoked"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RevokedReason   *string    `json:"revoked_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	// IsCurrentSession is set in the handler by comparing session IDs
	IsCurrentSession bool `json:"is_current_session"`
}

// ============================================================================
// MAPPER
// ============================================================================

func ToCLISessionResponse(s *models.CLISession, currentSessionID *uuid.UUID) *CLISessionResponse {
	isCurrent := false
	if currentSessionID != nil && *currentSessionID == s.ID {
		isCurrent = true
	}

	return &CLISessionResponse{
		ID:               s.ID,
		UserID:           s.UserID,
		OrganizationID:   s.OrganizationID,
		DeviceName:       s.DeviceName,
		CLIVersion:       s.CLIVersion,
		OSInfo:           s.OSInfo,
		ExpiresAt:        s.ExpiresAt,
		LastUsedAt:       s.LastUsedAt,
		LastUsedCommand:  s.LastUsedCommand,
		UsageCount:       s.UsageCount,
		IsExpired:        s.IsExpired(),
		IsRevoked:        s.IsRevoked(),
		RevokedAt:        s.RevokedAt,
		RevokedReason:    s.RevokedReason,
		CreatedAt:        s.CreatedAt,
		IsCurrentSession: isCurrent,
	}
}