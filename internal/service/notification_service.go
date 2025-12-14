package service

import (
	"context"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"bamboo-rescue/internal/repository"
	"go.uber.org/zap"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Create(ctx context.Context, notification *entity.Notification) error
	CreateForCase(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, notificationType enum.NotificationType, title string, body *string) error
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
	log              *zap.Logger
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(notificationRepo repository.NotificationRepository, log *zap.Logger) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		log:              log,
	}
}

func (s *notificationService) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Notification, int64, error) {
	notifications, total, err := s.notificationRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get notifications", zap.Error(err))
		return nil, 0, err
	}

	return notifications, total, nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	if err := s.notificationRepo.MarkAsRead(ctx, notificationID); err != nil {
		s.log.Error("Failed to mark notification as read", zap.Error(err))
		return err
	}
	return nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		s.log.Error("Failed to mark all notifications as read", zap.Error(err))
		return err
	}
	return nil
}

func (s *notificationService) Create(ctx context.Context, notification *entity.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to create notification", zap.Error(err))
		return err
	}
	return nil
}

func (s *notificationService) CreateForCase(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, notificationType enum.NotificationType, title string, body *string) error {
	notification := &entity.Notification{
		UserID:           userID,
		NotificationType: notificationType,
		Title:            title,
		Body:             body,
		CaseID:           &caseID,
		IsRead:           false,
	}

	return s.Create(ctx, notification)
}
