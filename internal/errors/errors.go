package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeAPI            ErrorType = "api"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeCircuitBreaker ErrorType = "circuit_breaker"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeConflict       ErrorType = "conflict"
)

// Severity represents the severity level of an error
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Error represents a comprehensive error with context
type Error struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	HTTPStatus int                    `json:"http_status,omitempty"`
	Severity   Severity               `json:"severity"`
	Retryable  bool                   `json:"retryable"`
	Timestamp  time.Time              `json:"timestamp"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Stack      string                 `json:"stack,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Message, e.Details, e.Code)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a target error type
func (e *Error) Is(target error) bool {
	if targetErr, ok := target.(*Error); ok {
		return e.Type == targetErr.Type && e.Code == targetErr.Code
	}
	return false
}

// WithContext adds context information to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// NewError creates a new Error instance
func NewError(errType ErrorType, code, message string) *Error {
	return &Error{
		Type:      errType,
		Code:      code,
		Message:   message,
		Severity:  SeverityMedium,
		Retryable: false,
		Timestamp: time.Now(),
	}
}

// Validation errors
func NewValidationError(field, message string) *Error {
	return NewError(ErrorTypeValidation, "VALIDATION_ERROR", fmt.Sprintf("Validation failed for field '%s': %s", field, message)).
		WithContext("field", field).
		WithSeverity(SeverityLow)
}

func NewRequiredFieldError(field string) *Error {
	return NewValidationError(field, "field is required")
}

func NewInvalidTypeError(field, expectedType string) *Error {
	return NewValidationError(field, fmt.Sprintf("expected type %s", expectedType))
}

// Authentication/Authorization errors
func NewAuthenticationError(message string) *Error {
	return NewError(ErrorTypeAuthentication, "AUTH_ERROR", message).
		WithHTTPStatus(http.StatusUnauthorized).
		WithSeverity(SeverityHigh)
}

func NewAuthorizationError(resource string) *Error {
	return NewError(ErrorTypeAuthorization, "FORBIDDEN", fmt.Sprintf("Access denied to resource: %s", resource)).
		WithHTTPStatus(http.StatusForbidden).
		WithSeverity(SeverityHigh).
		WithContext("resource", resource)
}

// Rate limiting errors
func NewRateLimitError(resetTime time.Time) *Error {
	return NewError(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded").
		WithHTTPStatus(http.StatusTooManyRequests).
		WithSeverity(SeverityMedium).
		WithRetryable(true).
		WithContext("reset_time", resetTime)
}

// Network errors
func NewNetworkError(operation string, cause error) *Error {
	return NewError(ErrorTypeNetwork, "NETWORK_ERROR", fmt.Sprintf("Network error during %s", operation)).
		WithCause(cause).
		WithRetryable(true).
		WithSeverity(SeverityMedium).
		WithContext("operation", operation)
}

func NewTimeoutError(operation string, timeout time.Duration) *Error {
	return NewError(ErrorTypeTimeout, "TIMEOUT", fmt.Sprintf("Operation %s timed out after %s", operation, timeout)).
		WithHTTPStatus(http.StatusRequestTimeout).
		WithRetryable(true).
		WithSeverity(SeverityMedium).
		WithContext("operation", operation).
		WithContext("timeout", timeout)
}

// API errors
func NewAPIError(httpStatus int, code, message string) *Error {
	severity := SeverityMedium
	if httpStatus >= 500 {
		severity = SeverityHigh
	}

	retryable := httpStatus == http.StatusTooManyRequests || httpStatus >= 500

	return NewError(ErrorTypeAPI, code, message).
		WithHTTPStatus(httpStatus).
		WithSeverity(severity).
		WithRetryable(retryable)
}

func NewNotFoundError(resource, id string) *Error {
	return NewError(ErrorTypeNotFound, "NOT_FOUND", fmt.Sprintf("%s with ID %s not found", resource, id)).
		WithHTTPStatus(http.StatusNotFound).
		WithSeverity(SeverityLow).
		WithContext("resource", resource).
		WithContext("id", id)
}

func NewConflictError(resource, reason string) *Error {
	return NewError(ErrorTypeConflict, "CONFLICT", fmt.Sprintf("Conflict with %s: %s", resource, reason)).
		WithHTTPStatus(http.StatusConflict).
		WithSeverity(SeverityMedium).
		WithContext("resource", resource).
		WithContext("reason", reason)
}

// Circuit breaker errors
func NewCircuitBreakerError(state string) *Error {
	return NewError(ErrorTypeCircuitBreaker, "CIRCUIT_BREAKER_OPEN", fmt.Sprintf("Circuit breaker is %s", state)).
		WithRetryable(true).
		WithSeverity(SeverityHigh).
		WithContext("circuit_breaker_state", state)
}

// Internal errors
func NewInternalError(operation string, cause error) *Error {
	return NewError(ErrorTypeInternal, "INTERNAL_ERROR", fmt.Sprintf("Internal error during %s", operation)).
		WithCause(cause).
		WithHTTPStatus(http.StatusInternalServerError).
		WithSeverity(SeverityCritical).
		WithContext("operation", operation)
}

// Helper methods for Error
func (e *Error) WithHTTPStatus(status int) *Error {
	e.HTTPStatus = status
	return e
}

func (e *Error) WithSeverity(severity Severity) *Error {
	e.Severity = severity
	return e
}

func (e *Error) WithRetryable(retryable bool) *Error {
	e.Retryable = retryable
	return e
}

func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger Logger
}

