package devicestest

import (
	"context"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDeviceQueries is a function-based mock implementation of devices.DeviceQueries.
type MockDeviceQueries struct {
	ListAvailableDevicesFunc func(ctx context.Context) ([]domain.Device, error)
	ListPlacementsFunc       func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error)
	GetPlacementFunc         func(ctx context.Context, id string) (domain.DevicePlacement, error)
}

func (m *MockDeviceQueries) ListAvailableDevices(ctx context.Context) ([]domain.Device, error) {
	if m.ListAvailableDevicesFunc == nil {
		panic("called not defined ListAvailableDevicesFunc")
	}
	return m.ListAvailableDevicesFunc(ctx)
}

func (m *MockDeviceQueries) ListPlacements(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	if m.ListPlacementsFunc == nil {
		panic("called not defined ListPlacementsFunc")
	}
	return m.ListPlacementsFunc(ctx, levelID)
}

func (m *MockDeviceQueries) GetPlacement(ctx context.Context, id string) (domain.DevicePlacement, error) {
	if m.GetPlacementFunc == nil {
		panic("called not defined GetPlacementFunc")
	}
	return m.GetPlacementFunc(ctx, id)
}

// MockDeviceCommands is a function-based mock implementation of devices.DeviceCommands.
type MockDeviceCommands struct {
	ListAvailableDevicesFunc func(ctx context.Context) ([]domain.Device, error)
	ListPlacementsFunc       func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error)
	GetPlacementFunc         func(ctx context.Context, id string) (domain.DevicePlacement, error)
	SavePlacementFunc        func(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error)
	DeletePlacementFunc      func(ctx context.Context, actor domain.Actor, levelID string, placementID string) error
}

func (m *MockDeviceCommands) ListAvailableDevices(ctx context.Context) ([]domain.Device, error) {
	if m.ListAvailableDevicesFunc == nil {
		panic("called not defined ListAvailableDevicesFunc")
	}
	return m.ListAvailableDevicesFunc(ctx)
}

func (m *MockDeviceCommands) ListPlacements(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	if m.ListPlacementsFunc == nil {
		panic("called not defined ListPlacementsFunc")
	}
	return m.ListPlacementsFunc(ctx, levelID)
}

func (m *MockDeviceCommands) GetPlacement(ctx context.Context, id string) (domain.DevicePlacement, error) {
	if m.GetPlacementFunc == nil {
		panic("called not defined GetPlacementFunc")
	}
	return m.GetPlacementFunc(ctx, id)
}

func (m *MockDeviceCommands) SavePlacement(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
	if m.SavePlacementFunc == nil {
		panic("called not defined SavePlacementFunc")
	}
	return m.SavePlacementFunc(ctx, actor, placement)
}

func (m *MockDeviceCommands) DeletePlacement(ctx context.Context, actor domain.Actor, levelID string, placementID string) error {
	if m.DeletePlacementFunc == nil {
		panic("called not defined DeletePlacementFunc")
	}
	return m.DeletePlacementFunc(ctx, actor, levelID, placementID)
}

// DeviceServiceContractTesting runs contract tests for DeviceCommands implementations.
func DeviceServiceContractTesting(t *testing.T, svc devices.DeviceCommands, testLevelID string) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-user"}

	t.Run("Contract: ListAvailableDevices returns supported devices", func(t *testing.T) {
		devs, err := svc.ListAvailableDevices(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, devs)
		for _, d := range devs {
			assert.True(t, domain.IsSupportedDomain(d.Domain))
		}
	})

	t.Run("Contract: SavePlacement creates and updates device placement", func(t *testing.T) {
		p := domain.DevicePlacement{
			LevelID:    testLevelID,
			DeviceID:   "light.living_room",
			X:          120.0,
			Y:          220.0,
			Icon:       "lightbulb",
			CustomName: "Plafonnier Salon",
		}

		created, err := svc.SavePlacement(ctx, actor, p)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, p.LevelID, created.LevelID)
		assert.Equal(t, p.DeviceID, created.DeviceID)
		assert.Equal(t, p.X, created.X)
		assert.Equal(t, p.Y, created.Y)

		// Update coordinates
		created.X = 140.0
		created.Y = 240.0
		updated, err := svc.SavePlacement(ctx, actor, created)
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, 140.0, updated.X)
		assert.Equal(t, 240.0, updated.Y)

		// Check ListPlacements contains it
		placements, err := svc.ListPlacements(ctx, testLevelID)
		require.NoError(t, err)
		assert.NotEmpty(t, placements)

		found := false
		for _, item := range placements {
			if item.ID == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Saved placement must be present in ListPlacements")

		// Clean up
		err = svc.DeletePlacement(ctx, actor, testLevelID, created.ID)
		require.NoError(t, err)
	})

	t.Run("Contract: SavePlacement returns ErrLevelNotFound for non-existent level", func(t *testing.T) {
		p := domain.DevicePlacement{
			LevelID:  "non-existent-level-id",
			DeviceID: "light.living_room",
			X:        100.0,
			Y:        100.0,
		}
		_, err := svc.SavePlacement(ctx, actor, p)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("Contract: DeletePlacement returns ErrPlacementNotFound for missing placement", func(t *testing.T) {
		err := svc.DeletePlacement(ctx, actor, testLevelID, "non-existent-placement-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlacementNotFound)
	})
}
