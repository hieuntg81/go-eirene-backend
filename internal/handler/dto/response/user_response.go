package response

import (
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
)

// UserResponse represents user data in response
type UserResponse struct {
	ID                 uuid.UUID         `json:"id"`
	Email              *string           `json:"email,omitempty"`
	Phone              *string           `json:"phone,omitempty"`
	DisplayName        string            `json:"displayName"`
	AvatarURL          *string           `json:"avatarUrl,omitempty"`
	Role               enum.UserRole     `json:"role"`
	IsAvailable        bool              `json:"isAvailable"`
	LastLocation       *GeoPointResponse `json:"lastLocation,omitempty"`
	LocationUpdatedAt  *time.Time        `json:"locationUpdatedAt,omitempty"`
	TotalCasesReported int               `json:"totalCasesReported"`
	TotalCasesResolved int               `json:"totalCasesResolved"`
	IsActive           bool              `json:"isActive"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

// GeoPointResponse represents a geographic point in response
type GeoPointResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ToUserResponse converts entity to response
func ToUserResponse(u *entity.User) *UserResponse {
	if u == nil {
		return nil
	}

	resp := &UserResponse{
		ID:                 u.ID,
		Email:              u.Email,
		Phone:              u.Phone,
		DisplayName:        u.DisplayName,
		AvatarURL:          u.AvatarURL,
		Role:               u.Role,
		IsAvailable:        u.IsAvailable,
		LocationUpdatedAt:  u.LocationUpdatedAt,
		TotalCasesReported: u.TotalCasesReported,
		TotalCasesResolved: u.TotalCasesResolved,
		IsActive:           u.IsActive,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}

	if u.Latitude != nil && u.Longitude != nil {
		resp.LastLocation = &GeoPointResponse{
			Latitude:  *u.Latitude,
			Longitude: *u.Longitude,
		}
	}

	return resp
}

// UserPreferencesResponse represents user preferences in response
type UserPreferencesResponse struct {
	ID                   uuid.UUID         `json:"id"`
	UserID               uuid.UUID         `json:"userId"`
	PushEnabled          bool              `json:"pushEnabled"`
	CaseTypes            []enum.CaseType   `json:"caseTypes"`
	NotificationRadiusKm int               `json:"notificationRadiusKm"`
	CenterLocation       *GeoPointResponse `json:"centerLocation,omitempty"`
	UseCurrentLocation   bool              `json:"useCurrentLocation"`
	QuietHoursStart      *string           `json:"quietHoursStart,omitempty"`
	QuietHoursEnd        *string           `json:"quietHoursEnd,omitempty"`
	CreatedAt            time.Time         `json:"createdAt"`
	UpdatedAt            time.Time         `json:"updatedAt"`
}

// ToUserPreferencesResponse converts entity to response
func ToUserPreferencesResponse(p *entity.UserPreferences) *UserPreferencesResponse {
	if p == nil {
		return nil
	}

	caseTypes := make([]enum.CaseType, len(p.CaseTypes))
	for i, ct := range p.CaseTypes {
		caseTypes[i] = enum.CaseType(ct)
	}

	resp := &UserPreferencesResponse{
		ID:                   p.ID,
		UserID:               p.UserID,
		PushEnabled:          p.PushEnabled,
		CaseTypes:            caseTypes,
		NotificationRadiusKm: p.NotificationRadiusKm,
		UseCurrentLocation:   p.UseCurrentLocation,
		QuietHoursStart:      p.QuietHoursStart,
		QuietHoursEnd:        p.QuietHoursEnd,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}

	if p.CenterLatitude != nil && p.CenterLongitude != nil {
		resp.CenterLocation = &GeoPointResponse{
			Latitude:  *p.CenterLatitude,
			Longitude: *p.CenterLongitude,
		}
	}

	return resp
}

// UserStatsResponse represents user statistics in response
type UserStatsResponse struct {
	CasesReported   int `json:"casesReported"`
	CasesAccepted   int `json:"casesAccepted"`
	CasesCompleted  int `json:"casesCompleted"`
	CasesInProgress int `json:"casesInProgress"`
}

// ToUserStatsResponse converts entity to response
func ToUserStatsResponse(s *entity.UserStats) *UserStatsResponse {
	if s == nil {
		return nil
	}

	return &UserStatsResponse{
		CasesReported:   s.CasesReported,
		CasesAccepted:   s.CasesAccepted,
		CasesCompleted:  s.CasesCompleted,
		CasesInProgress: s.CasesInProgress,
	}
}
