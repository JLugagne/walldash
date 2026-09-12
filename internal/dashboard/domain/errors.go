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
	ErrLevelNotFound = &Error{
		Code:    "LEVEL_NOT_FOUND",
		Message: "level not found",
	}
	ErrInvalidLevel = &Error{
		Code:    "INVALID_LEVEL",
		Message: "invalid level data",
	}
	ErrPlanNotFound = &Error{
		Code:    "PLAN_NOT_FOUND",
		Message: "plan not found",
	}
	ErrInvalidPlan = &Error{
		Code:    "INVALID_PLAN",
		Message: "invalid plan data",
	}
	ErrDeviceNotFound = &Error{
		Code:    "DEVICE_NOT_FOUND",
		Message: "device not found",
	}
	ErrPlacementNotFound = &Error{
		Code:    "PLACEMENT_NOT_FOUND",
		Message: "device placement not found",
	}
	ErrUnsupportedDomain = &Error{
		Code:    "UNSUPPORTED_DOMAIN",
		Message: "unsupported device domain",
	}
	ErrInvalidPlacement = &Error{
		Code:    "INVALID_PLACEMENT",
		Message: "invalid device placement",
	}
	ErrInvalidDevice = &Error{
		Code:    "INVALID_DEVICE",
		Message: "invalid device data",
	}
	ErrActionNotAllowed = &Error{
		Code:    "ACTION_NOT_ALLOWED",
		Message: "action not permitted by whitelist",
	}
	ErrDashboardNotFound = &Error{
		Code:    "DASHBOARD_NOT_FOUND",
		Message: "dashboard not found",
	}
	ErrInvalidDashboard = &Error{
		Code:    "INVALID_DASHBOARD",
		Message: "invalid dashboard data",
	}
	ErrWidgetNotFound = &Error{
		Code:    "WIDGET_NOT_FOUND",
		Message: "widget not found",
	}
	ErrInvalidWidget = &Error{
		Code:    "INVALID_WIDGET",
		Message: "invalid widget data",
	}
	ErrAutomationNotFound = &Error{
		Code:    "AUTOMATION_NOT_FOUND",
		Message: "automation not found",
	}
	ErrAccountNotFound = &Error{
		Code:    "ACCOUNT_NOT_FOUND",
		Message: "account not found",
	}
	ErrForbidden = &Error{
		Code:    "FORBIDDEN",
		Message: "operation not permitted",
	}
	ErrInvalidRole = &Error{
		Code:    "INVALID_ROLE",
		Message: "invalid account role",
	}
	ErrInvalidLabel = &Error{
		Code:    "INVALID_LABEL",
		Message: "invalid account label",
	}
)
