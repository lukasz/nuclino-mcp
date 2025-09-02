package nuclino

import (
	"fmt"
	"net/http"
)

// APIError represents an error from the Nuclino API
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Nuclino API error (status %d): %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("Nuclino API error (status %d): %s", e.StatusCode, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(statusCode int, message, details string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// IsNotFound checks if the error is a 404 not found error
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

// IsUnauthorized checks if the error is a 401 unauthorized error
func IsUnauthorized(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusUnauthorized
}

// IsForbidden checks if the error is a 403 forbidden error
func IsForbidden(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusForbidden
}

// IsRateLimited checks if the error is a 429 rate limit error
func IsRateLimited(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusTooManyRequests
}

// IsBadRequest checks if the error is a 400 bad request error
func IsBadRequest(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusBadRequest
}

// IsServerError checks if the error is a 5xx server error
func IsServerError(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode >= 500
}
