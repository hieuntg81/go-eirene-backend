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

// MediaHandler handles media requests
type MediaHandler struct {
	mediaService service.MediaService
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// Upload handles media upload
// @Summary Upload media
// @Description Upload media file for a case
// @Tags Media
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param case_id formData string true "Case ID"
// @Param file formData file true "Media file"
// @Success 201 {object} pkgresponse.Response{data=response.MediaUploadResponse}
// @Failure 400 {object} pkgresponse.Response
// @Failure 401 {object} pkgresponse.Response
// @Router /media/upload [post]
func (h *MediaHandler) Upload(c *gin.Context) {
	caseIDStr := c.PostForm("case_id")
	if caseIDStr == "" {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Case ID is required", 400))
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "File is required", 400))
		return
	}

	result, err := h.mediaService.Upload(c.Request.Context(), file, caseID)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusCreated, response.MediaUploadResponse{
		ID:           result.ID,
		URL:          result.URL,
		ThumbnailURL: result.ThumbnailURL,
		MediaType:    result.MediaType,
		FileSize:     result.FileSize,
	})
}

// UploadMultiple handles multiple media upload
// @Summary Upload multiple media
// @Description Upload multiple media files for a case
// @Tags Media
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param case_id formData string true "Case ID"
// @Param files formData file true "Media files"
// @Success 201 {object} pkgresponse.Response{data=[]response.MediaUploadResponse}
// @Failure 400 {object} pkgresponse.Response
// @Failure 401 {object} pkgresponse.Response
// @Router /media/upload-multiple [post]
func (h *MediaHandler) UploadMultiple(c *gin.Context) {
	caseIDStr := c.PostForm("case_id")
	if caseIDStr == "" {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Case ID is required", 400))
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid case ID", 400))
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Failed to parse multipart form", 400))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "At least one file is required", 400))
		return
	}

	results, err := h.mediaService.UploadMultiple(c.Request.Context(), files, caseID)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	responses := make([]response.MediaUploadResponse, len(results))
	for i, r := range results {
		responses[i] = response.MediaUploadResponse{
			ID:           r.ID,
			URL:          r.URL,
			ThumbnailURL: r.ThumbnailURL,
			MediaType:    r.MediaType,
			FileSize:     r.FileSize,
		}
	}

	pkgresponse.Success(c, http.StatusCreated, responses)
}

// Delete handles media deletion
// @Summary Delete media
// @Description Delete a media file
// @Tags Media
// @Security BearerAuth
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} pkgresponse.Response
// @Failure 401 {object} pkgresponse.Response
// @Failure 404 {object} pkgresponse.Response
// @Router /media/{id} [delete]
func (h *MediaHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	mediaID, err := uuid.Parse(idStr)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid media ID", 400))
		return
	}

	if err := h.mediaService.Delete(c.Request.Context(), mediaID); err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, gin.H{"message": "Media deleted successfully"})
}
