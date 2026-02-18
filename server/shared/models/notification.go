package models

import (
	"slices"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

// NotificationType represents notification type
type NotificationType string

const (
	// Invitation notifications
	NotificationInvitation         NotificationType = "invitation"
	NotificationInvitationAccepted NotificationType = "invitation_accepted"
	NotificationInvitationDeclined NotificationType = "invitation_declined"

	// Team notifications
	NotificationTeamMemberJoined  NotificationType = "team_member_joined"
	NotificationTeamMemberRemoved NotificationType = "team_member_removed"

	// Billing notifications
	NotificationBillingFailed         NotificationType = "billing_failed"
	NotificationSubscriptionExpiring  NotificationType = "subscription_expiring"
	NotificationSubscriptionRenewed   NotificationType = "subscription_renewed"
	NotificationSubscriptionCanceled  NotificationType = "subscription_canceled"
	NotificationPaymentSucceeded      NotificationType = "payment_succeeded"
	NotificationPaymentFailed         NotificationType = "payment_failed"

	// Usage notifications
	NotificationUsageLimitWarning  NotificationType = "usage_limit_warning"
	NotificationUsageLimitExceeded NotificationType = "usage_limit_exceeded"
)

// NotificationPriority represents notification priority level
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// ============================================================================
// NOTIFICATION METADATA
// ============================================================================

// NotificationMetadata represents notification metadata stored as JSONB
type NotificationMetadata struct {
	InviterName string `json:"inviter_name,omitempty"`
	OrgName     string `json:"org_name,omitempty"`
	PlanName    string `json:"plan_name,omitempty"`
	Amount      string `json:"amount,omitempty"`
	DaysLeft    int    `json:"days_left,omitempty"`
	UsagePercent float64 `json:"usage_percent,omitempty"`
}

// Value implements driver.Valuer for NotificationMetadata
func (m NotificationMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements sql.Scanner for NotificationMetadata
func (m *NotificationMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal NotificationMetadata value")
	}
	return json.Unmarshal(bytes, m)
}

// ============================================================================
// NOTIFICATION MODEL
// ============================================================================

// Notification represents an in-app notification
type Notification struct {
	// Primary identification
	ID     uuid.UUID `db:"id" json:"id"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`

	// Notification type
	Type NotificationType `db:"type" json:"type"`

	// Title and message
	Title   string `db:"title" json:"title"`
	Message string `db:"message" json:"message"`

	// Action button (optional)
	ActionURL   *string `db:"action_url" json:"action_url,omitempty"`     // e.g., '/invitations/accept/TOKEN'
	ActionLabel *string `db:"action_label" json:"action_label,omitempty"` // e.g., 'Accept Invitation'

	// Related entities (for routing and context)
	RelatedOrganizationID *uuid.UUID `db:"related_organization_id" json:"related_organization_id,omitempty"`
	RelatedInvitationID   *uuid.UUID `db:"related_invitation_id" json:"related_invitation_id,omitempty"`
	RelatedSubscriptionID *uuid.UUID `db:"related_subscription_id" json:"related_subscription_id,omitempty"`

	// Metadata
	Metadata NotificationMetadata `db:"metadata" json:"metadata"`

	// Status
	IsRead bool       `db:"is_read" json:"is_read"`
	ReadAt *time.Time `db:"read_at" json:"read_at,omitempty"`

	// Priority
	Priority NotificationPriority `db:"priority" json:"priority"`

	// Auto-expire
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// STATUS HELPERS
// ============================================================================

// IsUnread checks if notification is unread
func (n *Notification) IsUnread() bool {
	return !n.IsRead
}

// IsExpired checks if notification has expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// IsActive checks if notification is active (unread and not expired)
func (n *Notification) IsActive() bool {
	return n.IsUnread() && !n.IsExpired()
}

// ============================================================================
// PRIORITY HELPERS
// ============================================================================

// IsUrgent checks if notification is urgent
func (n *Notification) IsUrgent() bool {
	return n.Priority == PriorityUrgent
}

// IsHighPriority checks if notification is high or urgent priority
func (n *Notification) IsHighPriority() bool {
	return n.Priority == PriorityHigh || n.Priority == PriorityUrgent
}

// ============================================================================
// TYPE HELPERS
// ============================================================================

// IsInvitation checks if notification is invitation-related
func (n *Notification) IsInvitation() bool {
	return n.Type == NotificationInvitation ||
		n.Type == NotificationInvitationAccepted ||
		n.Type == NotificationInvitationDeclined
}

// IsBilling checks if notification is billing-related
func (n *Notification) IsBilling() bool {
	return n.Type == NotificationBillingFailed ||
		n.Type == NotificationSubscriptionExpiring ||
		n.Type == NotificationSubscriptionRenewed ||
		n.Type == NotificationSubscriptionCanceled ||
		n.Type == NotificationPaymentSucceeded ||
		n.Type == NotificationPaymentFailed
}

// IsUsage checks if notification is usage-related
func (n *Notification) IsUsage() bool {
	return n.Type == NotificationUsageLimitWarning ||
		n.Type == NotificationUsageLimitExceeded
}

// ============================================================================
// ACTIONS
// ============================================================================

// MarkAsRead marks notification as read
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
}

// MarkAsUnread marks notification as unread
func (n *Notification) MarkAsUnread() {
	n.IsRead = false
	n.ReadAt = nil
}

// SetExpiration sets expiration time
func (n *Notification) SetExpiration(expiresAt time.Time) {
	n.ExpiresAt = &expiresAt
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates the notification struct
func (n *Notification) Validate() error {
	if n.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if n.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if n.Message == "" {
		return fmt.Errorf("notification message is required")
	}
	// Validate type
	validTypes := []NotificationType{
		NotificationInvitation,
		NotificationInvitationAccepted,
		NotificationInvitationDeclined,
		NotificationTeamMemberJoined,
		NotificationTeamMemberRemoved,
		NotificationBillingFailed,
		NotificationSubscriptionExpiring,
		NotificationSubscriptionRenewed,
		NotificationSubscriptionCanceled,
		NotificationPaymentSucceeded,
		NotificationPaymentFailed,
		NotificationUsageLimitWarning,
		NotificationUsageLimitExceeded,
	}
	validType := slices.Contains(validTypes, n.Type)
	if !validType {
		return fmt.Errorf("invalid notification type")
	}
	return nil
}

// ============================================================================
// PUBLIC RESPONSE
// ============================================================================

// PublicNotification represents safe notification data for API responses
type PublicNotification struct {
	ID                    uuid.UUID            `json:"id"`
	Type                  NotificationType     `json:"type"`
	Title                 string               `json:"title"`
	Message               string               `json:"message"`
	ActionURL             *string              `json:"action_url,omitempty"`
	ActionLabel           *string              `json:"action_label,omitempty"`
	RelatedOrganizationID *uuid.UUID           `json:"related_organization_id,omitempty"`
	Metadata              NotificationMetadata `json:"metadata"`
	IsRead                bool                 `json:"is_read"`
	ReadAt                *time.Time           `json:"read_at,omitempty"`
	Priority              NotificationPriority `json:"priority"`
	ExpiresAt             *time.Time           `json:"expires_at,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
}

// ToPublic converts Notification to PublicNotification
func (n *Notification) ToPublic() *PublicNotification {
	return &PublicNotification{
		ID:                    n.ID,
		Type:                  n.Type,
		Title:                 n.Title,
		Message:               n.Message,
		ActionURL:             n.ActionURL,
		ActionLabel:           n.ActionLabel,
		RelatedOrganizationID: n.RelatedOrganizationID,
		Metadata:              n.Metadata,
		IsRead:                n.IsRead,
		ReadAt:                n.ReadAt,
		Priority:              n.Priority,
		ExpiresAt:             n.ExpiresAt,
		CreatedAt:             n.CreatedAt,
	}
}

// ============================================================================
// NOTIFICATION HELPERS
// ============================================================================

// GetNotificationIcon returns emoji/icon for notification type
func (n *Notification) GetNotificationIcon() string {
	switch n.Type {
	case NotificationInvitation:
		return "👥"
	case NotificationInvitationAccepted:
		return "✅"
	case NotificationInvitationDeclined:
		return "❌"
	case NotificationTeamMemberJoined:
		return "🎉"
	case NotificationTeamMemberRemoved:
		return "👋"
	case NotificationBillingFailed:
		return "⚠️"
	case NotificationSubscriptionExpiring:
		return "⏰"
	case NotificationSubscriptionRenewed:
		return "🔄"
	case NotificationSubscriptionCanceled:
		return "🚫"
	case NotificationPaymentSucceeded:
		return "💳"
	case NotificationPaymentFailed:
		return "❌"
	case NotificationUsageLimitWarning:
		return "📊"
	case NotificationUsageLimitExceeded:
		return "🚨"
	default:
		return "🔔"
	}
}