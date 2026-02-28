package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// CREATE ENDPOINT
// ============================================================================

type CreateEndpointRequest struct {
	Name                    string                `json:"name"`
	URL                     string                `json:"url"`
	Description             *string               `json:"description,omitempty"`
	MaxRetries              *int                  `json:"max_retries,omitempty"`              // default: 3
	RetryStrategy           *models.RetryStrategy `json:"retry_strategy,omitempty"`           // default: exponential_backoff
	RetryDelaySeconds       *int                  `json:"retry_delay_seconds,omitempty"`      // default: 5
	RetryMaxDelaySeconds    *int                  `json:"retry_max_delay_seconds,omitempty"`  // default: 300
	TimeoutSeconds          *int                  `json:"timeout_seconds,omitempty"`          // default: 30
	CustomHeaders           map[string]string     `json:"custom_headers,omitempty"`
	VerifySSL               *bool                 `json:"verify_ssl,omitempty"`               // default: true
	RateLimitPerSecond      *int                  `json:"rate_limit_per_second,omitempty"`    // default: 10
	CircuitFailureThreshold *int                  `json:"circuit_failure_threshold,omitempty"` // default: 5
	CircuitResetTimeoutSecs *int                  `json:"circuit_reset_timeout_secs,omitempty"` // default: 60
}

type CreateEndpointResponse struct {
	Endpoint *EndpointResponse `json:"endpoint"`
}

// ============================================================================
// LIST ENDPOINTS
// ============================================================================

type ListEndpointsResponse struct {
	Endpoints []*EndpointResponse `json:"endpoints"`
	Total     int                 `json:"total"`
}

// ============================================================================
// GET ENDPOINT
// ============================================================================

type GetEndpointResponse struct {
	Endpoint *EndpointResponse `json:"endpoint"`
}

// ============================================================================
// UPDATE ENDPOINT
// ============================================================================

type UpdateEndpointRequest struct {
	Name                 *string           `json:"name,omitempty"`
	URL                  *string           `json:"url,omitempty"`
	Description          *string           `json:"description,omitempty"`
	MaxRetries           *int              `json:"max_retries,omitempty"`
	TimeoutSeconds       *int              `json:"timeout_seconds,omitempty"`
	CustomHeaders        map[string]string `json:"custom_headers,omitempty"`
	VerifySSL            *bool             `json:"verify_ssl,omitempty"`
	RateLimitPerSecond   *int              `json:"rate_limit_per_second,omitempty"`
}

type UpdateEndpointResponse struct {
	Endpoint *EndpointResponse `json:"endpoint"`
}

// ============================================================================
// DELETE ENDPOINT
// ============================================================================

type DeleteEndpointResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// TOGGLE ACTIVE
// ============================================================================

type ToggleEndpointResponse struct {
	Endpoint *EndpointResponse `json:"endpoint"`
}

// ============================================================================
// RESPONSE MODEL
// ============================================================================

type EndpointResponse struct {
	ID                      uuid.UUID                   `json:"id"`
	ProjectID               uuid.UUID                   `json:"project_id"`
	OrganizationID          uuid.UUID                   `json:"organization_id"`
	Name                    string                      `json:"name"`
	URL                     string                      `json:"url"`
	Description             *string                     `json:"description,omitempty"`
	IsActive                bool                        `json:"is_active"`
	MaxRetries              int                         `json:"max_retries"`
	RetryStrategy           models.RetryStrategy        `json:"retry_strategy"`
	RetryDelaySeconds       int                         `json:"retry_delay_seconds"`
	RetryMaxDelaySeconds    int                         `json:"retry_max_delay_seconds"`
	TimeoutSeconds          int                         `json:"timeout_seconds"`
	CustomHeaders           map[string]string           `json:"custom_headers"`
	VerifySSL               bool                        `json:"verify_ssl"`
	RateLimitPerSecond      int                         `json:"rate_limit_per_second"`
	LastSuccessAt           *time.Time                  `json:"last_success_at,omitempty"`
	LastFailureAt           *time.Time                  `json:"last_failure_at,omitempty"`
	ConsecutiveFailures     int                         `json:"consecutive_failures"`
	TotalDeliveries         int64                       `json:"total_deliveries"`
	TotalFailures           int64                       `json:"total_failures"`
	CircuitBreakerStatus    models.CircuitBreakerStatus `json:"circuit_breaker_status"`
	CircuitOpenedAt         *time.Time                  `json:"circuit_opened_at,omitempty"`
	CircuitFailureThreshold int                         `json:"circuit_failure_threshold"`
	CircuitResetTimeoutSecs int                         `json:"circuit_reset_timeout_secs"`
	CreatedAt               time.Time                   `json:"created_at"`
	UpdatedAt               time.Time                   `json:"updated_at"`
}

// ============================================================================
// MAPPER
// ============================================================================

func ToEndpointResponse(e *models.Endpoint) *EndpointResponse {
	return &EndpointResponse{
		ID:                      e.ID,
		ProjectID:               e.ProjectID,
		OrganizationID:          e.OrganizationID,
		Name:                    e.Name,
		URL:                     e.URL,
		Description:             e.Description,
		IsActive:                e.IsActive,
		MaxRetries:              e.MaxRetries,
		RetryStrategy:           e.RetryStrategy,
		RetryDelaySeconds:       e.RetryDelaySeconds,
		RetryMaxDelaySeconds:    e.RetryMaxDelaySeconds,
		TimeoutSeconds:          e.TimeoutSeconds,
		CustomHeaders:           e.CustomHeaders,
		VerifySSL:               e.VerifySSL,
		RateLimitPerSecond:      e.RateLimitPerSecond,
		LastSuccessAt:           e.LastSuccessAt,
		LastFailureAt:           e.LastFailureAt,
		ConsecutiveFailures:     e.ConsecutiveFailures,
		TotalDeliveries:         e.TotalDeliveries,
		TotalFailures:           e.TotalFailures,
		CircuitBreakerStatus:    e.CircuitBreakerStatus,
		CircuitOpenedAt:         e.CircuitOpenedAt,
		CircuitFailureThreshold: e.CircuitFailureThreshold,
		CircuitResetTimeoutSecs: e.CircuitResetTimeoutSecs,
		CreatedAt:               e.CreatedAt,
		UpdatedAt:               e.UpdatedAt,
	}
}