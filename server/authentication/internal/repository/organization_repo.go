package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"authentication/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// Organization Repository
// ============================================================================

type OrganizationRepository struct {
	db *db.DB
}

func NewOrganizationRepository(database *db.DB) *OrganizationRepository {
	return &OrganizationRepository{db: database}
}

// ───────────────── Helpers (Optimization & Safety) ─────────────────

// withWriteContext creates a detached context for safe Write operations.
func (r *OrganizationRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// performUpdate executes an update query safely using a background context.
func (r *OrganizationRepository) performUpdate(query, opName string, resourceID uuid.UUID, args ...any) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to "+opName, err, "id", resourceID)
		return errors.Wrap(err, 500, "Failed to "+opName)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Resource not found or operation not applicable")
	}

	logger.Debug("Success: "+opName, "id", resourceID)
	return nil
}

// scanOrg handles scanning organization rows to avoid duplication
func (r *OrganizationRepository) scanOrg(scanner interface{ Scan(...any) error }) (*models.Organization, error) {
	org := &models.Organization{}
	err := scanner.Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID,
		&org.PlanTier, &org.BillingPeriod, &org.SubscriptionStatus,
		&org.LemonSqueezyCustomerID, &org.LemonSqueezySubscriptionID,
		&org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.MonthlyEventsLimit, &org.MonthlyEventsUsed, &org.RetentionDays,
		&org.TeamSeatsLimit, &org.RateLimitPerSecond, &org.ProjectsLimit,
		&org.Metadata, &org.DeletedAt, &org.CreatedAt, &org.UpdatedAt,
	)
	return org, err
}

// scanMember handles scanning member rows
func (r *OrganizationRepository) scanMember(scanner interface{ Scan(...any) error }) (*models.OrganizationMember, error) {
	member := &models.OrganizationMember{}
	err := scanner.Scan(
		&member.ID, &member.OrganizationID, &member.UserID, &member.Role,
		&member.InvitedBy, &member.InvitationAcceptedAt,
		&member.DeletedAt, &member.CreatedAt, &member.UpdatedAt,
	)
	return member, err
}

// ============================================================================
// CREATE (Safe Write)
// ============================================================================

func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO organizations (
			id, name, slug, owner_id,
			plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id,
			current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days,
			team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err := r.db.Exec(writeCtx, query,
		org.ID, org.Name, org.Slug, org.OwnerID,
		org.PlanTier, org.BillingPeriod, org.SubscriptionStatus,
		org.LemonSqueezyCustomerID, org.LemonSqueezySubscriptionID,
		org.CurrentPeriodStart, org.CurrentPeriodEnd,
		org.MonthlyEventsLimit, org.MonthlyEventsUsed, org.RetentionDays,
		org.TeamSeatsLimit, org.RateLimitPerSecond, org.ProjectsLimit,
		org.Metadata, org.CreatedAt, org.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create organization", err, "name", org.Name)
		return errors.Wrap(err, 500, "Failed to create organization")
	}

	logger.Debug("Organization created successfully", "org_id", org.ID)
	return nil
}

func (r *OrganizationRepository) CreateMembership(ctx context.Context, member *models.OrganizationMember) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO organization_members (
			id, organization_id, user_id, role,
			invited_by, invitation_accepted_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(writeCtx, query,
		member.ID, member.OrganizationID, member.UserID, member.Role,
		member.InvitedBy, member.InvitationAcceptedAt,
		member.CreatedAt, member.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create membership", err, "org_id", member.OrganizationID, "user_id", member.UserID)
		return errors.Wrap(err, 500, "Failed to create organization membership")
	}
	return nil
}

// ============================================================================
// READ (Fast Read - Request Context)
// ============================================================================

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	query := `
		SELECT 
			id, name, slug, owner_id, plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id, current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days, team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, deleted_at, created_at, updated_at
		FROM organizations WHERE slug = $1 AND deleted_at IS NULL`

	org, err := r.scanOrg(r.db.QueryRow(ctx, query, slug))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Organization not found")
	}
	if err != nil {
		logger.Error("Failed to get organization by slug", err, "slug", slug)
		return nil, errors.Internal(err)
	}
	return org, nil
}

func (r *OrganizationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	query := `
		SELECT 
			o.id, o.name, o.slug, o.owner_id, o.plan_tier, o.billing_period, o.subscription_status,
			o.lemon_squeezy_customer_id, o.lemon_squeezy_subscription_id, o.current_period_start, o.current_period_end,
			o.monthly_events_limit, o.monthly_events_used, o.retention_days, o.team_seats_limit, o.rate_limit_per_second, o.projects_limit,
			o.metadata, o.deleted_at, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1 AND o.deleted_at IS NULL AND om.deleted_at IS NULL
		ORDER BY o.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to list user organizations", err, "user_id", userID)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		org, err := r.scanOrg(rows)
		if err != nil {
			logger.Error("Failed to scan organization", err)
			return nil, errors.Internal(err)
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (r *OrganizationRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1 AND deleted_at IS NULL)`
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		logger.Error("Failed to check slug existence", err, "slug", slug)
		return false, errors.Internal(err)
	}
	return exists, nil
}

// ============================================================================
// MEMBERSHIP READS
// ============================================================================

func (r *OrganizationRepository) GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, invited_by, invitation_accepted_at, deleted_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	member, err := r.scanMember(r.db.QueryRow(ctx, query, orgID, userID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Membership not found")
	}
	if err != nil {
		logger.Error("Failed to get membership", err, "org_id", orgID)
		return nil, errors.Internal(err)
	}
	return member, nil
}