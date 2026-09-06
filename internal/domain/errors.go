package domain

import (
	"errors"
)

// Error represents a standardized domain error.
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

// Common shared domain errors
var (
	ErrInternalServer = &Error{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "internal server error",
	}
	ErrInvalidInput = &Error{
		Code:    "INVALID_INPUT",
		Message: "invalid input",
	}
	ErrUnauthorized = &Error{
		Code:    "UNAUTHORIZED",
		Message: "unauthorized access",
	}
	ErrNotFound = &Error{
		Code:    "NOT_FOUND",
		Message: "resource not found",
	}
)
