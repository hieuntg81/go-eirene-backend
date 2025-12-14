package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"bamboo-rescue/pkg/apperror"
	"bamboo-rescue/pkg/response"
	"go.uber.org/zap"
)

// Re-export from apperror for backward compatibility
type AppError = apperror.AppError

var (
	ErrNotFound           = apperror.ErrNotFound
	ErrUnauthorized       = apperror.ErrUnauthorized
	ErrForbidden          = apperror.ErrForbidden
	ErrBadRequest         = apperror.ErrBadRequest
	ErrConflict           = apperror.ErrConflict
	ErrInternalServer     = apperror.ErrInternalServer
	ErrValidation         = apperror.ErrValidation
	ErrInvalidToken       = apperror.ErrInvalidToken
	ErrUserNotFound       = apperror.ErrUserNotFound
	ErrCaseNotFound       = apperror.ErrCaseNotFound
	ErrInvalidCredentials = apperror.ErrInvalidCredentials
	ErrEmailExists        = apperror.ErrEmailExists
	ErrPhoneExists        = apperror.ErrPhoneExists
)

func NewAppError(code string, message string, status int) *AppError {
	return apperror.NewAppError(code, message, status)
}

func WrapError(err error, appErr *AppError) *AppError {
	return apperror.WrapError(err, appErr)
}

// ErrorHandler middleware handles panics and errors
func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", GetRequestID(c)),
				)
				response.InternalServerError(c, "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()

		// Handle errors set during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err, log)
		}
	}
}

func handleError(c *gin.Context, err error, log *zap.Logger) {
	// Check if it's an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		response.Error(c, appErr)
		return
	}

	// Check if it's a validation error
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		message := formatValidationErrors(validationErrors)
		response.BadRequest(c, message)
		return
	}

	// Log unknown errors
	log.Error("Unhandled error",
		zap.Error(err),
		zap.String("request_id", GetRequestID(c)),
	)

	response.InternalServerError(c, "An unexpected error occurred")
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "Validation failed"
	}

	// Return first error message
	err := errs[0]
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + err.Param() + " characters"
	case "max":
		return field + " must be at most " + err.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + err.Param()
	case "uuid":
		return field + " must be a valid UUID"
	case "url":
		return field + " must be a valid URL"
	default:
		return field + " is invalid"
	}
}

// HandleError is a helper to handle errors in handlers
func HandleError(c *gin.Context, err error, log *zap.Logger) {
	handleError(c, err, log)
}
