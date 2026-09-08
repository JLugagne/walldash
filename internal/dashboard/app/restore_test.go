package app_test

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	repohealthtest "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health/healthtest"
	repolevelstest "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels/levelstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/overviews/overviewstest"
	repoplacementstest "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/placements/placementstest"
	repoplanstest "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/plans/planstest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow/uowtest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_RestoreBackup(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}
	input := restore.RestoreInput{
		Levels: []domain.Level{
			{ID: "lvl-1", Name: "Ground floor", Order: 1, Layers: []domain.Layer{{Name: "controls"}}},
		},
		Plans: []domain.Plan{
			{LevelID: "lvl-1", Walls: []domain.WallSegment{
				{ID: "w1", X1: 0, Y1: 0, X2: 100, Y2: 0, Thickness: 4},
			}},
		},
		Placements: []domain.DevicePlacement{
			{ID: "p1", LevelID: "lvl-1", DeviceID: "light.kitchen", X: 10, Y: 10, Layer: "controls"},
		},
		Overviews: []domain.OverviewDashboard{
			{ID: "ov-1", Name: "Main", Order: 1, Cols: 4, Rows: 4, Widgets: []domain.Widget{
				{ID: "w1", DashboardID: "ov-1", Type: domain.WidgetTypeActuator, Title: "Kitchen", Order: 1, Col: 0, Row: 0, ColSpan: 1, RowSpan: 1, Config: domain.WidgetConfig{EntityIDs: []string{"light.kitchen"}, Display: domain.DisplayToggle}},
			}},
		},
		IncludeDevices: true,
	}
	var createdLevels int
	var createdPlacements int
	var createdWidgets int
	mockLevels := &repolevelstest.MockLevelRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.Level, error) {
			return []domain.Level{}, nil
		},
		CreateFunc: func(ctx context.Context, level domain.Level) (domain.Level, error) {
			createdLevels++
			assert.Equal(t, "lvl-1", level.ID)
			return level, nil
		},
	}
	mockPlans := &repoplanstest.MockPlanRepository{
		SaveFunc: func(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
			return plan, nil
		},
	}
	mockPlacements := &repoplacementstest.MockDevicePlacementRepository{
		FindPlacementsByLevelIDFunc: func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
			return []domain.DevicePlacement{}, nil
		},
		SavePlacementFunc: func(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
			createdPlacements++
			return placement, nil
		},
	}
	mockOverviews := &overviewstest.MockOverviewRepository{
		FindAllOverviewsFunc: func(ctx context.Context) ([]domain.OverviewDashboard, error) {
			return []domain.OverviewDashboard{}, nil
		},
		CreateOverviewFunc: func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
			return overview, nil
		},
	}
	mockWidgets := &overviewstest.MockWidgetRepository{
		DeleteWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) error {
			return nil
		},
		CreateWidgetFunc: func(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
			createdWidgets++
			return widget, nil
		},
	}
	mockHealth := &repohealthtest.MockRepository{}
	txRepos := uow.Repositories{
		Levels:     mockLevels,
		Plans:      mockPlans,
		Placements: mockPlacements,
		Overviews:  mockOverviews,
		Widgets:    mockWidgets,
	}
	mockUow := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(txRepos)
		},
	}
	service := app.New(mockHealth, mockLevels, mockPlans, mockPlacements, nil, mockOverviews, mockWidgets, mockUow, "0.3.0")
	summary, err := service.RestoreBackup(ctx, actor, input)
	require.NoError(t, err)
	assert.Equal(t, 1, summary.Levels)
	assert.Equal(t, 1, summary.Plans)
	assert.Equal(t, 1, summary.Placements)
	assert.Equal(t, 1, summary.Overviews)
	assert.Equal(t, 1, summary.Widgets)
	assert.Equal(t, 1, createdLevels)
	assert.Equal(t, 1, createdPlacements)
	assert.Equal(t, 1, createdWidgets)
}

func TestApp_RestoreBackupSkipsDevicesWhenOptedOut(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}
	input := restore.RestoreInput{
		Levels: []domain.Level{
			{ID: "lvl-1", Name: "Ground floor", Order: 1, Layers: []domain.Layer{{Name: "controls"}}},
		},
		Placements: []domain.DevicePlacement{
			{ID: "p1", LevelID: "lvl-1", DeviceID: "light.kitchen", X: 10, Y: 10, Layer: "controls"},
		},
		IncludeDevices: false,
	}
	saves := 0
	mockLevels := &repolevelstest.MockLevelRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.Level, error) {
			return []domain.Level{}, nil
		},
		CreateFunc: func(ctx context.Context, level domain.Level) (domain.Level, error) {
			return level, nil
		},
	}
	mockPlans := &repoplanstest.MockPlanRepository{}
	mockPlacements := &repoplacementstest.MockDevicePlacementRepository{
		FindPlacementsByLevelIDFunc: func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
			return []domain.DevicePlacement{}, nil
		},
		SavePlacementFunc: func(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
			saves++
			return placement, nil
		},
	}
	mockOverviews := &overviewstest.MockOverviewRepository{
		FindAllOverviewsFunc: func(ctx context.Context) ([]domain.OverviewDashboard, error) {
			return []domain.OverviewDashboard{}, nil
		},
	}
	mockWidgets := &overviewstest.MockWidgetRepository{
		DeleteWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) error {
			return nil
		},
	}
	mockHealth := &repohealthtest.MockRepository{}
	txRepos := uow.Repositories{
		Levels:     mockLevels,
		Plans:      mockPlans,
		Placements: mockPlacements,
		Overviews:  mockOverviews,
		Widgets:    mockWidgets,
	}
	mockUow := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(txRepos)
		},
	}
	service := app.New(mockHealth, mockLevels, mockPlans, mockPlacements, nil, mockOverviews, mockWidgets, mockUow, "0.3.0")
	summary, err := service.RestoreBackup(ctx, actor, input)
	require.NoError(t, err)
	assert.Equal(t, 0, summary.Placements)
	assert.Equal(t, 0, saves)
}

func TestApp_RestoreBackupRejectsInvalidSnapshot(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}
	input := restore.RestoreInput{
		Levels: []domain.Level{{ID: "lvl-1", Name: "", Order: 1}},
	}
	calls := 0
	mockUow := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			calls++
			return fn(uow.Repositories{})
		},
	}
	mockHealth := &repohealthtest.MockRepository{}
	service := app.New(mockHealth, nil, nil, nil, nil, nil, nil, mockUow, "0.3.0")
	_, err := service.RestoreBackup(ctx, actor, input)
	require.Error(t, err)
	assert.Equal(t, 0, calls, "invalid snapshot must fail before touching storage")
}
