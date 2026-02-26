package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// ProjectEnvironment represents the project environment
type ProjectEnvironment string

const (
	ProjectEnvironmentProduction  ProjectEnvironment = "production"
	ProjectEnvironmentDevelopment ProjectEnvironment = "development"
)

func (e ProjectEnvironment) IsValid() bool {
	return e == ProjectEnvironmentProduction || e == ProjectEnvironmentDevelopment
}

// ============================================================================
// PROJECT MODEL
// ============================================================================

// Project represents a logical grouping of webhooks within an organization
type Project struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Identification
	Name        string  `db:"name" json:"name"`
	Slug        string  `db:"slug" json:"slug"`
	Description *string `db:"description" json:"description,omitempty"`

	// Signing secret for inbound webhook verification (store encrypted)
	SigningSecret string `db:"signing_secret" json:"-"`

	// Environment
	Environment ProjectEnvironment `db:"environment" json:"environment"`

	// Status
	IsActive bool `db:"is_active" json:"is_active"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// HELPERS
// ============================================================================

func (p *Project) IsProduction() bool {
	return p.Environment == ProjectEnvironmentProduction
}

func (p *Project) IsDevelopment() bool {
	return p.Environment == ProjectEnvironmentDevelopment
}

func (p *Project) IsDeleted() bool {
	return p.DeletedAt != nil
}

func (p *Project) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.Slug = fmt.Sprintf("%s-deleted-%d", p.Slug, now.Unix())
}

// ============================================================================
// VALIDATION
// ============================================================================

func (p *Project) Validate() error {
	if p.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if err := utils.ValidateRequired(p.Name, "name"); err != nil {
		return err
	}
	if err := utils.ValidateLength(p.Name, 2, 255, "name"); err != nil {
		return err
	}
	if err := utils.ValidateRequired(p.Slug, "slug"); err != nil {
		return err
	}
	if err := utils.ValidateSlug(p.Slug); err != nil {
		return err
	}
	if !p.Environment.IsValid() {
		return fmt.Errorf("invalid environment: must be production or development")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicProject struct {
	ID             uuid.UUID          `json:"id"`
	OrganizationID uuid.UUID          `json:"organization_id"`
	Name           string             `json:"name"`
	Slug           string             `json:"slug"`
	Description    *string            `json:"description,omitempty"`
	Environment    ProjectEnvironment `json:"environment"`
	IsActive       bool               `json:"is_active"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

func (p *Project) ToPublic() *PublicProject {
	return &PublicProject{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		Name:           p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		Environment:    p.Environment,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}