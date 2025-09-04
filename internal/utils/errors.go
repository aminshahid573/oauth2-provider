package utils

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// AppError represents a custom application error.
type AppError struct {
	Code       string
	Title      string
	Message    string
	HTTPStatus int
	Err        error
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
	ErrInternal      = &AppError{Code: "INTERNAL_ERROR", Title: "Server Error", Message: "An internal server error occurred.", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound      = &AppError{Code: "NOT_FOUND", Title: "Not Found", Message: "The requested resource was not found.", HTTPStatus: http.StatusNotFound}
	ErrBadRequest    = &AppError{Code: "BAD_REQUEST", Title: "Bad Request", Message: "The request is invalid.", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Title: "Unauthorized", Message: "Authentication is required and has failed or has not yet been provided.", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden     = &AppError{Code: "FORBIDDEN", Title: "Forbidden", Message: "You do not have permission to access this resource.", HTTPStatus: http.StatusForbidden}
	ErrInvalidClient = &AppError{Code: "INVALID_CLIENT", Title: "Unauthorized", Message: "Client authentication failed.", HTTPStatus: http.StatusUnauthorized}
)

// HandleError is a centralized function to process errors and send a standardized response.
// It inspects the Accept header to decide whether to send JSON or render an HTML error page.
func HandleError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, tc TemplateCache, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = ErrInternal
		appErr.Err = err
	}

	// Log the detailed error for internal diagnostics.
	logAttr := []slog.Attr{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("error_code", appErr.Code),
		slog.String("error_message", appErr.Message),
	}
	if appErr.Err != nil {
		logAttr = append(logAttr, slog.String("underlying_error", appErr.Err.Error()))
	}
	logger.Error("HTTP handler error", logAttr)

	// Check if the client prefers an HTML response.
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		renderErrorPage(w, r, tc, appErr)
	} else {
		// Default to a JSON response for API clients.
		WriteJSONError(w, appErr)
	}
}

// HandleAPIError is a convenience wrapper for API handlers that only ever send JSON.
func HandleAPIError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = ErrInternal
		appErr.Err = err
	}
	logAttr := []slog.Attr{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("error_code", appErr.Code),
		slog.String("error_message", appErr.Message),
	}
	if appErr.Err != nil {
		logAttr = append(logAttr, slog.String("underlying_error", appErr.Err.Error()))
	}
	logger.Error("HTTP handler error", logAttr)
	WriteJSONError(w, appErr)
}

// writeJSONError sends a standard JSON error response.
func WriteJSONError(w http.ResponseWriter, appErr *AppError) {
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
	_ = json.NewEncoder(w).Encode(errorResponse)
}

// renderErrorPage renders the standard HTML error page.
func renderErrorPage(w http.ResponseWriter, r *http.Request, tc TemplateCache, appErr *AppError) {
	w.WriteHeader(appErr.HTTPStatus)
	data := map[string]any{
		"StatusCode": appErr.HTTPStatus,
		"Title":      appErr.Title,
		"Message":    appErr.Message,
	}
	// Use a fallback to plain text if the error template itself fails to render.
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "An unexpected error occurred while rendering the error page.", http.StatusInternalServerError)
		}
	}()
	tc.Render(w, r, "base.html", "error.html", data)
}
