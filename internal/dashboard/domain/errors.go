package domain

import (
	"errors"
)

// Error represents a dashboard domain error.
type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying error if one was attached.
func (e *Error) Unwrap() error {
	return e.Err
}

// IsDomainError checks if any error in the chain is a *domain.Error.
func IsDomainError(err error) bool {
	var domainErr *Error
	return errors.As(err, &domainErr)
}

// AsDomainError extracts the first *domain.Error from the error chain.
func AsDomainError(err error) (*Error, bool) {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

// Common domain errors for dashboard
var (
	ErrDatabaseUnavailable = &Error{
		Code:    "DATABASE_UNAVAILABLE",
		Message: "database connection is unavailable",
	}
	ErrHealthCheckFailed = &Error{
		Code:    "HEALTH_CHECK_FAILED",
		Message: "health check failed",
	}
	ErrInvalidRequest = &Error{
		Code:    "INVALID_REQUEST",
		Message: "request payload or parameters are invalid",
	}
	ErrNotFound = &Error{
		Code:    "NOT_FOUND",
		Message: "requested entity was not found",
	}
)
