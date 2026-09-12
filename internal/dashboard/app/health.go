package app

import (
	"context"
	"errors"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	repodashboards "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards"
	repoha "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha"
	repohealth "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health"
	repolevels "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels"
	repoplacements "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements"
	repoplans "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	svcactions "github.com/JLugagne/walldash/internal/dashboard/domain/service/actions"
	svcdashboards "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

// App encapsulates the dashboard core business logic and implements domain service interfaces.
type App struct {
	healthRepo     repohealth.Repository
	levelsRepo     repolevels.LevelRepository
	plansRepo      repoplans.PlanRepository
	placementsRepo repoplacements.DevicePlacementRepository
	haRepo         repoha.HomeAssistantRepository
	dashboardsRepo repodashboards.DashboardRepository
	widgetsRepo    repodashboards.WidgetRepository
	uow            uow.UnitOfWork
	broadcaster    domain.DeviceBroadcaster
	version        string
}

// Ensure App implements domain service interfaces.
var (
	_ svchealth.HealthQueries         = (*App)(nil)
	_ svchealth.HealthCommands        = (*App)(nil)
	_ svclevels.LevelQueries          = (*App)(nil)
	_ svclevels.LevelCommands         = (*App)(nil)
	_ svcdevices.DeviceQueries        = (*App)(nil)
	_ svcdevices.DeviceCommands       = (*App)(nil)
	_ svcactions.ActionCommands       = (*App)(nil)
	_ svcdashboards.DashboardQueries  = (*App)(nil)
	_ svcdashboards.DashboardCommands = (*App)(nil)
)

// SetBroadcaster configures the device state broadcaster (e.g. WebSocket hub).
func (a *App) SetBroadcaster(broadcaster domain.DeviceBroadcaster) {
	a.broadcaster = broadcaster
}

// New initializes the application service with repositories and configuration.
func New(
	healthRepo repohealth.Repository,
	levelsRepo repolevels.LevelRepository,
	plansRepo repoplans.PlanRepository,
	placementsRepo repoplacements.DevicePlacementRepository,
	haRepo repoha.HomeAssistantRepository,
	dashboardsRepo repodashboards.DashboardRepository,
	widgetsRepo repodashboards.WidgetRepository,
	uow uow.UnitOfWork,
	version string,
) *App {
	if version == "" {
		version = "0.3.0"
	}
	return &App{
		healthRepo:     healthRepo,
		levelsRepo:     levelsRepo,
		plansRepo:      plansRepo,
		placementsRepo: placementsRepo,
		haRepo:         haRepo,
		dashboardsRepo: dashboardsRepo,
		widgetsRepo:    widgetsRepo,
		uow:            uow,
		version:        version,
	}
}

// GetHealth checks repository connectivity and returns the system health status.
func (a *App) GetHealth(ctx context.Context) (domain.Health, error) {
	log := logger.LoggerFromContext(ctx)

	if err := a.healthRepo.Ping(ctx); err != nil {
		log.WithError(err).Error("health check failed: database unreachable")
		return domain.Health{}, errors.Join(domain.ErrHealthCheckFailed, err)
	}

	log.Info("health check succeeded")
	return domain.Health{
		Status:    domain.HealthStatusOK,
		Database:  "ok",
		Version:   a.version,
		Timestamp: time.Now().UTC(),
	}, nil
}
