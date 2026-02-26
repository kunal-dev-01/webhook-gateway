package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// DedicatedIPStatus represents the status of a dedicated IP
type DedicatedIPStatus string

const (
	DedicatedIPAvailable     DedicatedIPStatus = "available"      // Pool mein hai
	DedicatedIPAssigned      DedicatedIPStatus = "assigned"       // Org ko assign hai
	DedicatedIPReserved      DedicatedIPStatus = "reserved"       // System use
	DedicatedIPDecommissioned DedicatedIPStatus = "decommissioned" // Retired
)

func (s DedicatedIPStatus) IsValid() bool {
	switch s {
	case DedicatedIPAvailable, DedicatedIPAssigned, DedicatedIPReserved, DedicatedIPDecommissioned:
		return true
	}
	return false
}

// IPVersion represents IPv4 or IPv6
type IPVersion string

const (
	IPv4 IPVersion = "ipv4"
	IPv6 IPVersion = "ipv6"
)

// IPProvider represents the cloud provider
type IPProvider string

const (
	IPProviderAWS          IPProvider = "aws"
	IPProviderHetzner      IPProvider = "hetzner"
	IPProviderGCP          IPProvider = "gcp"
	IPProviderDigitalOcean IPProvider = "digitalocean"
	IPProviderVultr        IPProvider = "vultr"
)

func (p IPProvider) IsValid() bool {
	switch p {
	case IPProviderAWS, IPProviderHetzner, IPProviderGCP, IPProviderDigitalOcean, IPProviderVultr:
		return true
	}
	return false
}

// ============================================================================
// DEDICATED IP MODEL
// ============================================================================

// DedicatedIP represents a dedicated outgoing IP address (Scale plan only)
// Architecture: shared app server → dedicated NAT/proxy → this fixed IP
// Customers use this for IP allowlisting on their destination servers
type DedicatedIP struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"` // nil = pool

	// IP details
	IPAddress string    `db:"ip_address" json:"ip_address"`
	IPVersion IPVersion `db:"ip_version" json:"ip_version"`

	// Cloud provider info
	Provider           IPProvider `db:"provider" json:"provider"`
	ProviderResourceID *string    `db:"provider_resource_id" json:"provider_resource_id,omitempty"`
	// AWS: "eipalloc-xxx", Hetzner: floating IP ID
	ProviderRegion *string `db:"provider_region" json:"provider_region,omitempty"`
	// AWS: "us-east-1", Hetzner: "nbg1"

	// Assignment tracking
	Status     DedicatedIPStatus `db:"status" json:"status"`
	AssignedAt *time.Time        `db:"assigned_at" json:"assigned_at,omitempty"`
	ReleasedAt *time.Time        `db:"released_at" json:"released_at,omitempty"`

	// Internal cost tracking
	MonthlyCostUSD float64 `db:"monthly_cost_usd" json:"monthly_cost_usd"`

	Notes *string `db:"notes" json:"notes,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

func (d *DedicatedIP) IsAvailable() bool {
	return d.Status == DedicatedIPAvailable
}

func (d *DedicatedIP) IsAssigned() bool {
	return d.Status == DedicatedIPAssigned
}

func (d *DedicatedIP) IsReserved() bool {
	return d.Status == DedicatedIPReserved
}

func (d *DedicatedIP) IsDecommissioned() bool {
	return d.Status == DedicatedIPDecommissioned
}

// ============================================================================
// ACTIONS
// ============================================================================

// Assign assigns this IP to an organization
func (d *DedicatedIP) Assign(orgID uuid.UUID) error {
	if !d.IsAvailable() {
		return fmt.Errorf("IP is not available for assignment (status: %s)", d.Status)
	}
	now := time.Now()
	d.OrganizationID = &orgID
	d.Status = DedicatedIPAssigned
	d.AssignedAt = &now
	d.ReleasedAt = nil
	return nil
}

// Release releases the IP back to the pool
func (d *DedicatedIP) Release() error {
	if !d.IsAssigned() {
		return fmt.Errorf("IP is not currently assigned")
	}
	now := time.Now()
	d.OrganizationID = nil
	d.Status = DedicatedIPAvailable
	d.ReleasedAt = &now
	return nil
}

// Decommission permanently retires this IP
func (d *DedicatedIP) Decommission() {
	d.Status = DedicatedIPDecommissioned
	d.OrganizationID = nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (d *DedicatedIP) Validate() error {
	if d.IPAddress == "" {
		return fmt.Errorf("IP address is required")
	}
	if !d.Status.IsValid() {
		return fmt.Errorf("invalid IP status")
	}
	if !d.Provider.IsValid() {
		return fmt.Errorf("invalid provider")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicDedicatedIP struct {
	ID             uuid.UUID         `json:"id"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty"`
	IPAddress      string            `json:"ip_address"`
	IPVersion      IPVersion         `json:"ip_version"`
	Provider       IPProvider        `json:"provider"`
	ProviderRegion *string           `json:"provider_region,omitempty"`
	Status         DedicatedIPStatus `json:"status"`
	AssignedAt     *time.Time        `json:"assigned_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

func (d *DedicatedIP) ToPublic() *PublicDedicatedIP {
	return &PublicDedicatedIP{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		IPAddress:      d.IPAddress,
		IPVersion:      d.IPVersion,
		Provider:       d.Provider,
		ProviderRegion: d.ProviderRegion,
		Status:         d.Status,
		AssignedAt:     d.AssignedAt,
		CreatedAt:      d.CreatedAt,
	}
}