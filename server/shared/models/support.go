package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// TicketCategory represents the type of support issue
type TicketCategory string

const (
	TicketCategoryGeneral        TicketCategory = "general"
	TicketCategoryBilling        TicketCategory = "billing"
	TicketCategoryTechnical      TicketCategory = "technical"
	TicketCategoryAccount        TicketCategory = "account"
	TicketCategoryFeatureRequest TicketCategory = "feature_request"
	TicketCategoryBug            TicketCategory = "bug"
)

func (c TicketCategory) IsValid() bool {
	switch c {
	case TicketCategoryGeneral, TicketCategoryBilling, TicketCategoryTechnical,
		TicketCategoryAccount, TicketCategoryFeatureRequest, TicketCategoryBug:
		return true
	}
	return false
}

// TicketPriority represents urgency level
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

func (p TicketPriority) IsValid() bool {
	switch p {
	case TicketPriorityLow, TicketPriorityNormal, TicketPriorityHigh, TicketPriorityUrgent:
		return true
	}
	return false
}

// TicketStatus represents the current state of a ticket
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusWaiting    TicketStatus = "waiting" // Waiting for user response
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

func (s TicketStatus) IsValid() bool {
	switch s {
	case TicketStatusOpen, TicketStatusInProgress, TicketStatusWaiting,
		TicketStatusResolved, TicketStatusClosed:
		return true
	}
	return false
}

func (s TicketStatus) IsTerminal() bool {
	return s == TicketStatusResolved || s == TicketStatusClosed
}

// MessageSenderType represents who sent the message
type MessageSenderType string

const (
	SenderTypeUser   MessageSenderType = "user"
	SenderTypeAgent  MessageSenderType = "agent"
	SenderTypeSystem MessageSenderType = "system"
)

func (s MessageSenderType) IsValid() bool {
	switch s {
	case SenderTypeUser, SenderTypeAgent, SenderTypeSystem:
		return true
	}
	return false
}

// MessageType represents the visibility of a message
type MessageType string

const (
	MessageTypeReply        MessageType = "reply"         // Visible to everyone
	MessageTypeInternalNote MessageType = "internal_note" // Team only
	MessageTypeSystem       MessageType = "system"        // Auto-generated
)

func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypeReply, MessageTypeInternalNote, MessageTypeSystem:
		return true
	}
	return false
}

// ScanStatus represents file virus scan result
type ScanStatus string

const (
	ScanStatusPending  ScanStatus = "pending"
	ScanStatusClean    ScanStatus = "clean"
	ScanStatusInfected ScanStatus = "infected"
	ScanStatusError    ScanStatus = "error"
)

// ============================================================================
// TICKET METADATA
// ============================================================================

