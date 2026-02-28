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
// CLI Session Repository
// ============================================================================

type CLISessionRepository struct {
	db *db.DB
}

func NewCLISessionRepository(database *db.DB) *CLISessionRepository {
	return &CLISessionRepository{db: database}
}

func (r *CLISessionRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func cliSessionCols() string {
	return `id, user_id, organization_id,
		device_name, device_fingerprint, cli_version, os_info,
		session_token_hash, refresh_token_hash,
		expires_at, refresh_expires_at,
		last_used_at, last_used_ip, last_used_command, usage_count,
		revoked_at, revoked_reason,
		created_ip, created_at, updated_at`
}

func (r *CLISessionRepository) scanSession(row *sql.Row) (*models.CLISession, error) {
	s := &models.CLISession{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.OrganizationID,
		&s.DeviceName, &s.DeviceFingerprint, &s.CLIVersion, &s.OSInfo,
		&s.SessionTokenHash, &s.RefreshTokenHash,
		&s.ExpiresAt, &s.RefreshExpiresAt,
		&s.LastUsedAt, &s.LastUsedIP, &s.LastUsedCommand, &s.UsageCount,
		&s.RevokedAt, &s.RevokedReason,
		&s.CreatedIP, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

// ============================================================================
// READ
// ============================================================================

func (r *CLISessionRepository) GetByIDAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.CLISession, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM cli_sessions
		WHERE id = $1 AND user_id = $2`, cliSessionCols())

	s, err := r.scanSession(r.db.QueryRow(ctx, query, sessionID, userID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("CLI session not found")
	}
	if err != nil {
		logger.Error("Failed to get CLI session", err, "session_id", sessionID)
		return nil, errors.Internal(err)
	}
	return s, nil
}

// ListByUser returns all sessions for a user — active + expired + revoked
// (so users can see history and revoke)
func (r *CLISessionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.CLISession, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM cli_sessions WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s FROM cli_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50`, cliSessionCols()) // cap at 50 — old sessions pruned by cron

	rows, err := r.db.Query(ctx, listQuery, userID)
	if err != nil {
		logger.Error("Failed to list CLI sessions", err, "user_id", userID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var sessions []*models.CLISession
	for rows.Next() {
		s := &models.CLISession{}
		err := rows.Scan(
			&s.ID, &s.UserID, &s.OrganizationID,
			&s.DeviceName, &s.DeviceFingerprint, &s.CLIVersion, &s.OSInfo,
			&s.SessionTokenHash, &s.RefreshTokenHash,
			&s.ExpiresAt, &s.RefreshExpiresAt,
			&s.LastUsedAt, &s.LastUsedIP, &s.LastUsedCommand, &s.UsageCount,
			&s.RevokedAt, &s.RevokedReason,
			&s.CreatedIP, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal(err)
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// ============================================================================
// REVOKE
// ============================================================================

func (r *CLISessionRepository) Revoke(sessionID uuid.UUID, reason string) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	now := time.Now()
	query := `
		UPDATE cli_sessions
		SET revoked_at = $2, revoked_reason = $3, updated_at = $4
		WHERE id = $1 AND revoked_at IS NULL`

	result, err := r.db.Exec(ctx, query, sessionID, now, reason, now)
	if err != nil {
		logger.Error("Failed to revoke CLI session", err, "session_id", sessionID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Session not found or already revoked")
	}
	return nil
}

// RevokeAllByUser revokes all active sessions for a user (e.g., security incident)
// Returns count of revoked sessions
func (r *CLISessionRepository) RevokeAllByUser(userID uuid.UUID, reason string) (int, error) {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	now := time.Now()
	query := `
		UPDATE cli_sessions
		SET revoked_at = $2, revoked_reason = $3, updated_at = $4
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()`

	result, err := r.db.Exec(ctx, query, userID, now, reason, now)
	if err != nil {
		logger.Error("Failed to revoke all CLI sessions", err, "user_id", userID)
		return 0, errors.Internal(err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}