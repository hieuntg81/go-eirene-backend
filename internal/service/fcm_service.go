package service

import (
	"context"
	"errors"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"bamboo-rescue/internal/config"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// FCMService defines the interface for Firebase Cloud Messaging operations
type FCMService interface {
	SendToUser(ctx context.Context, userID uuid.UUID, notification *entity.NotificationPayload) error
	SendToUsers(ctx context.Context, userIDs []uuid.UUID, notification *entity.NotificationPayload) error
	SendToToken(ctx context.Context, token string, notification *entity.NotificationPayload) error
	SendToTokens(ctx context.Context, tokens []string, notification *entity.NotificationPayload) error
}

type fcmService struct {
	client   *messaging.Client
	userRepo repository.UserRepository
	log      *zap.Logger
	enabled  bool
}

// NewFCMService creates a new FCMService
func NewFCMService(cfg *config.Config, userRepo repository.UserRepository, log *zap.Logger) (FCMService, error) {
	svc := &fcmService{
		userRepo: userRepo,
		log:      log,
		enabled:  false,
	}

	if cfg.FCM.CredentialsFile == "" {
		log.Warn("FCM credentials file not configured, push notifications disabled")
		return svc, nil
	}

	opt := option.WithCredentialsFile(cfg.FCM.CredentialsFile)
	app, err := firebase.NewApp(ctx(), nil, opt)
	if err != nil {
		log.Error("Failed to initialize Firebase app", zap.Error(err))
		return svc, nil
	}

	client, err := app.Messaging(ctx())
	if err != nil {
		log.Error("Failed to initialize FCM client", zap.Error(err))
		return svc, nil
	}

	svc.client = client
	svc.enabled = true
	log.Info("FCM service initialized successfully")

	return svc, nil
}

func ctx() context.Context {
	return context.Background()
}

func (s *fcmService) SendToUser(ctx context.Context, userID uuid.UUID, notification *entity.NotificationPayload) error {
	if !s.enabled {
		s.log.Debug("FCM disabled, skipping notification")
		return nil
	}

	tokens, err := s.userRepo.GetPushTokens(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get push tokens", zap.Error(err))
		return err
	}

	if len(tokens) == 0 {
		s.log.Debug("No push tokens found for user", zap.String("user_id", userID.String()))
		return nil
	}

	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = t.Token
	}

	return s.SendToTokens(ctx, tokenStrings, notification)
}

func (s *fcmService) SendToUsers(ctx context.Context, userIDs []uuid.UUID, notification *entity.NotificationPayload) error {
	if !s.enabled {
		s.log.Debug("FCM disabled, skipping notifications")
		return nil
	}

	var allTokens []string
	for _, userID := range userIDs {
		tokens, err := s.userRepo.GetPushTokens(ctx, userID)
		if err != nil {
			s.log.Warn("Failed to get push tokens for user", zap.String("user_id", userID.String()), zap.Error(err))
			continue
		}
		for _, t := range tokens {
			allTokens = append(allTokens, t.Token)
		}
	}

	if len(allTokens) == 0 {
		return nil
	}

	return s.SendToTokens(ctx, allTokens, notification)
}

func (s *fcmService) SendToToken(ctx context.Context, token string, notification *entity.NotificationPayload) error {
	if !s.enabled {
		return nil
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data: notification.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound:            "default",
					ContentAvailable: true,
				},
			},
		},
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		s.log.Error("Failed to send FCM message", zap.Error(err))
		return err
	}

	return nil
}

func (s *fcmService) SendToTokens(ctx context.Context, tokens []string, notification *entity.NotificationPayload) error {
	if !s.enabled {
		return nil
	}

	if len(tokens) == 0 {
		return errors.New("no tokens provided")
	}

	// FCM allows max 500 tokens per batch
	batchSize := 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}

		batch := tokens[i:end]
		message := &messaging.MulticastMessage{
			Tokens: batch,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data: notification.Data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Sound:       "default",
					ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound:            "default",
						ContentAvailable: true,
					},
				},
			},
		}

		response, err := s.client.SendEachForMulticast(ctx, message)
		if err != nil {
			s.log.Error("Failed to send multicast FCM message", zap.Error(err))
			continue
		}

		if response.FailureCount > 0 {
			s.log.Warn("Some FCM messages failed",
				zap.Int("success", response.SuccessCount),
				zap.Int("failure", response.FailureCount),
			)
		}
	}

	return nil
}
