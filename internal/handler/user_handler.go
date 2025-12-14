package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"bamboo-rescue/internal/handler/dto/request"
	dto "bamboo-rescue/internal/handler/dto/response"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/service"
	"bamboo-rescue/pkg/response"
)

// UserHandler handles user requests
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile handles get current user profile
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 401 {object} response.Response
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserResponse(user))
}

// UpdateProfile handles update user profile
// @Summary Update user profile
// @Description Update the profile of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdateUserRequest true "Update profile request"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.Update(c.Request.Context(), *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserResponse(user))
}

// UpdateLocation handles update user location
// @Summary Update user location
// @Description Update the current location of the user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdateLocationRequest true "Update location request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me/location [put]
func (h *UserHandler) UpdateLocation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userService.UpdateLocation(c.Request.Context(), *userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Location updated successfully"})
}

// UpdateAvailability handles update user availability
// @Summary Update user availability
// @Description Update the availability status of the user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdateAvailabilityRequest true "Update availability request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me/availability [put]
func (h *UserHandler) UpdateAvailability(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userService.UpdateAvailability(c.Request.Context(), *userID, req.IsAvailable); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Availability updated successfully"})
}

// GetPreferences handles get user preferences
// @Summary Get user preferences
// @Description Get the notification preferences of the user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserPreferencesResponse}
// @Failure 401 {object} response.Response
// @Router /users/me/preferences [get]
func (h *UserHandler) GetPreferences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	prefs, err := h.userService.GetPreferences(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserPreferencesResponse(prefs))
}

// UpdatePreferences handles update user preferences
// @Summary Update user preferences
// @Description Update the notification preferences of the user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.UpdatePreferencesRequest true "Update preferences request"
// @Success 200 {object} response.Response{data=dto.UserPreferencesResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me/preferences [put]
func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	prefs, err := h.userService.UpdatePreferences(c.Request.Context(), *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserPreferencesResponse(prefs))
}

// GetStats handles get user statistics
// @Summary Get user statistics
// @Description Get the rescue statistics of the user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserStatsResponse}
// @Failure 401 {object} response.Response
// @Router /users/me/stats [get]
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	stats, err := h.userService.GetStats(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToUserStatsResponse(stats))
}

// RegisterPushToken handles push token registration
// @Summary Register push token
// @Description Register a device push token for notifications
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.RegisterPushTokenRequest true "Push token request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me/push-token [post]
func (h *UserHandler) RegisterPushToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.RegisterPushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userService.RegisterPushToken(c.Request.Context(), *userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Push token registered successfully"})
}

// DeletePushToken handles push token deletion
// @Summary Delete push token
// @Description Delete a device push token
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param token path string true "Push token"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/me/push-token/{token} [delete]
func (h *UserHandler) DeletePushToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Token is required", 400))
		return
	}

	if err := h.userService.DeletePushToken(c.Request.Context(), token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Push token deleted successfully"})
}
