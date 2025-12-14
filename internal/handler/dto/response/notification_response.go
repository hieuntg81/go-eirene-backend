package response

import (
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
)

// NotificationResponse represents a notification in response
type NotificationResponse struct {
	ID               uuid.UUID             `json:"id"`
	NotificationType enum.NotificationType `json:"notificationType"`
	Title            string                `json:"title"`
	Body             *string               `json:"body,omitempty"`
	CaseID           *uuid.UUID            `json:"caseId,omitempty"`
	IsRead           bool                  `json:"isRead"`
	CreatedAt        time.Time             `json:"createdAt"`
}

// ToNotificationResponse converts entity to response
func ToNotificationResponse(n *entity.Notification) *NotificationResponse {
	if n == nil {
		return nil
	}

	return &NotificationResponse{
		ID:               n.ID,
		NotificationType: n.NotificationType,
		Title:            n.Title,
		Body:             n.Body,
		CaseID:           n.CaseID,
		IsRead:           n.IsRead,
		CreatedAt:        n.CreatedAt,
	}
}

// ToNotificationListResponse converts a slice of notifications to response
func ToNotificationListResponse(notifications []entity.Notification) []NotificationResponse {
	result := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = *ToNotificationResponse(&n)
	}
	return result
}

// NotificationListResponse represents notification list with unread count
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unreadCount"`
}

// GeocodeResponse represents a geocoding result
type GeocodeResponse struct {
	Address   string           `json:"address"`
	Location  GeoPointResponse `json:"location"`
	PlaceID   string           `json:"placeId,omitempty"`
	PlaceType string           `json:"placeType,omitempty"`
}

// MediaUploadResponse represents media upload result
type MediaUploadResponse struct {
	ID           uuid.UUID      `json:"id"`
	URL          string         `json:"url"`
	ThumbnailURL *string        `json:"thumbnailUrl,omitempty"`
	MediaType    enum.MediaType `json:"mediaType"`
	FileSize     int64          `json:"fileSize"`
}

// SuccessMessageResponse represents a simple success response
type SuccessMessageResponse struct {
	Message string `json:"message"`
}

// IDResponse represents a response with just an ID
type IDResponse struct {
	ID uuid.UUID `json:"id"`
}
