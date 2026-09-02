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
	GetStatesFunc         func(ctx context.Context) ([]domain.Device, error)
	GetStateFunc          func(ctx context.Context, entityID string) (domain.Device, error)
	CallServiceFunc       func(ctx context.Context, domain string, service string, entityID string) error
	GetAutomationsFunc    func(ctx context.Context) ([]domain.Automation, error)
	GetAutomationFunc     func(ctx context.Context, entityID string) (domain.Automation, error)
	TriggerAutomationFunc func(ctx context.Context, entityID string) error
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

func (m *MockHomeAssistantRepository) GetAutomations(ctx context.Context) ([]domain.Automation, error) {
	if m.GetAutomationsFunc == nil {
		panic("called not defined GetAutomationsFunc")
	}
	return m.GetAutomationsFunc(ctx)
}

func (m *MockHomeAssistantRepository) GetAutomation(ctx context.Context, entityID string) (domain.Automation, error) {
	if m.GetAutomationFunc == nil {
		panic("called not defined GetAutomationFunc")
	}
	return m.GetAutomationFunc(ctx, entityID)
}

func (m *MockHomeAssistantRepository) TriggerAutomation(ctx context.Context, entityID string) error {
	if m.TriggerAutomationFunc == nil {
		panic("called not defined TriggerAutomationFunc")
	}
	return m.TriggerAutomationFunc(ctx, entityID)
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

	t.Run("Contract: GetAutomations returns list of automations", func(t *testing.T) {
		automations, err := repo.GetAutomations(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, automations)

		for _, a := range automations {
			assert.NotEmpty(t, a.ID)
			assert.Contains(t, a.ID, "automation.")
			assert.NotEmpty(t, a.Name)
		}
	})

	t.Run("Contract: GetAutomation returns single automation or ErrAutomationNotFound", func(t *testing.T) {
		automations, err := repo.GetAutomations(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, automations)

		firstID := automations[0].ID
		a, err := repo.GetAutomation(ctx, firstID)
		require.NoError(t, err)
		assert.Equal(t, firstID, a.ID)

		_, err = repo.GetAutomation(ctx, "automation.unknown_automation_id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAutomationNotFound)
	})

	t.Run("Contract: TriggerAutomation executes successfully", func(t *testing.T) {
		automations, err := repo.GetAutomations(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, automations)

		firstID := automations[0].ID
		err = repo.TriggerAutomation(ctx, firstID)
		require.NoError(t, err)
	})
}
