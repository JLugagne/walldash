package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha/hatest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels/levelstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements/placementstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans/planstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow/uowtest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/devices/devicestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAppWithMocks() (
	*app.App,
	*levelstest.MockLevelRepository,
	*placementstest.MockDevicePlacementRepository,
	*hatest.MockHomeAssistantRepository,
) {
	healthRepo := &healthtest.MockRepository{
		PingFunc: func(ctx context.Context) error { return nil },
	}
	levelsRepo := &levelstest.MockLevelRepository{}
	plansRepo := &planstest.MockPlanRepository{}
	placementsRepo := &placementstest.MockDevicePlacementRepository{}
	haRepo := &hatest.MockHomeAssistantRepository{}
	mockUOW := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(uow.Repositories{
				Health:     healthRepo,
				Levels:     levelsRepo,
				Plans:      plansRepo,
				Placements: placementsRepo,
			})
		},
	}

	application := app.New(
		healthRepo,
		levelsRepo,
		plansRepo,
		placementsRepo,
		haRepo,
		nil,
		nil,
		mockUOW,
		"1.0.0",
	)

	return application, levelsRepo, placementsRepo, haRepo
}

func TestDeviceServiceContract(t *testing.T) {
	// In-memory mock storage for contract testing
	levelsStore := map[string]domain.Level{
		"level-contract": {
			ID:   "level-contract",
			Name: "Ground Floor",
		},
	}
	placementsStore := map[string]domain.DevicePlacement{}

	levelsRepo := &levelstest.MockLevelRepository{
		FindByIDFunc: func(ctx context.Context, id string) (domain.Level, error) {
			if l, ok := levelsStore[id]; ok {
				return l, nil
			}
			return domain.Level{}, domain.ErrLevelNotFound
		},
	}

	placementsRepo := &placementstest.MockDevicePlacementRepository{
		SavePlacementFunc: func(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
			placementsStore[placement.ID] = placement
			return placement, nil
		},
		FindPlacementByIDFunc: func(ctx context.Context, id string) (domain.DevicePlacement, error) {
			if p, ok := placementsStore[id]; ok {
				return p, nil
			}
			return domain.DevicePlacement{}, domain.ErrPlacementNotFound
		},
		FindPlacementsByLevelIDFunc: func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
			var res []domain.DevicePlacement
			for _, p := range placementsStore {
				if p.LevelID == levelID {
					res = append(res, p)
				}
			}
			return res, nil
		},
		DeletePlacementFunc: func(ctx context.Context, id string) error {
			if _, ok := placementsStore[id]; !ok {
				return domain.ErrPlacementNotFound
			}
			delete(placementsStore, id)
			return nil
		},
	}

	haRepo := &hatest.MockHomeAssistantRepository{
		GetStatesFunc: func(ctx context.Context) ([]domain.Device, error) {
			return []domain.Device{
				{
					ID:     "light.salon",
					Name:   "Living Room Ceiling Light",
					Domain: domain.DomainLight,
					State:  "on",
				},
			}, nil
		},
		GetStateFunc: func(ctx context.Context, entityID string) (domain.Device, error) {
			if entityID == "light.salon" {
				return domain.Device{
					ID:     "light.salon",
					Name:   "Living Room Ceiling Light",
					Domain: domain.DomainLight,
					State:  "on",
				}, nil
			}
			return domain.Device{}, domain.ErrDeviceNotFound
		},
	}

	mockUOW := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(uow.Repositories{
				Levels:     levelsRepo,
				Placements: placementsRepo,
			})
		},
	}

	application := app.New(
		nil,
		levelsRepo,
		nil,
		placementsRepo,
		haRepo,
		nil,
		nil,
		mockUOW,
		"1.0.0",
	)

	devicestest.DeviceServiceContractTesting(t, application, "level-contract")
}

