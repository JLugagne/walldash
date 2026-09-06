package planstest

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPlanRepository is a function-based mock implementation of plans.PlanRepository.
type MockPlanRepository struct {
	SaveFunc            func(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	FindByLevelIDFunc   func(ctx context.Context, levelID string) (domain.Plan, error)
	DeleteByLevelIDFunc func(ctx context.Context, levelID string) error
}

func (m *MockPlanRepository) Save(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	if m.SaveFunc == nil {
		panic("called not defined SaveFunc")
	}
	return m.SaveFunc(ctx, plan)
}

func (m *MockPlanRepository) FindByLevelID(ctx context.Context, levelID string) (domain.Plan, error) {
	if m.FindByLevelIDFunc == nil {
		panic("called not defined FindByLevelIDFunc")
	}
	return m.FindByLevelIDFunc(ctx, levelID)
}

func (m *MockPlanRepository) DeleteByLevelID(ctx context.Context, levelID string) error {
	if m.DeleteByLevelIDFunc == nil {
		panic("called not defined DeleteByLevelIDFunc")
	}
	return m.DeleteByLevelIDFunc(ctx, levelID)
}

// PlanRepositoryContractTesting runs all contract tests for a PlanRepository implementation.
func PlanRepositoryContractTesting(t *testing.T, repo plans.PlanRepository) {
	ctx := context.Background()

	t.Run("Contract: Save and FindByLevelID retrieves stored plan", func(t *testing.T) {
		plan := domain.Plan{
			LevelID: "lvl-plan-contract-1",
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
					ID:    "z1",
					Name:  "Kitchen",
					Color: "#ef4444",
					Points: []domain.Point2D{
						{X: 0, Y: 0},
						{X: 50, Y: 0},
						{X: 50, Y: 50},
						{X: 0, Y: 50},
					},
					TempSensor:     "sensor.cuisine_temperature",
					HumiditySensor: "sensor.cuisine_humidity",
				},
			},
		}

		saved, err := repo.Save(ctx, plan)
		require.NoError(t, err)
		assert.Equal(t, plan.LevelID, saved.LevelID)
		assert.Equal(t, 1, len(saved.Walls))
		assert.Equal(t, 1, len(saved.Zones))

		retrieved, err := repo.FindByLevelID(ctx, plan.LevelID)
		require.NoError(t, err)
		assert.Equal(t, plan.LevelID, retrieved.LevelID)
		assert.Equal(t, 1, len(retrieved.Walls))
		assert.Equal(t, "w1", retrieved.Walls[0].ID)
		assert.Equal(t, 1, len(retrieved.Zones))
		assert.Equal(t, "z1", retrieved.Zones[0].ID)
		assert.Equal(t, 4, len(retrieved.Zones[0].Points))
		assert.Equal(t, "sensor.cuisine_temperature", retrieved.Zones[0].TempSensor)
		assert.Equal(t, "sensor.cuisine_humidity", retrieved.Zones[0].HumiditySensor)
	})

	t.Run("Contract: FindByLevelID returns ErrPlanNotFound for missing plan", func(t *testing.T) {
		_, err := repo.FindByLevelID(ctx, "non-existent-level-plan")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlanNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: Save updates an existing plan", func(t *testing.T) {
		levelID := "lvl-plan-contract-update"
		initialPlan := domain.Plan{
			LevelID: levelID,
			Walls: []domain.WallSegment{
				{ID: "w-old", X1: 0, Y1: 0, X2: 10, Y2: 10, Thickness: 5},
			},
			Zones: []domain.Zone{},
		}
		_, err := repo.Save(ctx, initialPlan)
		require.NoError(t, err)

		updatedPlan := domain.Plan{
			LevelID: levelID,
			Walls: []domain.WallSegment{
				{ID: "w-new-1", X1: 0, Y1: 0, X2: 200, Y2: 0, Thickness: 10},
				{ID: "w-new-2", X1: 200, Y1: 0, X2: 200, Y2: 200, Thickness: 10},
			},
			Zones: []domain.Zone{
				{
					ID:    "z-living",
					Name:  "Living Room",
					Color: "#10b981",
					Points: []domain.Point2D{
						{X: 0, Y: 0},
						{X: 200, Y: 0},
						{X: 200, Y: 200},
					},
				},
			},
		}
		saved, err := repo.Save(ctx, updatedPlan)
		require.NoError(t, err)
		assert.Equal(t, 2, len(saved.Walls))
		assert.Equal(t, 1, len(saved.Zones))

		found, err := repo.FindByLevelID(ctx, levelID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(found.Walls))
		assert.Equal(t, 1, len(found.Zones))
	})

	t.Run("Contract: DeleteByLevelID deletes stored plan", func(t *testing.T) {
		levelID := "lvl-plan-contract-delete"
		plan := domain.Plan{
			LevelID: levelID,
			Walls: []domain.WallSegment{
				{ID: "w-del", X1: 0, Y1: 0, X2: 10, Y2: 10, Thickness: 5},
			},
		}
		_, err := repo.Save(ctx, plan)
		require.NoError(t, err)

		err = repo.DeleteByLevelID(ctx, levelID)
		require.NoError(t, err)

		_, err = repo.FindByLevelID(ctx, levelID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	})
}
