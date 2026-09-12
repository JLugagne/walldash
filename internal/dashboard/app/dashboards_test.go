package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/app"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	repodashboards "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards/dashboardstest"
	repoha "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha/hatest"
	repohealthtest "github.com/JLugagne/walldash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow/uowtest"
	svcdashboardstest "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards/dashboardstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDashboardApp wires an App for dashboard-focused tests. unitOfWork may be nil for
// tests that never exercise a transactional path: it then defaults to a MockUnitOfWork
// with no DoFunc, which panics if App.UpdateLayout is invoked against it.
func setupTestDashboardApp(
	dashboardsRepo repodashboards.DashboardRepository,
	widgetsRepo repodashboards.WidgetRepository,
	haRepo repoha.HomeAssistantRepository,
	unitOfWork uow.UnitOfWork,
) *app.App {
	mockHealth := &repohealthtest.MockRepository{
		PingFunc: func(ctx context.Context) error { return nil },
	}
	if unitOfWork == nil {
		unitOfWork = &uowtest.MockUnitOfWork{}
	}

	return app.New(
		mockHealth,
		nil,
		nil,
		nil,
		haRepo,
		dashboardsRepo,
		widgetsRepo,
		unitOfWork,
		"0.1.0",
	)
}

func TestDashboardServiceContract(t *testing.T) {
	mockDashboards := &dashboardstest.MockDashboardRepository{
		FindAllDashboardsFunc: func(ctx context.Context) ([]domain.Dashboard, error) {
			return []domain.Dashboard{}, nil
		},
		FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
			return domain.Dashboard{}, domain.ErrDashboardNotFound
		},
		CreateDashboardFunc: func(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
			return dashboard, nil
		},
		UpdateDashboardFunc: func(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
			if dashboard.ID == "non-existent" {
				return domain.Dashboard{}, domain.ErrDashboardNotFound
			}
			return dashboard, nil
		},
		DeleteDashboardFunc: func(ctx context.Context, id string) error {
			if id == "non-existent" {
				return domain.ErrDashboardNotFound
			}
			return nil
		},
	}

	mockWidgets := &dashboardstest.MockWidgetRepository{
		CreateWidgetFunc: func(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
			return widget, nil
		},
		FindWidgetByIDFunc: func(ctx context.Context, id string) (domain.Widget, error) {
			if id == "non-existent" {
				return domain.Widget{}, domain.ErrWidgetNotFound
			}
			return domain.Widget{ID: id, DashboardID: "dashboard-id"}, nil
		},
		DeleteWidgetFunc: func(ctx context.Context, id string) error {
			if id == "non-existent" {
				return domain.ErrWidgetNotFound
			}
			return nil
		},
		DeleteWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) error {
			return nil
		},
	}

	mockHA := &hatest.MockHomeAssistantRepository{
		GetAutomationsFunc: func(ctx context.Context) ([]domain.Automation, error) {
			return []domain.Automation{}, nil
		},
		TriggerAutomationFunc: func(ctx context.Context, entityID string) error {
			return nil
		},
	}

	mockUow := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
		},
	}

	service := setupTestDashboardApp(mockDashboards, mockWidgets, mockHA, mockUow)
	svcdashboardstest.DashboardCommandsContractTesting(t, service)
}

