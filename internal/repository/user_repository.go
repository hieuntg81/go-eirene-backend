package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetByOAuth(ctx context.Context, provider, oauthID string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error
	UpdateAvailability(ctx context.Context, userID uuid.UUID, isAvailable bool) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Preferences
	GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.UserPreferences, error)
	CreatePreferences(ctx context.Context, prefs *entity.UserPreferences) error
	UpdatePreferences(ctx context.Context, prefs *entity.UserPreferences) error

	// Push Tokens
	GetPushTokens(ctx context.Context, userID uuid.UUID) ([]entity.PushToken, error)
	CreatePushToken(ctx context.Context, token *entity.PushToken) error
	DeletePushToken(ctx context.Context, token string) error
	DeletePushTokensByUser(ctx context.Context, userID uuid.UUID) error

	// Stats
	GetStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error)
	IncrementCasesReported(ctx context.Context, userID uuid.UUID) error
	IncrementCasesResolved(ctx context.Context, userID uuid.UUID) error

	// Volunteers
	FindAvailableVolunteers(ctx context.Context, lat, lng float64, radiusKm int, caseType string, limit int) ([]VolunteerWithDistance, error)
}

// VolunteerWithDistance represents a volunteer with their distance from a location
type VolunteerWithDistance struct {
	User       *entity.User
	DistanceKm float64
	PushTokens []entity.PushToken
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db interface{}) UserRepository {
	return &userRepository{db: db.(*gorm.DB)}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Preferences").
		First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		First(&user, "phone = ?", phone).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByOAuth(ctx context.Context, provider, oauthID string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		First(&user, "oauth_provider = ? AND oauth_id = ?", provider, oauthID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"latitude":            lat,
			"longitude":           lng,
			"location_updated_at": now,
		}).Error
}

func (r *userRepository) UpdateAvailability(ctx context.Context, userID uuid.UUID, isAvailable bool) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("is_available", isAvailable).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *userRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.UserPreferences, error) {
	var prefs entity.UserPreferences
	err := r.db.WithContext(ctx).
		First(&prefs, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &prefs, nil
}

func (r *userRepository) CreatePreferences(ctx context.Context, prefs *entity.UserPreferences) error {
	if prefs.ID == uuid.Nil {
		prefs.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(prefs).Error
}

func (r *userRepository) UpdatePreferences(ctx context.Context, prefs *entity.UserPreferences) error {
	return r.db.WithContext(ctx).Save(prefs).Error
}

func (r *userRepository) GetPushTokens(ctx context.Context, userID uuid.UUID) ([]entity.PushToken, error) {
	var tokens []entity.PushToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&tokens).Error
	return tokens, err
}

func (r *userRepository) CreatePushToken(ctx context.Context, token *entity.PushToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	// Upsert: if token exists, update it
	return r.db.WithContext(ctx).
		Where("token = ?", token.Token).
		Assign(map[string]interface{}{
			"user_id":      token.UserID,
			"platform":     token.Platform,
			"device_id":    token.DeviceID,
			"is_active":    true,
			"last_used_at": time.Now(),
		}).
		FirstOrCreate(token).Error
}

func (r *userRepository) DeletePushToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&entity.PushToken{}).Error
}

func (r *userRepository) DeletePushTokensByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.PushToken{}).Error
}

func (r *userRepository) GetStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Select("total_cases_reported", "total_cases_resolved").
		First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	// Count cases by status for this volunteer
	var accepted, inProgress int64
	r.db.WithContext(ctx).
		Model(&entity.CaseVolunteer{}).
		Where("volunteer_id = ?", userID).
		Count(&accepted)

	r.db.WithContext(ctx).
		Model(&entity.CaseVolunteer{}).
		Where("volunteer_id = ? AND status IN ?", userID, []string{"accepted", "en_route", "on_site", "handling"}).
		Count(&inProgress)

	return &entity.UserStats{
		CasesReported:   user.TotalCasesReported,
		CasesAccepted:   int(accepted),
		CasesCompleted:  user.TotalCasesResolved,
		CasesInProgress: int(inProgress),
	}, nil
}

func (r *userRepository) IncrementCasesReported(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		UpdateColumn("total_cases_reported", gorm.Expr("total_cases_reported + 1")).Error
}

