package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"project/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// Endpoint Repository
// ============================================================================

type EndpointRepository struct {
	db *db.DB
}

func NewEndpointRepository(database *db.DB) *EndpointRepository {
	return &EndpointRepository{db: database}
}

func (r *EndpointRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func endpointCols() string {
	return `id, project_id, organization_id,
		name, url, description, is_active,
		max_retries, retry_strategy, retry_delay_seconds, retry_max_delay_seconds,
		timeout_seconds, custom_headers, verify_ssl, signing_secret, rate_limit_per_second,
		last_success_at, last_failure_at, consecutive_failures, total_deliveries, total_failures,
		circuit_breaker_status, circuit_opened_at, circuit_failure_count,
		circuit_failure_threshold, circuit_reset_timeout_seconds,
		metadata, deleted_at, created_at, updated_at`
}

func (r *EndpointRepository) scanEndpoint(row *sql.Row) (*models.Endpoint, error) {
	e := &models.Endpoint{}
	err := row.Scan(
		&e.ID, &e.ProjectID, &e.OrganizationID,
		&e.Name, &e.URL, &e.Description, &e.IsActive,
		&e.MaxRetries, &e.RetryStrategy, &e.RetryDelaySeconds, &e.RetryMaxDelaySeconds,
		&e.TimeoutSeconds, &e.CustomHeaders, &e.VerifySSL, &e.SigningSecret, &e.RateLimitPerSecond,
		&e.LastSuccessAt, &e.LastFailureAt, &e.ConsecutiveFailures, &e.TotalDeliveries, &e.TotalFailures,
		&e.CircuitBreakerStatus, &e.CircuitOpenedAt, &e.CircuitFailureCount,
		&e.CircuitFailureThreshold, &e.CircuitResetTimeoutSecs,
		&e.Metadata, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *EndpointRepository) Create(e *models.Endpoint) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO endpoints (
			id, project_id, organization_id,
			name, url, description, is_active,
			max_retries, retry_strategy, retry_delay_seconds, retry_max_delay_seconds,
			timeout_seconds, custom_headers, verify_ssl, rate_limit_per_second,
			circuit_breaker_status, circuit_failure_threshold, circuit_reset_timeout_seconds,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21
		)`

	_, err := r.db.Exec(ctx, query,
		e.ID, e.ProjectID, e.OrganizationID,
		e.Name, e.URL, e.Description, e.IsActive,
		e.MaxRetries, e.RetryStrategy, e.RetryDelaySeconds, e.RetryMaxDelaySeconds,
		e.TimeoutSeconds, e.CustomHeaders, e.VerifySSL, e.RateLimitPerSecond,
		e.CircuitBreakerStatus, e.CircuitFailureThreshold, e.CircuitResetTimeoutSecs,
		e.Metadata, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create endpoint", err, "project_id", e.ProjectID)
		return errors.Internal(err)
	}

	logger.Debug("Endpoint created", "endpoint_id", e.ID, "url", e.URL)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *EndpointRepository) GetByIDAndProject(ctx context.Context, endpointID, projectID uuid.UUID) (*models.Endpoint, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM endpoints
		WHERE id = $1 AND project_id = $2 AND deleted_at IS NULL`, endpointCols())

	e, err := r.scanEndpoint(r.db.QueryRow(ctx, query, endpointID, projectID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Endpoint not found")
	}
	if err != nil {
		logger.Error("Failed to get endpoint", err, "endpoint_id", endpointID)
		return nil, errors.Internal(err)
	}
	return e, nil
}

func (r *EndpointRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*models.Endpoint, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM endpoints WHERE project_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRow(ctx, countQuery, projectID).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s FROM endpoints
		WHERE project_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`, endpointCols())

	rows, err := r.db.Query(ctx, listQuery, projectID)
	if err != nil {
		logger.Error("Failed to list endpoints", err, "project_id", projectID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var endpoints []*models.Endpoint
	for rows.Next() {
		e := &models.Endpoint{}
		err := rows.Scan(
			&e.ID, &e.ProjectID, &e.OrganizationID,
			&e.Name, &e.URL, &e.Description, &e.IsActive,
			&e.MaxRetries, &e.RetryStrategy, &e.RetryDelaySeconds, &e.RetryMaxDelaySeconds,
			&e.TimeoutSeconds, &e.CustomHeaders, &e.VerifySSL, &e.SigningSecret, &e.RateLimitPerSecond,
			&e.LastSuccessAt, &e.LastFailureAt, &e.ConsecutiveFailures, &e.TotalDeliveries, &e.TotalFailures,
			&e.CircuitBreakerStatus, &e.CircuitOpenedAt, &e.CircuitFailureCount,
			&e.CircuitFailureThreshold, &e.CircuitResetTimeoutSecs,
			&e.Metadata, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		endpoints = append(endpoints, e)
	}

	return endpoints, total, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *EndpointRepository) Update(e *models.Endpoint) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE endpoints
		SET name = $2, url = $3, description = $4,
		    max_retries = $5, timeout_seconds = $6,
		    custom_headers = $7, verify_ssl = $8, rate_limit_per_second = $9,
		    updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		e.ID, e.Name, e.URL, e.Description,
		e.MaxRetries, e.TimeoutSeconds,
		e.CustomHeaders, e.VerifySSL, e.RateLimitPerSecond,
		time.Now(),
	)
	if err != nil {
		logger.Error("Failed to update endpoint", err, "endpoint_id", e.ID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Endpoint not found")
	}
	return nil
}

func (r *EndpointRepository) SetActive(endpointID uuid.UUID, isActive bool) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `UPDATE endpoints SET is_active = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, endpointID, isActive, time.Now())
	if err != nil {
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Endpoint not found")
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *EndpointRepository) SoftDelete(endpointID uuid.UUID) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	now := time.Now()
	query := `
		UPDATE endpoints
		SET deleted_at = $2, is_active = false, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, endpointID, now, now)
	if err != nil {
		logger.Error("Failed to delete endpoint", err, "endpoint_id", endpointID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Endpoint not found")
	}
	return nil
}