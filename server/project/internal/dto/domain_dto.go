package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// ADD DOMAIN
// ============================================================================

type AddDomainRequest struct {
	Domain    string     `json:"domain"`     // webhooks.acmecorp.com
	ProjectID *uuid.UUID `json:"project_id"` // nil = applies to all org projects
}

type AddDomainResponse struct {
	Domain *DomainResponse `json:"domain"`
}

// ============================================================================
// LIST DOMAINS
// ============================================================================

type ListDomainsResponse struct {
	Domains []*DomainResponse `json:"domains"`
	Total   int               `json:"total"`
}

// ============================================================================
// GET DOMAIN
// ============================================================================

type GetDomainResponse struct {
	Domain      *DomainResponse `json:"domain"`
	CNAMETarget string          `json:"cname_target"`  // verify-TOKEN.hooks.yourdomain.com
	Instruction string          `json:"instruction"`   // Human-readable setup guide
}

// ============================================================================
// DELETE DOMAIN
// ============================================================================

type DeleteDomainResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// VERIFY DOMAIN (manual trigger)
// ============================================================================

type TriggerVerificationResponse struct {
	Domain  *DomainResponse `json:"domain"`
	Message string          `json:"message"`
}

// ============================================================================
// RESPONSE MODEL
// ============================================================================

type DomainResponse struct {
	ID                 uuid.UUID                        `json:"id"`
	OrganizationID     uuid.UUID                        `json:"organization_id"`
	ProjectID          *uuid.UUID                       `json:"project_id,omitempty"`
	Domain             string                           `json:"domain"`
	DomainType         models.DomainType                `json:"domain_type"`
	VerificationToken  string                           `json:"verification_token"`
	VerificationStatus models.DomainVerificationStatus  `json:"verification_status"`
	VerifiedAt         *time.Time                       `json:"verified_at,omitempty"`
	DNSCheckAttempts   int                              `json:"dns_check_attempts"`
	DNSCheckError      *string                          `json:"dns_check_error,omitempty"`
	SSLStatus          models.DomainSSLStatus           `json:"ssl_status"`
	SSLExpiresAt       *time.Time                       `json:"ssl_expires_at,omitempty"`
	SSLAutoRenew       bool                             `json:"ssl_auto_renew"`
	SSLError           *string                          `json:"ssl_error,omitempty"`
	IsActive           bool                             `json:"is_active"`
	CreatedAt          time.Time                        `json:"created_at"`
}

// ============================================================================
// MAPPER
// ============================================================================

func ToDomainResponse(d *models.CustomDomain) *DomainResponse {
	return &DomainResponse{
		ID:                 d.ID,
		OrganizationID:     d.OrganizationID,
		ProjectID:          d.ProjectID,
		Domain:             d.Domain,
		DomainType:         d.DomainType,
		VerificationToken:  d.VerificationToken,
		VerificationStatus: d.VerificationStatus,
		VerifiedAt:         d.VerifiedAt,
		DNSCheckAttempts:   d.DNSCheckAttempts,
		DNSCheckError:      d.DNSCheckError,
		SSLStatus:          d.SSLStatus,
		SSLExpiresAt:       d.SSLExpiresAt,
		SSLAutoRenew:       d.SSLAutoRenew,
		SSLError:           d.SSLError,
		IsActive:           d.IsActive,
		CreatedAt:          d.CreatedAt,
	}
}