package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"bamboo-rescue/internal/domain/enum"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Email              *string       `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	Phone              *string       `gorm:"type:varchar(20);uniqueIndex" json:"phone,omitempty"`
	PasswordHash       *string       `gorm:"type:varchar(255)" json:"-"`
	OAuthProvider      *string       `gorm:"column:oauth_provider;type:varchar(50)" json:"oauth_provider,omitempty"`
	OAuthID            *string       `gorm:"column:oauth_id;type:varchar(255)" json:"-"`
	DisplayName        string        `gorm:"type:varchar(100);not null" json:"display_name"`
	AvatarURL          *string       `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	Role               enum.UserRole `gorm:"type:varchar(20);not null;default:'both'" json:"role"`
	IsAvailable        bool          `gorm:"default:false" json:"is_available"`
	Latitude           *float64      `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude          *float64      `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`
	LocationUpdatedAt  *time.Time    `json:"location_updated_at,omitempty"`
	TotalCasesReported int           `gorm:"default:0" json:"total_cases_reported"`
	TotalCasesResolved int           `gorm:"default:0" json:"total_cases_resolved"`
	IsActive           bool          `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Preferences *UserPreferences `gorm:"foreignKey:UserID" json:"preferences,omitempty"`
	PushTokens  []PushToken      `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// GetLocation returns a GeoPoint from the user's coordinates
func (u *User) GetLocation() *GeoPoint {
	if u.Latitude == nil || u.Longitude == nil {
		return nil
	}
	return NewGeoPoint(*u.Latitude, *u.Longitude)
}

// SetLocation sets the user's coordinates from a GeoPoint
func (u *User) SetLocation(loc *GeoPoint) {
	if loc == nil {
		u.Latitude = nil
		u.Longitude = nil
		return
	}
	u.Latitude = &loc.Latitude
	u.Longitude = &loc.Longitude
}

// UserPreferences represents user notification preferences
type UserPreferences struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	PushEnabled          bool           `gorm:"default:true" json:"push_enabled"`
	CaseTypes            pq.StringArray `gorm:"type:text[];default:'{animal,flood,accident}'" json:"case_types"`
	NotificationRadiusKm int            `gorm:"default:10" json:"notification_radius_km"`
	CenterLatitude       *float64       `gorm:"type:decimal(10,8)" json:"center_latitude,omitempty"`
	CenterLongitude      *float64       `gorm:"type:decimal(11,8)" json:"center_longitude,omitempty"`
	UseCurrentLocation   bool           `gorm:"default:true" json:"use_current_location"`
	QuietHoursStart      *string        `gorm:"type:time" json:"quiet_hours_start,omitempty"`
	QuietHoursEnd        *string        `gorm:"type:time" json:"quiet_hours_end,omitempty"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for UserPreferences
func (UserPreferences) TableName() string {
	return "user_preferences"
}

// GetCenterLocation returns a GeoPoint from the preferences' center coordinates
func (p *UserPreferences) GetCenterLocation() *GeoPoint {
	if p.CenterLatitude == nil || p.CenterLongitude == nil {
		return nil
	}
	return NewGeoPoint(*p.CenterLatitude, *p.CenterLongitude)
}

// SetCenterLocation sets the preferences' center coordinates from a GeoPoint
func (p *UserPreferences) SetCenterLocation(loc *GeoPoint) {
	if loc == nil {
		p.CenterLatitude = nil
		p.CenterLongitude = nil
		return
	}
	p.CenterLatitude = &loc.Latitude
	p.CenterLongitude = &loc.Longitude
}

// UserStats represents user statistics
type UserStats struct {
	CasesReported   int `json:"cases_reported"`
	CasesAccepted   int `json:"cases_accepted"`
	CasesCompleted  int `json:"cases_completed"`
	CasesInProgress int `json:"cases_in_progress"`
}

// PushToken represents a push notification token for a user's device
type PushToken struct {
	ID         uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Token      string              `gorm:"type:varchar(500);not null;uniqueIndex" json:"token"`
	Platform   enum.DevicePlatform `gorm:"type:varchar(20);not null" json:"platform"`
	DeviceID   *string             `gorm:"type:varchar(255)" json:"device_id,omitempty"`
	IsActive   bool                `gorm:"default:true" json:"is_active"`
	LastUsedAt time.Time           `gorm:"autoUpdateTime" json:"last_used_at"`
	CreatedAt  time.Time           `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for PushToken
func (PushToken) TableName() string {
	return "push_tokens"
}

// RefreshToken represents a refresh token for authentication
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:varchar(500);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	// Relations
	User *User `gorm:"foreignKey:UserID"`
}

// TableName returns the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired returns true if the refresh token has expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
