// Package apperror defines a structured application error so the service
// layer can raise domain errors that handlers translate into consistent
// HTTP responses without each handler re-deciding status codes/messages.
package apperror

import "net/http"

// AppError is a structured error carrying the HTTP status and a stable code
// alongside the human-readable message.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Errors     map[string]string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func WithErrors(code, message string, httpStatus int, errs map[string]string) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, Errors: errs}
}

func NotFound(message string) *AppError {
	return New("NOT_FOUND", message, http.StatusNotFound)
}

func BadRequest(message string) *AppError {
	return New("BAD_REQUEST", message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New("FORBIDDEN", message, http.StatusForbidden)
}

func Conflict(message string) *AppError {
	return New("CONFLICT", message, http.StatusConflict)
}

func Internal(message string) *AppError {
	return New("INTERNAL_ERROR", message, http.StatusInternalServerError)
}

func TooManyRequests(message string) *AppError {
	return New("TOO_MANY_REQUESTS", message, http.StatusTooManyRequests)
}

func ValidationFailed(errs map[string]string) *AppError {
	return WithErrors("VALIDATION_FAILED", "Validation failed", http.StatusBadRequest, errs)
}
