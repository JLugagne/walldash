package domain_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelValidation(t *testing.T) {
	t.Run("valid level passes validation", func(t *testing.T) {
		level := domain.Level{
			ID:        "level-1",
			Name:      "Rez-de-chaussée",
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
			Name: "Rez-de-chaussée",
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
		Name:  "Salon",
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
}
