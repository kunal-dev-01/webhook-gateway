package services

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"

	"project/internal/dto"
	"project/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

// maxEndpointsPerProject is the max endpoints allowed per project
const maxEndpointsPerProject = 20

// ============================================================================
// Endpoint Service
// ============================================================================

type EndpointService struct {
	endpointRepo *repository.EndpointRepository
	projectRepo  *repository.ProjectRepository
	log          *logger.Logger
}

func NewEndpointService(
	endpointRepo *repository.EndpointRepository,
	projectRepo *repository.ProjectRepository,
	log *logger.Logger,
) *EndpointService {
	return &EndpointService{
		endpointRepo: endpointRepo,
		projectRepo:  projectRepo,
		log:          log,
	}
}

// ============================================================================
// CREATE ENDPOINT
// ============================================================================

func (s *EndpointService) CreateEndpoint(ctx context.Context, projectID, orgID uuid.UUID, req *dto.CreateEndpointRequest) (*dto.CreateEndpointResponse, error) {
	// Verify project belongs to org
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	// Validate
	if err := sharedUtils.ValidateRequired(req.Name, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := sharedUtils.ValidateLength(req.Name, 2, 255, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := validateURL(req.URL); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Enforce per-project limit
	_, total, err := s.endpointRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if total >= maxEndpointsPerProject {
		return nil, errors.BadRequest(fmt.Sprintf("endpoint limit reached (%d/%d) per project", total, maxEndpointsPerProject))
	}

	// Apply defaults
	maxRetries := 3
	if req.MaxRetries != nil {
		maxRetries = *req.MaxRetries
	}
	retryStrategy := models.RetryStrategyExponentialBackoff
	if req.RetryStrategy != nil {
		if !req.RetryStrategy.IsValid() {
			return nil, errors.BadRequest("invalid retry_strategy")
		}
		retryStrategy = *req.RetryStrategy
	}
	retryDelay := 5
	if req.RetryDelaySeconds != nil {
		retryDelay = *req.RetryDelaySeconds
	}
	retryMaxDelay := 300
	if req.RetryMaxDelaySeconds != nil {
		retryMaxDelay = *req.RetryMaxDelaySeconds
	}
	timeout := 30
	if req.TimeoutSeconds != nil {
		timeout = *req.TimeoutSeconds
	}
	verifySSL := true
	if req.VerifySSL != nil {
		verifySSL = *req.VerifySSL
	}
	rateLimit := 10
	if req.RateLimitPerSecond != nil {
		rateLimit = *req.RateLimitPerSecond
	}
	circuitThreshold := 5
	if req.CircuitFailureThreshold != nil {
		circuitThreshold = *req.CircuitFailureThreshold
	}
	circuitReset := 60
	if req.CircuitResetTimeoutSecs != nil {
		circuitReset = *req.CircuitResetTimeoutSecs
	}
	customHeaders := models.EndpointHeaders{}
	if req.CustomHeaders != nil {
		customHeaders = req.CustomHeaders
	}

	now := time.Now()
	endpoint := &models.Endpoint{
		ID:                      uuid.New(),
		ProjectID:               projectID,
		OrganizationID:          orgID,
		Name:                    req.Name,
		URL:                     req.URL,
		Description:             req.Description,
		IsActive:                true,
		MaxRetries:              maxRetries,
		RetryStrategy:           retryStrategy,
		RetryDelaySeconds:       retryDelay,
		RetryMaxDelaySeconds:    retryMaxDelay,
		TimeoutSeconds:          timeout,
		CustomHeaders:           customHeaders,
		VerifySSL:               verifySSL,
		RateLimitPerSecond:      rateLimit,
		CircuitBreakerStatus:    models.CircuitClosed,
		CircuitFailureThreshold: circuitThreshold,
		CircuitResetTimeoutSecs: circuitReset,
		Metadata:                map[string]interface{}{},
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if err := endpoint.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.endpointRepo.Create(endpoint); err != nil {
		return nil, err
	}

	s.log.Info("Endpoint created",
		"endpoint_id", endpoint.ID,
		"project_id", projectID,
		"url", endpoint.URL,
	)

	return &dto.CreateEndpointResponse{
		Endpoint: dto.ToEndpointResponse(endpoint),
	}, nil
}

// ============================================================================
// LIST ENDPOINTS
// ============================================================================

func (s *EndpointService) ListEndpoints(ctx context.Context, projectID, orgID uuid.UUID) (*dto.ListEndpointsResponse, error) {
	// Verify project belongs to org
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	endpoints, total, err := s.endpointRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.EndpointResponse, 0, len(endpoints))
	for _, e := range endpoints {
		responses = append(responses, dto.ToEndpointResponse(e))
	}

	return &dto.ListEndpointsResponse{
		Endpoints: responses,
		Total:     total,
	}, nil
}

// ============================================================================
// GET ENDPOINT
// ============================================================================

func (s *EndpointService) GetEndpoint(ctx context.Context, endpointID, projectID, orgID uuid.UUID) (*dto.GetEndpointResponse, error) {
	// Verify project belongs to org first
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	endpoint, err := s.endpointRepo.GetByIDAndProject(ctx, endpointID, projectID)
	if err != nil {
		return nil, err
	}

	return &dto.GetEndpointResponse{
		Endpoint: dto.ToEndpointResponse(endpoint),
	}, nil
}

// ============================================================================
// UPDATE ENDPOINT
// ============================================================================

func (s *EndpointService) UpdateEndpoint(ctx context.Context, endpointID, projectID, orgID uuid.UUID, req *dto.UpdateEndpointRequest) (*dto.UpdateEndpointResponse, error) {
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	endpoint, err := s.endpointRepo.GetByIDAndProject(ctx, endpointID, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if err := sharedUtils.ValidateLength(*req.Name, 2, 255, "name"); err != nil {
			return nil, errors.BadRequest(err.Error())
		}
		endpoint.Name = *req.Name
	}
	if req.URL != nil {
		if err := validateURL(*req.URL); err != nil {
			return nil, errors.BadRequest(err.Error())
		}
		endpoint.URL = *req.URL
	}
	if req.Description != nil {
		endpoint.Description = req.Description
	}
	if req.MaxRetries != nil {
		if *req.MaxRetries < 0 || *req.MaxRetries > 10 {
			return nil, errors.BadRequest("max_retries must be between 0 and 10")
		}
		endpoint.MaxRetries = *req.MaxRetries
	}
	if req.TimeoutSeconds != nil {
		if *req.TimeoutSeconds < 1 || *req.TimeoutSeconds > 300 {
			return nil, errors.BadRequest("timeout_seconds must be between 1 and 300")
		}
		endpoint.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.CustomHeaders != nil {
		endpoint.CustomHeaders = req.CustomHeaders
	}
	if req.VerifySSL != nil {
		endpoint.VerifySSL = *req.VerifySSL
	}
	if req.RateLimitPerSecond != nil {
		endpoint.RateLimitPerSecond = *req.RateLimitPerSecond
	}

	if err := s.endpointRepo.Update(endpoint); err != nil {
		return nil, err
	}

	s.log.Info("Endpoint updated", "endpoint_id", endpointID, "project_id", projectID)

	return &dto.UpdateEndpointResponse{
		Endpoint: dto.ToEndpointResponse(endpoint),
	}, nil
}

// ============================================================================
// DELETE ENDPOINT
// ============================================================================

func (s *EndpointService) DeleteEndpoint(ctx context.Context, endpointID, projectID, orgID uuid.UUID) (*dto.DeleteEndpointResponse, error) {
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	if _, err := s.endpointRepo.GetByIDAndProject(ctx, endpointID, projectID); err != nil {
		return nil, err
	}

	if err := s.endpointRepo.SoftDelete(endpointID); err != nil {
		return nil, err
	}

	s.log.Info("Endpoint deleted", "endpoint_id", endpointID, "project_id", projectID)

	return &dto.DeleteEndpointResponse{
		Message: "Endpoint deleted successfully",
	}, nil
}

// ============================================================================
// TOGGLE ACTIVE
// ============================================================================

func (s *EndpointService) ToggleEndpoint(ctx context.Context, endpointID, projectID, orgID uuid.UUID) (*dto.ToggleEndpointResponse, error) {
	if _, err := s.projectRepo.GetByIDAndOrg(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	endpoint, err := s.endpointRepo.GetByIDAndProject(ctx, endpointID, projectID)
	if err != nil {
		return nil, err
	}

	newState := !endpoint.IsActive
	if err := s.endpointRepo.SetActive(endpointID, newState); err != nil {
		return nil, err
	}
	endpoint.IsActive = newState

	s.log.Info("Endpoint toggled",
		"endpoint_id", endpointID,
		"is_active", newState,
	)

	return &dto.ToggleEndpointResponse{
		Endpoint: dto.ToEndpointResponse(endpoint),
	}, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("url must be a valid http or https URL")
	}
	if parsed.Host == "" {
		return fmt.Errorf("url must have a valid host")
	}
	return nil
}