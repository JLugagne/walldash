package dashboard

import (
	"time"

	"github.com/JLugagne/walldash/domain"
)

// HealthResponse represents the response format for the health query API.
type HealthResponse struct {
	Status    string    `json:"status" validate:"required,oneof=ok degraded down"`
	Database  string    `json:"database" validate:"required"`
	Version   string    `json:"version" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

// Validation errors for dashboard public types
var (
	ErrInvalidHealthResponse = &domain.Error{
		Code:    "INVALID_HEALTH_RESPONSE",
		Message: "invalid health response data",
	}
)
