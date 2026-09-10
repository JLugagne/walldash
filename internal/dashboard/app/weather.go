package app

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcweather "github.com/JLugagne/walldash/internal/dashboard/domain/service/weather"
)

var _ svcweather.WeatherQueries = (*App)(nil)

// GetWeatherConfig returns the Home Assistant location and unit configuration.
func (a *App) GetWeatherConfig(ctx context.Context) (domain.HomeConfig, error) {
	return a.haRepo.GetConfig(ctx)
}
