package health

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// HealthQueries defines query operations for application health inspection.
type HealthQueries interface {
	GetHealth(ctx context.Context) (domain.Health, error)
}

// HealthCommands defines command operations for application health management, extending queries.
type HealthCommands interface {
	HealthQueries
}
