package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"bamboo-rescue/internal/handler/dto/request"
	dto "bamboo-rescue/internal/handler/dto/response"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/service"
	"bamboo-rescue/pkg/response"
)

// CaseHandler handles case requests
type CaseHandler struct {
	caseService service.CaseService
}

// NewCaseHandler creates a new CaseHandler
func NewCaseHandler(caseService service.CaseService) *CaseHandler {
	return &CaseHandler{
		caseService: caseService,
	}
}

// Create handles case creation
// @Summary Create a new case
// @Description Create a new rescue case
// @Tags Cases
// @Accept json
// @Produce json
// @Param request body request.CreateCaseRequest true "Create case request"
// @Success 201 {object} response.Response{data=dto.CaseResponse}
// @Failure 400 {object} response.Response
// @Router /cases [post]
func (h *CaseHandler) Create(c *gin.Context) {
	var req request.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := middleware.GetUserID(c)

	caseEntity, err := h.caseService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToCaseResponse(caseEntity))
}

// GetByID handles get case by ID
// @Summary Get case by ID
// @Description Get detailed information about a case
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} response.Response{data=dto.CaseResponse}
// @Failure 404 {object} response.Response
// @Router /cases/{id} [get]
func (h *CaseHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	caseEntity, err := h.caseService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCaseResponse(caseEntity))
}

