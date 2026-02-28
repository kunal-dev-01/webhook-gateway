package services

import (
	"context"
	"fmt"
	"bytes"
	"io"
	"time"

	"github.com/google/uuid"

	"support/internal/dto"
	"support/internal/repository"
	"support/internal/utils"

	"shared/errors"
	"shared/logger"
	"shared/models"
)

// Max attachments per ticket (across all messages)
const maxAttachmentsPerTicket = 20

// ============================================================================
// Support Service
// ============================================================================

type SupportService struct {
	ticketRepo     *repository.TicketRepository
	messageRepo    *repository.MessageRepository
	attachmentRepo *repository.AttachmentRepository
	r2             *utils.R2Client
	log            *logger.Logger
}

func NewSupportService(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.MessageRepository,
	attachmentRepo *repository.AttachmentRepository,
	r2 *utils.R2Client,
	log *logger.Logger,
) *SupportService {
	return &SupportService{
		ticketRepo:     ticketRepo,
		messageRepo:    messageRepo,
		attachmentRepo: attachmentRepo,
		r2:             r2,
		log:            log,
	}
}

// ============================================================================
// CREATE TICKET
// ============================================================================

func (s *SupportService) CreateTicket(ctx context.Context, userID, orgID uuid.UUID, req *dto.CreateTicketRequest) (*dto.CreateTicketResponse, error) {
	// Validate input
	if req.Subject == "" {
		return nil, errors.BadRequest("subject is required")
	}
	if len(req.Subject) > 500 {
		return nil, errors.BadRequest("subject must be under 500 characters")
	}
	if req.Body == "" {
		return nil, errors.BadRequest("message body is required")
	}
	if !req.Category.IsValid() {
		return nil, errors.BadRequest("invalid category — must be one of: general, billing, technical, account, feature_request, bug")
	}

	// Default priority to normal if not set
	if req.Priority == "" {
		req.Priority = models.TicketPriorityNormal
	}
	if !req.Priority.IsValid() {
		return nil, errors.BadRequest("invalid priority — must be one of: low, normal, high, urgent")
	}

	now := time.Now()

	// Build ticket
	ticket := &models.SupportTicket{
		ID:                uuid.New(),
		UserID:            userID,
		OrganizationID:    orgID,
		Subject:           req.Subject,
		Category:          req.Category,
		Priority:          req.Priority,
		Status:            models.TicketStatusOpen,
		RelatedProjectID:  req.RelatedProjectID,
		RelatedEndpointID: req.RelatedEndpointID,
		Metadata:          req.Metadata,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := ticket.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Persist ticket (TicketNumber auto-assigned via DB sequence)
	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, err
	}

	// Create first message (user's opening message)
	msg := &models.SupportMessage{
		ID:          uuid.New(),
		TicketID:    ticket.ID,
		SenderID:    &userID,
		SenderType:  models.SenderTypeUser,
		MessageType: models.MessageTypeReply,
		Body:        req.Body,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		// Ticket created but message failed — log and return partial success
		s.log.Error("Failed to create opening message for ticket", err, "ticket_id", ticket.ID)
		return nil, err
	}

	s.log.Info("Support ticket created",
		"ticket_id", ticket.ID,
		"ticket_code", ticket.GetTicketCode(),
		"user_id", userID,
		"org_id", orgID,
	)

	return &dto.CreateTicketResponse{
		Ticket:  dto.ToTicketResponse(ticket),
		Message: dto.ToMessageResponse(msg, nil),
	}, nil
}

// ============================================================================
// LIST TICKETS
// ============================================================================

func (s *SupportService) ListTickets(ctx context.Context, orgID uuid.UUID, query *dto.ListTicketsQuery) (*dto.ListTicketsResponse, error) {
	// Sanitize pagination
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}
	offset := (query.Page - 1) * query.Limit

	// Validate status/category filters if provided
	if query.Status != "" {
		s := models.TicketStatus(query.Status)
		if !s.IsValid() {
			return nil, errors.BadRequest("invalid status filter")
		}
	}
	if query.Category != "" {
		c := models.TicketCategory(query.Category)
		if !c.IsValid() {
			return nil, errors.BadRequest("invalid category filter")
		}
	}

	tickets, total, err := s.ticketRepo.ListByOrg(ctx, orgID, query.Status, query.Category, query.Limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		responses = append(responses, dto.ToTicketResponse(t))
	}

	totalPages := (total + query.Limit - 1) / query.Limit

	return &dto.ListTicketsResponse{
		Tickets:    responses,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}

// ============================================================================
// GET TICKET
// ============================================================================

func (s *SupportService) GetTicket(ctx context.Context, ticketID, orgID uuid.UUID) (*dto.GetTicketResponse, error) {
	ticket, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}

	return &dto.GetTicketResponse{
		Ticket: dto.ToTicketResponse(ticket),
	}, nil
}