func (r *userRepository) IncrementCasesResolved(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", userID).
		UpdateColumn("total_cases_resolved", gorm.Expr("total_cases_resolved + 1")).Error
}

func (r *userRepository) FindAvailableVolunteers(ctx context.Context, lat, lng float64, radiusKm int, caseType string, limit int) ([]VolunteerWithDistance, error) {
	// Create bounding box for initial filtering
	bbox := entity.NewBoundingBox(lat, lng, float64(radiusKm))

	// Get available users with their preferences
	type UserWithPrefs struct {
		entity.User
		PushEnabled          *bool    `gorm:"column:push_enabled"`
		CaseTypes            []string `gorm:"column:case_types;type:text[]"`
		NotificationRadiusKm *int     `gorm:"column:notification_radius_km"`
		CenterLatitude       *float64 `gorm:"column:center_latitude"`
		CenterLongitude      *float64 `gorm:"column:center_longitude"`
		UseCurrentLocation   *bool    `gorm:"column:use_current_location"`
		QuietHoursStart      *string  `gorm:"column:quiet_hours_start"`
		QuietHoursEnd        *string  `gorm:"column:quiet_hours_end"`
	}

	var usersWithPrefs []UserWithPrefs

	query := r.db.WithContext(ctx).
		Table("users u").
		Select(`u.*,
			up.push_enabled,
			up.case_types,
			up.notification_radius_km,
			up.center_latitude,
			up.center_longitude,
			up.use_current_location,
			up.quiet_hours_start,
			up.quiet_hours_end`).
		Joins("LEFT JOIN user_preferences up ON up.user_id = u.id").
		Where("u.is_available = true").
		Where("u.is_active = true").
		Where("(up.push_enabled IS NULL OR up.push_enabled = true)").
		Where("u.latitude IS NOT NULL AND u.longitude IS NOT NULL").
		Where("u.latitude BETWEEN ? AND ?", bbox.MinLat, bbox.MaxLat).
		Where("u.longitude BETWEEN ? AND ?", bbox.MinLng, bbox.MaxLng)

	if err := query.Find(&usersWithPrefs).Error; err != nil {
		return nil, err
	}

	// Calculate distances and filter
	centerPoint := entity.NewGeoPoint(lat, lng)
	var volunteers []VolunteerWithDistance

	for _, uwp := range usersWithPrefs {
		user := uwp.User

		// Determine the effective location for distance calculation
		var effectiveLat, effectiveLng float64
		if uwp.UseCurrentLocation != nil && !*uwp.UseCurrentLocation && uwp.CenterLatitude != nil && uwp.CenterLongitude != nil {
			effectiveLat = *uwp.CenterLatitude
			effectiveLng = *uwp.CenterLongitude
		} else if user.Latitude != nil && user.Longitude != nil {
			effectiveLat = *user.Latitude
			effectiveLng = *user.Longitude
		} else {
			continue // Skip users without location
		}

		// Calculate distance
		userPoint := entity.NewGeoPoint(effectiveLat, effectiveLng)
		distance := centerPoint.DistanceKm(userPoint)

		// Check against user's notification radius
		userRadiusKm := 10 // default
		if uwp.NotificationRadiusKm != nil {
			userRadiusKm = *uwp.NotificationRadiusKm
		}

		if distance > float64(userRadiusKm) {
			continue
		}

		// Check case type preference
		if caseType != "" && uwp.CaseTypes != nil {
			found := false
			for _, ct := range uwp.CaseTypes {
				if ct == caseType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// TODO: Check quiet hours (would require current time comparison)

		// Fetch push tokens
		tokens, _ := r.GetPushTokens(ctx, user.ID)

		volunteers = append(volunteers, VolunteerWithDistance{
			User:       &user,
			DistanceKm: distance,
			PushTokens: tokens,
		})
	}

	// Sort by distance
	sort.Slice(volunteers, func(i, j int) bool {
		return volunteers[i].DistanceKm < volunteers[j].DistanceKm
	})

	// Apply limit
	if limit > 0 && len(volunteers) > limit {
		volunteers = volunteers[:limit]
	}

	return volunteers, nil
}
