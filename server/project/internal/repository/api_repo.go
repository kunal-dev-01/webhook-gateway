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
// API Key Repository
// ============================================================================

type APIKeyRepository struct {
	db *db.DB
}

func NewAPIKeyRepository(database *db.DB) *APIKeyRepository {
	return &APIKeyRepository{db: database}
}

func (r *APIKeyRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func apiKeyCols() string {
	return `id, organization_id, project_id, created_by,
		name, description, key_type, environment,
		key_prefix, key_hash, scopes,
		rate_limit_per_minute,
		last_used_at, last_used_ip, usage_count,
		expires_at, is_active,
		revoked_at, revoked_by, revoked_reason,
		deleted_at, created_at, updated_at`
}

func (r *APIKeyRepository) scanAPIKey(row *sql.Row) (*models.APIKey, error) {
	k := &models.APIKey{}
	err := row.Scan(
		&k.ID, &k.OrganizationID, &k.ProjectID, &k.CreatedBy,
		&k.Name, &k.Description, &k.KeyType, &k.Environment,
		&k.KeyPrefix, &k.KeyHash, &k.Scopes,
		&k.RateLimitPerMinute,
		&k.LastUsedAt, &k.LastUsedIP, &k.UsageCount,
		&k.ExpiresAt, &k.IsActive,
		&k.RevokedAt, &k.RevokedBy, &k.RevokedReason,
		&k.DeletedAt, &k.CreatedAt, &k.UpdatedAt,
	)
	return k, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *APIKeyRepository) Create(k *models.APIKey) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO api_keys (
			id, organization_id, project_id, created_by,
			name, description, key_type, environment,
			key_prefix, key_hash, scopes,
			rate_limit_per_minute, expires_at, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14,
			$15, $16
		)`

	_, err := r.db.Exec(ctx, query,
		k.ID, k.OrganizationID, k.ProjectID, k.CreatedBy,
		k.Name, k.Description, k.KeyType, k.Environment,
		k.KeyPrefix, k.KeyHash, k.Scopes,
		k.RateLimitPerMinute, k.ExpiresAt, k.IsActive,
		k.CreatedAt, k.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create API key", err, "org_id", k.OrganizationID)
		return errors.Internal(err)
	}

	logger.Debug("API key created", "key_id", k.ID, "prefix", k.KeyPrefix)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *APIKeyRepository) GetByIDAndOrg(ctx context.Context, keyID, orgID uuid.UUID) (*models.APIKey, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM api_keys
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`, apiKeyCols())

	k, err := r.scanAPIKey(r.db.QueryRow(ctx, query, keyID, orgID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("API key not found")
	}
	if err != nil {
		logger.Error("Failed to get API key", err, "key_id", keyID)
		return nil, errors.Internal(err)
	}
	return k, nil
}

// GetByPrefix fetches a key by its prefix — used during auth flow
func (r *APIKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM api_keys
		WHERE key_prefix = $1 AND deleted_at IS NULL`, apiKeyCols())

	k, err := r.scanAPIKey(r.db.QueryRow(ctx, query, prefix))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("API key not found")
	}
	if err != nil {
		return nil, errors.Internal(err)
	}
	return k, nil
}

func (r *APIKeyRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.APIKey, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE organization_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s FROM api_keys
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`, apiKeyCols())

	rows, err := r.db.Query(ctx, listQuery, orgID)
	if err != nil {
		logger.Error("Failed to list API keys", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		k := &models.APIKey{}
		err := rows.Scan(
			&k.ID, &k.OrganizationID, &k.ProjectID, &k.CreatedBy,
			&k.Name, &k.Description, &k.KeyType, &k.Environment,
			&k.KeyPrefix, &k.KeyHash, &k.Scopes,
			&k.RateLimitPerMinute,
			&k.LastUsedAt, &k.LastUsedIP, &k.UsageCount,
			&k.ExpiresAt, &k.IsActive,
			&k.RevokedAt, &k.RevokedBy, &k.RevokedReason,
			&k.DeletedAt, &k.CreatedAt, &k.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		keys = append(keys, k)
	}

	return keys, total, nil
}

func (r *APIKeyRepository) CountActiveByOrg(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM api_keys
		WHERE organization_id = $1 AND deleted_at IS NULL
		  AND is_active = true AND revoked_at IS NULL`
	if err := r.db.QueryRow(ctx, query, orgID).Scan(&count); err != nil {
		return 0, errors.Internal(err)
	}
	return count, nil
}

// ============================================================================
// REVOKE
// ============================================================================

func (r *APIKeyRepository) Revoke(k *models.APIKey) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE api_keys
		SET is_active = false,
		    revoked_at = $2,
		    revoked_by = $3,
		    revoked_reason = $4,
		    updated_at = $5
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		k.ID, k.RevokedAt, k.RevokedBy, k.RevokedReason, time.Now(),
	)
	if err != nil {
		logger.Error("Failed to revoke API key", err, "key_id", k.ID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("API key not found")
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *APIKeyRepository) SoftDelete(keyID uuid.UUID) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	now := time.Now()
	query := `
		UPDATE api_keys
		SET deleted_at = $2, is_active = false, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, keyID, now, now)
	if err != nil {
		logger.Error("Failed to delete API key", err, "key_id", keyID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("API key not found")
	}
	return nil
}