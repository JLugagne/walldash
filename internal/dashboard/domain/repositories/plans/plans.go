package plans

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// PlanRepository defines repository operations for Plan domain entities.
type PlanRepository interface {
	Save(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	FindByLevelID(ctx context.Context, levelID string) (domain.Plan, error)
	DeleteByLevelID(ctx context.Context, levelID string) error
}
