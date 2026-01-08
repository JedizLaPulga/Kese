package kese

import (
	"errors"
	"fmt"
)

// ErrorHandler is a function that handles errors returned by handlers.
// It receives the context and the error, and should write an appropriate response.
type ErrorHandler func(err error) (int, interface{})

// DefaultErrorHandler is the default error handler that returns appropriate status codes.
// It does not expose internal error details to clients for security reasons.
// The actual error is logged by the framework in kese.go ServeHTTP.
func DefaultErrorHandler(err error) (int, interface{}) {
	// Check for common error types
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return 400, map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErr.Errors,
		}
	}

	// Default to 500 Internal Server Error
	// Don't expose internal error details to clients in production
	return 500, map[string]string{
		"error": "Internal Server Error",
	}
}

// ValidationError represents validation errors for struct fields.
type ValidationError struct {
	Errors map[string]string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(v.Errors))
}

// NewValidationError creates a new validation error.
func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

// Add adds a field error to the validation error.
func (v *ValidationError) Add(field, message string) {
	v.Errors[field] = message
}

// HasErrors returns true if there are any validation errors.
func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}