// ============================================================================
// ADD MESSAGE
// ============================================================================

func (s *SupportService) AddMessage(ctx context.Context, ticketID, orgID, userID uuid.UUID, req *dto.AddMessageRequest) (*dto.AddMessageResponse, error) {
	if req.Body == "" {
		return nil, errors.BadRequest("message body is required")
	}
	if len(req.Body) > 50000 {
		return nil, errors.BadRequest("message body is too long (max 50,000 characters)")
	}

	// Verify ticket belongs to org
	ticket, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}

	// Cannot reply to a closed ticket
	if ticket.IsClosed() {
		return nil, errors.BadRequest("cannot add messages to a closed ticket — please open a new ticket")
	}

	now := time.Now()
	msg := &models.SupportMessage{
		ID:          uuid.New(),
		TicketID:    ticketID,
		SenderID:    &userID,
		SenderType:  models.SenderTypeUser,
		MessageType: models.MessageTypeReply,
		Body:        req.Body,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		return nil, err
	}

	// If ticket was in "waiting" state (waiting for user), move it back to in_progress
	if ticket.IsWaiting() {
		_ = s.ticketRepo.UpdateStatus(ticketID, models.TicketStatusInProgress, nil, nil)
	}

	// If ticket was resolved, auto-reopen on new user message
	if ticket.IsResolved() {
		_ = s.ticketRepo.UpdateStatus(ticketID, models.TicketStatusOpen, nil, nil)
	}

	s.log.Info("User added message to ticket",
		"ticket_id", ticketID,
		"message_id", msg.ID,
		"user_id", userID,
	)

	return &dto.AddMessageResponse{
		Message: dto.ToMessageResponse(msg, nil),
	}, nil
}

// ============================================================================
// LIST MESSAGES
// ============================================================================

func (s *SupportService) ListMessages(ctx context.Context, ticketID, orgID uuid.UUID) (*dto.ListMessagesResponse, error) {
	// Verify ticket belongs to org
	if _, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID); err != nil {
		return nil, err
	}

	messages, err := s.messageRepo.ListByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Bulk-load attachments for all messages
	messageIDs := make([]uuid.UUID, 0, len(messages))
	for _, m := range messages {
		messageIDs = append(messageIDs, m.ID)
	}

	attachments, err := s.attachmentRepo.ListByMessageIDs(ctx, messageIDs)
	if err != nil {
		s.log.Error("Failed to load attachments for messages", err, "ticket_id", ticketID)
		// Non-fatal: return messages without attachments
		attachments = nil
	}

	// Group attachments by message ID
	attachmentsByMessage := make(map[uuid.UUID][]*models.SupportAttachment)
	for _, a := range attachments {
		attachmentsByMessage[a.MessageID] = append(attachmentsByMessage[a.MessageID], a)
	}

	responses := make([]*dto.MessageResponse, 0, len(messages))
	for _, m := range messages {
		responses = append(responses, dto.ToMessageResponse(m, attachmentsByMessage[m.ID]))
	}

	// Mark all agent/system messages as read now that user viewed them
	go func() {
		_ = s.messageRepo.MarkAllReadByUser(ticketID)
	}()

	return &dto.ListMessagesResponse{
		Messages: responses,
		Total:    len(responses),
	}, nil
}

// ============================================================================
// UPLOAD ATTACHMENT
// ============================================================================

