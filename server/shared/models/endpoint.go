package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

// RetryStrategy represents the retry algorithm
type RetryStrategy string

const (
	RetryStrategyNone               RetryStrategy = "no_retry"
	RetryStrategyLinear             RetryStrategy = "linear"
	RetryStrategyExponentialBackoff RetryStrategy = "exponential_backoff"
	RetryStrategySmartBackoff       RetryStrategy = "smart_backoff"
)

func (r RetryStrategy) IsValid() bool {
	switch r {
	case RetryStrategyNone, RetryStrategyLinear, RetryStrategyExponentialBackoff, RetryStrategySmartBackoff:
		return true
	}
	return false
}

// CircuitBreakerStatus represents circuit breaker state
type CircuitBreakerStatus string

const (
	CircuitClosed   CircuitBreakerStatus = "closed"    // Normal - requests flowing
	CircuitOpen     CircuitBreakerStatus = "open"      // Blocked - too many failures
	CircuitHalfOpen CircuitBreakerStatus = "half_open" // Testing - one request allowed
)

func (c CircuitBreakerStatus) IsValid() bool {
	switch c {
	case CircuitClosed, CircuitOpen, CircuitHalfOpen:
		return true
	}
	return false
}

// ============================================================================
// CUSTOM HEADERS
// ============================================================================

// EndpointHeaders represents custom headers sent with every delivery
type EndpointHeaders map[string]string

