package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/app"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha/hatest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/levels/levelstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/placements/placementstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/plans/planstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow/uowtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/actions/actionstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBroadcaster struct {
	mu        sync.Mutex
	broadcast []domain.Device
}

func (m *mockBroadcaster) BroadcastDevice(device domain.Device) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcast = append(m.broadcast, device)
}

func setupTestAppWithHA(mockHA *hatest.MockHomeAssistantRepository) *app.App {
	mockHealth := &healthtest.MockRepository{
		PingFunc: func(ctx context.Context) error { return nil },
	}
	mockLevels := &levelstest.MockLevelRepository{}
	mockPlans := &planstest.MockPlanRepository{}
	mockPlacements := &placementstest.MockDevicePlacementRepository{}
	mockUOW := &uowtest.MockUnitOfWork{}

	return app.New(
		mockHealth,
		mockLevels,
		mockPlans,
		mockPlacements,
		mockHA,
		mockUOW,
		"0.1.0",
	)
}

func TestActionsContract(t *testing.T) {
	deviceState := domain.Device{
		ID:     "light.living_room",
		Name:   "Living Room Light",
		Domain: domain.DomainLight,
		State:  "on",
	}

	mockHA := &hatest.MockHomeAssistantRepository{
		CallServiceFunc: func(ctx context.Context, domainStr string, service string, entityID string) error {
			if service == "toggle" {
				if deviceState.State == "on" {
					deviceState.State = "off"
				} else {
					deviceState.State = "on"
				}
			}
			return nil
		},
		GetStateFunc: func(ctx context.Context, entityID string) (domain.Device, error) {
			if entityID == "light.living_room" {
				return deviceState, nil
			}
			return domain.Device{}, domain.ErrDeviceNotFound
		},
	}

	testApp := setupTestAppWithHA(mockHA)
	actionstest.ActionCommandsContractTesting(t, testApp, "light.living_room")
}

func TestExecuteAction(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-user"}

	t.Run("ExecuteAction strictly enforces whitelist", func(t *testing.T) {
		haCalled := false
		mockHA := &hatest.MockHomeAssistantRepository{
			CallServiceFunc: func(ctx context.Context, domainStr string, service string, entityID string) error {
				haCalled = true
				return nil
			},
		}
		testApp := setupTestAppWithHA(mockHA)

		// Disallowed action
		err := testApp.ExecuteAction(ctx, actor, domain.ActionCommand{
			EntityID: "light.living_room",
			Action:   "set_brightness",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.False(t, haCalled)

		// Disallowed domain
		err = testApp.ExecuteAction(ctx, actor, domain.ActionCommand{
			EntityID: "sensor.temperature",
			Action:   "toggle",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.False(t, haCalled)
	})

	t.Run("ExecuteAction delegates to HA and notifies broadcaster", func(t *testing.T) {
		var calledDomain, calledService, calledEntity string
		mockHA := &hatest.MockHomeAssistantRepository{
			CallServiceFunc: func(ctx context.Context, domainStr string, service string, entityID string) error {
				calledDomain = domainStr
				calledService = service
				calledEntity = entityID
				return nil
			},
			GetStateFunc: func(ctx context.Context, entityID string) (domain.Device, error) {
				return domain.Device{
					ID:     entityID,
					Name:   "Plafonnier Salon",
					Domain: domain.DomainLight,
					State:  "off",
				}, nil
			},
		}

		testApp := setupTestAppWithHA(mockHA)
		broadcaster := &mockBroadcaster{}
		testApp.SetBroadcaster(broadcaster)

		err := testApp.ExecuteAction(ctx, actor, domain.ActionCommand{
			EntityID: "light.salon_plafond",
			Action:   "toggle",
		})
		require.NoError(t, err)
		assert.Equal(t, "light", calledDomain)
		assert.Equal(t, "toggle", calledService)
		assert.Equal(t, "light.salon_plafond", calledEntity)

		require.Len(t, broadcaster.broadcast, 1)
		assert.Equal(t, "light.salon_plafond", broadcaster.broadcast[0].ID)
		assert.Equal(t, "off", broadcaster.broadcast[0].State)
	})

	t.Run("ExecuteAction propagates HA errors", func(t *testing.T) {
		mockHA := &hatest.MockHomeAssistantRepository{
			CallServiceFunc: func(ctx context.Context, domainStr string, service string, entityID string) error {
				return errors.New("ha connection failed")
			},
		}

		testApp := setupTestAppWithHA(mockHA)
		err := testApp.ExecuteAction(ctx, actor, domain.ActionCommand{
			EntityID: "switch.coffee_maker",
			Action:   "turn_on",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ha connection failed")
	})
}
