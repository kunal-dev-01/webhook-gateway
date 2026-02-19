package dto

import (
	"time"

	"github.com/google/uuid"

	"shared/models"
)

// ============================================================================
// REQUEST MODELS
// ============================================================================

type ListNotificationsRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

type NotificationResponse struct {
	ID                    uuid.UUID                    `json:"id"`
	Type                  models.NotificationType      `json:"type"`
	Title                 string                       `json:"title"`
	Message               string                       `json:"message"`
	ActionURL             *string                      `json:"action_url,omitempty"`
	ActionLabel           *string                      `json:"action_label,omitempty"`
	RelatedOrganizationID *uuid.UUID                   `json:"related_organization_id,omitempty"`
	Metadata              models.NotificationMetadata  `json:"metadata"`
	IsRead                bool                         `json:"is_read"`
	ReadAt                *time.Time                   `json:"read_at,omitempty"`
	Priority              models.NotificationPriority  `json:"priority"`
	ExpiresAt             *time.Time                   `json:"expires_at,omitempty"`
	CreatedAt             time.Time                    `json:"created_at"`
}

type ListNotificationsResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
	Total         int                     `json:"total"`
	Limit         int                     `json:"limit"`
	Offset        int                     `json:"offset"`
}

// ============================================================================
// HELPERS
// ============================================================================

func ToNotificationResponse(n *models.Notification) *NotificationResponse {
	return &NotificationResponse{
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