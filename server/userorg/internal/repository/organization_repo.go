package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"userorg/internal/db"
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

// ───────────────── Helpers ─────────────────

func (r *OrganizationRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

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
// ORGANIZATION — CREATE
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

	logger.Debug("Organization created", "org_id", org.ID)
	return nil
}

// ============================================================================
// ORGANIZATION — READ
// ============================================================================

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT
			id, name, slug, owner_id, plan_tier, billing_period, subscription_status,
			lemon_squeezy_customer_id, lemon_squeezy_subscription_id, current_period_start, current_period_end,
			monthly_events_limit, monthly_events_used, retention_days, team_seats_limit, rate_limit_per_second, projects_limit,
			metadata, deleted_at, created_at, updated_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL`

	org, err := r.scanOrg(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Organization not found")
	}
	if err != nil {
		logger.Error("Failed to get organization by ID", err, "org_id", id)
		return nil, errors.Internal(err)
	}
	return org, nil
}

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
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1 AND deleted_at IS NULL)`,
		slug,
	).Scan(&exists); err != nil {
		logger.Error("Failed to check slug existence", err, "slug", slug)
		return false, errors.Internal(err)
	}
	return exists, nil
}

// ============================================================================
// ORGANIZATION — UPDATE
// ============================================================================

func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()
	query := `UPDATE organizations SET name = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	return r.performUpdate(query, "update organization", org.ID, org.ID, org.Name, org.UpdatedAt)
}

func (r *OrganizationRepository) UpdateSubscription(ctx context.Context, org *models.Organization) error {
	org.UpdatedAt = time.Now()
	query := `
		UPDATE organizations SET
			plan_tier = $2, billing_period = $3, subscription_status = $4,
			lemon_squeezy_customer_id = $5, lemon_squeezy_subscription_id = $6,
			current_period_start = $7, current_period_end = $8,
			monthly_events_limit = $9, retention_days = $10,
			team_seats_limit = $11, rate_limit_per_second = $12, projects_limit = $13,
			updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL`

	return r.performUpdate(query, "update subscription", org.ID,
		org.ID, org.PlanTier, org.BillingPeriod, org.SubscriptionStatus,
		org.LemonSqueezyCustomerID, org.LemonSqueezySubscriptionID,
		org.CurrentPeriodStart, org.CurrentPeriodEnd,
		org.MonthlyEventsLimit, org.RetentionDays,
		org.TeamSeatsLimit, org.RateLimitPerSecond, org.ProjectsLimit, org.UpdatedAt,
	)
}

func (r *OrganizationRepository) IncrementEventsUsed(ctx context.Context, orgID uuid.UUID, count int64) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	_, err := r.db.Exec(writeCtx,
		`UPDATE organizations SET monthly_events_used = monthly_events_used + $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`,
		orgID, count, time.Now(),
	)
	if err != nil {
		logger.Error("Failed to increment events", err, "org_id", orgID)
		return errors.Internal(err)
	}
	return nil
}

func (r *OrganizationRepository) ResetMonthlyUsage(ctx context.Context, orgID uuid.UUID) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	_, err := r.db.Exec(writeCtx,
		`UPDATE organizations SET monthly_events_used = 0, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`,
		orgID, time.Now(),
	)
	if err != nil {
		logger.Error("Failed to reset monthly usage", err, "org_id", orgID)
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// ORGANIZATION — DELETE
// ============================================================================

func (r *OrganizationRepository) SoftDelete(ctx context.Context, orgID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE organizations SET deleted_at = $2, slug = CONCAT(slug, '-deleted-', $3), updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	return r.performUpdate(query, "delete organization", orgID, orgID, now, now.Unix(), now)
}

// ============================================================================
// ORGANIZATION — TRANSACTION
// ============================================================================

func (r *OrganizationRepository) TransferOwnership(ctx context.Context, orgID, newOwnerID, oldOwnerID uuid.UUID) error {
	writeCtx, cancel := r.withWriteContext()
	defer cancel()

	tx, err := r.db.BeginTx(writeCtx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", err)
		return errors.Internal(err)
	}
	defer tx.Rollback()

	now := time.Now()

	if _, err := tx.ExecContext(writeCtx, `UPDATE organizations SET owner_id = $1, updated_at = $2 WHERE id = $3`, newOwnerID, now, orgID); err != nil {
		logger.Error("Failed to update org owner", err)
		return errors.Internal(err)
	}
	if _, err := tx.ExecContext(writeCtx, `UPDATE organization_members SET role = $1, updated_at = $2 WHERE organization_id = $3 AND user_id = $4`, models.RoleOwner, now, orgID, newOwnerID); err != nil {
		logger.Error("Failed to promote new owner", err)
		return errors.Internal(err)
	}
	if _, err := tx.ExecContext(writeCtx, `UPDATE organization_members SET role = $1, updated_at = $2 WHERE organization_id = $3 AND user_id = $4`, models.RoleAdmin, now, orgID, oldOwnerID); err != nil {
		logger.Error("Failed to demote old owner", err)
		return errors.Internal(err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit ownership transfer", err)
		return errors.Internal(err)
	}

	logger.Info("Ownership transferred", "org_id", orgID, "new_owner", newOwnerID)
	return nil
}

// ============================================================================
// MEMBER — CREATE
// ============================================================================

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
// MEMBER — READ
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
		logger.Error("Failed to get membership", err, "org_id", orgID, "user_id", userID)
		return nil, errors.Internal(err)
	}
	return member, nil
}

// ExistsMembership is a lightweight bool check — avoids scanning a full row.
func (r *OrganizationRepository) ExistsMembership(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM organization_members WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL)`,
		orgID, userID,
	).Scan(&exists); err != nil {
		logger.Error("Failed to check membership existence", err, "org_id", orgID, "user_id", userID)
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.OrganizationMember, int, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND deleted_at IS NULL`,
		orgID,
	).Scan(&total); err != nil {
		logger.Error("Failed to count members", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}

	query := `
		SELECT id, organization_id, user_id, role, invited_by, invitation_accepted_at, deleted_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		logger.Error("Failed to list members", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var members []*models.OrganizationMember
	for rows.Next() {
		member, err := r.scanMember(rows)
		if err != nil {
			logger.Error("Failed to scan member", err)
			return nil, 0, errors.Internal(err)
		}
		members = append(members, member)
	}
	return members, total, rows.Err()
}

func (r *OrganizationRepository) GetMemberCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM organization_members WHERE organization_id = $1 AND deleted_at IS NULL`,
		orgID,
	).Scan(&count); err != nil {
		logger.Error("Failed to get member count", err, "org_id", orgID)
		return 0, errors.Internal(err)
	}
	return count, nil
}

// ============================================================================
// MEMBER — UPDATE
// ============================================================================

func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role models.OrganizationMemberRole) error {
	query := `
		UPDATE organization_members SET role = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	return r.performUpdate(query, "update member role", userID,
		orgID, userID, role, time.Now(),
	)
}

// AcceptInvitation marks invitation_accepted_at for a pending member.
func (r *OrganizationRepository) AcceptInvitation(ctx context.Context, orgID, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE organization_members
		SET invitation_accepted_at = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2
		  AND deleted_at IS NULL AND invitation_accepted_at IS NULL`

	return r.performUpdate(query, "accept invitation", userID,
		orgID, userID, now, now,
	)
}

// ============================================================================
// MEMBER — DELETE
// ============================================================================

func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE organization_members SET deleted_at = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	return r.performUpdate(query, "remove member", userID,
		orgID, userID, now, now,
	)
}