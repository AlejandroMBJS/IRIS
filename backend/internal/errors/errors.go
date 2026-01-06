/*
Package errors - Custom Error Types for IRIS Payroll System

==============================================================================
FILE: internal/errors/errors.go
==============================================================================

DESCRIPTION:
    Provides typed error definitions for consistent error handling across the
    application. Replaces string-based error checking with type assertions,
    making error handling more robust and maintainable.

USAGE:
    // In service layer:
    return errors.ErrInvalidCredentials

    // In handler layer:
    if errors.Is(err, errors.ErrInvalidCredentials) {
        c.JSON(http.StatusUnauthorized, ...)
    }

    // For wrapped errors:
    return errors.Wrap(err, errors.ErrDatabaseOperation)

DEVELOPER GUIDELINES:
    OK to modify: Add new error types as needed
    CAUTION: Changing error messages may affect frontend error display
    DO NOT modify: Error interface implementation

==============================================================================
*/
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Re-export standard library functions for convenience
var (
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
)

// AppError represents an application-level error with HTTP status code
type AppError struct {
	Code       string // Machine-readable error code
	Message    string // Human-readable message
	HTTPStatus int    // HTTP status code for API responses
	Err        error  // Underlying error (optional)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Is implements error matching for errors.Is()
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAppError creates a new application error
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Wrap wraps an underlying error with an AppError
func Wrap(err error, appErr *AppError) *AppError {
	return &AppError{
		Code:       appErr.Code,
		Message:    appErr.Message,
		HTTPStatus: appErr.HTTPStatus,
		Err:        err,
	}
}

// WithMessage creates a copy of the error with a custom message
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// ============================================================================
// Authentication Errors
// ============================================================================

var (
	ErrInvalidCredentials = NewAppError(
		"AUTH_INVALID_CREDENTIALS",
		"Invalid email or password",
		http.StatusUnauthorized,
	)

	ErrAccountDeactivated = NewAppError(
		"AUTH_ACCOUNT_DEACTIVATED",
		"Account is deactivated",
		http.StatusUnauthorized,
	)

	ErrInvalidToken = NewAppError(
		"AUTH_INVALID_TOKEN",
		"Invalid or expired token",
		http.StatusUnauthorized,
	)

	ErrTokenExpired = NewAppError(
		"AUTH_TOKEN_EXPIRED",
		"Token has expired",
		http.StatusUnauthorized,
	)

	ErrRefreshTokenInvalid = NewAppError(
		"AUTH_REFRESH_TOKEN_INVALID",
		"Invalid refresh token",
		http.StatusUnauthorized,
	)

	ErrUnauthorized = NewAppError(
		"AUTH_UNAUTHORIZED",
		"Unauthorized access",
		http.StatusUnauthorized,
	)

	ErrForbidden = NewAppError(
		"AUTH_FORBIDDEN",
		"Insufficient permissions",
		http.StatusForbidden,
	)
)

// ============================================================================
// Validation Errors
// ============================================================================

var (
	ErrValidationFailed = NewAppError(
		"VALIDATION_FAILED",
		"Validation failed",
		http.StatusBadRequest,
	)

	ErrInvalidInput = NewAppError(
		"VALIDATION_INVALID_INPUT",
		"Invalid input provided",
		http.StatusBadRequest,
	)

	ErrMissingField = NewAppError(
		"VALIDATION_MISSING_FIELD",
		"Required field is missing",
		http.StatusBadRequest,
	)

	ErrInvalidEmail = NewAppError(
		"VALIDATION_INVALID_EMAIL",
		"Invalid email format",
		http.StatusBadRequest,
	)

	ErrPasswordTooWeak = NewAppError(
		"VALIDATION_PASSWORD_WEAK",
		"Password does not meet requirements",
		http.StatusBadRequest,
	)

	ErrPasswordMismatch = NewAppError(
		"VALIDATION_PASSWORD_MISMATCH",
		"Current password is incorrect",
		http.StatusBadRequest,
	)
)

// ============================================================================
// Resource Errors
// ============================================================================

var (
	ErrNotFound = NewAppError(
		"RESOURCE_NOT_FOUND",
		"Resource not found",
		http.StatusNotFound,
	)

	ErrAlreadyExists = NewAppError(
		"RESOURCE_ALREADY_EXISTS",
		"Resource already exists",
		http.StatusConflict,
	)

	ErrEmailAlreadyExists = NewAppError(
		"RESOURCE_EMAIL_EXISTS",
		"Email already registered",
		http.StatusConflict,
	)

	ErrRFCAlreadyExists = NewAppError(
		"RESOURCE_RFC_EXISTS",
		"RFC already registered",
		http.StatusConflict,
	)

	ErrEmployeeNumberExists = NewAppError(
		"RESOURCE_EMPLOYEE_NUMBER_EXISTS",
		"Employee number already exists",
		http.StatusConflict,
	)
)

// ============================================================================
// Database Errors
// ============================================================================

var (
	ErrDatabaseOperation = NewAppError(
		"DATABASE_ERROR",
		"Database operation failed",
		http.StatusInternalServerError,
	)

	ErrRecordNotFound = NewAppError(
		"DATABASE_RECORD_NOT_FOUND",
		"Record not found",
		http.StatusNotFound,
	)

	ErrDuplicateKey = NewAppError(
		"DATABASE_DUPLICATE_KEY",
		"Duplicate key violation",
		http.StatusConflict,
	)
)

// ============================================================================
// Business Logic Errors
// ============================================================================

var (
	ErrPayrollAlreadyProcessed = NewAppError(
		"PAYROLL_ALREADY_PROCESSED",
		"Payroll period has already been processed",
		http.StatusConflict,
	)

	ErrPayrollPeriodClosed = NewAppError(
		"PAYROLL_PERIOD_CLOSED",
		"Payroll period is closed",
		http.StatusConflict,
	)

	ErrIncidenceAlreadyApproved = NewAppError(
		"INCIDENCE_ALREADY_APPROVED",
		"Incidence has already been approved",
		http.StatusConflict,
	)

	ErrInsufficientVacationDays = NewAppError(
		"VACATION_INSUFFICIENT_DAYS",
		"Insufficient vacation days available",
		http.StatusBadRequest,
	)

	ErrInvalidDateRange = NewAppError(
		"BUSINESS_INVALID_DATE_RANGE",
		"Invalid date range",
		http.StatusBadRequest,
	)
)

// ============================================================================
// File/Upload Errors
// ============================================================================

var (
	ErrFileTooLarge = NewAppError(
		"FILE_TOO_LARGE",
		"File size exceeds maximum allowed",
		http.StatusBadRequest,
	)

	ErrInvalidFileType = NewAppError(
		"FILE_INVALID_TYPE",
		"File type not allowed",
		http.StatusBadRequest,
	)

	ErrFileUploadFailed = NewAppError(
		"FILE_UPLOAD_FAILED",
		"Failed to upload file",
		http.StatusInternalServerError,
	)
)

// ============================================================================
// Rate Limiting Errors
// ============================================================================

var (
	ErrRateLimitExceeded = NewAppError(
		"RATE_LIMIT_EXCEEDED",
		"Too many requests, please try again later",
		http.StatusTooManyRequests,
	)
)

// ============================================================================
// Internal Errors
// ============================================================================

var (
	ErrInternal = NewAppError(
		"INTERNAL_ERROR",
		"An internal error occurred",
		http.StatusInternalServerError,
	)

	ErrServiceUnavailable = NewAppError(
		"SERVICE_UNAVAILABLE",
		"Service temporarily unavailable",
		http.StatusServiceUnavailable,
	)
)

// ============================================================================
// Helper Functions
// ============================================================================

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
func GetErrorCode(err error) string {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetErrorMessage returns the user-friendly message for an error
func GetErrorMessage(err error) string {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}