func TestDashboardService_Operations(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}

	t.Run("CreateDashboard generates UUID and auto-increments order", func(t *testing.T) {
		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindAllDashboardsFunc: func(ctx context.Context) ([]domain.Dashboard, error) {
				return []domain.Dashboard{
					{ID: "o-1", Name: "Existing", Order: 3},
				}, nil
			},
			CreateDashboardFunc: func(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
				return dashboard, nil
			},
		}

		service := setupTestDashboardApp(mockDashboards, nil, nil, nil)
		created, err := service.CreateDashboard(ctx, actor, domain.Dashboard{Name: "New Dashboard"})
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, 4, created.Order)
	})

	t.Run("CreateWidget verifies parent dashboard existence", func(t *testing.T) {
		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				return domain.Dashboard{}, domain.ErrDashboardNotFound
			},
		}
		mockWidgets := &dashboardstest.MockWidgetRepository{}
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "missing-dash",
			Title:       "Widget 1",
			Type:        domain.WidgetTypeAutomationList,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
	})

	t.Run("DeleteWidget verifies dashboardID match", func(t *testing.T) {
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetByIDFunc: func(ctx context.Context, id string) (domain.Widget, error) {
				return domain.Widget{ID: id, DashboardID: "dash-A"}, nil
			},
		}

		service := setupTestDashboardApp(nil, mockWidgets, nil, nil)
		err := service.DeleteWidget(ctx, actor, "dash-B", "widget-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("DeleteDashboard deletes widgets and the dashboard inside a single unit of work", func(t *testing.T) {
		var calls []string
		mockWidgets := &dashboardstest.MockWidgetRepository{
			DeleteWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) error {
				calls = append(calls, "widgets")
				assert.Equal(t, "dash-1", dashboardID)
				return nil
			},
		}
		mockDashboards := &dashboardstest.MockDashboardRepository{
			DeleteDashboardFunc: func(ctx context.Context, id string) error {
				calls = append(calls, "dashboard")
				assert.Equal(t, "dash-1", id)
				return nil
			},
		}
		doCalled := false
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				doCalled = true
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
		err := service.DeleteDashboard(ctx, actor, "dash-1")
		require.NoError(t, err)
		assert.True(t, doCalled)
		assert.Equal(t, []string{"widgets", "dashboard"}, calls)
	})

	t.Run("DeleteDashboard aborts before deleting the dashboard when widget delete fails", func(t *testing.T) {
		expectedErr := errors.New("widget delete boom")
		dashboardDeleteCalled := false
		mockWidgets := &dashboardstest.MockWidgetRepository{
			DeleteWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) error {
				return expectedErr
			},
		}
		mockDashboards := &dashboardstest.MockDashboardRepository{
			DeleteDashboardFunc: func(ctx context.Context, id string) error {
				dashboardDeleteCalled = true
				return nil
			},
		}
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
		err := service.DeleteDashboard(ctx, actor, "dash-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.False(t, dashboardDeleteCalled)
	})

	t.Run("TriggerAutomation successfully delegates to HA repository", func(t *testing.T) {
		var triggeredID string
		mockHA := &hatest.MockHomeAssistantRepository{
			TriggerAutomationFunc: func(ctx context.Context, entityID string) error {
				triggeredID = entityID
				return nil
			},
		}

		service := setupTestDashboardApp(nil, nil, mockHA, nil)
		err := service.TriggerAutomation(ctx, actor, "automation.eteindre_lumieres")
		require.NoError(t, err)
		assert.Equal(t, "automation.eteindre_lumieres", triggeredID)
	})

	t.Run("ListAutomations retrieves list from HA repository", func(t *testing.T) {
		now := time.Now().UTC()
		mockHA := &hatest.MockHomeAssistantRepository{
			GetAutomationsFunc: func(ctx context.Context) ([]domain.Automation, error) {
				return []domain.Automation{
					{
						ID:            "automation.cinema",
						Name:          "Cinema Mode",
						State:         "on",
						Current:       0,
						LastTriggered: &now,
					},
				}, nil
			},
		}

		service := setupTestDashboardApp(nil, nil, mockHA, nil)
		automations, err := service.ListAutomations(ctx)
		require.NoError(t, err)
		require.Len(t, automations, 1)
		assert.Equal(t, "automation.cinema", automations[0].ID)
	})

	t.Run("UpdateWidget overlays title and config without touching position", func(t *testing.T) {
		existingWidget := domain.Widget{
			ID:          "w-1",
			DashboardID: "dash-1",
			Type:        domain.WidgetTypeSensor,
			Title:       "Old title",
			Col:         2,
			Row:         3,
			ColSpan:     1,
			RowSpan:     1,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"sensor.temp"},
				Display:   domain.DisplayNumber,
			},
		}

		var saved domain.Widget
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetByIDFunc: func(ctx context.Context, id string) (domain.Widget, error) {
				return existingWidget, nil
			},
			UpdateWidgetFunc: func(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
				saved = widget
				return widget, nil
			},
		}
		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				return domain.Dashboard{ID: "dash-1", Name: "Dash", Cols: domain.DefaultGridCols, Rows: domain.DefaultGridRows}, nil
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, nil)
		updated, err := service.UpdateWidget(ctx, actor, "dash-1", domain.Widget{
			ID:    "w-1",
			Title: "New title",
			Config: domain.WidgetConfig{
				EntityIDs: []string{"sensor.temp"},
				Display:   domain.DisplayNumber,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "New title", updated.Title)
		assert.Equal(t, 2, updated.Col)
		assert.Equal(t, 3, updated.Row)
		assert.Equal(t, "New title", saved.Title)
		assert.Equal(t, 2, saved.Col)
	})

	t.Run("UpdateWidget fails when widget does not belong to dashboard", func(t *testing.T) {
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetByIDFunc: func(ctx context.Context, id string) (domain.Widget, error) {
				return domain.Widget{ID: id, DashboardID: "dash-A"}, nil
			},
		}

		service := setupTestDashboardApp(nil, mockWidgets, nil, nil)
		_, err := service.UpdateWidget(ctx, actor, "dash-B", domain.Widget{ID: "w-1"})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("UpdateLayout validates the candidate layout before replacing positions", func(t *testing.T) {
		widgets := []domain.Widget{
			{
				ID: "w-1", DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "T1",
				Col: 0, Row: 0, ColSpan: 1, RowSpan: 1,
				Config: domain.WidgetConfig{EntityIDs: []string{"sensor.a"}, Display: domain.DisplayNumber},
			},
			{
				ID: "w-2", DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "T2",
				Col: 1, Row: 0, ColSpan: 1, RowSpan: 1,
				Config: domain.WidgetConfig{EntityIDs: []string{"sensor.b"}, Display: domain.DisplayNumber},
			},
		}

		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				return domain.Dashboard{ID: "dash-1", Name: "Dash", Cols: domain.DefaultGridCols, Rows: domain.DefaultGridRows}, nil
			},
		}
		var replaced []domain.WidgetPosition
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
				return widgets, nil
			},
			ReplaceWidgetPositionsFunc: func(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error {
				replaced = positions
				return nil
			},
		}
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
		err := service.UpdateLayout(ctx, actor, "dash-1", []domain.WidgetPosition{
			{ID: "w-1", Col: 5, Row: 5, ColSpan: 1, RowSpan: 1},
		})
		require.NoError(t, err)
		require.Len(t, replaced, 1)
		assert.Equal(t, "w-1", replaced[0].ID)
		assert.Equal(t, 5, replaced[0].Col)
	})

	t.Run("UpdateLayout rejects a layout violating the grid invariant without writing", func(t *testing.T) {
		widgets := []domain.Widget{
			{
				ID: "w-1", DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "T1",
				Col: 0, Row: 0, ColSpan: 1, RowSpan: 1,
				Config: domain.WidgetConfig{EntityIDs: []string{"sensor.a"}, Display: domain.DisplayNumber},
			},
		}

		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				return domain.Dashboard{ID: "dash-1", Name: "Dash", Cols: domain.DefaultGridCols, Rows: domain.DefaultGridRows}, nil
			},
		}
		replaceCalled := false
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
				return widgets, nil
			},
			ReplaceWidgetPositionsFunc: func(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error {
				replaceCalled = true
				return nil
			},
		}
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}

		service := setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
		err := service.UpdateLayout(ctx, actor, "dash-1", []domain.WidgetPosition{
			{ID: "w-1", Col: domain.DefaultGridCols, Row: 0, ColSpan: 1, RowSpan: 1},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
		assert.False(t, replaceCalled)
	})
}

