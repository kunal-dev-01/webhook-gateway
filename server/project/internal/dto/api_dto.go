package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// CREATE API KEY
// ============================================================================

type CreateAPIKeyRequest struct {
	Name               string              `json:"name"`
	Description        *string             `json:"description,omitempty"`
	KeyType            models.APIKeyType   `json:"key_type"`            // api_key | cli_pat | ci_cd
	Environment        models.ProjectEnvironment `json:"environment"`   // production | development
	ProjectID          *uuid.UUID          `json:"project_id,omitempty"` // nil = org-level
	Scopes             []string            `json:"scopes"`
	ExpiresAt          *time.Time          `json:"expires_at,omitempty"` // nil = never
	RateLimitPerMinute *int                `json:"rate_limit_per_minute,omitempty"` // default: 1000
}

// CreateAPIKeyResponse includes the plaintext key ONCE — never returned again
type CreateAPIKeyResponse struct {
	APIKey    *APIKeyResponse `json:"api_key"`
	PlainKey  string          `json:"key"` // whk_live_xxxx — shown only once
}

// ============================================================================
// LIST API KEYS
// ============================================================================

type ListAPIKeysResponse struct {
	APIKeys []*APIKeyResponse `json:"api_keys"`
	Total   int               `json:"total"`
}

// ============================================================================
// GET API KEY
// ============================================================================

type GetAPIKeyResponse struct {
	APIKey *APIKeyResponse `json:"api_key"`
}

// ============================================================================
// REVOKE API KEY
// ============================================================================

type RevokeAPIKeyRequest struct {
	Reason string `json:"reason"`
}

type RevokeAPIKeyResponse struct {
	APIKey  *APIKeyResponse `json:"api_key"`
	Message string          `json:"message"`
}

// ============================================================================
// DELETE API KEY
// ============================================================================

type DeleteAPIKeyResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// RESPONSE MODEL
// ============================================================================

type APIKeyResponse struct {
	ID                 uuid.UUID                `json:"id"`
	OrganizationID     uuid.UUID                `json:"organization_id"`
	ProjectID          *uuid.UUID               `json:"project_id,omitempty"`
	CreatedBy          uuid.UUID                `json:"created_by"`
	Name               string                   `json:"name"`
	Description        *string                  `json:"description,omitempty"`
	KeyType            models.APIKeyType        `json:"key_type"`
	Environment        models.ProjectEnvironment `json:"environment"`
	KeyPrefix          string                   `json:"key_prefix"` // whk_live_abc1... shown for identification
	Scopes             []string                 `json:"scopes"`
	RateLimitPerMinute int                      `json:"rate_limit_per_minute"`
	LastUsedAt         *time.Time               `json:"last_used_at,omitempty"`
	UsageCount         int64                    `json:"usage_count"`
	ExpiresAt          *time.Time               `json:"expires_at,omitempty"`
	IsActive           bool                     `json:"is_active"`
	IsRevoked          bool                     `json:"is_revoked"`
	IsExpired          bool                     `json:"is_expired"`
	RevokedAt          *time.Time               `json:"revoked_at,omitempty"`
	RevokedReason      *string                  `json:"revoked_reason,omitempty"`
	CreatedAt          time.Time                `json:"created_at"`
}

// ============================================================================
// MAPPER
// ============================================================================

func ToAPIKeyResponse(k *models.APIKey) *APIKeyResponse {
	scopes := make([]string, len(k.Scopes))
	copy(scopes, k.Scopes)

	return &APIKeyResponse{
		ID:                 k.ID,
		OrganizationID:     k.OrganizationID,
		ProjectID:          k.ProjectID,
		CreatedBy:          k.CreatedBy,
		Name:               k.Name,
		Description:        k.Description,
		KeyType:            k.KeyType,
		Environment:        k.Environment,
		KeyPrefix:          k.KeyPrefix,
		Scopes:             scopes,
		RateLimitPerMinute: k.RateLimitPerMinute,
		LastUsedAt:         k.LastUsedAt,
		UsageCount:         k.UsageCount,
		ExpiresAt:          k.ExpiresAt,
		IsActive:           k.IsActive,
		IsRevoked:          k.IsRevoked(),
		IsExpired:          k.IsExpired(),
		RevokedAt:          k.RevokedAt,
		RevokedReason:      k.RevokedReason,
		CreatedAt:          k.CreatedAt,
	}
}