// GetCases handles get cases with search and pagination
// @Summary Get cases
// @Description Get cases with optional search and filters
// @Tags Cases
// @Produce json
// @Param q query string false "Search query"
// @Param type query string false "Case type filter"
// @Param status query string false "Status filter"
// @Param urgency query string false "Urgency filter"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {object} response.Response{data=[]dto.CaseResponse}
// @Failure 400 {object} response.Response
// @Router /cases [get]
func (h *CaseHandler) GetCases(c *gin.Context) {
	var req request.GetCasesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	cases, total, err := h.caseService.GetCases(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	meta := response.NewMeta(req.GetDefaultPage(), req.GetDefaultLimit(), total)
	response.SuccessWithMeta(c, dto.ToCaseListResponse(cases), meta)
}

// GetNearby handles get nearby cases
// @Summary Get nearby cases
// @Description Get cases near a location
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Param radius_km query number false "Radius in km (default 10)"
// @Param case_type query string false "Case type filter"
// @Param status query string false "Status filter"
// @Success 200 {object} response.Response{data=[]dto.CaseNearbyResponse}
// @Failure 400 {object} response.Response
// @Router /cases/nearby [get]
func (h *CaseHandler) GetNearby(c *gin.Context) {
	var req request.GetNearbyCasesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	cases, err := h.caseService.GetNearby(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCaseNearbyListResponse(cases))
}

// GetMyCases handles get user's reported cases
// @Summary Get my reported cases
// @Description Get cases reported by the current user
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.Response{data=[]dto.CaseResponse}
// @Failure 401 {object} response.Response
// @Router /cases/my-cases [get]
func (h *CaseHandler) GetMyCases(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	cases, total, err := h.caseService.GetUserReportedCases(c.Request.Context(), *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	meta := response.NewMeta(req.GetDefaultPage(), req.GetDefaultLimit(), total)
	response.SuccessWithMeta(c, dto.ToCaseListResponse(cases), meta)
}

// GetMyVolunteerCases handles get cases user volunteered for
// @Summary Get my volunteer cases
// @Description Get cases where the current user is a volunteer
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.Response{data=[]dto.CaseResponse}
// @Failure 401 {object} response.Response
// @Router /cases/my-volunteer-cases [get]
func (h *CaseHandler) GetMyVolunteerCases(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	cases, total, err := h.caseService.GetUserAcceptedCases(c.Request.Context(), *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	meta := response.NewMeta(req.GetDefaultPage(), req.GetDefaultLimit(), total)
	response.SuccessWithMeta(c, dto.ToCaseListResponse(cases), meta)
}

// Update handles case update
// @Summary Update a case
// @Description Update an existing case
// @Tags Cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param request body request.UpdateCaseRequest true "Update case request"
// @Success 200 {object} response.Response{data=dto.CaseResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id} [put]
func (h *CaseHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	caseEntity, err := h.caseService.Update(c.Request.Context(), id, *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCaseResponse(caseEntity))
}

// Delete handles case deletion
// @Summary Delete a case
// @Description Delete an existing case
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id} [delete]
func (h *CaseHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	if err := h.caseService.Delete(c.Request.Context(), id, *userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Case deleted successfully"})
}

// Accept handles volunteer accepting a case
// @Summary Accept a case
// @Description Accept a case as a volunteer
// @Tags Cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param request body request.AcceptCaseRequest true "Accept case request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/accept [post]
func (h *CaseHandler) Accept(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.AcceptCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.caseService.Accept(c.Request.Context(), caseID, *userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Case accepted successfully"})
}

// Withdraw handles volunteer withdrawing from a case
// @Summary Withdraw from a case
// @Description Withdraw from a case as a volunteer
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/withdraw [post]
func (h *CaseHandler) Withdraw(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	if err := h.caseService.Withdraw(c.Request.Context(), caseID, *userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Withdrawn from case successfully"})
}

// UpdateVolunteerStatus handles updating volunteer status
// @Summary Update volunteer status
// @Description Update volunteer status (arrived, completed)
// @Tags Cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param request body request.UpdateVolunteerStatusRequest true "Update status request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/volunteer-status [put]
func (h *CaseHandler) UpdateVolunteerStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.UpdateVolunteerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.caseService.UpdateVolunteerStatus(c.Request.Context(), caseID, *userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// CreateUpdate handles creating a case update
// @Summary Create case update
// @Description Add an update to a case
// @Tags Cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param request body request.CreateCaseUpdateRequest true "Create update request"
// @Success 201 {object} response.Response{data=dto.CaseUpdateResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/updates [post]
func (h *CaseHandler) CreateUpdate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.CreateCaseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	update, err := h.caseService.CreateUpdate(c.Request.Context(), caseID, *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToCaseUpdateResponse(update))
}

// GetVolunteers handles getting volunteers for a case
// @Summary Get case volunteers
// @Description Get all active volunteers for a case
// @Tags Cases
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} response.Response{data=[]dto.VolunteerResponse}
// @Failure 404 {object} response.Response
// @Router /cases/{id}/volunteers [get]
func (h *CaseHandler) GetVolunteers(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	volunteers, err := h.caseService.GetVolunteers(c.Request.Context(), caseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToVolunteerListResponse(volunteers))
}

// GetUpdates handles getting case updates
// @Summary Get case updates
// @Description Get all updates for a case
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Case ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.Response{data=[]dto.CaseUpdateResponse}
// @Failure 404 {object} response.Response
// @Router /cases/{id}/updates [get]
func (h *CaseHandler) GetUpdates(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates, _, err := h.caseService.GetUpdates(c.Request.Context(), caseID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]dto.CaseUpdateResponse, len(updates))
	for i, u := range updates {
		result[i] = *dto.ToCaseUpdateResponse(&u)
	}

	response.Success(c, http.StatusOK, result)
}

// GetComments handles getting case comments
// @Summary Get case comments
// @Description Get all comments for a case
// @Tags Cases
// @Produce json
// @Param id path string true "Case ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response{data=[]dto.CommentResponse}
// @Failure 404 {object} response.Response
// @Router /cases/{id}/comments [get]
func (h *CaseHandler) GetComments(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Get case to determine reporter ID for role assignment
	caseEntity, err := h.caseService.GetByID(c.Request.Context(), caseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	comments, _, err := h.caseService.GetComments(c.Request.Context(), caseID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToCommentListResponse(comments, caseEntity.ReporterID))
}

// CreateComment handles creating a case comment
// @Summary Create case comment
// @Description Add a comment to a case
// @Tags Cases
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param request body request.CreateCommentRequest true "Create comment request"
// @Success 201 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/comments [post]
func (h *CaseHandler) CreateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	var req request.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	comment, err := h.caseService.CreateComment(c.Request.Context(), caseID, *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Get case to determine reporter ID for role assignment
	caseEntity, _ := h.caseService.GetByID(c.Request.Context(), caseID)
	var reporterID *uuid.UUID
	if caseEntity != nil {
		reporterID = caseEntity.ReporterID
	}

	response.Success(c, http.StatusCreated, dto.ToCommentResponse(comment, reporterID))
}

// DeleteComment handles deleting a case comment
// @Summary Delete case comment
// @Description Delete a comment (only author can delete)
// @Tags Cases
// @Security BearerAuth
// @Produce json
// @Param id path string true "Case ID"
// @Param commentId path string true "Comment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /cases/{id}/comments/{commentId} [delete]
func (h *CaseHandler) DeleteComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == nil {
		response.Error(c, middleware.ErrUnauthorized)
		return
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid comment ID", 400))
		return
	}

	if err := h.caseService.DeleteComment(c.Request.Context(), commentID, *userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
