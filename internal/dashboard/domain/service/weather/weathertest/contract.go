package weathertest

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcweather "github.com/JLugagne/walldash/internal/dashboard/domain/service/weather"
)

// MockWeatherQueries is a function-based mock implementation of weather.WeatherQueries.
type MockWeatherQueries struct {
	GetWeatherConfigFunc func(ctx context.Context) (domain.HomeConfig, error)
}

// GetWeatherConfig delegates to GetWeatherConfigFunc.
func (m *MockWeatherQueries) GetWeatherConfig(ctx context.Context) (domain.HomeConfig, error) {
	if m.GetWeatherConfigFunc == nil {
		panic("called not defined GetWeatherConfigFunc")
	}
	return m.GetWeatherConfigFunc(ctx)
}

var _ svcweather.WeatherQueries = (*MockWeatherQueries)(nil)
