package sqlite_test

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanRepository_ZoneSensorsAndAlertThresholds(t *testing.T) {
	ctx := context.Background()
	adapter := setupTestDB(t)

	t.Run("stores and retrieves zone sensor attributes and thresholds in zones_json", func(t *testing.T) {
		tempMin := 17.5
		tempMax := 24.0
		humMin := 35.0
		humMax := 65.0
		plan := domain.Plan{
			LevelID: "lvl-zone-sensors-test",
			Walls: []domain.WallSegment{
				{
					ID:        "w1",
					X1:        0,
					Y1:        0,
					X2:        100,
					Y2:        0,
					Thickness: 10,
				},
			},
			Zones: []domain.Zone{
				{
					ID:             "zone-with-sensors",
					Name:           "Suite Parentale",
					Color:          "#8b5cf6",
					Points:         []domain.Point2D{{X: 0, Y: 0}, {X: 50, Y: 0}, {X: 50, Y: 50}},
					TempSensor:     "sensor.bedroom_temperature",
					TempMin:        &tempMin,
					TempMax:        &tempMax,
					HumiditySensor: "sensor.bedroom_humidity",
					HumidityMin:    &humMin,
					HumidityMax:    &humMax,
				},
			},
		}

		saved, err := adapter.Save(ctx, plan)
		require.NoError(t, err)
		require.Len(t, saved.Zones, 1)
		assert.Equal(t, "sensor.bedroom_temperature", saved.Zones[0].TempSensor)
		require.NotNil(t, saved.Zones[0].TempMin)
		assert.Equal(t, tempMin, *saved.Zones[0].TempMin)
		require.NotNil(t, saved.Zones[0].TempMax)
		assert.Equal(t, tempMax, *saved.Zones[0].TempMax)
		assert.Equal(t, "sensor.bedroom_humidity", saved.Zones[0].HumiditySensor)
		require.NotNil(t, saved.Zones[0].HumidityMin)
		assert.Equal(t, humMin, *saved.Zones[0].HumidityMin)
		require.NotNil(t, saved.Zones[0].HumidityMax)
		assert.Equal(t, humMax, *saved.Zones[0].HumidityMax)

		retrieved, err := adapter.FindByLevelID(ctx, plan.LevelID)
		require.NoError(t, err)
		require.Len(t, retrieved.Zones, 1)
		rz := retrieved.Zones[0]
		assert.Equal(t, "zone-with-sensors", rz.ID)
		assert.Equal(t, "Suite Parentale", rz.Name)
		assert.Equal(t, "#8b5cf6", rz.Color)
		assert.Equal(t, "sensor.bedroom_temperature", rz.TempSensor)
		require.NotNil(t, rz.TempMin)
		assert.Equal(t, tempMin, *rz.TempMin)
		require.NotNil(t, rz.TempMax)
		assert.Equal(t, tempMax, *rz.TempMax)
		assert.Equal(t, "sensor.bedroom_humidity", rz.HumiditySensor)
		require.NotNil(t, rz.HumidityMin)
		assert.Equal(t, humMin, *rz.HumidityMin)
		require.NotNil(t, rz.HumidityMax)
		assert.Equal(t, humMax, *rz.HumidityMax)
	})

	t.Run("backward compatibility with legacy zones without sensor fields", func(t *testing.T) {
		levelID := "lvl-legacy-plan"
		legacyPlan := domain.Plan{
			LevelID: levelID,
			Walls: []domain.WallSegment{
				{ID: "w-leg", X1: 0, Y1: 0, X2: 50, Y2: 0, Thickness: 10},
			},
			Zones: []domain.Zone{
				{
					ID:     "zone-legacy",
					Name:   "Kitchen",
					Color:  "#ef4444",
					Points: []domain.Point2D{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}},
				},
			},
		}

		_, err := adapter.Save(ctx, legacyPlan)
		require.NoError(t, err)

		retrieved, err := adapter.FindByLevelID(ctx, levelID)
		require.NoError(t, err)
		require.Len(t, retrieved.Zones, 1)
		rz := retrieved.Zones[0]
		assert.Equal(t, "zone-legacy", rz.ID)
		assert.Empty(t, rz.TempSensor)
		assert.Nil(t, rz.TempMin)
		assert.Nil(t, rz.TempMax)
		assert.Empty(t, rz.HumiditySensor)
		assert.Nil(t, rz.HumidityMin)
		assert.Nil(t, rz.HumidityMax)
	})
}
