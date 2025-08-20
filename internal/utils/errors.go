package utils

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// AppError represents a custom application error.
type AppError struct {
	Code       string // A unique, machine-readable error code
	Message    string // A human-readable message for the client
	HTTPStatus int    // The HTTP status code to return
	Err        error  // The underlying wrapped error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap provides compatibility for errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Pre-defined application errors
var (
	ErrInternal      = &AppError{Code: "INTERNAL_ERROR", Message: "An internal server error occurred.", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound      = &AppError{Code: "NOT_FOUND", Message: "The requested resource was not found.", HTTPStatus: http.StatusNotFound}
	ErrBadRequest    = &AppError{Code: "BAD_REQUEST", Message: "The request is invalid.", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Message: "Authentication is required and has failed or has not yet been provided.", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden     = &AppError{Code: "FORBIDDEN", Message: "You do not have permission to access this resource.", HTTPStatus: http.StatusForbidden}
	ErrInvalidClient = &AppError{Code: "INVALID_CLIENT", Message: "Client authentication failed.", HTTPStatus: http.StatusUnauthorized}
)

// HandleError is a centralized function to process errors and send a standardized JSON response.
func HandleError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	var appErr *AppError
	// Use errors.As to check if the error is our custom AppError type.
	if !errors.As(err, &appErr) {
		// If it's not an AppError, wrap it in a generic internal error.
		appErr = ErrInternal
		appErr.Err = err // Keep the original error for logging
	}

	// Log the error with structured details.
	logAttr := []slog.Attr{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("error_code", appErr.Code),
		slog.String("error_message", appErr.Message),
	}
	if appErr.Err != nil {
		// Include the underlying error message for better debugging.
		logAttr = append(logAttr, slog.String("underlying_error", appErr.Err.Error()))
	}
	logger.Error("HTTP handler error", logAttr)

	// Prepare the JSON response.
	// For internal errors, we don't expose the specific message to the client.
	responseMessage := appErr.Message
	if appErr.HTTPStatus >= 500 {
		responseMessage = "An internal server error occurred."
	}

	errorResponse := map[string]string{
		"error":             appErr.Code,
		"error_description": responseMessage,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(appErr.HTTPStatus)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		logger.Error("Failed to write error response", slog.String("error", err.Error()))
	}
}
