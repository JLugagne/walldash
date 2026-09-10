package app_test

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha/hatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_GetWeatherConfig(t *testing.T) {
	mockHA := &hatest.MockHomeAssistantRepository{
		GetConfigFunc: func(ctx context.Context) (domain.HomeConfig, error) {
			return domain.HomeConfig{
				Configured:      true,
				Latitude:        48.85,
				Longitude:       2.35,
				LocationName:    "Paris",
				TemperatureUnit: "°C",
				LengthUnit:      "km",
			}, nil
		},
	}
	service := app.New(nil, nil, nil, nil, mockHA, nil, nil, nil, "0.1.0")

	cfg, err := service.GetWeatherConfig(context.Background())
	require.NoError(t, err)
	assert.True(t, cfg.Configured)
	assert.Equal(t, "Paris", cfg.LocationName)
}
