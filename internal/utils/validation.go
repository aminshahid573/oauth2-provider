package utils

import (
	"net/url"
	"strings"
)

// Validator holds validation errors.
type Validator struct {
	Errors map[string]string
}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if there are no validation errors.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the map.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check is a generic helper for adding errors.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// NotBlank checks that a string value is not just whitespace.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars checks that a string value does not exceed a specific length.
func MaxChars(value string, n int) bool {
	return len(value) <= n
}

// IsValidURL checks if a string is a valid URL.
func IsValidURL(value string) bool {
	_, err := url.ParseRequestURI(value)
	return err == nil
}
