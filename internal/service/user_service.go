package service

import (
	"context"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/handler/dto/request"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/repository"
	"go.uber.org/zap"
)

// UserService defines the interface for user operations
type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, userID uuid.UUID, req *request.UpdateUserRequest) (*entity.User, error)
	UpdateLocation(ctx context.Context, userID uuid.UUID, req *request.UpdateLocationRequest) error
	UpdateAvailability(ctx context.Context, userID uuid.UUID, isAvailable bool) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, req *request.UpdatePreferencesRequest) (*entity.UserPreferences, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error)
	RegisterPushToken(ctx context.Context, userID uuid.UUID, req *request.RegisterPushTokenRequest) error
	DeletePushToken(ctx context.Context, token string) error
}

type userService struct {
	userRepo repository.UserRepository
	log      *zap.Logger
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository, log *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, middleware.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, userID uuid.UUID, req *request.UpdateUserRequest) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, middleware.ErrUserNotFound
	}

	// Update fields
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Phone != nil {
		// Check if phone is already taken by another user
		existing, err := s.userRepo.GetByPhone(ctx, *req.Phone)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != userID {
			return nil, middleware.ErrPhoneExists
		}
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateLocation(ctx context.Context, userID uuid.UUID, req *request.UpdateLocationRequest) error {
	if err := s.userRepo.UpdateLocation(ctx, userID, req.Latitude, req.Longitude); err != nil {
		s.log.Error("Failed to update location", zap.Error(err))
		return err
	}

	return nil
}

func (s *userService) UpdateAvailability(ctx context.Context, userID uuid.UUID, isAvailable bool) error {
	if err := s.userRepo.UpdateAvailability(ctx, userID, isAvailable); err != nil {
		s.log.Error("Failed to update availability", zap.Error(err))
		return err
	}

	s.log.Info("User availability updated",
		zap.String("user_id", userID.String()),
		zap.Bool("is_available", isAvailable),
	)

	return nil
}

func (s *userService) GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.UserPreferences, error) {
	prefs, err := s.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If preferences don't exist, create default ones
	if prefs == nil {
		prefs = &entity.UserPreferences{
			UserID:               userID,
			PushEnabled:          true,
			CaseTypes:            []string{"animal", "flood", "accident"},
			NotificationRadiusKm: 10,
			UseCurrentLocation:   true,
		}
		if err := s.userRepo.CreatePreferences(ctx, prefs); err != nil {
			s.log.Error("Failed to create default preferences", zap.Error(err))
			return nil, err
		}
	}

	return prefs, nil
}

func (s *userService) UpdatePreferences(ctx context.Context, userID uuid.UUID, req *request.UpdatePreferencesRequest) (*entity.UserPreferences, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.PushEnabled != nil {
		prefs.PushEnabled = *req.PushEnabled
	}
	if req.CaseTypes != nil {
		caseTypes := make([]string, len(req.CaseTypes))
		for i, ct := range req.CaseTypes {
			caseTypes[i] = string(ct)
		}
		prefs.CaseTypes = caseTypes
	}
	if req.NotificationRadiusKm != nil {
		prefs.NotificationRadiusKm = *req.NotificationRadiusKm
	}
	if req.CenterLocation != nil {
		prefs.CenterLatitude = &req.CenterLocation.Latitude
		prefs.CenterLongitude = &req.CenterLocation.Longitude
	}
	if req.UseCurrentLocation != nil {
		prefs.UseCurrentLocation = *req.UseCurrentLocation
	}
	if req.QuietHoursStart != nil {
		prefs.QuietHoursStart = req.QuietHoursStart
	}
	if req.QuietHoursEnd != nil {
		prefs.QuietHoursEnd = req.QuietHoursEnd
	}

	if err := s.userRepo.UpdatePreferences(ctx, prefs); err != nil {
		s.log.Error("Failed to update preferences", zap.Error(err))
		return nil, err
	}

	return prefs, nil
}

func (s *userService) GetStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error) {
	stats, err := s.userRepo.GetStats(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user stats", zap.Error(err))
		return nil, err
	}
	return stats, nil
}

func (s *userService) RegisterPushToken(ctx context.Context, userID uuid.UUID, req *request.RegisterPushTokenRequest) error {
	token := &entity.PushToken{
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
		DeviceID: req.DeviceID,
		IsActive: true,
	}

	if err := s.userRepo.CreatePushToken(ctx, token); err != nil {
		s.log.Error("Failed to register push token", zap.Error(err))
		return err
	}

	s.log.Info("Push token registered",
		zap.String("user_id", userID.String()),
		zap.String("platform", string(req.Platform)),
	)

	return nil
}

func (s *userService) DeletePushToken(ctx context.Context, token string) error {
	if err := s.userRepo.DeletePushToken(ctx, token); err != nil {
		s.log.Error("Failed to delete push token", zap.Error(err))
		return err
	}
	return nil
}
