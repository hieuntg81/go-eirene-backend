package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"bamboo-rescue/internal/handler/dto/response"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/service"
	pkgresponse "bamboo-rescue/pkg/response"
)

// NotificationHandler handles notification requests
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles get user notifications
// @Summary Get notifications
// @Description Get notifications for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} pkgresponse.Response{data=response.NotificationListResponse}
// @Failure 401 {object} pkgresponse.Response
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		pkgresponse.Error(c, middleware.ErrUnauthorized)
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if _, err := c.GetQuery("limit"); err {
			limit = 20
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := c.GetQuery("offset"); err {
			offset = 0
		}
	}

	notifications, unreadCount, err := h.notificationService.GetByUser(c.Request.Context(), *userID, limit, offset)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, response.NotificationListResponse{
		Notifications: response.ToNotificationListResponse(notifications),
		UnreadCount:   unreadCount,
	})
}

// GetUnreadCount handles get unread notification count
// @Summary Get unread count
// @Description Get unread notification count for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} pkgresponse.Response{data=map[string]int64}
// @Failure 401 {object} pkgresponse.Response
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		pkgresponse.Error(c, middleware.ErrUnauthorized)
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), *userID)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsRead handles mark notification as read
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} pkgresponse.Response
// @Failure 401 {object} pkgresponse.Response
// @Failure 404 {object} pkgresponse.Response
// @Router /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		pkgresponse.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid notification ID", 400))
		return
	}

	if err := h.notificationService.MarkAsRead(c.Request.Context(), *userID, notificationID); err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead handles mark all notifications as read
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} pkgresponse.Response
// @Failure 401 {object} pkgresponse.Response
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		pkgresponse.Error(c, middleware.ErrUnauthorized)
		return
	}

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), *userID); err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
