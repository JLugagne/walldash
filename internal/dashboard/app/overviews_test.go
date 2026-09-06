package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/app"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	repoha "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha/hatest"
	repohealthtest "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	repooverviews "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/overviews"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/overviews/overviewstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow/uowtest"
	svcoverviewstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/overviews/overviewstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestOverviewApp(
	overviewsRepo repooverviews.OverviewRepository,
	widgetsRepo repooverviews.WidgetRepository,
	haRepo repoha.HomeAssistantRepository,
) *app.App {
	mockHealth := &repohealthtest.MockRepository{
		PingFunc: func(ctx context.Context) error { return nil },
	}
	mockUow := &uowtest.MockUnitOfWork{}

	return app.New(
		mockHealth,
		nil,
		nil,
		nil,
		haRepo,
		overviewsRepo,
		widgetsRepo,
		mockUow,
		"0.1.0",
	)
}

func TestOverviewServiceContract(t *testing.T) {
	mockOverviews := &overviewstest.MockOverviewRepository{
		FindAllOverviewsFunc: func(ctx context.Context) ([]domain.OverviewDashboard, error) {
			return []domain.OverviewDashboard{}, nil
		},
		FindOverviewByIDFunc: func(ctx context.Context, id string) (domain.OverviewDashboard, error) {
			return domain.OverviewDashboard{}, domain.ErrOverviewNotFound
		},
		CreateOverviewFunc: func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
			return overview, nil
		},
		UpdateOverviewFunc: func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
			if overview.ID == "non-existent" {
				return domain.OverviewDashboard{}, domain.ErrOverviewNotFound
			}
			return overview, nil
		},
		DeleteOverviewFunc: func(ctx context.Context, id string) error {
			if id == "non-existent" {
				return domain.ErrOverviewNotFound
			}
			return nil
		},
	}

	mockWidgets := &overviewstest.MockWidgetRepository{
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
	}

	mockHA := &hatest.MockHomeAssistantRepository{
		GetAutomationsFunc: func(ctx context.Context) ([]domain.Automation, error) {
			return []domain.Automation{}, nil
		},
		TriggerAutomationFunc: func(ctx context.Context, entityID string) error {
			return nil
		},
	}

	service := setupTestOverviewApp(mockOverviews, mockWidgets, mockHA)
	svcoverviewstest.OverviewCommandsContractTesting(t, service)
}

func TestOverviewService_Operations(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-1"}

	t.Run("CreateOverview generates UUID and auto-increments order", func(t *testing.T) {
		mockOverviews := &overviewstest.MockOverviewRepository{
			FindAllOverviewsFunc: func(ctx context.Context) ([]domain.OverviewDashboard, error) {
				return []domain.OverviewDashboard{
					{ID: "o-1", Name: "Existing", Order: 3},
				}, nil
			},
			CreateOverviewFunc: func(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
				return overview, nil
			},
		}

		service := setupTestOverviewApp(mockOverviews, nil, nil)
		created, err := service.CreateOverview(ctx, actor, domain.OverviewDashboard{Name: "New Dashboard"})
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, 4, created.Order)
	})

	t.Run("CreateWidget verifies parent dashboard existence", func(t *testing.T) {
		mockOverviews := &overviewstest.MockOverviewRepository{
			FindOverviewByIDFunc: func(ctx context.Context, id string) (domain.OverviewDashboard, error) {
				return domain.OverviewDashboard{}, domain.ErrOverviewNotFound
			},
		}
		mockWidgets := &overviewstest.MockWidgetRepository{}

		service := setupTestOverviewApp(mockOverviews, mockWidgets, nil)
		_, err := service.CreateWidget(ctx, actor, domain.Widget{
			DashboardID: "missing-dash",
			Title:       "Widget 1",
			Type:        domain.WidgetTypeAutomationList,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
	})

	t.Run("DeleteWidget verifies dashboardID match", func(t *testing.T) {
		mockWidgets := &overviewstest.MockWidgetRepository{
			FindWidgetByIDFunc: func(ctx context.Context, id string) (domain.Widget, error) {
				return domain.Widget{ID: id, DashboardID: "dash-A"}, nil
			},
		}

		service := setupTestOverviewApp(nil, mockWidgets, nil)
		err := service.DeleteWidget(ctx, actor, "dash-B", "widget-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("TriggerAutomation successfully delegates to HA repository", func(t *testing.T) {
		var triggeredID string
		mockHA := &hatest.MockHomeAssistantRepository{
			TriggerAutomationFunc: func(ctx context.Context, entityID string) error {
				triggeredID = entityID
				return nil
			},
		}

		service := setupTestOverviewApp(nil, nil, mockHA)
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
						Name:          "Mode Cinéma",
						State:         "on",
						Current:       0,
						LastTriggered: &now,
					},
				}, nil
			},
		}

		service := setupTestOverviewApp(nil, nil, mockHA)
		automations, err := service.ListAutomations(ctx)
		require.NoError(t, err)
		require.Len(t, automations, 1)
		assert.Equal(t, "automation.cinema", automations[0].ID)
	})
}
