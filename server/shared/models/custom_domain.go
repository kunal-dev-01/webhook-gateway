package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// DomainVerificationStatus represents DNS verification state
type DomainVerificationStatus string

const (
	DomainVerificationPending  DomainVerificationStatus = "pending"
	DomainVerificationVerified DomainVerificationStatus = "verified"
	DomainVerificationFailed   DomainVerificationStatus = "failed"
	DomainVerificationExpired  DomainVerificationStatus = "expired"
)

func (s DomainVerificationStatus) IsValid() bool {
	switch s {
	case DomainVerificationPending, DomainVerificationVerified,
		DomainVerificationFailed, DomainVerificationExpired:
		return true
	}
	return false
}

// DomainSSLStatus represents SSL certificate state
type DomainSSLStatus string

const (
	SSLPending      DomainSSLStatus = "pending"
	SSLProvisioning DomainSSLStatus = "provisioning"
	SSLActive       DomainSSLStatus = "active"
	SSLFailed       DomainSSLStatus = "failed"
	SSLExpired      DomainSSLStatus = "expired"
	SSLRenewing     DomainSSLStatus = "renewing"
)

func (s DomainSSLStatus) IsValid() bool {
	switch s {
	case SSLPending, SSLProvisioning, SSLActive, SSLFailed, SSLExpired, SSLRenewing:
		return true
	}
	return false
}

// DomainType represents the direction of the domain
type DomainType string

const (
	DomainTypeInbound  DomainType = "inbound"  // Receive webhooks
	DomainTypeOutbound DomainType = "outbound" // Future use
)

// ============================================================================
// CUSTOM DOMAIN MODEL
// ============================================================================

