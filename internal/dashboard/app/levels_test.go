package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/app"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	repohealthtest "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	repolevelstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/levels/levelstest"
	repoplanstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/plans/planstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow/uowtest"
	svclevelstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels/levelstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelServiceContract(t *testing.T) {
	mockLevels := &repolevelstest.MockLevelRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.Level, error) {
			return []domain.Level{}, nil
		},
		FindByIDFunc: func(ctx context.Context, id string) (domain.Level, error) {
			return domain.Level{}, domain.ErrLevelNotFound
		},
	}
	mockPlans := &repoplanstest.MockPlanRepository{
		FindByLevelIDFunc: func(ctx context.Context, levelID string) (domain.Plan, error) {
			return domain.Plan{}, domain.ErrPlanNotFound
		},
	}
	mockHealth := &repohealthtest.MockRepository{}
	mockUow := &uowtest.MockUnitOfWork{}

	service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
	svclevelstest.LevelQueriesContractTesting(t, service)
}

func TestLevelService_Operations(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}

	t.Run("CreateLevel succeeds and assigns generated ID", func(t *testing.T) {
		mockLevels := &repolevelstest.MockLevelRepository{
			FindAllFunc: func(ctx context.Context) ([]domain.Level, error) {
				return []domain.Level{}, nil
			},
			CreateFunc: func(ctx context.Context, level domain.Level) (domain.Level, error) {
				assert.NotEmpty(t, level.ID)
				assert.Equal(t, "Étage 1", level.Name)
				return level, nil
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		created, err := service.CreateLevel(ctx, actor, domain.Level{Name: "Étage 1"})
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "Étage 1", created.Name)
	})

	t.Run("CreateLevel rejects empty name", func(t *testing.T) {
		mockLevels := &repolevelstest.MockLevelRepository{}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		_, err := service.CreateLevel(ctx, actor, domain.Level{Name: ""})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidLevel)
	})

	t.Run("UpdateLevel validates and updates existing level", func(t *testing.T) {
		mockLevels := &repolevelstest.MockLevelRepository{
			UpdateFunc: func(ctx context.Context, level domain.Level) (domain.Level, error) {
				return level, nil
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		lvl := domain.Level{
			ID:        "lvl-1",
			Name:      "RDC Renommé",
			Order:     0,
			CreatedAt: time.Now(),
		}
		updated, err := service.UpdateLevel(ctx, actor, lvl)
		require.NoError(t, err)
		assert.Equal(t, "RDC Renommé", updated.Name)
	})

	t.Run("DeleteLevel deletes level", func(t *testing.T) {
		deletedID := ""
		mockLevels := &repolevelstest.MockLevelRepository{
			DeleteFunc: func(ctx context.Context, id string) error {
				deletedID = id
				return nil
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		err := service.DeleteLevel(ctx, actor, "lvl-123")
		require.NoError(t, err)
		assert.Equal(t, "lvl-123", deletedID)
	})

	t.Run("ReorderLevels reorders levels", func(t *testing.T) {
		var ordered []string
		mockLevels := &repolevelstest.MockLevelRepository{
			ReorderFunc: func(ctx context.Context, orderedIDs []string) error {
				ordered = orderedIDs
				return nil
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		err := service.ReorderLevels(ctx, actor, []string{"lvl-2", "lvl-1"})
		require.NoError(t, err)
		assert.Equal(t, []string{"lvl-2", "lvl-1"}, ordered)
	})

	t.Run("SavePlan fails if level does not exist", func(t *testing.T) {
		mockLevels := &repolevelstest.MockLevelRepository{
			FindByIDFunc: func(ctx context.Context, id string) (domain.Level, error) {
				return domain.Level{}, domain.ErrLevelNotFound
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		_, err := service.SavePlan(ctx, actor, domain.Plan{
			LevelID: "lvl-missing",
			Walls:   []domain.WallSegment{},
			Zones:   []domain.Zone{},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("SavePlan saves valid plan when level exists", func(t *testing.T) {
		mockLevels := &repolevelstest.MockLevelRepository{
			FindByIDFunc: func(ctx context.Context, id string) (domain.Level, error) {
				return domain.Level{ID: id, Name: "RDC"}, nil
			},
		}
		mockPlans := &repoplanstest.MockPlanRepository{
			SaveFunc: func(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
				return plan, nil
			},
		}
		mockHealth := &repohealthtest.MockRepository{}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockHealth, mockLevels, mockPlans, mockUow, "0.1.0")
		plan := domain.Plan{
			LevelID: "lvl-exists",
			Walls: []domain.WallSegment{
				{ID: "w1", X1: 0, Y1: 0, X2: 100, Y2: 0, Thickness: 10},
			},
			Zones: []domain.Zone{},
		}
		saved, err := service.SavePlan(ctx, actor, plan)
		require.NoError(t, err)
		assert.Equal(t, "lvl-exists", saved.LevelID)
		assert.Equal(t, 1, len(saved.Walls))
	})
}
