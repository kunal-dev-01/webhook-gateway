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
// Domain Repository
// ============================================================================

type DomainRepository struct {
	db *db.DB
}

func NewDomainRepository(database *db.DB) *DomainRepository {
	return &DomainRepository{db: database}
}

func (r *DomainRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func domainCols() string {
	return `id, organization_id, project_id, domain, domain_type,
		verification_token, verification_status, verified_at, last_dns_check_at,
		dns_check_attempts, dns_check_error,
		ssl_status, ssl_provider, ssl_issued_at, ssl_expires_at,
		ssl_auto_renew, ssl_error,
		is_active, deleted_at, created_at, updated_at`
}

func (r *DomainRepository) scanDomain(row *sql.Row) (*models.CustomDomain, error) {
	d := &models.CustomDomain{}
	err := row.Scan(
		&d.ID, &d.OrganizationID, &d.ProjectID, &d.Domain, &d.DomainType,
		&d.VerificationToken, &d.VerificationStatus, &d.VerifiedAt, &d.LastDNSCheckAt,
		&d.DNSCheckAttempts, &d.DNSCheckError,
		&d.SSLStatus, &d.SSLProvider, &d.SSLIssuedAt, &d.SSLExpiresAt,
		&d.SSLAutoRenew, &d.SSLError,
		&d.IsActive, &d.DeletedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *DomainRepository) Create(d *models.CustomDomain) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO custom_domains (
			id, organization_id, project_id, domain, domain_type,
			verification_token, verification_status,
			ssl_status, ssl_provider, ssl_auto_renew,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		d.ID, d.OrganizationID, d.ProjectID, d.Domain, d.DomainType,
		d.VerificationToken, d.VerificationStatus,
		d.SSLStatus, d.SSLProvider, d.SSLAutoRenew,
		d.IsActive, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create custom domain", err, "domain", d.Domain, "org_id", d.OrganizationID)
		return errors.Internal(err)
	}

	logger.Debug("Custom domain added", "domain_id", d.ID, "domain", d.Domain)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *DomainRepository) GetByIDAndOrg(ctx context.Context, domainID, orgID uuid.UUID) (*models.CustomDomain, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM custom_domains
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`, domainCols())

	d, err := r.scanDomain(r.db.QueryRow(ctx, query, domainID, orgID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Domain not found")
	}
	if err != nil {
		logger.Error("Failed to get domain", err, "domain_id", domainID)
		return nil, errors.Internal(err)
	}
	return d, nil
}

func (r *DomainRepository) GetByDomain(ctx context.Context, domain string) (*models.CustomDomain, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM custom_domains
		WHERE domain = $1 AND deleted_at IS NULL`, domainCols())

	d, err := r.scanDomain(r.db.QueryRow(ctx, query, domain))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Domain not found")
	}
	if err != nil {
		return nil, errors.Internal(err)
	}
	return d, nil
}

func (r *DomainRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM custom_domains WHERE domain = $1 AND deleted_at IS NULL)`
	if err := r.db.QueryRow(ctx, query, domain).Scan(&exists); err != nil {
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *DomainRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.CustomDomain, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM custom_domains WHERE organization_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s FROM custom_domains
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`, domainCols())

	rows, err := r.db.Query(ctx, listQuery, orgID)
	if err != nil {
		logger.Error("Failed to list domains", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var domains []*models.CustomDomain
	for rows.Next() {
		d := &models.CustomDomain{}
		err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.ProjectID, &d.Domain, &d.DomainType,
			&d.VerificationToken, &d.VerificationStatus, &d.VerifiedAt, &d.LastDNSCheckAt,
			&d.DNSCheckAttempts, &d.DNSCheckError,
			&d.SSLStatus, &d.SSLProvider, &d.SSLIssuedAt, &d.SSLExpiresAt,
			&d.SSLAutoRenew, &d.SSLError,
			&d.IsActive, &d.DeletedAt, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		domains = append(domains, d)
	}

	return domains, total, nil
}

// ============================================================================
// UPDATE — Domain state changes (called by cron worker AND user trigger)
// ============================================================================

func (r *DomainRepository) UpdateVerificationStatus(d *models.CustomDomain) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE custom_domains
		SET verification_status = $2,
		    verified_at = $3,
		    last_dns_check_at = $4,
		    dns_check_attempts = $5,
		    dns_check_error = $6,
		    ssl_status = $7,
		    updated_at = $8
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		d.ID,
		d.VerificationStatus,
		d.VerifiedAt,
		d.LastDNSCheckAt,
		d.DNSCheckAttempts,
		d.DNSCheckError,
		d.SSLStatus,
		time.Now(),
	)
	if err != nil {
		logger.Error("Failed to update domain verification status", err, "domain_id", d.ID)
		return errors.Internal(err)
	}
	return nil
}

func (r *DomainRepository) UpdateSSLStatus(d *models.CustomDomain) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE custom_domains
		SET ssl_status = $2, ssl_issued_at = $3, ssl_expires_at = $4,
		    ssl_error = $5, is_active = $6, updated_at = $7
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		d.ID,
		d.SSLStatus,
		d.SSLIssuedAt,
		d.SSLExpiresAt,
		d.SSLError,
		d.IsActive,
		time.Now(),
	)
	if err != nil {
		logger.Error("Failed to update domain SSL status", err, "domain_id", d.ID)
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *DomainRepository) SoftDelete(domainID uuid.UUID) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	now := time.Now()
	query := `
		UPDATE custom_domains
		SET deleted_at = $2, is_active = false, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, domainID, now, now)
	if err != nil {
		logger.Error("Failed to delete domain", err, "domain_id", domainID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Domain not found")
	}
	return nil
}