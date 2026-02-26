package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// APIKeyType represents the type/purpose of an API key
type APIKeyType string

const (
	APIKeyTypeAPIKey APIKeyType = "api_key" // Dashboard se banaya
	APIKeyTypeCLIPAT APIKeyType = "cli_pat" // CLI Personal Access Token
	APIKeyTypeCICD   APIKeyType = "ci_cd"   // CI/CD pipelines ke liye
)

func (k APIKeyType) IsValid() bool {
	switch k {
	case APIKeyTypeAPIKey, APIKeyTypeCLIPAT, APIKeyTypeCICD:
		return true
	}
	return false
}

// ============================================================================
// SCOPES
// ============================================================================

// APIKeyScope represents a permission scope
type APIKeyScope string

const (
	ScopeWebhooksRead   APIKeyScope = "webhooks:read"
	ScopeWebhooksWrite  APIKeyScope = "webhooks:write"
	ScopeEndpointsRead  APIKeyScope = "endpoints:read"
	ScopeEndpointsWrite APIKeyScope = "endpoints:write"
	ScopeProjectsRead   APIKeyScope = "projects:read"
	ScopeProjectsWrite  APIKeyScope = "projects:write"
	ScopeAnalyticsRead  APIKeyScope = "analytics:read"
	ScopeAdmin          APIKeyScope = "admin"
)

// AllScopes returns all valid scope values
func AllScopes() []APIKeyScope {
	return []APIKeyScope{
		ScopeWebhooksRead, ScopeWebhooksWrite,
		ScopeEndpointsRead, ScopeEndpointsWrite,
		ScopeProjectsRead, ScopeProjectsWrite,
		ScopeAnalyticsRead, ScopeAdmin,
	}
}

// ============================================================================
// API KEY MODEL
// ============================================================================

// APIKey represents an authentication key for API/CLI access
// Security: Only key_prefix is stored readable. Full key is bcrypt hashed.
// Format: whk_live_{random} (production) or whk_test_{random} (development)
type APIKey struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	ProjectID      *uuid.UUID `db:"project_id" json:"project_id,omitempty"` // nil = org-level

	// Creator
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`

	// Identification
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description,omitempty"`

	// Type and environment
	KeyType     APIKeyType         `db:"key_type" json:"key_type"`
	Environment ProjectEnvironment `db:"environment" json:"environment"`

	// Key storage (NEVER store plaintext key after creation)
	// KeyPrefix: shown to users for identification ("whk_live_abc1...")
	// KeyHash:   bcrypt hash for verification
	KeyPrefix string `db:"key_prefix" json:"key_prefix"`
	KeyHash   string `db:"key_hash" json:"-"` // Never expose

	// Permissions
	Scopes pq.StringArray `db:"scopes" json:"scopes"`

	// Rate limiting
	RateLimitPerMinute int `db:"rate_limit_per_minute" json:"rate_limit_per_minute"`

	// Usage tracking
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	LastUsedIP *string    `db:"last_used_ip" json:"last_used_ip,omitempty"`
	UsageCount int64      `db:"usage_count" json:"usage_count"`

	// Expiration (nil = never expires)
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`

	// Status
	IsActive bool `db:"is_active" json:"is_active"`

	// Revocation
	RevokedAt     *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	RevokedBy     *uuid.UUID `db:"revoked_by" json:"revoked_by,omitempty"`
	RevokedReason *string    `db:"revoked_reason" json:"revoked_reason,omitempty"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsUsable checks if key can be used for auth
func (k *APIKey) IsUsable() bool {
	return k.IsActive && !k.IsRevoked() && !k.IsExpired() && k.DeletedAt == nil
}

func (k *APIKey) IsOrgLevel() bool {
	return k.ProjectID == nil
}

// ============================================================================
// SCOPE HELPERS
// ============================================================================

// HasScope checks if key has a specific permission
func (k *APIKey) HasScope(scope APIKeyScope) bool {
	// Admin scope = full access
	for _, s := range k.Scopes {
		if s == string(ScopeAdmin) || s == string(scope) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if key has all required permissions
func (k *APIKey) HasAllScopes(scopes ...APIKeyScope) bool {
	for _, scope := range scopes {
		if !k.HasScope(scope) {
			return false
		}
	}
	return true
}

// ============================================================================
// ACTIONS
// ============================================================================

// Revoke revokes the API key
func (k *APIKey) Revoke(revokedBy uuid.UUID, reason string) error {
	if k.IsRevoked() {
		return fmt.Errorf("key is already revoked")
	}
	now := time.Now()
	k.RevokedAt = &now
	k.RevokedBy = &revokedBy
	k.RevokedReason = &reason
	k.IsActive = false
	return nil
}

// RecordUsage updates last used tracking
func (k *APIKey) RecordUsage(ip string) {
	now := time.Now()
	k.LastUsedAt = &now
	k.LastUsedIP = &ip
	k.UsageCount++
}

// ============================================================================
// VALIDATION
// ============================================================================

func (k *APIKey) Validate() error {
	if k.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if k.CreatedBy == uuid.Nil {
		return fmt.Errorf("created by is required")
	}
	if err := utils.ValidateRequired(k.Name, "name"); err != nil {
		return err
	}
	if !k.KeyType.IsValid() {
		return fmt.Errorf("invalid key type")
	}
	if !k.Environment.IsValid() {
		return fmt.Errorf("invalid environment")
	}
	if k.KeyPrefix == "" {
		return fmt.Errorf("key prefix is required")
	}
	if k.KeyHash == "" {
		return fmt.Errorf("key hash is required")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicAPIKey struct {
	ID                 uuid.UUID          `json:"id"`
	OrganizationID     uuid.UUID          `json:"organization_id"`
	ProjectID          *uuid.UUID         `json:"project_id,omitempty"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	KeyType            APIKeyType         `json:"key_type"`
	Environment        ProjectEnvironment `json:"environment"`
	KeyPrefix          string             `json:"key_prefix"` // e.g., "whk_live_abc1..."
	Scopes             pq.StringArray     `json:"scopes"`
	RateLimitPerMinute int                `json:"rate_limit_per_minute"`
	LastUsedAt         *time.Time         `json:"last_used_at,omitempty"`
	UsageCount         int64              `json:"usage_count"`
	ExpiresAt          *time.Time         `json:"expires_at,omitempty"`
	IsActive           bool               `json:"is_active"`
	RevokedAt          *time.Time         `json:"revoked_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
}

func (k *APIKey) ToPublic() *PublicAPIKey {
	return &PublicAPIKey{
		ID:                 k.ID,
		OrganizationID:     k.OrganizationID,
		ProjectID:          k.ProjectID,
		Name:               k.Name,
		Description:        k.Description,
		KeyType:            k.KeyType,
		Environment:        k.Environment,
		KeyPrefix:          k.KeyPrefix,
		Scopes:             k.Scopes,
		RateLimitPerMinute: k.RateLimitPerMinute,
		LastUsedAt:         k.LastUsedAt,
		UsageCount:         k.UsageCount,
		ExpiresAt:          k.ExpiresAt,
		IsActive:           k.IsActive,
		RevokedAt:          k.RevokedAt,
		CreatedAt:          k.CreatedAt,
	}
}