func TestApp_Devices(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin"}

	t.Run("ListAvailableDevices returns devices from repository", func(t *testing.T) {
		application, _, _, haRepo := setupTestAppWithMocks()
		haRepo.GetStatesFunc = func(ctx context.Context) ([]domain.Device, error) {
			return []domain.Device{
				{ID: "light.1", Name: "L1", Domain: domain.DomainLight},
			}, nil
		}

		devs, err := application.ListAvailableDevices(ctx)
		require.NoError(t, err)
		assert.Len(t, devs, 1)
		assert.Equal(t, "light.1", devs[0].ID)
	})

	t.Run("ListAvailableDevices propagates error", func(t *testing.T) {
		application, _, _, haRepo := setupTestAppWithMocks()
		haRepo.GetStatesFunc = func(ctx context.Context) ([]domain.Device, error) {
			return nil, errors.New("ha connection failed")
		}

		_, err := application.ListAvailableDevices(ctx)
		require.Error(t, err)
	})

	t.Run("ListPlacements checks level existence", func(t *testing.T) {
		application, levelsRepo, _, _ := setupTestAppWithMocks()
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{}, domain.ErrLevelNotFound
		}

		_, err := application.ListPlacements(ctx, "non-existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("SavePlacement returns ErrLevelNotFound when level does not exist", func(t *testing.T) {
		application, levelsRepo, _, _ := setupTestAppWithMocks()
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{}, domain.ErrLevelNotFound
		}

		_, err := application.SavePlacement(ctx, actor, domain.DevicePlacement{
			LevelID:  "missing-level",
			DeviceID: "light.1",
			X:        10,
			Y:        10,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("SavePlacement auto-generates ID if empty", func(t *testing.T) {
		application, levelsRepo, placementsRepo, _ := setupTestAppWithMocks()
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{ID: id}, nil
		}
		placementsRepo.SavePlacementFunc = func(ctx context.Context, p domain.DevicePlacement) (domain.DevicePlacement, error) {
			return p, nil
		}

		saved, err := application.SavePlacement(ctx, actor, domain.DevicePlacement{
			LevelID:  "lvl-1",
			DeviceID: "light.1",
			X:        50,
			Y:        60,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, saved.ID)
		assert.Equal(t, 50.0, saved.X)
		assert.Equal(t, "controls", saved.Layer)
	})

	t.Run("SavePlacement auto-appends new layer to level layers if not present", func(t *testing.T) {
		application, levelsRepo, placementsRepo, _ := setupTestAppWithMocks()
		levelObj := domain.Level{
			ID:     "lvl-1",
			Layers: []string{"controls", "sensors"},
		}
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return levelObj, nil
		}
		var updatedLevel domain.Level
		levelsRepo.UpdateFunc = func(ctx context.Context, l domain.Level) (domain.Level, error) {
			updatedLevel = l
			return l, nil
		}
		placementsRepo.SavePlacementFunc = func(ctx context.Context, p domain.DevicePlacement) (domain.DevicePlacement, error) {
			return p, nil
		}

		saved, err := application.SavePlacement(ctx, actor, domain.DevicePlacement{
			LevelID:  "lvl-1",
			DeviceID: "light.1",
			X:        50,
			Y:        60,
			Layer:    "hvac",
		})
		require.NoError(t, err)
		assert.Equal(t, "hvac", saved.Layer)
		assert.Contains(t, updatedLevel.Layers, "hvac")
	})

	t.Run("DeletePlacement returns ErrPlacementNotFound when placement belongs to another level", func(t *testing.T) {
		application, levelsRepo, placementsRepo, _ := setupTestAppWithMocks()
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{ID: id}, nil
		}
		placementsRepo.FindPlacementByIDFunc = func(ctx context.Context, id string) (domain.DevicePlacement, error) {
			return domain.DevicePlacement{
				ID:      id,
				LevelID: "different-level-id",
			}, nil
		}

		err := application.DeletePlacement(ctx, actor, "lvl-1", "p-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlacementNotFound)
	})

	t.Run("DeletePlacement deletes when valid", func(t *testing.T) {
		application, levelsRepo, placementsRepo, _ := setupTestAppWithMocks()
		levelsRepo.FindByIDFunc = func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{ID: id}, nil
		}
		placementsRepo.FindPlacementByIDFunc = func(ctx context.Context, id string) (domain.DevicePlacement, error) {
			return domain.DevicePlacement{
				ID:      id,
				LevelID: "lvl-1",
			}, nil
		}
		deleted := false
		placementsRepo.DeletePlacementFunc = func(ctx context.Context, id string) error {
			deleted = true
			return nil
		}

		err := application.DeletePlacement(ctx, actor, "lvl-1", "p-1")
		require.NoError(t, err)
		assert.True(t, deleted)
	})
}
