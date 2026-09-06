package domain

import (
	"time"
)

// HealthStatus represents the overall health state.
type HealthStatus string

const (
	HealthStatusOK       HealthStatus = "ok"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusDown     HealthStatus = "down"
)

// Health describes the current operational health of the application and its dependencies.
type Health struct {
	Status    HealthStatus
	Database  string
	Version   string
	Timestamp time.Time
}
