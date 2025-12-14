package apperror

import "net/http"

// AppError represents an application error
type AppError struct {
	Code       string
	Message    string
	Status     int
	StatusCode int // Alias for Status for backward compatibility
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Common application errors
var (
	ErrNotFound           = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: http.StatusNotFound, StatusCode: http.StatusNotFound}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", Status: http.StatusUnauthorized, StatusCode: http.StatusUnauthorized}
	ErrForbidden          = &AppError{Code: "FORBIDDEN", Message: "Forbidden", Status: http.StatusForbidden, StatusCode: http.StatusForbidden}
	ErrBadRequest         = &AppError{Code: "BAD_REQUEST", Message: "Bad request", Status: http.StatusBadRequest, StatusCode: http.StatusBadRequest}
	ErrConflict           = &AppError{Code: "CONFLICT", Message: "Resource already exists", Status: http.StatusConflict, StatusCode: http.StatusConflict}
	ErrInternalServer     = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error", Status: http.StatusInternalServerError, StatusCode: http.StatusInternalServerError}
	ErrValidation         = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed", Status: http.StatusBadRequest, StatusCode: http.StatusBadRequest}
	ErrInvalidToken       = &AppError{Code: "INVALID_TOKEN", Message: "Invalid or expired token", Status: http.StatusUnauthorized, StatusCode: http.StatusUnauthorized}
	ErrUserNotFound       = &AppError{Code: "USER_NOT_FOUND", Message: "User not found", Status: http.StatusNotFound, StatusCode: http.StatusNotFound}
	ErrCaseNotFound       = &AppError{Code: "CASE_NOT_FOUND", Message: "Case not found", Status: http.StatusNotFound, StatusCode: http.StatusNotFound}
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password", Status: http.StatusUnauthorized, StatusCode: http.StatusUnauthorized}
	ErrEmailExists        = &AppError{Code: "EMAIL_EXISTS", Message: "Email already registered", Status: http.StatusConflict, StatusCode: http.StatusConflict}
	ErrPhoneExists        = &AppError{Code: "PHONE_EXISTS", Message: "Phone number already registered", Status: http.StatusConflict, StatusCode: http.StatusConflict}
)

// NewAppError creates a new AppError with a custom message
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Status:     status,
		StatusCode: status,
	}
}

// WrapError wraps an error with an AppError
func WrapError(err error, appErr *AppError) *AppError {
	return &AppError{
		Code:       appErr.Code,
		Message:    appErr.Message,
		Status:     appErr.Status,
		StatusCode: appErr.Status,
		Err:        err,
	}
}
