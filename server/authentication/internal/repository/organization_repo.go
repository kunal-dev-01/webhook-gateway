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

// ============================================================================
// CREATE
// ============================================================================

// Create creates a new organization with default settings
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (
			id, name, slug, owner_id,
			plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id,
			current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days,
			team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := r.db.Exec(ctx, query,
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

	return nil
}

// CreateMembership creates organization membership (adds user to org)
func (r *OrganizationRepository) CreateMembership(ctx context.Context, member *models.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (
			id, organization_id, user_id, role,
			invited_by, invitation_accepted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.db.Exec(ctx, query,
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
// READ
// ============================================================================

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	org := &models.Organization{}
	query := `
		SELECT 
			id, name, slug, owner_id,
			plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id,
			current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days,
			team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, deleted_at, created_at, updated_at
		FROM organizations 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID,
		&org.PlanTier, &org.BillingPeriod, &org.SubscriptionStatus,
		&org.LemonSqueezyCustomerID, &org.LemonSqueezySubscriptionID,
		&org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.MonthlyEventsLimit, &org.MonthlyEventsUsed, &org.RetentionDays,
		&org.TeamSeatsLimit, &org.RateLimitPerSecond, &org.ProjectsLimit,
		&org.Metadata, &org.DeletedAt, &org.CreatedAt, &org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Organization not found")
	}
	if err != nil {
		logger.Error("Failed to get organization by ID", err, "org_id", id)
		return nil, errors.Internal(err)
	}

	return org, nil
}

// GetBySlug retrieves an organization by slug
func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	org := &models.Organization{}
	query := `
		SELECT 
			id, name, slug, owner_id,
			plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id,
			current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days,
			team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, deleted_at, created_at, updated_at
		FROM organizations 
		WHERE slug = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID,
		&org.PlanTier, &org.BillingPeriod, &org.SubscriptionStatus,
		&org.LemonSqueezyCustomerID, &org.LemonSqueezySubscriptionID,
		&org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.MonthlyEventsLimit, &org.MonthlyEventsUsed, &org.RetentionDays,
		&org.TeamSeatsLimit, &org.RateLimitPerSecond, &org.ProjectsLimit,
		&org.Metadata, &org.DeletedAt, &org.CreatedAt, &org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Organization not found")
	}
	if err != nil {
		logger.Error("Failed to get organization by slug", err, "slug", slug)
		return nil, errors.Internal(err)
	}

	return org, nil
}

// ListByUser retrieves all organizations where user is a member
func (r *OrganizationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	query := `
		SELECT 
			o.id, o.name, o.slug, o.owner_id,
			o.plan_tier, o.billing_period, o.subscription_status,
			o.lemon_squeezy_customer_id, o.lemon_squeezy_subscription_id,
			o.current_period_start, o.current_period_end,
			o.monthly_events_limit, o.monthly_events_used, o.retention_days,
			o.team_seats_limit, o.rate_limit_per_second, o.projects_limit,
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
		org := &models.Organization{}
		if err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.OwnerID,
			&org.PlanTier, &org.BillingPeriod, &org.SubscriptionStatus,
			&org.LemonSqueezyCustomerID, &org.LemonSqueezySubscriptionID,
			&org.CurrentPeriodStart, &org.CurrentPeriodEnd,
			&org.MonthlyEventsLimit, &org.MonthlyEventsUsed, &org.RetentionDays,
			&org.TeamSeatsLimit, &org.RateLimitPerSecond, &org.ProjectsLimit,
			&org.Metadata, &org.DeletedAt, &org.CreatedAt, &org.UpdatedAt,
		); err != nil {
			logger.Error("Failed to scan organization", err)
			return nil, errors.Internal(err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// ExistsBySlug checks if organization exists with given slug
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
// MEMBERSHIP QUERIES
// ============================================================================

// GetMembership retrieves user's membership in organization
func (r *OrganizationRepository) GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
	member := &models.OrganizationMember{}
	query := `
		SELECT 
			id, organization_id, user_id, role,
			invited_by, invitation_accepted_at,
			deleted_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, orgID, userID).Scan(
		&member.ID, &member.OrganizationID, &member.UserID, &member.Role,
		&member.InvitedBy, &member.InvitationAcceptedAt,
		&member.DeletedAt, &member.CreatedAt, &member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Membership not found")
	}
	if err != nil {
		logger.Error("Failed to get membership", err, "org_id", orgID, "user_id", userID)
		return nil, errors.Internal(err)
	}

	return member, nil
}

// ListMembers retrieves all members of an organization
func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.OrganizationMember, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM organization_members 
		WHERE organization_id = $1 AND deleted_at IS NULL`
	
	var total int
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		logger.Error("Failed to count members", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}

	// Get members
	query := `
		SELECT 
			id, organization_id, user_id, role,
			invited_by, invitation_accepted_at,
			deleted_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		logger.Error("Failed to list members", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var members []*models.OrganizationMember
	for rows.Next() {
		member := &models.OrganizationMember{}
		if err := rows.Scan(
			&member.ID, &member.OrganizationID, &member.UserID, &member.Role,
			&member.InvitedBy, &member.InvitationAcceptedAt,
			&member.DeletedAt, &member.CreatedAt, &member.UpdatedAt,
		); err != nil {
			logger.Error("Failed to scan member", err)
			return nil, 0, errors.Internal(err)
		}
		members = append(members, member)
	}

	return members, total, nil
}

// GetMemberCount returns active member count for an organization
func (r *OrganizationRepository) GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM organization_members 
		WHERE organization_id = $1 AND deleted_at IS NULL`
	
	if err := r.db.QueryRow(ctx, query, orgID).Scan(&count); err != nil {
		logger.Error("Failed to get member count", err, "org_id", orgID)
		return 0, errors.Internal(err)
	}
	
	return count, nil
}

// ============================================================================
// UPDATE
// ============================================================================

// Update updates organization details
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()
	
	query := `
		UPDATE organizations SET
			name = $2,
			slug = $3,
			metadata = $4,
			updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		org.ID, org.Name, org.Slug, org.Metadata, org.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update organization", err, "org_id", org.ID)
		return errors.Wrap(err, 500, "Failed to update organization")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Organization not found")
	}

	return nil
}

