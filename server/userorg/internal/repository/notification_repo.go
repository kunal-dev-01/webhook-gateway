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
// Notification Repository
// ============================================================================

type NotificationRepository struct {
	db *db.DB
}

func NewNotificationRepository(database *db.DB) *NotificationRepository {
	return &NotificationRepository{db: database}
}

// ───────────────── Helpers ─────────────────

func (r *NotificationRepository) withWriteCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *NotificationRepository) scanNotification(scanner interface{ Scan(...any) error }) (*models.Notification, error) {
	n := &models.Notification{}
	err := scanner.Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
		&n.ActionURL, &n.ActionLabel,
		&n.RelatedOrganizationID, &n.RelatedInvitationID, &n.RelatedSubscriptionID,
		&n.Metadata, &n.IsRead, &n.ReadAt, &n.Priority, &n.ExpiresAt,
		&n.CreatedAt, &n.UpdatedAt,
	)
	return n, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	query := `
		INSERT INTO notifications (
			id, user_id, type, title, message,
			action_url, action_label,
			related_organization_id, related_invitation_id, related_subscription_id,
			metadata, is_read, read_at, priority, expires_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`

	_, err := r.db.Exec(writeCtx, query,
		n.ID, n.UserID, n.Type, n.Title, n.Message,
		n.ActionURL, n.ActionLabel,
		n.RelatedOrganizationID, n.RelatedInvitationID, n.RelatedSubscriptionID,
		n.Metadata, n.IsRead, n.ReadAt, n.Priority, n.ExpiresAt,
		n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create notification", err, "user_id", n.UserID, "type", n.Type)
		return errors.Wrap(err, 500, "Failed to create notification")
	}

	logger.Debug("Notification created", "notification_id", n.ID, "user_id", n.UserID)
	return nil
}

// CreateBulk inserts multiple notifications in a single transaction.
func (r *NotificationRepository) CreateBulk(ctx context.Context, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	tx, err := r.db.BeginTx(writeCtx, nil)
	if err != nil {
		return errors.Internal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(writeCtx, `
		INSERT INTO notifications (
			id, user_id, type, title, message,
			action_url, action_label,
			related_organization_id, related_invitation_id, related_subscription_id,
			metadata, is_read, read_at, priority, expires_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`)
	if err != nil {
		return errors.Internal(err)
	}
	defer stmt.Close()

	for _, n := range notifications {
		if _, err := stmt.ExecContext(writeCtx,
			n.ID, n.UserID, n.Type, n.Title, n.Message,
			n.ActionURL, n.ActionLabel,
			n.RelatedOrganizationID, n.RelatedInvitationID, n.RelatedSubscriptionID,
			n.Metadata, n.IsRead, n.ReadAt, n.Priority, n.ExpiresAt,
			n.CreatedAt, n.UpdatedAt,
		); err != nil {
			logger.Error("Failed to insert notification in bulk", err, "user_id", n.UserID)
			return errors.Wrap(err, 500, "Failed to create notifications")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *NotificationRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message,
		       action_url, action_label,
		       related_organization_id, related_invitation_id, related_subscription_id,
		       metadata, is_read, read_at, priority, expires_at,
		       created_at, updated_at
		FROM notifications WHERE id = $1 AND user_id = $2`

	n, err := r.scanNotification(r.db.QueryRow(ctx, query, id, userID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Notification not found")
	}
	if err != nil {
		return nil, errors.Internal(err)
	}
	return n, nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, title, message,
		       action_url, action_label,
		       related_organization_id, related_invitation_id, related_subscription_id,
		       metadata, is_read, read_at, priority, expires_at,
		       created_at, updated_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		logger.Error("Failed to list notifications", err, "user_id", userID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		n, err := r.scanNotification(rows)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		notifications = append(notifications, n)
	}
	return notifications, total, rows.Err()
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID,
	).Scan(&count)
	if err != nil {
		return 0, errors.Internal(err)
	}
	return count, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	now := time.Now()
	result, err := r.db.Exec(writeCtx,
		`UPDATE notifications SET is_read = true, read_at = $3, updated_at = $3 WHERE id = $1 AND user_id = $2`,
		id, userID, now,
	)
	if err != nil {
		return errors.Internal(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.NotFound("Notification not found")
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	now := time.Now()
	_, err := r.db.Exec(writeCtx,
		`UPDATE notifications SET is_read = true, read_at = $2, updated_at = $2 WHERE user_id = $1 AND is_read = false`,
		userID, now,
	)
	if err != nil {
		logger.Error("Failed to mark all notifications read", err, "user_id", userID)
		return errors.Internal(err)
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *NotificationRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	writeCtx, cancel := r.withWriteCtx()
	defer cancel()

	result, err := r.db.Exec(writeCtx,
		`DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		logger.Error("Failed to delete notification", err, "id", id)
		return errors.Internal(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.NotFound("Notification not found")
	}
	return nil
}