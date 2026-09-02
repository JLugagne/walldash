package hatest

import (
	"context"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHomeAssistantRepository is a function-based mock implementation of ha.HomeAssistantRepository.
type MockHomeAssistantRepository struct {
	GetStatesFunc   func(ctx context.Context) ([]domain.Device, error)
	GetStateFunc    func(ctx context.Context, entityID string) (domain.Device, error)
	CallServiceFunc func(ctx context.Context, domain string, service string, entityID string) error
}

func (m *MockHomeAssistantRepository) GetStates(ctx context.Context) ([]domain.Device, error) {
	if m.GetStatesFunc == nil {
		panic("called not defined GetStatesFunc")
	}
	return m.GetStatesFunc(ctx)
}

func (m *MockHomeAssistantRepository) GetState(ctx context.Context, entityID string) (domain.Device, error) {
	if m.GetStateFunc == nil {
		panic("called not defined GetStateFunc")
	}
	return m.GetStateFunc(ctx, entityID)
}

func (m *MockHomeAssistantRepository) CallService(ctx context.Context, domain string, service string, entityID string) error {
	if m.CallServiceFunc == nil {
		panic("called not defined CallServiceFunc")
	}
	return m.CallServiceFunc(ctx, domain, service, entityID)
}

// HomeAssistantRepositoryContractTesting runs contract tests for HomeAssistantRepository implementations.
func HomeAssistantRepositoryContractTesting(t *testing.T, repo ha.HomeAssistantRepository) {
	ctx := context.Background()

	t.Run("Contract: GetStates returns list of devices with only supported domains", func(t *testing.T) {
		devices, err := repo.GetStates(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, devices)

		for _, d := range devices {
			assert.NotEmpty(t, d.ID)
			assert.NotEmpty(t, d.Name)
			assert.True(t, domain.IsSupportedDomain(d.Domain), "Device domain must be supported: "+d.Domain)
		}
	})

	t.Run("Contract: GetState returns device for existing entity or ErrDeviceNotFound", func(t *testing.T) {
		devices, err := repo.GetStates(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, devices)

		existingID := devices[0].ID
		dev, err := repo.GetState(ctx, existingID)
		require.NoError(t, err)
		assert.Equal(t, existingID, dev.ID)

		_, err = repo.GetState(ctx, "non_existent_entity_domain.does_not_exist")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDeviceNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: CallService toggles or executes service for entity", func(t *testing.T) {
		err := repo.CallService(ctx, "light", "toggle", "light.salon_plafond")
		require.NoError(t, err)
	})
}
