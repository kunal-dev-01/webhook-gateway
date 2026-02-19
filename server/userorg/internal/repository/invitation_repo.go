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
// Invitation Repository
// ============================================================================

type InvitationRepository struct {
	db *db.DB
}

func NewInvitationRepository(database *db.DB) *InvitationRepository {
	return &InvitationRepository{db: database}
}

// ───────────────── Helpers ─────────────────

func (r *InvitationRepository) withWriteCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *InvitationRepository) scanInvitation(scanner interface{ Scan(...any) error }) (*models.Invitation, error) {
	inv := &models.Invitation{}
	err := scanner.Scan(
		&inv.ID, &inv.OrganizationID, &inv.Email, &inv.Role,
		&inv.InvitedBy, &inv.Token, &inv.Status,
		&inv.AcceptedBy, &inv.AcceptedAt, &inv.ExpiresAt,
		&inv.Metadata, &inv.CreatedAt, &inv.UpdatedAt,
	)
	return inv, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *InvitationRepository) Create(ctx context.Context, inv *models.Invitation) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	query := `
		INSERT INTO invitations (
			id, organization_id, email, role,
			invited_by, token, status,
			accepted_by, accepted_at, expires_at,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	_, err := r.db.Exec(writeCtx, query,
		inv.ID, inv.OrganizationID, inv.Email, inv.Role,
		inv.InvitedBy, inv.Token, inv.Status,
		inv.AcceptedBy, inv.AcceptedAt, inv.ExpiresAt,
		inv.Metadata, inv.CreatedAt, inv.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create invitation", err, "email", inv.Email, "org_id", inv.OrganizationID)
		return errors.Wrap(err, 500, "Failed to create invitation")
	}

	logger.Debug("Invitation created", "invitation_id", inv.ID)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *InvitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Invitation, error) {
	query := `
		SELECT id, organization_id, email, role, invited_by, token, status,
		       accepted_by, accepted_at, expires_at, metadata, created_at, updated_at
		FROM invitations WHERE id = $1`

	inv, err := r.scanInvitation(r.db.QueryRow(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Invitation not found")
	}
	if err != nil {
		logger.Error("Failed to get invitation by ID", err, "id", id)
		return nil, errors.Internal(err)
	}
	return inv, nil
}

func (r *InvitationRepository) GetByToken(ctx context.Context, token string) (*models.Invitation, error) {
	query := `
		SELECT id, organization_id, email, role, invited_by, token, status,
		       accepted_by, accepted_at, expires_at, metadata, created_at, updated_at
		FROM invitations WHERE token = $1`

	inv, err := r.scanInvitation(r.db.QueryRow(ctx, query, token))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Invitation not found")
	}
	if err != nil {
		logger.Error("Failed to get invitation by token", err)
		return nil, errors.Internal(err)
	}
	return inv, nil
}

// ExistsPendingForEmail checks if a pending invitation already exists for this email+org combo.
func (r *InvitationRepository) ExistsPendingForEmail(ctx context.Context, orgID uuid.UUID, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM invitations
			WHERE organization_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()
		)`,
		orgID, email,
	).Scan(&exists)
	if err != nil {
		logger.Error("Failed to check pending invitation", err, "org_id", orgID, "email", email)
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *InvitationRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.Invitation, int, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM invitations WHERE organization_id = $1`, orgID,
	).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, email, role, invited_by, token, status,
		       accepted_by, accepted_at, expires_at, metadata, created_at, updated_at
		FROM invitations WHERE organization_id = $1
		ORDER BY created_at DESC`, orgID)
	if err != nil {
		logger.Error("Failed to list invitations", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		inv, err := r.scanInvitation(rows)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, total, rows.Err()
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *InvitationRepository) UpdateStatus(ctx context.Context, inv *models.Invitation) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	inv.UpdatedAt = time.Now()
	_, err := r.db.Exec(writeCtx, `
		UPDATE invitations
		SET status = $2, accepted_by = $3, accepted_at = $4, updated_at = $5
		WHERE id = $1`,
		inv.ID, inv.Status, inv.AcceptedBy, inv.AcceptedAt, inv.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to update invitation status", err, "id", inv.ID)
		return errors.Wrap(err, 500, "Failed to update invitation")
	}
	return nil
}

// MarkExpiredBulk marks all past-deadline pending invitations as expired (for cron).
func (r *InvitationRepository) MarkExpiredBulk(ctx context.Context) (int64, error) {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	result, err := r.db.Exec(writeCtx, `
		UPDATE invitations SET status = 'expired', updated_at = $1
		WHERE status = 'pending' AND expires_at <= $1`,
		time.Now(),
	)
	if err != nil {
		logger.Error("Failed to bulk-expire invitations", err)
		return 0, errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}