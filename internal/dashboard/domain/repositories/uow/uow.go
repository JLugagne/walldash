package uow

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/levels"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/overviews"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/placements"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/plans"
)

// Repositories aggregates domain repositories available within a unit of work.
type Repositories struct {
	Health     health.Repository
	Levels     levels.LevelRepository
	Plans      plans.PlanRepository
	Placements placements.DevicePlacementRepository
	Overviews  overviews.OverviewRepository
	Widgets    overviews.WidgetRepository
}

// UnitOfWork defines the contract for executing business logic within a single transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
}