// TicketMetadata stores additional context about the ticket
type TicketMetadata struct {
	Browser    string `json:"browser,omitempty"`
	OS         string `json:"os,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

func (m TicketMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *TicketMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TicketMetadata value")
	}
	return json.Unmarshal(bytes, m)
}

// ============================================================================
// SUPPORT TICKET MODEL
// ============================================================================

// SupportTicket represents a customer support request
// ticket_number is auto-increment displayed as TKT-1001, TKT-1002...
type SupportTicket struct {
	ID             uuid.UUID `db:"id" json:"id"`

	// Human-readable ID (display as TKT-{TicketNumber})
	TicketNumber   int       `db:"ticket_number" json:"ticket_number"`

	// Who raised it
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`

	// Ticket details
	Subject  string         `db:"subject" json:"subject"`
	Category TicketCategory `db:"category" json:"category"`
	Priority TicketPriority `db:"priority" json:"priority"`
	Status   TicketStatus   `db:"status" json:"status"`

	// Assignment (nil = unassigned)
	AssignedTo *uuid.UUID `db:"assigned_to" json:"assigned_to,omitempty"`

	// Related context
	RelatedProjectID  *uuid.UUID `db:"related_project_id" json:"related_project_id,omitempty"`
	RelatedEndpointID *uuid.UUID `db:"related_endpoint_id" json:"related_endpoint_id,omitempty"`

	// Plan snapshot for SLA
	PlanTier *string `db:"plan_tier" json:"plan_tier,omitempty"`

	// SLA tracking
	FirstResponseAt *time.Time `db:"first_response_at" json:"first_response_at,omitempty"`
	ResolvedAt      *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ClosedAt        *time.Time `db:"closed_at" json:"closed_at,omitempty"`

	// User satisfaction after resolution (1-5)
	SatisfactionRating  *int    `db:"satisfaction_rating" json:"satisfaction_rating,omitempty"`
	SatisfactionComment *string `db:"satisfaction_comment" json:"satisfaction_comment,omitempty"`

	// Metadata
	Metadata TicketMetadata `db:"metadata" json:"metadata"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// TICKET HELPERS
// ============================================================================

// GetTicketCode returns human-readable ticket identifier
func (t *SupportTicket) GetTicketCode() string {
	return fmt.Sprintf("TKT-%d", t.TicketNumber)
}

func (t *SupportTicket) IsOpen() bool       { return t.Status == TicketStatusOpen }
func (t *SupportTicket) IsInProgress() bool { return t.Status == TicketStatusInProgress }
func (t *SupportTicket) IsWaiting() bool    { return t.Status == TicketStatusWaiting }
func (t *SupportTicket) IsResolved() bool   { return t.Status == TicketStatusResolved }
func (t *SupportTicket) IsClosed() bool     { return t.Status == TicketStatusClosed }
func (t *SupportTicket) IsTerminal() bool   { return t.Status.IsTerminal() }
func (t *SupportTicket) IsAssigned() bool   { return t.AssignedTo != nil }
func (t *SupportTicket) IsUrgent() bool     { return t.Priority == TicketPriorityUrgent }

// IsRatedPositively checks if user gave a good rating (4 or 5 stars)
func (t *SupportTicket) IsRatedPositively() bool {
	return t.SatisfactionRating != nil && *t.SatisfactionRating >= 4
}

// GetFirstResponseSLA returns expected SLA duration based on plan
// Scale: 4h, Growth: 12h, Starter: 24h, Developer: 48h
func (t *SupportTicket) GetFirstResponseSLA() time.Duration {
	if t.PlanTier == nil {
		return 48 * time.Hour
	}
	switch PlanTier(*t.PlanTier) {
	case PlanScale:
		return 4 * time.Hour
	case PlanGrowth:
		return 12 * time.Hour
	case PlanStarter:
		return 24 * time.Hour
	default:
		return 48 * time.Hour
	}
}

// IsBreachingSLA checks if ticket has exceeded first response SLA
func (t *SupportTicket) IsBreachingSLA() bool {
	if t.FirstResponseAt != nil || t.IsTerminal() {
		return false // Already responded or closed
	}
	sla := t.GetFirstResponseSLA()
	return time.Since(t.CreatedAt) > sla
}

// ============================================================================
// TICKET ACTIONS
// ============================================================================

func (t *SupportTicket) Assign(agentID uuid.UUID) {
	t.AssignedTo = &agentID
	if t.Status == TicketStatusOpen {
		t.Status = TicketStatusInProgress
	}
}

func (t *SupportTicket) Unassign() {
	t.AssignedTo = nil
}

func (t *SupportTicket) MarkFirstResponse() {
	if t.FirstResponseAt == nil {
		now := time.Now()
		t.FirstResponseAt = &now
	}
}

func (t *SupportTicket) Resolve() error {
	if t.IsTerminal() {
		return fmt.Errorf("ticket is already %s", t.Status)
	}
	now := time.Now()
	t.Status = TicketStatusResolved
	t.ResolvedAt = &now
	return nil
}

func (t *SupportTicket) Close() {
	now := time.Now()
	t.Status = TicketStatusClosed
	t.ClosedAt = &now
}

func (t *SupportTicket) Reopen() error {
	if t.Status == TicketStatusClosed {
		return fmt.Errorf("closed tickets cannot be reopened, create a new ticket")
	}
	t.Status = TicketStatusOpen
	t.ResolvedAt = nil
	return nil
}

func (t *SupportTicket) WaitingForUser() {
	t.Status = TicketStatusWaiting
}

func (t *SupportTicket) SetSatisfaction(rating int, comment string) error {
	if !t.IsResolved() && !t.IsClosed() {
		return fmt.Errorf("satisfaction can only be set on resolved/closed tickets")
	}
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	t.SatisfactionRating = &rating
	if comment != "" {
		t.SatisfactionComment = &comment
	}
	return nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (t *SupportTicket) Validate() error {
	if t.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if t.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if t.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if len(t.Subject) > 500 {
		return fmt.Errorf("subject must be under 500 characters")
	}
	if !t.Category.IsValid() {
		return fmt.Errorf("invalid category")
	}
	if !t.Priority.IsValid() {
		return fmt.Errorf("invalid priority")
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("invalid status")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicSupportTicket struct {
	ID                  uuid.UUID      `json:"id"`
	TicketNumber        int            `json:"ticket_number"`
	TicketCode          string         `json:"ticket_code"` // TKT-1001
	UserID              uuid.UUID      `json:"user_id"`
	OrganizationID      uuid.UUID      `json:"organization_id"`
	Subject             string         `json:"subject"`
	Category            TicketCategory `json:"category"`
	Priority            TicketPriority `json:"priority"`
	Status              TicketStatus   `json:"status"`
	AssignedTo          *uuid.UUID     `json:"assigned_to,omitempty"`
	RelatedProjectID    *uuid.UUID     `json:"related_project_id,omitempty"`
	FirstResponseAt     *time.Time     `json:"first_response_at,omitempty"`
	ResolvedAt          *time.Time     `json:"resolved_at,omitempty"`
	SatisfactionRating  *int           `json:"satisfaction_rating,omitempty"`
	IsBreachingSLA      bool           `json:"is_breaching_sla"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (t *SupportTicket) ToPublic() *PublicSupportTicket {
	return &PublicSupportTicket{
		ID:                 t.ID,
		TicketNumber:       t.TicketNumber,
		TicketCode:         t.GetTicketCode(),
		UserID:             t.UserID,
		OrganizationID:     t.OrganizationID,
		Subject:            t.Subject,
		Category:           t.Category,
		Priority:           t.Priority,
		Status:             t.Status,
		AssignedTo:         t.AssignedTo,
		RelatedProjectID:   t.RelatedProjectID,
		FirstResponseAt:    t.FirstResponseAt,
		ResolvedAt:         t.ResolvedAt,
		SatisfactionRating: t.SatisfactionRating,
		IsBreachingSLA:     t.IsBreachingSLA(),
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
	}
}

// ============================================================================
// SUPPORT MESSAGE MODEL
// ============================================================================

// SupportMessage represents a single message in a ticket conversation
// Three visibility levels: reply (all), internal_note (team only), system (auto)
type SupportMessage struct {
	ID       uuid.UUID `db:"id" json:"id"`
	TicketID uuid.UUID `db:"ticket_id" json:"ticket_id"`

	// Sender (nil for system messages)
	SenderID   *uuid.UUID        `db:"sender_id" json:"sender_id,omitempty"`
	SenderType MessageSenderType `db:"sender_type" json:"sender_type"`

	// Message content
	MessageType MessageType `db:"message_type" json:"message_type"`
	Body        string      `db:"body" json:"body"` // Markdown supported

	// Read tracking
	IsReadByUser   bool       `db:"is_read_by_user" json:"is_read_by_user"`
	ReadByUserAt   *time.Time `db:"read_by_user_at" json:"read_by_user_at,omitempty"`
	IsReadByAgent  bool       `db:"is_read_by_agent" json:"is_read_by_agent"`
	ReadByAgentAt  *time.Time `db:"read_by_agent_at" json:"read_by_agent_at,omitempty"`

	// Edit tracking
	IsEdited     bool       `db:"is_edited" json:"is_edited"`
	EditedAt     *time.Time `db:"edited_at" json:"edited_at,omitempty"`
	OriginalBody *string    `db:"original_body" json:"-"` // Not exposed in API

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// MESSAGE HELPERS
// ============================================================================

func (m *SupportMessage) IsFromUser() bool    { return m.SenderType == SenderTypeUser }
func (m *SupportMessage) IsFromAgent() bool   { return m.SenderType == SenderTypeAgent }
func (m *SupportMessage) IsSystemMsg() bool   { return m.SenderType == SenderTypeSystem }
func (m *SupportMessage) IsInternalNote() bool { return m.MessageType == MessageTypeInternalNote }
func (m *SupportMessage) IsPublicReply() bool  { return m.MessageType == MessageTypeReply }

// IsVisibleToUser checks if this message should be shown to the customer
func (m *SupportMessage) IsVisibleToUser() bool {
	return m.MessageType != MessageTypeInternalNote
}

// ============================================================================
// MESSAGE ACTIONS
// ============================================================================

func (m *SupportMessage) MarkReadByUser() {
	if !m.IsReadByUser {
		now := time.Now()
		m.IsReadByUser = true
		m.ReadByUserAt = &now
	}
}

func (m *SupportMessage) MarkReadByAgent() {
	if !m.IsReadByAgent {
		now := time.Now()
		m.IsReadByAgent = true
		m.ReadByAgentAt = &now
	}
}

func (m *SupportMessage) Edit(newBody string) error {
	if m.IsSystemMsg() {
		return fmt.Errorf("system messages cannot be edited")
	}
	now := time.Now()
	// Preserve original body on first edit
	if !m.IsEdited {
		m.OriginalBody = &m.Body
	}
	m.Body = newBody
	m.IsEdited = true
	m.EditedAt = &now
	return nil
}

// ============================================================================
// VALIDATION
// ============================================================================

func (m *SupportMessage) Validate() error {
	if m.TicketID == uuid.Nil {
		return fmt.Errorf("ticket ID is required")
	}
	if m.Body == "" {
		return fmt.Errorf("message body is required")
	}
	if !m.SenderType.IsValid() {
		return fmt.Errorf("invalid sender type")
	}
	if !m.MessageType.IsValid() {
		return fmt.Errorf("invalid message type")
	}
	// System messages must have nil sender
	if m.SenderType == SenderTypeSystem && m.SenderID != nil {
		return fmt.Errorf("system messages cannot have a sender")
	}
	// Non-system messages must have a sender
	if m.SenderType != SenderTypeSystem && m.SenderID == nil {
		return fmt.Errorf("sender ID is required for non-system messages")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicSupportMessage struct {
	ID            uuid.UUID         `json:"id"`
	TicketID      uuid.UUID         `json:"ticket_id"`
	SenderID      *uuid.UUID        `json:"sender_id,omitempty"`
	SenderType    MessageSenderType `json:"sender_type"`
	MessageType   MessageType       `json:"message_type"`
	Body          string            `json:"body"`
	IsReadByUser  bool              `json:"is_read_by_user"`
	IsReadByAgent bool              `json:"is_read_by_agent"`
	IsEdited      bool              `json:"is_edited"`
	EditedAt      *time.Time        `json:"edited_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

func (m *SupportMessage) ToPublic() *PublicSupportMessage {
	return &PublicSupportMessage{
		ID:            m.ID,
		TicketID:      m.TicketID,
		SenderID:      m.SenderID,
		SenderType:    m.SenderType,
		MessageType:   m.MessageType,
		Body:          m.Body,
		IsReadByUser:  m.IsReadByUser,
		IsReadByAgent: m.IsReadByAgent,
		IsEdited:      m.IsEdited,
		EditedAt:      m.EditedAt,
		CreatedAt:     m.CreatedAt,
	}
}

// ============================================================================
// SUPPORT ATTACHMENT MODEL
// ============================================================================

// SupportAttachment represents a file attached to a support message
// Actual file stored in S3/R2, this is just metadata
type SupportAttachment struct {
	ID         uuid.UUID `db:"id" json:"id"`
	TicketID   uuid.UUID `db:"ticket_id" json:"ticket_id"`
	MessageID  uuid.UUID `db:"message_id" json:"message_id"`
	UploadedBy uuid.UUID `db:"uploaded_by" json:"uploaded_by"`

	// File info
	FileName      string `db:"file_name" json:"file_name"`
	FileSizeBytes int64  `db:"file_size_bytes" json:"file_size_bytes"`
	MimeType      string `db:"mime_type" json:"mime_type"`

	// S3/R2 storage info
	StorageKey    string `db:"storage_key" json:"-"`    // Internal path, never expose
	StorageBucket string `db:"storage_bucket" json:"-"` // Internal, never expose

	// Virus scan
	ScanStatus ScanStatus `db:"scan_status" json:"scan_status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ============================================================================
// ATTACHMENT HELPERS
// ============================================================================

func (a *SupportAttachment) IsClean() bool {
	return a.ScanStatus == ScanStatusClean
}

func (a *SupportAttachment) IsInfected() bool {
	return a.ScanStatus == ScanStatusInfected
}

func (a *SupportAttachment) IsScanPending() bool {
	return a.ScanStatus == ScanStatusPending
}

// IsDownloadable checks if file can be served to user
func (a *SupportAttachment) IsDownloadable() bool {
	return a.ScanStatus == ScanStatusClean
}

// GetFileSizeMB returns file size in MB
func (a *SupportAttachment) GetFileSizeMB() float64 {
	return float64(a.FileSizeBytes) / (1024 * 1024)
}

// IsImage checks if attachment is an image
func (a *SupportAttachment) IsImage() bool {
	switch a.MimeType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	}
	return false
}

// ============================================================================
// VALIDATION
// ============================================================================

func (a *SupportAttachment) Validate() error {
	if a.TicketID == uuid.Nil {
		return fmt.Errorf("ticket ID is required")
	}
	if a.MessageID == uuid.Nil {
		return fmt.Errorf("message ID is required")
	}
	if a.UploadedBy == uuid.Nil {
		return fmt.Errorf("uploader ID is required")
	}
	if a.FileName == "" {
		return fmt.Errorf("file name is required")
	}
	if a.FileSizeBytes <= 0 {
		return fmt.Errorf("file size must be positive")
	}
	if a.FileSizeBytes > 10485760 { // 10MB
		return fmt.Errorf("file size cannot exceed 10MB")
	}
	if a.StorageKey == "" {
		return fmt.Errorf("storage key is required")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

type PublicSupportAttachment struct {
	ID            uuid.UUID  `json:"id"`
	TicketID      uuid.UUID  `json:"ticket_id"`
	MessageID     uuid.UUID  `json:"message_id"`
	UploadedBy    uuid.UUID  `json:"uploaded_by"`
	FileName      string     `json:"file_name"`
	FileSizeBytes int64      `json:"file_size_bytes"`
	FileSizeMB    float64    `json:"file_size_mb"`
	MimeType      string     `json:"mime_type"`
	IsImage       bool       `json:"is_image"`
	ScanStatus    ScanStatus `json:"scan_status"`
	IsDownloadable bool      `json:"is_downloadable"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (a *SupportAttachment) ToPublic() *PublicSupportAttachment {
	return &PublicSupportAttachment{
		ID:             a.ID,
		TicketID:       a.TicketID,
		MessageID:      a.MessageID,
		UploadedBy:     a.UploadedBy,
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