package request

import "bamboo-rescue/internal/domain/enum"

// UpdateUserRequest represents user profile update request
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=2,max=100"`
	Phone       *string `json:"phone" validate:"omitempty,min=10,max=20"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,url"`
}

// UpdateLocationRequest represents location update request
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// UpdateAvailabilityRequest represents availability update request
type UpdateAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

// UpdatePreferencesRequest represents preferences update request
type UpdatePreferencesRequest struct {
	PushEnabled          *bool            `json:"push_enabled"`
	CaseTypes            []enum.CaseType  `json:"case_types" validate:"omitempty,dive,oneof=animal flood accident"`
	NotificationRadiusKm *int             `json:"notification_radius_km" validate:"omitempty,min=1,max=100"`
	CenterLocation       *LocationRequest `json:"center_location"`
	UseCurrentLocation   *bool            `json:"use_current_location"`
	QuietHoursStart      *string          `json:"quiet_hours_start" validate:"omitempty"`
	QuietHoursEnd        *string          `json:"quiet_hours_end" validate:"omitempty"`
}

// LocationRequest represents a location in request
type LocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// RegisterPushTokenRequest represents push token registration request
type RegisterPushTokenRequest struct {
	Token    string              `json:"token" validate:"required"`
	Platform enum.DevicePlatform `json:"platform" validate:"required,oneof=ios android web"`
	DeviceID *string             `json:"device_id"`
}
