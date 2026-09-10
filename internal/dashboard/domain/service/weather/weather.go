// Package weather defines read-only weather configuration queries.
package weather

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// WeatherQueries defines the weather configuration query operations.
type WeatherQueries interface {
	GetWeatherConfig(ctx context.Context) (domain.HomeConfig, error)
}
