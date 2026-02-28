package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"support/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// Attachment Repository
// ============================================================================

type AttachmentRepository struct {
	db *db.DB
}

func NewAttachmentRepository(database *db.DB) *AttachmentRepository {
	return &AttachmentRepository{db: database}
}

func (r *AttachmentRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *AttachmentRepository) scanAttachment(row *sql.Row) (*models.SupportAttachment, error) {
	a := &models.SupportAttachment{}
	err := row.Scan(
		&a.ID, &a.TicketID, &a.MessageID, &a.UploadedBy,
		&a.FileName, &a.FileSizeBytes, &a.MimeType,
		&a.StorageKey, &a.StorageBucket,
		&a.ScanStatus,
		&a.CreatedAt,
	)
	return a, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *AttachmentRepository) Create(attachment *models.SupportAttachment) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO support_attachments (
			id, ticket_id, message_id, uploaded_by,
			file_name, file_size_bytes, mime_type,
			storage_key, storage_bucket,
			scan_status,
			created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10,
			$11
		)`

	_, err := r.db.Exec(ctx, query,
		attachment.ID, attachment.TicketID, attachment.MessageID, attachment.UploadedBy,
		attachment.FileName, attachment.FileSizeBytes, attachment.MimeType,
		attachment.StorageKey, attachment.StorageBucket,
		attachment.ScanStatus,
		attachment.CreatedAt,
	)
	if err != nil {
		logger.Error("Failed to create attachment", err, "ticket_id", attachment.TicketID)
		return errors.Internal(err)
	}

	logger.Debug("Attachment created", "attachment_id", attachment.ID, "ticket_id", attachment.TicketID)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *AttachmentRepository) GetByID(ctx context.Context, attachmentID uuid.UUID) (*models.SupportAttachment, error) {
	query := `
		SELECT id, ticket_id, message_id, uploaded_by,
			file_name, file_size_bytes, mime_type,
			storage_key, storage_bucket,
			scan_status, created_at
		FROM support_attachments WHERE id = $1`

	a, err := r.scanAttachment(r.db.QueryRow(ctx, query, attachmentID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Attachment not found")
	}
	if err != nil {
		logger.Error("Failed to get attachment", err, "attachment_id", attachmentID)
		return nil, errors.Internal(err)
	}
	return a, nil
}

// GetByIDAndTicket fetches attachment only if it belongs to the given ticket (scoped access)
func (r *AttachmentRepository) GetByIDAndTicket(ctx context.Context, attachmentID, ticketID uuid.UUID) (*models.SupportAttachment, error) {
	query := `
		SELECT id, ticket_id, message_id, uploaded_by,
			file_name, file_size_bytes, mime_type,
			storage_key, storage_bucket,
			scan_status, created_at
		FROM support_attachments WHERE id = $1 AND ticket_id = $2`

	a, err := r.scanAttachment(r.db.QueryRow(ctx, query, attachmentID, ticketID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Attachment not found")
	}
	if err != nil {
		logger.Error("Failed to get attachment", err, "attachment_id", attachmentID)
		return nil, errors.Internal(err)
	}
	return a, nil
}

// ListByMessageIDs fetches all attachments for a list of message IDs (for bulk loading)
func (r *AttachmentRepository) ListByMessageIDs(ctx context.Context, messageIDs []uuid.UUID) ([]*models.SupportAttachment, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	// Build $1,$2,$3,... placeholder
	placeholders := ""
	args := make([]any, len(messageIDs))
	for i, id := range messageIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "$" + itoa(i+1)
		args[i] = id
	}

	query := `
		SELECT id, ticket_id, message_id, uploaded_by,
			file_name, file_size_bytes, mime_type,
			storage_key, storage_bucket,
			scan_status, created_at
		FROM support_attachments
		WHERE message_id IN (` + placeholders + `)
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to list attachments by message IDs", err)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var attachments []*models.SupportAttachment
	for rows.Next() {
		a := &models.SupportAttachment{}
		err := rows.Scan(
			&a.ID, &a.TicketID, &a.MessageID, &a.UploadedBy,
			&a.FileName, &a.FileSizeBytes, &a.MimeType,
			&a.StorageKey, &a.StorageBucket,
			&a.ScanStatus, &a.CreatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan attachment row", err)
			return nil, errors.Internal(err)
		}
		attachments = append(attachments, a)
	}

	return attachments, nil
}

// CountByTicket returns the total number of attachments for a ticket (for limit checks)
func (r *AttachmentRepository) CountByTicket(ctx context.Context, ticketID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM support_attachments WHERE ticket_id = $1`
	if err := r.db.QueryRow(ctx, query, ticketID).Scan(&count); err != nil {
		return 0, errors.Internal(err)
	}
	return count, nil
}

// ============================================================================
// Helpers
// ============================================================================

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}