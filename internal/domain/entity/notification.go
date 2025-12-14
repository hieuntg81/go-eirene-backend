package entity

import (
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/enum"
)

// Notification represents a notification sent to a user
type Notification struct {
	ID               uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	NotificationType enum.NotificationType `gorm:"type:varchar(30);not null" json:"notification_type"`
	Title            string                `gorm:"type:varchar(200);not null" json:"title"`
	Body             *string               `gorm:"type:text" json:"body,omitempty"`
	CaseID           *uuid.UUID            `gorm:"type:uuid" json:"case_id,omitempty"`
	IsRead           bool                  `gorm:"default:false" json:"is_read"`
	IsPushed         bool                  `gorm:"default:false" json:"is_pushed"`
	PushedAt         *time.Time            `json:"pushed_at,omitempty"`
	CreatedAt        time.Time             `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"-"`
	Case *Case `gorm:"foreignKey:CaseID" json:"case,omitempty"`
}

// TableName returns the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}

// MarkAsPushed marks the notification as pushed
func (n *Notification) MarkAsPushed() {
	now := time.Now()
	n.IsPushed = true
	n.PushedAt = &now
}

// NotificationPayload represents the payload for push notifications
type NotificationPayload struct {
	Type       enum.NotificationType `json:"type"`
	Title      string                `json:"title"`
	Body       string                `json:"body"`
	CaseID     *uuid.UUID            `json:"case_id,omitempty"`
	CaseType   *enum.CaseType        `json:"case_type,omitempty"`
	Urgency    *enum.UrgencyLevel    `json:"urgency,omitempty"`
	DistanceKm *float64              `json:"distance_km,omitempty"`
	Data       map[string]string     `json:"data,omitempty"`
}

// NewCaseNotificationPayload creates a payload for new case notifications
func NewCaseNotificationPayload(c *Case, distanceKm float64) *NotificationPayload {
	body := c.Title
	if distanceKm > 0 {
		body = c.Title + " - " + formatDistance(distanceKm)
	}

	return &NotificationPayload{
		Type:       enum.NotificationTypeNewCaseNearby,
		Title:      "Case mới gần bạn",
		Body:       body,
		CaseID:     &c.ID,
		CaseType:   &c.CaseType,
		Urgency:    &c.Urgency,
		DistanceKm: &distanceKm,
	}
}

// CaseAcceptedNotificationPayload creates a payload for case accepted notifications
func CaseAcceptedNotificationPayload(c *Case, volunteerName string) *NotificationPayload {
	return &NotificationPayload{
		Type:   enum.NotificationTypeCaseAccepted,
		Title:  "Case đã có người nhận",
		Body:   volunteerName + " đã nhận case của bạn",
		CaseID: &c.ID,
	}
}

// CaseResolvedNotificationPayload creates a payload for case resolved notifications
func CaseResolvedNotificationPayload(c *Case) *NotificationPayload {
	return &NotificationPayload{
		Type:   enum.NotificationTypeCaseResolved,
		Title:  "Case đã hoàn thành",
		Body:   "Cảm ơn bạn đã tham gia cứu hộ!",
		CaseID: &c.ID,
	}
}

// VolunteerJoinedNotificationPayload creates a payload for volunteer joined notifications
func VolunteerJoinedNotificationPayload(c *Case, volunteerName string) *NotificationPayload {
	return &NotificationPayload{
		Type:   enum.NotificationTypeVolunteerJoined,
		Title:  "Tình nguyện viên mới",
		Body:   volunteerName + " đã tham gia case",
		CaseID: &c.ID,
	}
}

func formatDistance(km float64) string {
	if km < 1 {
		return "< 1km"
	}
	return formatFloat(km, 1) + "km"
}

func formatFloat(f float64, precision int) string {
	format := "%." + string(rune('0'+precision)) + "f"
	return sprintf(format, f)
}

func sprintf(format string, f float64) string {
	// Simple implementation for formatting
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 10)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	return itoa(intPart) + "." + itoa(fracPart)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	negative := i < 0
	if negative {
		i = -i
	}
	var result []byte
	for i > 0 {
		result = append([]byte{byte('0' + i%10)}, result...)
		i /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
