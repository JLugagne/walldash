package uow

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans"
)

// Repositories aggregates domain repositories available within a unit of work.
type Repositories struct {
	Health     health.Repository
	Levels     levels.LevelRepository
	Plans      plans.PlanRepository
	Placements placements.DevicePlacementRepository
	Dashboards dashboards.DashboardRepository
	Widgets    dashboards.WidgetRepository
}

// UnitOfWork defines the contract for executing business logic within a single transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
}