func (h EndpointHeaders) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (h *EndpointHeaders) Scan(value interface{}) error {
	if value == nil {
		*h = EndpointHeaders{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal EndpointHeaders value")
	}
	return json.Unmarshal(bytes, h)
}

// ============================================================================
// ENDPOINT MODEL
// ============================================================================

// Endpoint represents a destination URL for webhook delivery
type Endpoint struct {
	ID             uuid.UUID `db:"id" json:"id"`
	ProjectID      uuid.UUID `db:"project_id" json:"project_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Destination
	Name        string  `db:"name" json:"name"`
	URL         string  `db:"url" json:"url"`
	Description *string `db:"description" json:"description,omitempty"`

	// Status
	IsActive bool `db:"is_active" json:"is_active"`

	// Retry config
	MaxRetries            int           `db:"max_retries" json:"max_retries"`
	RetryStrategy         RetryStrategy `db:"retry_strategy" json:"retry_strategy"`
	RetryDelaySeconds     int           `db:"retry_delay_seconds" json:"retry_delay_seconds"`
	RetryMaxDelaySeconds  int           `db:"retry_max_delay_seconds" json:"retry_max_delay_seconds"`

	// Request config
	TimeoutSeconds     int             `db:"timeout_seconds" json:"timeout_seconds"`
	CustomHeaders      EndpointHeaders `db:"custom_headers" json:"custom_headers"`
	VerifySSL          bool            `db:"verify_ssl" json:"verify_ssl"`
	SigningSecret      *string         `db:"signing_secret" json:"-"` // Never expose
	RateLimitPerSecond int             `db:"rate_limit_per_second" json:"rate_limit_per_second"`

	// Health tracking (updated by worker)
	LastSuccessAt       *time.Time `db:"last_success_at" json:"last_success_at,omitempty"`
	LastFailureAt       *time.Time `db:"last_failure_at" json:"last_failure_at,omitempty"`
	ConsecutiveFailures int        `db:"consecutive_failures" json:"consecutive_failures"`
	TotalDeliveries     int64      `db:"total_deliveries" json:"total_deliveries"`
	TotalFailures       int64      `db:"total_failures" json:"total_failures"`

	// Circuit breaker
	CircuitBreakerStatus    CircuitBreakerStatus `db:"circuit_breaker_status" json:"circuit_breaker_status"`
	CircuitOpenedAt         *time.Time           `db:"circuit_opened_at" json:"circuit_opened_at,omitempty"`
	CircuitFailureCount     int                  `db:"circuit_failure_count" json:"circuit_failure_count"`
	CircuitFailureThreshold int                  `db:"circuit_failure_threshold" json:"circuit_failure_threshold"`
	CircuitResetTimeoutSecs int                  `db:"circuit_reset_timeout_seconds" json:"circuit_reset_timeout_seconds"`

	// Metadata
	Metadata map[string]interface{} `db:"metadata" json:"metadata"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// CIRCUIT BREAKER HELPERS
// ============================================================================

func (e *Endpoint) IsCircuitOpen() bool {
	return e.CircuitBreakerStatus == CircuitOpen
}

func (e *Endpoint) IsCircuitClosed() bool {
	return e.CircuitBreakerStatus == CircuitClosed
}

func (e *Endpoint) IsCircuitHalfOpen() bool {
	return e.CircuitBreakerStatus == CircuitHalfOpen
}

// OpenCircuit trips the circuit breaker
func (e *Endpoint) OpenCircuit() {
	now := time.Now()
	e.CircuitBreakerStatus = CircuitOpen
	e.CircuitOpenedAt = &now
}

// HalfOpenCircuit moves to half-open state for testing
func (e *Endpoint) HalfOpenCircuit() {
	e.CircuitBreakerStatus = CircuitHalfOpen
}

// CloseCircuit resets circuit breaker to normal
func (e *Endpoint) CloseCircuit() {
	e.CircuitBreakerStatus = CircuitClosed
	e.CircuitOpenedAt = nil
	e.CircuitFailureCount = 0
	e.ConsecutiveFailures = 0
}

// ShouldAttemptReset checks if enough time has passed to try half-open
func (e *Endpoint) ShouldAttemptReset() bool {
	if e.CircuitBreakerStatus != CircuitOpen || e.CircuitOpenedAt == nil {
		return false
	}
	resetAfter := time.Duration(e.CircuitResetTimeoutSecs) * time.Second
	return time.Since(*e.CircuitOpenedAt) >= resetAfter
}

// RecordSuccess records a successful delivery
func (e *Endpoint) RecordSuccess() {
	now := time.Now()
	e.LastSuccessAt = &now
	e.ConsecutiveFailures = 0
	e.TotalDeliveries++
	if e.IsCircuitHalfOpen() {
		e.CloseCircuit()
	}
}

// RecordFailure records a failed delivery and trips circuit if needed
func (e *Endpoint) RecordFailure() {
	now := time.Now()
	e.LastFailureAt = &now
	e.ConsecutiveFailures++
	e.CircuitFailureCount++
	e.TotalDeliveries++
	e.TotalFailures++
	if e.ConsecutiveFailures >= e.CircuitFailureThreshold {
		e.OpenCircuit()
	}
}

// CanDeliver checks if endpoint can receive deliveries
func (e *Endpoint) CanDeliver() bool {
	return e.IsActive && !e.IsCircuitOpen()
}

// ============================================================================
// SOFT DELETE
// ============================================================================

func (e *Endpoint) IsDeleted() bool {
	return e.DeletedAt != nil
}

func (e *Endpoint) SoftDelete() {
	now := time.Now()
	e.DeletedAt = &now
	e.IsActive = false
}

// ============================================================================
// VALIDATION
// ============================================================================

func (e *Endpoint) Validate() error {
	if e.ProjectID == uuid.Nil {
		return fmt.Errorf("project ID is required")
	}
	if e.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if err := utils.ValidateRequired(e.Name, "name"); err != nil {
		return err
	}
	if err := utils.ValidateRequired(e.URL, "url"); err != nil {
		return err
	}
	if !e.RetryStrategy.IsValid() {
		return fmt.Errorf("invalid retry strategy")
	}
	if e.MaxRetries < 0 || e.MaxRetries > 10 {
		return fmt.Errorf("max retries must be between 0 and 10")
	}
	if e.TimeoutSeconds < 1 || e.TimeoutSeconds > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicEndpoint struct {
	ID                      uuid.UUID            `json:"id"`
	ProjectID               uuid.UUID            `json:"project_id"`
	OrganizationID          uuid.UUID            `json:"organization_id"`
	Name                    string               `json:"name"`
	URL                     string               `json:"url"`
	Description             *string              `json:"description,omitempty"`
	IsActive                bool                 `json:"is_active"`
	MaxRetries              int                  `json:"max_retries"`
	RetryStrategy           RetryStrategy        `json:"retry_strategy"`
	TimeoutSeconds          int                  `json:"timeout_seconds"`
	RateLimitPerSecond      int                  `json:"rate_limit_per_second"`
	VerifySSL               bool                 `json:"verify_ssl"`
	LastSuccessAt           *time.Time           `json:"last_success_at,omitempty"`
	LastFailureAt           *time.Time           `json:"last_failure_at,omitempty"`
	ConsecutiveFailures     int                  `json:"consecutive_failures"`
	TotalDeliveries         int64                `json:"total_deliveries"`
	CircuitBreakerStatus    CircuitBreakerStatus `json:"circuit_breaker_status"`
	CreatedAt               time.Time            `json:"created_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
}

func (e *Endpoint) ToPublic() *PublicEndpoint {
	return &PublicEndpoint{
		ID:                   e.ID,
		ProjectID:            e.ProjectID,
		OrganizationID:       e.OrganizationID,
		Name:                 e.Name,
		URL:                  e.URL,
		Description:          e.Description,
		IsActive:             e.IsActive,
		MaxRetries:           e.MaxRetries,
		RetryStrategy:        e.RetryStrategy,
		TimeoutSeconds:       e.TimeoutSeconds,
		RateLimitPerSecond:   e.RateLimitPerSecond,
		VerifySSL:            e.VerifySSL,
		LastSuccessAt:        e.LastSuccessAt,
		LastFailureAt:        e.LastFailureAt,
		ConsecutiveFailures:  e.ConsecutiveFailures,
		TotalDeliveries:      e.TotalDeliveries,
		CircuitBreakerStatus: e.CircuitBreakerStatus,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}