func (s *SupportService) UploadAttachment(ctx context.Context, ticketID, orgID, messageID, userID uuid.UUID, file io.Reader, fileName, contentType string, fileSize int64) (*dto.UploadAttachmentResponse, error) {
	// Verify ticket ownership
	ticket, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}

	if ticket.IsClosed() {
		return nil, errors.BadRequest("cannot attach files to a closed ticket")
	}

	// Verify message belongs to this ticket
	msg, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg.TicketID != ticketID {
		return nil, errors.NotFound("Message not found")
	}

	// Size check (10MB)
	if fileSize > 10485760 {
		return nil, errors.BadRequest("file size cannot exceed 10MB")
	}

	// Enforce total attachment limit per ticket
	count, err := s.attachmentRepo.CountByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if count >= maxAttachmentsPerTicket {
		return nil, errors.BadRequest(fmt.Sprintf("ticket cannot have more than %d attachments", maxAttachmentsPerTicket))
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to read file: %w", err))
	}

	// Generate R2 storage key
	objectKey := fmt.Sprintf("support/%s/%s/%d_%s",
		ticketID.String(),
		messageID.String(),
		time.Now().UnixMilli(),
		sanitizeFileName(fileName),
	)

	// Upload to R2 — Upload takes io.Reader, returns (url, error)
	if _, err := s.r2.Upload(ctx, objectKey, bytes.NewReader(fileData), contentType); err != nil {
		s.log.Error("Failed to upload attachment to R2", err, "ticket_id", ticketID)
		return nil, errors.Internal(fmt.Errorf("failed to upload file: %w", err))
	}

	now := time.Now()
	attachment := &models.SupportAttachment{
		ID:            uuid.New(),
		TicketID:      ticketID,
		MessageID:     messageID,
		UploadedBy:    userID,
		FileName:      fileName,
		FileSizeBytes: fileSize,
		MimeType:      contentType,
		StorageKey:    objectKey,
		StorageBucket: s.r2.BucketName(),
		ScanStatus:    models.ScanStatusPending, // Virus scan happens async
		CreatedAt:     now,
	}

	if err := attachment.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.attachmentRepo.Create(attachment); err != nil {
		// R2 upload succeeded but DB failed — log for cleanup
		s.log.Error("Attachment DB insert failed after R2 upload",
			err,
			"storage_key", objectKey,
			"ticket_id", ticketID,
		)
		return nil, err
	}

	s.log.Info("Attachment uploaded",
		"attachment_id", attachment.ID,
		"ticket_id", ticketID,
		"file_name", fileName,
	)

	return &dto.UploadAttachmentResponse{
		Attachment: dto.ToAttachmentResponse(attachment),
	}, nil
}

// ============================================================================
// GET ATTACHMENT DOWNLOAD URL
// ============================================================================

func (s *SupportService) GetAttachmentDownloadURL(ctx context.Context, ticketID, orgID, attachmentID uuid.UUID) (*dto.GetAttachmentDownloadURLResponse, error) {
	// Verify ticket ownership
	if _, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID); err != nil {
		return nil, err
	}

	attachment, err := s.attachmentRepo.GetByIDAndTicket(ctx, attachmentID, ticketID)
	if err != nil {
		return nil, err
	}

	if !attachment.IsDownloadable() {
		if attachment.IsInfected() {
			return nil, errors.BadRequest("file has been flagged and cannot be downloaded")
		}
		return nil, errors.BadRequest("file is pending security scan — please try again shortly")
	}

	// Generate presigned URL (valid for 15 minutes)
	ttl := 15 * time.Minute
	presignedURL, err := s.r2.GeneratePresignedURL(ctx, attachment.StorageKey, attachment.FileName, ttl)
	if err != nil {
		s.log.Error("Failed to generate presigned URL", err, "attachment_id", attachmentID)
		return nil, errors.Internal(fmt.Errorf("failed to generate download link: %w", err))
	}

	return &dto.GetAttachmentDownloadURLResponse{
		DownloadURL: presignedURL,
		ExpiresAt:   time.Now().Add(ttl),
	}, nil
}

// ============================================================================
// SUBMIT SATISFACTION
// ============================================================================

func (s *SupportService) SubmitSatisfaction(ctx context.Context, ticketID, orgID uuid.UUID, req *dto.SubmitSatisfactionRequest) (*dto.SubmitSatisfactionResponse, error) {
	ticket, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}

	if err := ticket.SetSatisfaction(req.Rating, req.Comment); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.ticketRepo.SetSatisfaction(ticketID, req.Rating, req.Comment); err != nil {
		return nil, err
	}

	s.log.Info("Satisfaction submitted",
		"ticket_id", ticketID,
		"rating", req.Rating,
	)

	return &dto.SubmitSatisfactionResponse{
		Rating:  req.Rating,
		Comment: req.Comment,
	}, nil
}

// ============================================================================
// REOPEN TICKET
// ============================================================================

func (s *SupportService) ReopenTicket(ctx context.Context, ticketID, orgID uuid.UUID) (*dto.ReopenTicketResponse, error) {
	ticket, err := s.ticketRepo.GetByIDAndOrg(ctx, ticketID, orgID)
	if err != nil {
		return nil, err
	}

	if err := ticket.Reopen(); err != nil {
		// ticket.Reopen() returns error for closed tickets
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, models.TicketStatusOpen, nil, nil); err != nil {
		return nil, err
	}

	s.log.Info("Ticket reopened", "ticket_id", ticketID, "org_id", orgID)

	return &dto.ReopenTicketResponse{
		Ticket: dto.ToTicketResponse(ticket),
	}, nil
}

// ============================================================================
// HELPERS
// ============================================================================

// sanitizeFileName removes path separators from user-provided file names
func sanitizeFileName(name string) string {
	safe := ""
	for _, r := range name {
		if r == '/' || r == '\\' || r == '\x00' {
			continue
		}
		safe += string(r)
	}
	if len(safe) > 100 {
		safe = safe[:100]
	}
	return safe
}