func TestDashboardService_CreateWidgetEnforcesDomainRules(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}

	dashboard := domain.Dashboard{
		ID: "dash-1", Name: "Dash", Cols: domain.DefaultGridCols, Rows: domain.DefaultGridRows,
	}
	existing := []domain.Widget{
		{
			ID: "w-1", DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "T1",
			Col: 0, Row: 0, ColSpan: 2, RowSpan: 2,
			Config: domain.WidgetConfig{EntityIDs: []string{"sensor.a"}, Display: domain.DisplayArc, Min: new(0.0), Max: new(100.0)},
		},
	}

	newApp := func(created *domain.Widget) *app.App {
		mockDashboards := &dashboardstest.MockDashboardRepository{
			FindDashboardByIDFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				return dashboard, nil
			},
		}
		mockWidgets := &dashboardstest.MockWidgetRepository{
			FindWidgetsByDashboardIDFunc: func(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
				return existing, nil
			},
			CreateWidgetFunc: func(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
				if created != nil {
					*created = widget
				}
				return widget, nil
			},
		}
		mockUow := &uowtest.MockUnitOfWork{
			DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
				return fn(uow.Repositories{Dashboards: mockDashboards, Widgets: mockWidgets})
			},
		}
		return setupTestDashboardApp(mockDashboards, mockWidgets, nil, mockUow)
	}

	t.Run("accepts a widget with an empty title", func(t *testing.T) {
		var created domain.Widget
		service := newApp(&created)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "",
			Col: 4, Row: 0, ColSpan: 1, RowSpan: 1,
			Config: domain.WidgetConfig{EntityIDs: []string{"sensor.b"}, Display: domain.DisplayNumber},
		})
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
	})

	t.Run("rejects a widget overflowing the grid", func(t *testing.T) {
		service := newApp(nil)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "Over",
			Col: domain.DefaultGridCols - 1, Row: 0, ColSpan: 2, RowSpan: 1,
			Config: domain.WidgetConfig{EntityIDs: []string{"sensor.b"}, Display: domain.DisplayBar, Min: new(0.0), Max: new(10.0)},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("rejects a widget overlapping an existing one", func(t *testing.T) {
		service := newApp(nil)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "dash-1", Type: domain.WidgetTypeSensor, Title: "Overlap",
			Col: 1, Row: 1, ColSpan: 1, RowSpan: 1,
			Config: domain.WidgetConfig{EntityIDs: []string{"sensor.b"}, Display: domain.DisplayNumber},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDashboard)
	})

	t.Run("rejects an illegal type and display pair", func(t *testing.T) {
		service := newApp(nil)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "dash-1", Type: domain.WidgetTypeActuator, Title: "Bad",
			Col: 4, Row: 0, ColSpan: 2, RowSpan: 2,
			Config: domain.WidgetConfig{EntityIDs: []string{"light.a"}, Display: domain.DisplayArc, Min: new(0.0), Max: new(1.0)},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("rejects an actuator bound to a non actionable domain", func(t *testing.T) {
		service := newApp(nil)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "dash-1", Type: domain.WidgetTypeActuator, Title: "Bad",
			Col: 4, Row: 0, ColSpan: 1, RowSpan: 1,
			Config: domain.WidgetConfig{EntityIDs: []string{"sensor.a"}, Display: domain.DisplayToggle},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})
}
