package domain_test

import (
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelValidation(t *testing.T) {
	t.Run("valid level passes validation", func(t *testing.T) {
		level := domain.Level{
			ID:        "level-1",
			Name:      "Ground Floor",
			Order:     1,
			IsOutdoor: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, level.Validate())
	})

	t.Run("empty level ID fails validation", func(t *testing.T) {
		level := domain.Level{
			ID:   "",
			Name: "Ground Floor",
		}
		err := level.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidLevel)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("empty level name fails validation", func(t *testing.T) {
		level := domain.Level{
			ID:   "level-1",
			Name: "   ",
		}
		err := level.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidLevel)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("level supports custom layers", func(t *testing.T) {
		level := domain.Level{
			ID:        "level-1",
			Name:      "Ground Floor",
			Layers:    []domain.Layer{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}, {Name: "hvac", HideGauges: false}},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, level.Validate())
		assert.Equal(t, []domain.Layer{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}, {Name: "hvac", HideGauges: false}}, level.Layers)
		assert.Equal(t, []domain.Layer{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}}, domain.DefaultLayers)
	})
}

func TestPlanValidation(t *testing.T) {
	validWall := domain.WallSegment{
		ID:        "wall-1",
		X1:        0,
		Y1:        0,
		X2:        100,
		Y2:        0,
		Thickness: 10,
	}

	validZone := domain.Zone{
		ID:    "zone-1",
		Name:  "Living Room",
		Color: "#3b82f6",
		Points: []domain.Point2D{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
			{X: 100, Y: 100},
			{X: 0, Y: 100},
		},
	}

	t.Run("valid plan passes validation", func(t *testing.T) {
		plan := domain.Plan{
			LevelID: "level-1",
			Walls:   []domain.WallSegment{validWall},
			Zones:   []domain.Zone{validZone},
		}
		require.NoError(t, plan.Validate())
	})

	t.Run("empty level ID fails validation", func(t *testing.T) {
		plan := domain.Plan{
			LevelID: "",
			Walls:   []domain.WallSegment{validWall},
		}
		err := plan.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})

	t.Run("invalid wall fails plan validation", func(t *testing.T) {
		invalidWall := validWall
		invalidWall.Thickness = -1
		plan := domain.Plan{
			LevelID: "level-1",
			Walls:   []domain.WallSegment{invalidWall},
		}
		err := plan.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})

	t.Run("zero-length wall fails validation", func(t *testing.T) {
		zeroWall := domain.WallSegment{
			ID:        "w2",
			X1:        50,
			Y1:        50,
			X2:        50,
			Y2:        50,
			Thickness: 5,
		}
		err := zeroWall.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})

	t.Run("zone with fewer than 3 points fails validation", func(t *testing.T) {
		invalidZone := domain.Zone{
			ID:    "z2",
			Name:  "Terrasse",
			Color: "#22c55e",
			Points: []domain.Point2D{
				{X: 0, Y: 0},
				{X: 10, Y: 10},
			},
		}
		err := invalidZone.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})

	t.Run("zone with valid sensors and thresholds passes validation", func(t *testing.T) {
		minTemp := 18.5
		maxTemp := 24.0
		minHum := 40.0
		maxHum := 65.0
		validSensorZone := domain.Zone{
			ID:             "z-sensor-1",
			Name:           "Chambre",
			Color:          "#3b82f6",
			Points:         validZone.Points,
			TempSensor:     "sensor.chambre_temperature",
			TempMin:        &minTemp,
			TempMax:        &maxTemp,
			HumiditySensor: "sensor.chambre_humidity",
			HumidityMin:    &minHum,
			HumidityMax:    &maxHum,
		}
		require.NoError(t, validSensorZone.Validate())
	})

	t.Run("zone with temp_min greater than temp_max fails validation", func(t *testing.T) {
		minTemp := 25.0
		maxTemp := 19.0
		invalidTempZone := domain.Zone{
			ID:      "z-temp-inv",
			Name:    "Living Room",
			Color:   "#3b82f6",
			Points:  validZone.Points,
			TempMin: &minTemp,
			TempMax: &maxTemp,
		}
		err := invalidTempZone.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("zone with temp_min equal to temp_max passes validation", func(t *testing.T) {
		temp := 21.0
		equalTempZone := domain.Zone{
			ID:      "z-temp-eq",
			Name:    "Bureau",
			Color:   "#3b82f6",
			Points:  validZone.Points,
			TempMin: &temp,
			TempMax: &temp,
		}
		require.NoError(t, equalTempZone.Validate())
	})

	t.Run("zone with humidity_min greater than humidity_max fails validation", func(t *testing.T) {
		minHum := 70.0
		maxHum := 40.0
		invalidHumZone := domain.Zone{
			ID:          "z-hum-inv",
			Name:        "Bathroom",
			Color:       "#3b82f6",
			Points:      validZone.Points,
			HumidityMin: &minHum,
			HumidityMax: &maxHum,
		}
		err := invalidHumZone.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("zone with humidity_min equal to humidity_max passes validation", func(t *testing.T) {
		hum := 50.0
		equalHumZone := domain.Zone{
			ID:          "z-hum-eq",
			Name:        "Bureau",
			Color:       "#3b82f6",
			Points:      validZone.Points,
			HumidityMin: &hum,
			HumidityMax: &hum,
		}
		require.NoError(t, equalHumZone.Validate())
	})

	t.Run("zone with only one threshold passes validation", func(t *testing.T) {
		temp := 18.0
		onlyMinZone := domain.Zone{
			ID:      "z-min-only",
			Name:    "Cave",
			Color:   "#3b82f6",
			Points:  validZone.Points,
			TempMin: &temp,
		}
		require.NoError(t, onlyMinZone.Validate())

		onlyMaxZone := domain.Zone{
			ID:      "z-max-only",
			Name:    "Grenier",
			Color:   "#3b82f6",
			Points:  validZone.Points,
			TempMax: &temp,
		}
		require.NoError(t, onlyMaxZone.Validate())

		hum := 45.0
		onlyMinHumZone := domain.Zone{
			ID:          "z-min-hum-only",
			Name:        "Cellar",
			Color:       "#3b82f6",
			Points:      validZone.Points,
			HumidityMin: &hum,
		}
		require.NoError(t, onlyMinHumZone.Validate())

		onlyMaxHumZone := domain.Zone{
			ID:          "z-max-hum-only",
			Name:        "Attic",
			Color:       "#3b82f6",
			Points:      validZone.Points,
			HumidityMax: &hum,
		}
		require.NoError(t, onlyMaxHumZone.Validate())
	})

	t.Run("zone with malformed sensor entity ID fails validation", func(t *testing.T) {
		badTempSensorZone := domain.Zone{
			ID:         "z-bad-temp",
			Name:       "Living Room",
			Color:      "#3b82f6",
			Points:     validZone.Points,
			TempSensor: "not_a_valid_entity_id",
		}
		err := badTempSensorZone.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
		assert.True(t, domain.IsDomainError(err))

		badHumSensorZone := domain.Zone{
			ID:             "z-bad-hum",
			Name:           "Living Room",
			Color:          "#3b82f6",
			Points:         validZone.Points,
			HumiditySensor: "no_dot",
		}
		err = badHumSensorZone.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("valid wall opening passes validation", func(t *testing.T) {
		validWindow := domain.WallOpening{
			ID:     "win-1",
			Type:   "window",
			Offset: 50,
			Width:  30,
		}
		require.NoError(t, validWindow.Validate())

		validDoor := domain.WallOpening{
			ID:     "door-1",
			Type:   "door",
			Offset: 20,
			Width:  25,
		}
		require.NoError(t, validDoor.Validate())
	})

	t.Run("invalid wall opening fails validation", func(t *testing.T) {
		emptyID := domain.WallOpening{
			ID:     "",
			Type:   "window",
			Offset: 10,
			Width:  20,
		}
		err := emptyID.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)

		invalidType := domain.WallOpening{
			ID:     "op-1",
			Type:   "chimney",
			Offset: 10,
			Width:  20,
		}
		err = invalidType.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)

		nonPositiveWidth := domain.WallOpening{
			ID:     "op-2",
			Type:   "door",
			Offset: 10,
			Width:  0,
		}
		err = nonPositiveWidth.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)

		negativeOffset := domain.WallOpening{
			ID:     "op-3",
			Type:   "door",
			Offset: -5,
			Width:  20,
		}
		err = negativeOffset.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})

	t.Run("wall with valid openings passes validation", func(t *testing.T) {
		wallWithOpenings := validWall
		wallWithOpenings.Openings = []domain.WallOpening{
			{
				ID:     "win-1",
				Type:   "window",
				Offset: 30,
				Width:  20,
			},
			{
				ID:        "door-1",
				Type:      "door",
				Offset:    70,
				Width:     20,
				FlipSide:  true,
				FlipHinge: true,
			},
		}
		require.NoError(t, wallWithOpenings.Validate())
		assert.True(t, wallWithOpenings.Openings[1].FlipSide)
		assert.True(t, wallWithOpenings.Openings[1].FlipHinge)
	})

	t.Run("wall with invalid opening fails validation", func(t *testing.T) {
		wallWithBadOpening := validWall
		wallWithBadOpening.Openings = []domain.WallOpening{
			{
				ID:     "",
				Type:   "window",
				Offset: 30,
				Width:  20,
			},
		}
		err := wallWithBadOpening.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlan)
	})
}
