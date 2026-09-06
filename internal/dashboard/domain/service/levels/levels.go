package levels

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// LevelQueries defines read-only operations for levels and plans.
type LevelQueries interface {
	ListLevels(ctx context.Context) ([]domain.Level, error)
	GetLevel(ctx context.Context, id string) (domain.Level, error)
	GetPlan(ctx context.Context, levelID string) (domain.Plan, error)
}

// LevelCommands defines mutating operations for levels and plans, extending LevelQueries.
type LevelCommands interface {
	LevelQueries
	CreateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error)
	UpdateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error)
	DeleteLevel(ctx context.Context, actor domain.Actor, id string) error
	ReorderLevels(ctx context.Context, actor domain.Actor, orderedIDs []string) error
	SavePlan(ctx context.Context, actor domain.Actor, plan domain.Plan) (domain.Plan, error)
}
