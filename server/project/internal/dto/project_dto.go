package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// CREATE PROJECT
// ============================================================================

type CreateProjectRequest struct {
	Name        string                    `json:"name"`
	Slug        string                    `json:"slug"`
	Description *string                   `json:"description,omitempty"`
	Environment models.ProjectEnvironment `json:"environment"` // production | development
}

type CreateProjectResponse struct {
	Project *ProjectResponse `json:"project"`
}

// ============================================================================
// LIST PROJECTS
// ============================================================================

type ListProjectsResponse struct {
	Projects []*ProjectResponse `json:"projects"`
	Total    int                `json:"total"`
}

// ============================================================================
// GET PROJECT
// ============================================================================

type GetProjectResponse struct {
	Project *ProjectResponse `json:"project"`
}

// ============================================================================
// UPDATE PROJECT
// ============================================================================

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateProjectResponse struct {
	Project *ProjectResponse `json:"project"`
}

// ============================================================================
// DELETE PROJECT
// ============================================================================

type DeleteProjectResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// TOGGLE ACTIVE
// ============================================================================

type ToggleProjectResponse struct {
	Project *ProjectResponse `json:"project"`
}

// ============================================================================
// RESPONSE MODEL
// ============================================================================

type ProjectResponse struct {
	ID             uuid.UUID                 `json:"id"`
	OrganizationID uuid.UUID                 `json:"organization_id"`
	Name           string                    `json:"name"`
	Slug           string                    `json:"slug"`
	Description    *string                   `json:"description,omitempty"`
	Environment    models.ProjectEnvironment `json:"environment"`
	IsActive       bool                      `json:"is_active"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

// ============================================================================
// MAPPER
// ============================================================================

func ToProjectResponse(p *models.Project) *ProjectResponse {
	return &ProjectResponse{
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