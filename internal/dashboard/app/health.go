package app

import (
	"context"
	"errors"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	repohealth "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow"
	svchealth "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

// App encapsulates the dashboard core business logic and implements domain service interfaces.
type App struct {
	healthRepo repohealth.Repository
	uow        uow.UnitOfWork
	version    string
}

// Ensure App implements svchealth.HealthQueries and svchealth.HealthCommands.
var (
	_ svchealth.HealthQueries  = (*App)(nil)
	_ svchealth.HealthCommands = (*App)(nil)
)

// New initializes the application service with repositories and configuration.
func New(healthRepo repohealth.Repository, uow uow.UnitOfWork, version string) *App {
	if version == "" {
		version = "0.1.0"
	}
	return &App{
		healthRepo: healthRepo,
		uow:        uow,
		version:    version,
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
