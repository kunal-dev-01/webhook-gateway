package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// WebhookSourceType represents the source platform
type WebhookSourceType string

const (
	SourceTypeGeneric  WebhookSourceType = "generic"
	SourceTypeStripe   WebhookSourceType = "stripe"
	SourceTypeGitHub   WebhookSourceType = "github"
	SourceTypeShopify  WebhookSourceType = "shopify"
	SourceTypeSvix     WebhookSourceType = "svix"
	SourceTypeCustom   WebhookSourceType = "custom"
)

func (s WebhookSourceType) IsValid() bool {
	switch s {
	case SourceTypeGeneric, SourceTypeStripe, SourceTypeGitHub,
		SourceTypeShopify, SourceTypeSvix, SourceTypeCustom:
		return true
	}
	return false
}

// VerificationMethod represents how incoming webhooks are verified
type VerificationMethod string

const (
	VerifyHMACSHA256 VerificationMethod = "hmac_sha256"
	VerifyHMACSHA1   VerificationMethod = "hmac_sha1"
	VerifyBasicAuth  VerificationMethod = "basic_auth"
	VerifyAPIKey     VerificationMethod = "api_key"
	VerifyNone       VerificationMethod = "none"
)

func (v VerificationMethod) IsValid() bool {
	switch v {
	case VerifyHMACSHA256, VerifyHMACSHA1, VerifyBasicAuth, VerifyAPIKey, VerifyNone:
		return true
	}
	return false
}

// ============================================================================
// TRANSFORM CONFIG
// ============================================================================

// TransformConfig represents payload transformation rules (future feature)
type TransformConfig map[string]interface{}

func (t TransformConfig) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TransformConfig) Scan(value interface{}) error {
	if value == nil {
		*t = TransformConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TransformConfig value")
	}
	return json.Unmarshal(bytes, t)
}

// ============================================================================
// WEBHOOK SOURCE MODEL
// ============================================================================

// WebhookSource represents an inbound webhook source
// URL format: https://hooks.yourdomain.com/in/{IngestToken}
type WebhookSource struct {
	ID             uuid.UUID `db:"id" json:"id"`
	ProjectID      uuid.UUID `db:"project_id" json:"project_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Identification
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description,omitempty"`

	// Unique token for inbound URL
	// Full URL: https://hooks.yourdomain.com/in/{IngestToken}
	IngestToken string `db:"ingest_token" json:"ingest_token"`

	// Source config
	SourceType         WebhookSourceType  `db:"source_type" json:"source_type"`
	VerificationMethod VerificationMethod `db:"verification_method" json:"verification_method"`

	// Verification secret (store encrypted, never expose)
	VerificationSecret *string `db:"verification_secret" json:"-"`

	// Header name for verification
	// e.g., "Stripe-Signature", "X-Hub-Signature-256", "X-API-Key"
	VerificationHeader *string `db:"verification_header" json:"verification_header,omitempty"`

	// IP allowlist (nil = all IPs allowed)
	AllowedIPs pq.StringArray `db:"allowed_ips" json:"allowed_ips,omitempty"`

	// Payload transformation (future feature)
	TransformConfig TransformConfig `db:"transform_config" json:"transform_config"`

	// Status
	IsActive bool `db:"is_active" json:"is_active"`

	// Stats (denormalized for quick display)
	LastReceivedAt *time.Time `db:"last_received_at" json:"last_received_at,omitempty"`
	TotalReceived  int64      `db:"total_received" json:"total_received"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// HELPERS
// ============================================================================

// GetIngestURL returns the full inbound URL for this source
func (w *WebhookSource) GetIngestURL(baseURL string) string {
	return fmt.Sprintf("%s/in/%s", baseURL, w.IngestToken)
}

// IsIPAllowed checks if a given IP is in the allowlist
// Returns true if no allowlist is configured (all IPs allowed)
func (w *WebhookSource) IsIPAllowed(ip string) bool {
	if len(w.AllowedIPs) == 0 {
		return true
	}
	for _, allowed := range w.AllowedIPs {
		if allowed == ip {
			return true
		}
	}
	return false
}

// HasIPAllowlist checks if IP allowlisting is configured
func (w *WebhookSource) HasIPAllowlist() bool {
	return len(w.AllowedIPs) > 0
}

// RequiresVerification checks if requests need signature verification
func (w *WebhookSource) RequiresVerification() bool {
	return w.VerificationMethod != VerifyNone
}

// RecordReceived updates stats when a webhook is received
func (w *WebhookSource) RecordReceived() {
	now := time.Now()
	w.LastReceivedAt = &now
	w.TotalReceived++
}

func (w *WebhookSource) IsDeleted() bool {
	return w.DeletedAt != nil
}

func (w *WebhookSource) SoftDelete() {
	now := time.Now()
	w.DeletedAt = &now
	w.IsActive = false
}

// ============================================================================
// VALIDATION
// ============================================================================

func (w *WebhookSource) Validate() error {
	if w.ProjectID == uuid.Nil {
		return fmt.Errorf("project ID is required")
	}
	if w.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if err := utils.ValidateRequired(w.Name, "name"); err != nil {
		return err
	}
	if w.IngestToken == "" {
		return fmt.Errorf("ingest token is required")
	}
	if !w.SourceType.IsValid() {
		return fmt.Errorf("invalid source type")
	}
	if !w.VerificationMethod.IsValid() {
		return fmt.Errorf("invalid verification method")
	}
	// Secret required for all methods except none
	if w.VerificationMethod != VerifyNone && w.VerificationSecret == nil {
		return fmt.Errorf("verification secret is required for method: %s", w.VerificationMethod)
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicWebhookSource struct {
	ID                 uuid.UUID          `json:"id"`
	ProjectID          uuid.UUID          `json:"project_id"`
	OrganizationID     uuid.UUID          `json:"organization_id"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	IngestToken        string             `json:"ingest_token"`
	IngestURL          string             `json:"ingest_url"` // Computed
	SourceType         WebhookSourceType  `json:"source_type"`
	VerificationMethod VerificationMethod `json:"verification_method"`
	VerificationHeader *string            `json:"verification_header,omitempty"`
	HasIPAllowlist     bool               `json:"has_ip_allowlist"`
	IsActive           bool               `json:"is_active"`
	LastReceivedAt     *time.Time         `json:"last_received_at,omitempty"`
	TotalReceived      int64              `json:"total_received"`
	CreatedAt          time.Time          `json:"created_at"`
}

func (w *WebhookSource) ToPublic(baseURL string) *PublicWebhookSource {
	return &PublicWebhookSource{
		ID:                 w.ID,
		ProjectID:          w.ProjectID,
		OrganizationID:     w.OrganizationID,
		Name:               w.Name,
		Description:        w.Description,
		IngestToken:        w.IngestToken,
		IngestURL:          w.GetIngestURL(baseURL),
		SourceType:         w.SourceType,
		VerificationMethod: w.VerificationMethod,
		VerificationHeader: w.VerificationHeader,
		HasIPAllowlist:     w.HasIPAllowlist(),
		IsActive:           w.IsActive,
		LastReceivedAt:     w.LastReceivedAt,
		TotalReceived:      w.TotalReceived,
		CreatedAt:          w.CreatedAt,
	}
}