// UpdateSubscription updates subscription details
func (r *OrganizationRepository) UpdateSubscription(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()
	
	query := `
		UPDATE organizations SET
			plan_tier = $2,
			billing_period = $3,
			subscription_status = $4,
			lemon_squeezy_customer_id = $5,
			lemon_squeezy_subscription_id = $6,
			current_period_start = $7,
			current_period_end = $8,
			monthly_events_limit = $9,
			retention_days = $10,
			team_seats_limit = $11,
			rate_limit_per_second = $12,
			projects_limit = $13,
			updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		org.ID, org.PlanTier, org.BillingPeriod, org.SubscriptionStatus,
		org.LemonSqueezyCustomerID, org.LemonSqueezySubscriptionID,
		org.CurrentPeriodStart, org.CurrentPeriodEnd,
		org.MonthlyEventsLimit, org.RetentionDays, org.TeamSeatsLimit,
		org.RateLimitPerSecond, org.ProjectsLimit, org.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update subscription", err, "org_id", org.ID)
		return errors.Wrap(err, 500, "Failed to update subscription")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Organization not found")
	}

	return nil
}

// IncrementEventsUsed increments monthly events counter
func (r *OrganizationRepository) IncrementEventsUsed(ctx context.Context, orgID uuid.UUID, count int64) error {
	query := `
		UPDATE organizations 
		SET 
			monthly_events_used = monthly_events_used + $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, orgID, count, now)
	
	if err != nil {
		logger.Error("Failed to increment events used", err, "org_id", orgID)
		return errors.Internal(err)
	}

	return nil
}

// ResetMonthlyUsage resets monthly events counter (called by cron)
func (r *OrganizationRepository) ResetMonthlyUsage(ctx context.Context, orgID uuid.UUID) error {
	query := `
		UPDATE organizations 
		SET 
			monthly_events_used = 0,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, orgID, now)
	
	if err != nil {
		logger.Error("Failed to reset monthly usage", err, "org_id", orgID)
		return errors.Internal(err)
	}

	return nil
}

// UpdateMemberRole updates member's role in organization
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role models.OrganizationMemberRole) error {
	now := time.Now()
	query := `
		UPDATE organization_members 
		SET role = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, orgID, userID, role, now)
	
	if err != nil {
		logger.Error("Failed to update member role", err, "org_id", orgID, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Member not found")
	}

	return nil
}

// TransferOwnership transfers organization ownership to another member
func (r *OrganizationRepository) TransferOwnership(ctx context.Context, orgID, newOwnerID, oldOwnerID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", err)
		return errors.Internal(err)
	}
	defer tx.Rollback() // pgx requires context in Rollback

	now := time.Now()

	// Update organization owner
	query1 := `UPDATE organizations SET owner_id = $2, updated_at = $3 WHERE id = $1`
	if _, err := tx.Exec(query1, orgID, newOwnerID, now); err != nil {
		logger.Error("Failed to update organization owner", err)
		return errors.Internal(err)
	}

	// Update new owner's role to owner
	query2 := `
		UPDATE organization_members 
		SET role = $3, updated_at = $4 
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	if _, err := tx.Exec(query2, orgID, newOwnerID, models.RoleOwner, now); err != nil {
		logger.Error("Failed to update new owner role", err)
		return errors.Internal(err)
	}

	// Downgrade old owner to admin (optional but recommended)
	query3 := `
		UPDATE organization_members 
		SET role = $3, updated_at = $4 
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	if _, err := tx.Exec(query3, orgID, oldOwnerID, models.RoleAdmin, now); err != nil {
		logger.Error("Failed to downgrade old owner role", err)
		return errors.Internal(err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit ownership transfer", err)
		return errors.Internal(err)
	}

	return nil
}

// ============================================================================
// DELETE
// ============================================================================

// RemoveMember soft deletes a member from organization
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE organization_members 
		SET deleted_at = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, orgID, userID, now, now)
	
	if err != nil {
		logger.Error("Failed to remove member", err, "org_id", orgID, "user_id", userID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Member not found")
	}

	return nil
}

// SoftDelete soft deletes an organization
func (r *OrganizationRepository) SoftDelete(ctx context.Context, orgID uuid.UUID) error {
	now := time.Now()
	
	// Append timestamp to slug to allow recreation
	query := `
		UPDATE organizations 
		SET 
			deleted_at = $2,
			slug = CONCAT(slug, '-deleted-', $3),
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, orgID, now, now.Unix(), now)
	
	if err != nil {
		logger.Error("Failed to soft delete organization", err, "org_id", orgID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Organization not found or already deleted")
	}

	return nil
}