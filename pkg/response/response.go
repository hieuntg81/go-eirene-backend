package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"bamboo-rescue/pkg/apperror"
)

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool       `json:"success"`
	Error   *ErrorInfo `json:"error"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta contains pagination information
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items interface{} `json:"items"`
	Meta  Meta        `json:"meta"`
}

// Success sends a successful response with data directly (no wrapper)
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// SuccessWithMeta sends a successful response with data and pagination meta
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, ListResponse{
		Items: data,
		Meta:  *meta,
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// NoContent sends a 204 no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error handles errors and sends appropriate response
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.StatusCode, ErrorResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	// Default internal server error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		},
	})
}

// ValidationError sends a 400 error with validation details
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		},
	})
}

// BadRequest sends a 400 bad request error
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

// Unauthorized sends a 401 unauthorized error
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a 403 forbidden error
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a 404 not found error
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// InternalServerError sends a 500 internal server error
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: message,
		},
	})
}

// TooManyRequests sends a 429 too many requests error
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    "TOO_MANY_REQUESTS",
			Message: message,
		},
	})
}

// CalculateTotalPages calculates total pages for pagination
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}
	return pages
}

// NewMeta creates a new Meta instance
func NewMeta(page, limit int, total int64) *Meta {
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: CalculateTotalPages(total, limit),
	}
}
