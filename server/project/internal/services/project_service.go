package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"project/internal/dto"
	"project/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

// ============================================================================
// Project Service
// ============================================================================

type ProjectService struct {
	projectRepo *repository.ProjectRepository
	log         *logger.Logger
}

func NewProjectService(projectRepo *repository.ProjectRepository, log *logger.Logger) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		log:         log,
	}
}

// ============================================================================
// CREATE PROJECT
// ============================================================================

func (s *ProjectService) CreateProject(ctx context.Context, orgID uuid.UUID, projectsLimit int, req *dto.CreateProjectRequest) (*dto.CreateProjectResponse, error) {
	// Validate input
	if err := sharedUtils.ValidateRequired(req.Name, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := sharedUtils.ValidateLength(req.Name, 2, 255, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if req.Slug == "" {
		req.Slug = sharedUtils.GenerateSlug(req.Name)
	}
	if err := sharedUtils.ValidateSlug(req.Slug); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if !req.Environment.IsValid() {
		return nil, errors.BadRequest("environment must be 'production' or 'development'")
	}

	// Plan limit check
	count, err := s.projectRepo.CountByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if count >= projectsLimit {
		return nil, errors.BadRequest(fmt.Sprintf("project limit reached (%d/%d) — upgrade your plan to create more", count, projectsLimit))
	}

	// Slug uniqueness check
	exists, err := s.projectRepo.ExistsBySlug(ctx, orgID, req.Slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("a project with this slug already exists in your organization")
	}

	// Generate signing secret
	signingSecret, err := generateSecret(32)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to generate signing secret: %w", err))
	}

	now := time.Now()
	project := &models.Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		SigningSecret:  signingSecret,
		Environment:    req.Environment,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := project.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	s.log.Info("Project created",
		"project_id", project.ID,
		"slug", project.Slug,
		"org_id", orgID,
	)

	return &dto.CreateProjectResponse{
		Project: dto.ToProjectResponse(project),
	}, nil
}

// ============================================================================
// LIST PROJECTS
// ============================================================================

func (s *ProjectService) ListProjects(ctx context.Context, orgID uuid.UUID) (*dto.ListProjectsResponse, error) {
	projects, total, err := s.projectRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		responses = append(responses, dto.ToProjectResponse(p))
	}

	return &dto.ListProjectsResponse{
		Projects: responses,
		Total:    total,
	}, nil
}

// ============================================================================
// GET PROJECT
// ============================================================================

func (s *ProjectService) GetProject(ctx context.Context, projectID, orgID uuid.UUID) (*dto.GetProjectResponse, error) {
	project, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID)
	if err != nil {
		return nil, err
	}

	return &dto.GetProjectResponse{
		Project: dto.ToProjectResponse(project),
	}, nil
}

// ============================================================================
// UPDATE PROJECT
// ============================================================================

func (s *ProjectService) UpdateProject(ctx context.Context, projectID, orgID uuid.UUID, req *dto.UpdateProjectRequest) (*dto.UpdateProjectResponse, error) {
	// Verify ownership
	project, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		if err := sharedUtils.ValidateLength(*req.Name, 2, 255, "name"); err != nil {
			return nil, errors.BadRequest(err.Error())
		}
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}

	if err := s.projectRepo.Update(projectID, project.Name, project.Description); err != nil {
		return nil, err
	}

	s.log.Info("Project updated", "project_id", projectID, "org_id", orgID)

	return &dto.UpdateProjectResponse{
		Project: dto.ToProjectResponse(project),
	}, nil
}

// ============================================================================
// DELETE PROJECT
// ============================================================================

func (s *ProjectService) DeleteProject(ctx context.Context, projectID, orgID uuid.UUID) (*dto.DeleteProjectResponse, error) {
	project, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID)
	if err != nil {
		return nil, err
	}

	project.SoftDelete()

	if err := s.projectRepo.SoftDelete(projectID, project.Slug); err != nil {
		return nil, err
	}

	s.log.Info("Project deleted", "project_id", projectID, "org_id", orgID)

	return &dto.DeleteProjectResponse{
		Message: "Project deleted successfully",
	}, nil
}

// ============================================================================
// TOGGLE ACTIVE
// ============================================================================

func (s *ProjectService) ToggleProject(ctx context.Context, projectID, orgID uuid.UUID) (*dto.ToggleProjectResponse, error) {
	project, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID)
	if err != nil {
		return nil, err
	}

	newState := !project.IsActive
	if err := s.projectRepo.SetActive(projectID, newState); err != nil {
		return nil, err
	}

	project.IsActive = newState

	s.log.Info("Project toggled",
		"project_id", projectID,
		"is_active", newState,
	)

	return &dto.ToggleProjectResponse{
		Project: dto.ToProjectResponse(project),
	}, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func generateSecret(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper("whsec_" + hex.EncodeToString(b)), nil
}