// CustomDomain represents a customer's custom webhook domain
// Normal URL:  https://hooks.yourdomain.com/in/TOKEN
// Custom URL:  https://webhooks.acmecorp.com/in/TOKEN
//
// Setup flow:
//  1. Customer adds domain → verification_token generated
//  2. Customer sets CNAME: domain → verify-{token}.hooks.yourdomain.com
//  3. Cron checks DNS → verification_status = verified
//  4. SSL provisioned via Let's Encrypt → ssl_status = active
//  5. is_active = TRUE → domain fully working
type CustomDomain struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organization_id"`
	ProjectID      *uuid.UUID `db:"project_id" json:"project_id,omitempty"` // nil = all projects

	// Domain details
	Domain     string     `db:"domain" json:"domain"`          // webhooks.acmecorp.com
	DomainType DomainType `db:"domain_type" json:"domain_type"` // inbound/outbound

	// DNS Verification
	VerificationToken  string                   `db:"verification_token" json:"verification_token"`
	VerificationStatus DomainVerificationStatus `db:"verification_status" json:"verification_status"`
	VerifiedAt         *time.Time               `db:"verified_at" json:"verified_at,omitempty"`
	LastDNSCheckAt     *time.Time               `db:"last_dns_check_at" json:"last_dns_check_at,omitempty"`
	DNSCheckAttempts   int                      `db:"dns_check_attempts" json:"dns_check_attempts"`
	DNSCheckError      *string                  `db:"dns_check_error" json:"dns_check_error,omitempty"`

	// SSL Certificate
	SSLStatus    DomainSSLStatus `db:"ssl_status" json:"ssl_status"`
	SSLProvider  string          `db:"ssl_provider" json:"ssl_provider"` // "letsencrypt"
	SSLIssuedAt  *time.Time      `db:"ssl_issued_at" json:"ssl_issued_at,omitempty"`
	SSLExpiresAt *time.Time      `db:"ssl_expires_at" json:"ssl_expires_at,omitempty"`
	SSLAutoRenew bool            `db:"ssl_auto_renew" json:"ssl_auto_renew"`
	// Certificate and key stored encrypted - never expose via API
	SSLCertificate *string `db:"ssl_certificate" json:"-"`
	SSLPrivateKey  *string `db:"ssl_private_key" json:"-"`
	SSLError       *string `db:"ssl_error" json:"ssl_error,omitempty"`

	// Active only when DNS verified + SSL active
	IsActive bool `db:"is_active" json:"is_active"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// VERIFICATION HELPERS
// ============================================================================

// GetCNAMETarget returns the expected CNAME target customer should set
func (d *CustomDomain) GetCNAMETarget(baseDomain string) string {
	return fmt.Sprintf("verify-%s.%s", d.VerificationToken, baseDomain)
}

// GetCNAMEInstruction returns DNS setup instructions for customer
func (d *CustomDomain) GetCNAMEInstruction(baseDomain string) string {
	return fmt.Sprintf("Add CNAME: %s → %s", d.Domain, d.GetCNAMETarget(baseDomain))
}

func (d *CustomDomain) IsDNSVerified() bool {
	return d.VerificationStatus == DomainVerificationVerified
}

func (d *CustomDomain) IsDNSPending() bool {
	return d.VerificationStatus == DomainVerificationPending
}

// ============================================================================
// SSL HELPERS
// ============================================================================

func (d *CustomDomain) IsSSLActive() bool {
	return d.SSLStatus == SSLActive
}

func (d *CustomDomain) IsSSLProvisioning() bool {
	return d.SSLStatus == SSLProvisioning
}

// IsSSLExpiringSoon checks if SSL cert expires within given days
func (d *CustomDomain) IsSSLExpiringSoon(days int) bool {
	if d.SSLExpiresAt == nil {
		return false
	}
	threshold := time.Now().AddDate(0, 0, days)
	return d.SSLExpiresAt.Before(threshold)
}

// NeedsSSLRenewal checks if cert should be renewed (< 30 days left)
func (d *CustomDomain) NeedsSSLRenewal() bool {
	return d.IsSSLActive() && d.SSLAutoRenew && d.IsSSLExpiringSoon(30)
}

// ============================================================================
// STATUS ACTIONS
// ============================================================================

// MarkDNSVerified marks DNS as verified and starts SSL provisioning
func (d *CustomDomain) MarkDNSVerified() {
	now := time.Now()
	d.VerificationStatus = DomainVerificationVerified
	d.VerifiedAt = &now
	d.SSLStatus = SSLProvisioning // Start SSL next
}

// MarkDNSFailed records a DNS verification failure
func (d *CustomDomain) MarkDNSFailed(errMsg string) {
	now := time.Now()
	d.LastDNSCheckAt = &now
	d.DNSCheckAttempts++
	d.VerificationStatus = DomainVerificationFailed
	d.DNSCheckError = &errMsg
}

// MarkSSLActive marks SSL as successfully provisioned
func (d *CustomDomain) MarkSSLActive(expiresAt time.Time) {
	now := time.Now()
	d.SSLStatus = SSLActive
	d.SSLIssuedAt = &now
	d.SSLExpiresAt = &expiresAt
	d.SSLError = nil
	d.IsActive = true // Domain fully ready
}

// MarkSSLFailed records SSL provisioning failure
func (d *CustomDomain) MarkSSLFailed(errMsg string) {
	d.SSLStatus = SSLFailed
	d.SSLError = &errMsg
	d.IsActive = false
}

// MarkSSLRenewing marks SSL as being renewed
func (d *CustomDomain) MarkSSLRenewing() {
	d.SSLStatus = SSLRenewing
}

// Deactivate disables this domain
func (d *CustomDomain) Deactivate() {
	d.IsActive = false
}

func (d *CustomDomain) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.IsActive = false
}

func (d *CustomDomain) IsDeleted() bool {
	return d.DeletedAt != nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (d *CustomDomain) Validate() error {
	if d.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if d.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	// Ensure lowercase
	if d.Domain != strings.ToLower(d.Domain) {
		return fmt.Errorf("domain must be lowercase")
	}
	// Basic domain format check
	if strings.HasPrefix(d.Domain, "http") {
		return fmt.Errorf("domain must not include protocol (no http/https)")
	}
	if !strings.Contains(d.Domain, ".") {
		return fmt.Errorf("invalid domain format")
	}
	if d.VerificationToken == "" {
		return fmt.Errorf("verification token is required")
	}
	if !d.VerificationStatus.IsValid() {
		return fmt.Errorf("invalid verification status")
	}
	if !d.SSLStatus.IsValid() {
		return fmt.Errorf("invalid SSL status")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicCustomDomain struct {
	ID                 uuid.UUID                `json:"id"`
	OrganizationID     uuid.UUID                `json:"organization_id"`
	ProjectID          *uuid.UUID               `json:"project_id,omitempty"`
	Domain             string                   `json:"domain"`
	DomainType         DomainType               `json:"domain_type"`
	VerificationToken  string                   `json:"verification_token"`
	VerificationStatus DomainVerificationStatus `json:"verification_status"`
	VerifiedAt         *time.Time               `json:"verified_at,omitempty"`
	DNSCheckAttempts   int                      `json:"dns_check_attempts"`
	DNSCheckError      *string                  `json:"dns_check_error,omitempty"`
	SSLStatus          DomainSSLStatus          `json:"ssl_status"`
	SSLExpiresAt       *time.Time               `json:"ssl_expires_at,omitempty"`
	SSLAutoRenew       bool                     `json:"ssl_auto_renew"`
	SSLError           *string                  `json:"ssl_error,omitempty"`
	IsActive           bool                     `json:"is_active"`
	CreatedAt          time.Time                `json:"created_at"`
}

func (d *CustomDomain) ToPublic() *PublicCustomDomain {
	return &PublicCustomDomain{
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