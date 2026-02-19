// ============================================================================
// FILE: services/notification_service.go
// ============================================================================
package services

import (
	"context"

	"github.com/google/uuid"

	"userorg/internal/dto"
	"userorg/internal/repository"
	"shared/errors"
	"shared/logger"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
	log       *logger.Logger
}

func NewNotificationService(notifRepo *repository.NotificationRepository, log *logger.Logger) *NotificationService {
	return &NotificationService{notifRepo: notifRepo, log: log}
}

func (s *NotificationService) ListNotifications(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.ListNotificationsRequest,
) (*dto.ListNotificationsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	notifications, total, err := s.notifRepo.ListByUser(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		resp = append(resp, dto.ToNotificationResponse(n))
	}

	return &dto.ListNotificationsResponse{
		Notifications: resp,
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.notifRepo.MarkRead(ctx, notifID, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.notifRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, errors.Internal(err)
	}
	return count, nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.notifRepo.Delete(ctx, notifID, userID)
}