// Logger interface for error logging
type Logger interface {
	Error(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// Handle processes an error and logs it appropriately
func (h *ErrorHandler) Handle(err error) *Error {
	var appErr *Error

	// Convert to our Error type if needed
	if e, ok := err.(*Error); ok {
		appErr = e
	} else {
		appErr = NewInternalError("unknown", err)
	}

	// Log based on severity
	fields := map[string]interface{}{
		"error_type": appErr.Type,
		"error_code": appErr.Code,
		"retryable":  appErr.Retryable,
		"timestamp":  appErr.Timestamp,
	}

	// Add context if available
	for k, v := range appErr.Context {
		fields[k] = v
	}

	switch appErr.Severity {
	case SeverityCritical, SeverityHigh:
		h.logger.Error(appErr.Error(), fields)
	case SeverityMedium:
		h.logger.Warn(appErr.Error(), fields)
	case SeverityLow:
		h.logger.Info(appErr.Error(), fields)
	}

	return appErr
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*Error); ok {
		return appErr.Retryable
	}
	return false
}

// GetHTTPStatus extracts HTTP status code from error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*Error); ok && appErr.HTTPStatus > 0 {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorType extracts error type
func GetErrorType(err error) ErrorType {
	if appErr, ok := err.(*Error); ok {
		return appErr.Type
	}
	return ErrorTypeInternal
}

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []ErrorType
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []ErrorType{
			ErrorTypeNetwork,
			ErrorTypeTimeout,
			ErrorTypeRateLimit,
			ErrorTypeCircuitBreaker,
		},
	}
}

// ShouldRetry checks if an error should be retried based on configuration
func (c RetryConfig) ShouldRetry(err error, attempt int) bool {
	if attempt >= c.MaxRetries {
		return false
	}

	if !IsRetryable(err) {
		return false
	}

	errType := GetErrorType(err)
	for _, retryableType := range c.RetryableErrors {
		if errType == retryableType {
			return true
		}
	}

	return false
}

// CalculateDelay calculates the delay for a retry attempt
func (c RetryConfig) CalculateDelay(attempt int) time.Duration {
	delay := float64(c.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= c.BackoffFactor
	}

	if time.Duration(delay) > c.MaxDelay {
		return c.MaxDelay
	}

	return time.Duration(delay)
}
