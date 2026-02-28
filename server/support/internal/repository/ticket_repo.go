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
// Ticket Repository
// ============================================================================

type TicketRepository struct {
	db *db.DB
}

func NewTicketRepository(database *db.DB) *TicketRepository {
	return &TicketRepository{db: database}
}

// ───────────────── Helpers ─────────────────

func (r *TicketRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// scanTicket handles scanning a single ticket row
func (r *TicketRepository) scanTicket(row *sql.Row) (*models.SupportTicket, error) {
	t := &models.SupportTicket{}
	err := row.Scan(
		&t.ID, &t.TicketNumber, &t.UserID, &t.OrganizationID,
		&t.Subject, &t.Category, &t.Priority, &t.Status,
		&t.AssignedTo,
		&t.RelatedProjectID, &t.RelatedEndpointID,
		&t.PlanTier,
		&t.FirstResponseAt, &t.ResolvedAt, &t.ClosedAt,
		&t.SatisfactionRating, &t.SatisfactionComment,
		&t.Metadata,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

func ticketCols() string {
	return `id, ticket_number, user_id, organization_id,
		subject, category, priority, status,
		assigned_to,
		related_project_id, related_endpoint_id,
		plan_tier,
		first_response_at, resolved_at, closed_at,
		satisfaction_rating, satisfaction_comment,
		metadata,
		created_at, updated_at`
}

// ============================================================================
// CREATE
// ============================================================================

func (r *TicketRepository) Create(ticket *models.SupportTicket) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO support_tickets (
			id, user_id, organization_id,
			subject, category, priority, status,
			related_project_id, related_endpoint_id,
			plan_tier, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9,
			$10, $11,
			$12, $13
		)
		RETURNING ticket_number`

	err := r.db.QueryRow(ctx, query,
		ticket.ID, ticket.UserID, ticket.OrganizationID,
		ticket.Subject, ticket.Category, ticket.Priority, ticket.Status,
		ticket.RelatedProjectID, ticket.RelatedEndpointID,
		ticket.PlanTier, ticket.Metadata,
		ticket.CreatedAt, ticket.UpdatedAt,
	).Scan(&ticket.TicketNumber)

	if err != nil {
		logger.Error("Failed to create support ticket", err, "user_id", ticket.UserID)
		return errors.Internal(err)
	}

	logger.Debug("Support ticket created", "ticket_id", ticket.ID, "ticket_number", ticket.TicketNumber)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *TicketRepository) GetByID(ctx context.Context, ticketID uuid.UUID) (*models.SupportTicket, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_tickets WHERE id = $1`, ticketCols())

	ticket, err := r.scanTicket(r.db.QueryRow(ctx, query, ticketID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Ticket not found")
	}
	if err != nil {
		logger.Error("Failed to get ticket by ID", err, "ticket_id", ticketID)
		return nil, errors.Internal(err)
	}
	return ticket, nil
}

// GetByIDAndOrg fetches ticket only if it belongs to the given org (user-scoped access)
func (r *TicketRepository) GetByIDAndOrg(ctx context.Context, ticketID, orgID uuid.UUID) (*models.SupportTicket, error) {
	query := fmt.Sprintf(`SELECT %s FROM support_tickets WHERE id = $1 AND organization_id = $2`, ticketCols())

	ticket, err := r.scanTicket(r.db.QueryRow(ctx, query, ticketID, orgID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Ticket not found")
	}
	if err != nil {
		logger.Error("Failed to get ticket", err, "ticket_id", ticketID, "org_id", orgID)
		return nil, errors.Internal(err)
	}
	return ticket, nil
}

// ListByOrg returns all tickets for an org (user sees all org tickets) with pagination
func (r *TicketRepository) ListByOrg(ctx context.Context, orgID uuid.UUID, status, category string, limit, offset int) ([]*models.SupportTicket, int, error) {
	// Build dynamic WHERE conditions
	conditions := "WHERE organization_id = $1"
	args := []any{orgID}
	argIdx := 2

	if status != "" {
		conditions += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if category != "" {
		conditions += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM support_tickets %s`, conditions)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		logger.Error("Failed to count tickets", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}

	// Fetch page
	listQuery := fmt.Sprintf(`
		SELECT %s FROM support_tickets
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		ticketCols(), conditions, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		logger.Error("Failed to list tickets", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var tickets []*models.SupportTicket
	for rows.Next() {
		t := &models.SupportTicket{}
		err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.UserID, &t.OrganizationID,
			&t.Subject, &t.Category, &t.Priority, &t.Status,
			&t.AssignedTo,
			&t.RelatedProjectID, &t.RelatedEndpointID,
			&t.PlanTier,
			&t.FirstResponseAt, &t.ResolvedAt, &t.ClosedAt,
			&t.SatisfactionRating, &t.SatisfactionComment,
			&t.Metadata,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan ticket row", err)
			return nil, 0, errors.Internal(err)
		}
		tickets = append(tickets, t)
	}

	return tickets, total, nil
}

// ============================================================================
// UPDATE
// ============================================================================

// UpdateStatus updates ticket status and relevant timestamps
func (r *TicketRepository) UpdateStatus(ticketID uuid.UUID, status models.TicketStatus, resolvedAt, closedAt *time.Time) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE support_tickets
		SET status = $2, resolved_at = $3, closed_at = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, ticketID, status, resolvedAt, closedAt, time.Now())
	if err != nil {
		logger.Error("Failed to update ticket status", err, "ticket_id", ticketID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Ticket not found")
	}
	return nil
}

// SetSatisfaction saves rating and comment on a resolved/closed ticket
func (r *TicketRepository) SetSatisfaction(ticketID uuid.UUID, rating int, comment string) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	query := `
		UPDATE support_tickets
		SET satisfaction_rating = $2, satisfaction_comment = $3, updated_at = $4
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, ticketID, rating, commentPtr, time.Now())
	if err != nil {
		logger.Error("Failed to set satisfaction", err, "ticket_id", ticketID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Ticket not found")
	}
	return nil
}