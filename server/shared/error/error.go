package errors

import (
	"fmt"
	"net/http"
)

// AppError is our custom error struct
type AppError struct {
	Code    int               `json:"code"`              // HTTP Status Code
	Message string            `json:"message"`           // User-facing message
	Err     error             `json:"-"`                 // Internal error (not sent to user)
	Fields  map[string]string `json:"fields,omitempty"`  // Validation errors (e.g. "email": "invalid")
}

// Error implements the standard error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows checking the internal error type
func (e *AppError) Unwrap() error {
	return e.Err
}

// ───────────────── Constructors ─────────────────

// New creates a simple error
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an internal error (e.g., DB error) with a status code
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ───────────────── Helper Methods (Shortcuts) ─────────────────

// 400: Bad Request (Simple message)
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// 400: Bad Request (With specific field errors)
func Validation(message string, fields map[string]string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Fields:  fields,
	}
}

// 401: Unauthorized (Login required)
func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message)
}

// 403: Forbidden (Permission denied)
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

// 404: Not Found
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

// 409: Conflict (Duplicate email, etc.)
func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

// 422: Unprocessable Entity
func Unprocessable(message string) *AppError {
	return New(http.StatusUnprocessableEntity, message)
}

// 500: Internal Server Error (Auto-hides the real error from user)
func Internal(err error) *AppError {
	// Default message for user, original error stored in Err
	return Wrap(err, http.StatusInternalServerError, "Something went wrong, please try again later")
}