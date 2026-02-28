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
// Message Repository
// ============================================================================

type MessageRepository struct {
	db *db.DB
}

func NewMessageRepository(database *db.DB) *MessageRepository {
	return &MessageRepository{db: database}
}

func (r *MessageRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func messageCols() string {
	return `id, ticket_id, sender_id, sender_type, message_type, body,
		is_read_by_user, read_by_user_at, is_read_by_agent, read_by_agent_at,
		is_edited, edited_at, original_body,
		created_at, updated_at`
}

func (r *MessageRepository) scanMessage(row *sql.Row) (*models.SupportMessage, error) {
	m := &models.SupportMessage{}
	err := row.Scan(
		&m.ID, &m.TicketID, &m.SenderID, &m.SenderType, &m.MessageType, &m.Body,
		&m.IsReadByUser, &m.ReadByUserAt, &m.IsReadByAgent, &m.ReadByAgentAt,
		&m.IsEdited, &m.EditedAt, &m.OriginalBody,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *MessageRepository) Create(msg *models.SupportMessage) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO support_messages (
			id, ticket_id, sender_id, sender_type, message_type, body,
			is_read_by_user, is_read_by_agent,
			is_edited,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8,
			$9,
			$10, $11
		)`

	_, err := r.db.Exec(ctx, query,
		msg.ID, msg.TicketID, msg.SenderID, msg.SenderType, msg.MessageType, msg.Body,
		msg.IsReadByUser, msg.IsReadByAgent,
		msg.IsEdited,
		msg.CreatedAt, msg.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create support message", err, "ticket_id", msg.TicketID)
		return errors.Internal(err)
	}

	logger.Debug("Support message created", "message_id", msg.ID, "ticket_id", msg.TicketID)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *MessageRepository) GetByID(ctx context.Context, messageID uuid.UUID) (*models.SupportMessage, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_messages WHERE id = $1`, messageCols())

	msg, err := r.scanMessage(r.db.QueryRow(ctx, query, messageID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Message not found")
	}
	if err != nil {
		logger.Error("Failed to get message by ID", err, "message_id", messageID)
		return nil, errors.Internal(err)
	}
	return msg, nil
}

// ListByTicket returns all user-visible messages for a ticket (excludes internal_notes)
func (r *MessageRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]*models.SupportMessage, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM support_messages
		WHERE ticket_id = $1 AND message_type != 'internal_note'
		ORDER BY created_at ASC`,
		messageCols(),
	)

	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		logger.Error("Failed to list messages for ticket", err, "ticket_id", ticketID)
		return nil, errors.Internal(err)
	}
	defer rows.Close()

	var messages []*models.SupportMessage
	for rows.Next() {
		m := &models.SupportMessage{}
		err := rows.Scan(
			&m.ID, &m.TicketID, &m.SenderID, &m.SenderType, &m.MessageType, &m.Body,
			&m.IsReadByUser, &m.ReadByUserAt, &m.IsReadByAgent, &m.ReadByAgentAt,
			&m.IsEdited, &m.EditedAt, &m.OriginalBody,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan message row", err)
			return nil, errors.Internal(err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

// ============================================================================
// UPDATE
// ============================================================================

// MarkReadByUser bulk-marks all agent/system messages in a ticket as read by user
func (r *MessageRepository) MarkAllReadByUser(ticketID uuid.UUID) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE support_messages
		SET is_read_by_user = true, read_by_user_at = $2, updated_at = $3
		WHERE ticket_id = $1
		  AND sender_type != 'user'
		  AND is_read_by_user = false
		  AND message_type != 'internal_note'`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, ticketID, now, now)
	if err != nil {
		logger.Error("Failed to mark messages read", err, "ticket_id", ticketID)
		return errors.Internal(err)
	}
	return nil
}