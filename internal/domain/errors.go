package domain

import (
	"fmt"
	"net/http"
)

// ErrDetail represents a field-level error detail.
type ErrDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// AppError represents a domain-level application error.
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    []ErrDetail `json:"details,omitempty"`
	StatusCode int         `json:"-"`
	Err        error       `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// --- Constructor Helpers ---

// ErrNotFound returns a 404 not found error.
func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// ErrValidation returns a 422 validation error.
func ErrValidation(details []ErrDetail) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "One or more fields are invalid",
		Details:    details,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// ErrConflict returns a 409 conflict error.
func ErrConflict(msg string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    msg,
		StatusCode: http.StatusConflict,
	}
}

// ErrUnauthorized returns a 401 unauthorized error.
func ErrUnauthorized() *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		StatusCode: http.StatusUnauthorized,
	}
}

// ErrForbidden returns a 403 forbidden error.
func ErrForbidden() *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    "You do not have permission to perform this action",
		StatusCode: http.StatusForbidden,
	}
}

// ErrInternal returns a 500 internal server error.
func ErrInternal(err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An unexpected error occurred",
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// ErrBadRequest returns a 400 bad request error.
func ErrBadRequest(msg string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    msg,
		StatusCode: http.StatusBadRequest,
	}
}

// ErrBookingConflict returns a 409 booking conflict error.
func ErrBookingConflict() *AppError {
	return &AppError{
		Code:       "BOOKING_CONFLICT",
		Message:    "The requested dates are unavailable",
		StatusCode: http.StatusConflict,
	}
}

// ErrBookingNotPending returns a 409 error when booking is not in pending status.
func ErrBookingNotPending() *AppError {
	return &AppError{
		Code:       "BOOKING_NOT_PENDING",
		Message:    "Booking status does not allow this action",
		StatusCode: http.StatusConflict,
	}
}

// ErrPaymentFailed returns a 402 payment failed error.
func ErrPaymentFailed(msg string) *AppError {
	return &AppError{
		Code:       "PAYMENT_FAILED",
		Message:    msg,
		StatusCode: http.StatusPaymentRequired,
	}
}

// ErrPaymentTimeout returns a 408 payment timeout error.
func ErrPaymentTimeout() *AppError {
	return &AppError{
		Code:       "PAYMENT_TIMEOUT",
		Message:    "Payment provider did not respond in time",
		StatusCode: http.StatusRequestTimeout,
	}
}

// ErrOTPExpired returns a 400 OTP expired error.
func ErrOTPExpired() *AppError {
	return &AppError{
		Code:       "OTP_EXPIRED",
		Message:    "OTP code has expired",
		StatusCode: http.StatusBadRequest,
	}
}

// ErrOTPInvalid returns a 400 OTP invalid error.
func ErrOTPInvalid() *AppError {
	return &AppError{
		Code:       "OTP_INVALID",
		Message:    "OTP code is incorrect",
		StatusCode: http.StatusBadRequest,
	}
}

// ErrRateLimited returns a 429 rate limited error.
func ErrRateLimited() *AppError {
	return &AppError{
		Code:       "RATE_LIMITED",
		Message:    "Too many requests, please try again later",
		StatusCode: http.StatusTooManyRequests,
	}
}

// ErrServiceUnavailable returns a 503 service unavailable error.
func ErrServiceUnavailable(service string) *AppError {
	return &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    fmt.Sprintf("%s service is currently unavailable", service),
		StatusCode: http.StatusServiceUnavailable,
	}
}
