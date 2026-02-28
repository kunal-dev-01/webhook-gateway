package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// ============================================================================
// CREATE TICKET
// ============================================================================

type CreateTicketRequest struct {
	Subject           string                `json:"subject"`
	Category          models.TicketCategory `json:"category"`
	Priority          models.TicketPriority `json:"priority"` // default: normal
	Body              string                `json:"body"`     // First message body (required)
	RelatedProjectID  *uuid.UUID            `json:"related_project_id,omitempty"`
	RelatedEndpointID *uuid.UUID            `json:"related_endpoint_id,omitempty"`
	Metadata          models.TicketMetadata `json:"metadata,omitempty"`
}

type CreateTicketResponse struct {
	Ticket  *TicketResponse  `json:"ticket"`
	Message *MessageResponse `json:"message"` // First message auto-created
}

// ============================================================================
// LIST TICKETS
// ============================================================================

// ListTicketsQuery represents URL query params for filtering
type ListTicketsQuery struct {
	Status   string `query:"status"`   // optional filter
	Category string `query:"category"` // optional filter
	Page     int    `query:"page"`     // default: 1
	Limit    int    `query:"limit"`    // default: 20, max: 100
}

type ListTicketsResponse struct {
	Tickets    []*TicketResponse `json:"tickets"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// ============================================================================
// GET TICKET
// ============================================================================

type GetTicketResponse struct {
	Ticket *TicketResponse `json:"ticket"`
}

// ============================================================================
// REOPEN TICKET
// ============================================================================

type ReopenTicketResponse struct {
	Ticket *TicketResponse `json:"ticket"`
}

// ============================================================================
// ADD MESSAGE
// ============================================================================

type AddMessageRequest struct {
	Body string `json:"body"` // Markdown supported
}

type AddMessageResponse struct {
	Message *MessageResponse `json:"message"`
}

// ============================================================================
// LIST MESSAGES
// ============================================================================

type ListMessagesResponse struct {
	Messages []*MessageResponse `json:"messages"`
	Total    int                `json:"total"`
}

// ============================================================================
// ATTACHMENTS
// ============================================================================

type UploadAttachmentResponse struct {
	Attachment *AttachmentResponse `json:"attachment"`
}

type GetAttachmentDownloadURLResponse struct {
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ============================================================================
// SATISFACTION
// ============================================================================

type SubmitSatisfactionRequest struct {
	Rating  int    `json:"rating"`  // 1-5
	Comment string `json:"comment"` // optional
}

type SubmitSatisfactionResponse struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment,omitempty"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

type TicketResponse struct {
	ID                  uuid.UUID             `json:"id"`
	TicketCode          string                `json:"ticket_code"` // TKT-1001
	TicketNumber        int                   `json:"ticket_number"`
	Subject             string                `json:"subject"`
	Category            models.TicketCategory `json:"category"`
	Priority            models.TicketPriority `json:"priority"`
	Status              models.TicketStatus   `json:"status"`
	RelatedProjectID    *uuid.UUID            `json:"related_project_id,omitempty"`
	RelatedEndpointID   *uuid.UUID            `json:"related_endpoint_id,omitempty"`
	FirstResponseAt     *time.Time            `json:"first_response_at,omitempty"`
	ResolvedAt          *time.Time            `json:"resolved_at,omitempty"`
	SatisfactionRating  *int                  `json:"satisfaction_rating,omitempty"`
	SatisfactionComment *string               `json:"satisfaction_comment,omitempty"`
	IsBreachingSLA      bool                  `json:"is_breaching_sla"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

type MessageResponse struct {
	ID          uuid.UUID               `json:"id"`
	TicketID    uuid.UUID               `json:"ticket_id"`
	SenderID    *uuid.UUID              `json:"sender_id,omitempty"`
	SenderType  models.MessageSenderType `json:"sender_type"`
	Body        string                  `json:"body"`
	IsEdited    bool                    `json:"is_edited"`
	EditedAt    *time.Time              `json:"edited_at,omitempty"`
	Attachments []*AttachmentResponse   `json:"attachments,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
}

type AttachmentResponse struct {
	ID             uuid.UUID          `json:"id"`
	TicketID       uuid.UUID          `json:"ticket_id"`
	MessageID      uuid.UUID          `json:"message_id"`
	FileName       string             `json:"file_name"`
	FileSizeBytes  int64              `json:"file_size_bytes"`
	FileSizeMB     float64            `json:"file_size_mb"`
	MimeType       string             `json:"mime_type"`
	IsImage        bool               `json:"is_image"`
	ScanStatus     models.ScanStatus  `json:"scan_status"`
	IsDownloadable bool               `json:"is_downloadable"`
	CreatedAt      time.Time          `json:"created_at"`
}

// ============================================================================
// MAPPERS
// ============================================================================

func ToTicketResponse(t *models.SupportTicket) *TicketResponse {
	return &TicketResponse{
		ID:                  t.ID,
		TicketCode:          t.GetTicketCode(),
		TicketNumber:        t.TicketNumber,
		Subject:             t.Subject,
		Category:            t.Category,
		Priority:            t.Priority,
		Status:              t.Status,
		RelatedProjectID:    t.RelatedProjectID,
		RelatedEndpointID:   t.RelatedEndpointID,
		FirstResponseAt:     t.FirstResponseAt,
		ResolvedAt:          t.ResolvedAt,
		SatisfactionRating:  t.SatisfactionRating,
		SatisfactionComment: t.SatisfactionComment,
		IsBreachingSLA:      t.IsBreachingSLA(),
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

func ToMessageResponse(m *models.SupportMessage, attachments []*models.SupportAttachment) *MessageResponse {
	resp := &MessageResponse{
		ID:         m.ID,
		TicketID:   m.TicketID,
		SenderID:   m.SenderID,
		SenderType: m.SenderType,
		Body:       m.Body,
		IsEdited:   m.IsEdited,
		EditedAt:   m.EditedAt,
		CreatedAt:  m.CreatedAt,
	}
	if len(attachments) > 0 {
		resp.Attachments = make([]*AttachmentResponse, 0, len(attachments))
		for _, a := range attachments {
			resp.Attachments = append(resp.Attachments, ToAttachmentResponse(a))
		}
	}
	return resp
}

func ToAttachmentResponse(a *models.SupportAttachment) *AttachmentResponse {
	return &AttachmentResponse{
		ID:             a.ID,
		TicketID:       a.TicketID,
		MessageID:      a.MessageID,
		FileName:       a.FileName,
		FileSizeBytes:  a.FileSizeBytes,
		FileSizeMB:     a.GetFileSizeMB(),
		MimeType:       a.MimeType,
		IsImage:        a.IsImage(),
		ScanStatus:     a.ScanStatus,
		IsDownloadable: a.IsDownloadable(),
		CreatedAt:      a.CreatedAt,
	}
}