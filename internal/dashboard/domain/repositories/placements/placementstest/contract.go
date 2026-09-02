package placementstest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/placements"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDevicePlacementRepository is a function-based mock implementation of placements.DevicePlacementRepository.
type MockDevicePlacementRepository struct {
	SavePlacementFunc           func(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error)
	FindPlacementByIDFunc       func(ctx context.Context, id string) (domain.DevicePlacement, error)
	FindPlacementsByLevelIDFunc func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error)
	DeletePlacementFunc         func(ctx context.Context, id string) error
}

func (m *MockDevicePlacementRepository) SavePlacement(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
	if m.SavePlacementFunc == nil {
		panic("called not defined SavePlacementFunc")
	}
	return m.SavePlacementFunc(ctx, placement)
}

func (m *MockDevicePlacementRepository) FindPlacementByID(ctx context.Context, id string) (domain.DevicePlacement, error) {
	if m.FindPlacementByIDFunc == nil {
		panic("called not defined FindPlacementByIDFunc")
	}
	return m.FindPlacementByIDFunc(ctx, id)
}

func (m *MockDevicePlacementRepository) FindPlacementsByLevelID(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	if m.FindPlacementsByLevelIDFunc == nil {
		panic("called not defined FindPlacementsByLevelIDFunc")
	}
	return m.FindPlacementsByLevelIDFunc(ctx, levelID)
}

func (m *MockDevicePlacementRepository) DeletePlacement(ctx context.Context, id string) error {
	if m.DeletePlacementFunc == nil {
		panic("called not defined DeletePlacementFunc")
	}
	return m.DeletePlacementFunc(ctx, id)
}

// DevicePlacementRepositoryContractTesting runs contract tests for any DevicePlacementRepository implementation.
// LevelID passed in tests should refer to a valid existing level if foreign keys are enforced.
func DevicePlacementRepositoryContractTesting(t *testing.T, repo placements.DevicePlacementRepository, testLevelID string) {
	ctx := context.Background()

	t.Run("Contract: SavePlacement stores and FindPlacementByID retrieves device placement", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		p := domain.DevicePlacement{
			ID:         "placement-contract-1",
			LevelID:    testLevelID,
			DeviceID:   "light.salon",
			X:          150.0,
			Y:          200.0,
			Icon:       "lightbulb",
			CustomName: "Spot Salon",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		saved, err := repo.SavePlacement(ctx, p)
		require.NoError(t, err)
		assert.Equal(t, p.ID, saved.ID)
		assert.Equal(t, p.LevelID, saved.LevelID)
		assert.Equal(t, p.DeviceID, saved.DeviceID)
		assert.Equal(t, p.X, saved.X)
		assert.Equal(t, p.Y, saved.Y)
		assert.Equal(t, p.Icon, saved.Icon)
		assert.Equal(t, p.CustomName, saved.CustomName)

		found, err := repo.FindPlacementByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, p.ID, found.ID)
		assert.Equal(t, p.LevelID, found.LevelID)
		assert.Equal(t, p.DeviceID, found.DeviceID)
		assert.Equal(t, p.X, found.X)
		assert.Equal(t, p.Y, found.Y)
	})

	t.Run("Contract: SavePlacement updates existing placement on conflict", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		p := domain.DevicePlacement{
			ID:        "placement-contract-update",
			LevelID:   testLevelID,
			DeviceID:  "switch.tv",
			X:         100.0,
			Y:         100.0,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err := repo.SavePlacement(ctx, p)
		require.NoError(t, err)

		// Update coordinates and name
		p.X = 180.0
		p.Y = 220.0
		p.CustomName = "Prise Télévision"
		updated, err := repo.SavePlacement(ctx, p)
		require.NoError(t, err)
		assert.Equal(t, 180.0, updated.X)
		assert.Equal(t, 220.0, updated.Y)
		assert.Equal(t, "Prise Télévision", updated.CustomName)

		found, err := repo.FindPlacementByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 180.0, found.X)
		assert.Equal(t, 220.0, found.Y)
		assert.Equal(t, "Prise Télévision", found.CustomName)
	})

	t.Run("Contract: FindPlacementByID returns ErrPlacementNotFound for non-existent placement", func(t *testing.T) {
		_, err := repo.FindPlacementByID(ctx, "non-existent-placement-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlacementNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindPlacementsByLevelID returns placements for given level and empty slice when none", func(t *testing.T) {
		list, err := repo.FindPlacementsByLevelID(ctx, testLevelID)
		require.NoError(t, err)
		assert.NotEmpty(t, list)

		emptyList, err := repo.FindPlacementsByLevelID(ctx, "level-with-no-devices")
		require.NoError(t, err)
		assert.Empty(t, emptyList)
	})

	t.Run("Contract: DeletePlacement removes placement and returns ErrPlacementNotFound when deleting non-existent", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		p := domain.DevicePlacement{
			ID:        "placement-to-delete",
			LevelID:   testLevelID,
			DeviceID:  "sensor.temp",
			X:         50.0,
			Y:         50.0,
			CreatedAt: now,
			UpdatedAt: now,
		}
		_, err := repo.SavePlacement(ctx, p)
		require.NoError(t, err)

		err = repo.DeletePlacement(ctx, p.ID)
		require.NoError(t, err)

		_, err = repo.FindPlacementByID(ctx, p.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlacementNotFound)

		// Deleting already deleted or non-existent returns ErrPlacementNotFound
		err = repo.DeletePlacement(ctx, p.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlacementNotFound)
	})
}
