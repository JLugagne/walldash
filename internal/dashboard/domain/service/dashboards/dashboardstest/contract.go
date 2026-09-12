package dashboardstest

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDashboardQueries is a function-based mock implementation of dashboards.DashboardQueries.
type MockDashboardQueries struct {
	ListDashboardsFunc  func(ctx context.Context) ([]domain.Dashboard, error)
	GetDashboardFunc    func(ctx context.Context, id string) (domain.Dashboard, error)
	ListAutomationsFunc func(ctx context.Context) ([]domain.Automation, error)
}

func (m *MockDashboardQueries) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	if m.ListDashboardsFunc == nil {
		panic("called not defined ListDashboardsFunc")
	}
	return m.ListDashboardsFunc(ctx)
}

func (m *MockDashboardQueries) GetDashboard(ctx context.Context, id string) (domain.Dashboard, error) {
	if m.GetDashboardFunc == nil {
		panic("called not defined GetDashboardFunc")
	}
	return m.GetDashboardFunc(ctx, id)
}

func (m *MockDashboardQueries) ListAutomations(ctx context.Context) ([]domain.Automation, error) {
	if m.ListAutomationsFunc == nil {
		panic("called not defined ListAutomationsFunc")
	}
	return m.ListAutomationsFunc(ctx)
}

// MockDashboardCommands is a function-based mock implementation of dashboards.DashboardCommands.
type MockDashboardCommands struct {
	ListDashboardsFunc    func(ctx context.Context) ([]domain.Dashboard, error)
	GetDashboardFunc      func(ctx context.Context, id string) (domain.Dashboard, error)
	ListAutomationsFunc   func(ctx context.Context) ([]domain.Automation, error)
	CreateDashboardFunc   func(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error)
	UpdateDashboardFunc   func(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error)
	DeleteDashboardFunc   func(ctx context.Context, actor domain.Actor, id string) error
	CreateWidgetFunc      func(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error)
	UpdateWidgetFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error)
	DeleteWidgetFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error
	UpdateLayoutFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error
	TriggerAutomationFunc func(ctx context.Context, actor domain.Actor, id string) error
}

func (m *MockDashboardCommands) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	if m.ListDashboardsFunc == nil {
		panic("called not defined ListDashboardsFunc")
	}
	return m.ListDashboardsFunc(ctx)
}

func (m *MockDashboardCommands) GetDashboard(ctx context.Context, id string) (domain.Dashboard, error) {
	if m.GetDashboardFunc == nil {
		panic("called not defined GetDashboardFunc")
	}
	return m.GetDashboardFunc(ctx, id)
}

func (m *MockDashboardCommands) ListAutomations(ctx context.Context) ([]domain.Automation, error) {
	if m.ListAutomationsFunc == nil {
		panic("called not defined ListAutomationsFunc")
	}
	return m.ListAutomationsFunc(ctx)
}

func (m *MockDashboardCommands) CreateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
	if m.CreateDashboardFunc == nil {
		panic("called not defined CreateDashboardFunc")
	}
	return m.CreateDashboardFunc(ctx, actor, dashboard)
}

func (m *MockDashboardCommands) UpdateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
	if m.UpdateDashboardFunc == nil {
		panic("called not defined UpdateDashboardFunc")
	}
	return m.UpdateDashboardFunc(ctx, actor, dashboard)
}

func (m *MockDashboardCommands) DeleteDashboard(ctx context.Context, actor domain.Actor, id string) error {
	if m.DeleteDashboardFunc == nil {
		panic("called not defined DeleteDashboardFunc")
	}
	return m.DeleteDashboardFunc(ctx, actor, id)
}

func (m *MockDashboardCommands) CreateWidget(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error) {
	if m.CreateWidgetFunc == nil {
		panic("called not defined CreateWidgetFunc")
	}
	return m.CreateWidgetFunc(ctx, actor, widget)
}

func (m *MockDashboardCommands) UpdateWidget(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error) {
	if m.UpdateWidgetFunc == nil {
		panic("called not defined UpdateWidgetFunc")
	}
	return m.UpdateWidgetFunc(ctx, actor, dashboardID, widget)
}

func (m *MockDashboardCommands) DeleteWidget(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error {
	if m.DeleteWidgetFunc == nil {
		panic("called not defined DeleteWidgetFunc")
	}
	return m.DeleteWidgetFunc(ctx, actor, dashboardID, widgetID)
}

func (m *MockDashboardCommands) UpdateLayout(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error {
	if m.UpdateLayoutFunc == nil {
		panic("called not defined UpdateLayoutFunc")
	}
	return m.UpdateLayoutFunc(ctx, actor, dashboardID, positions)
}

func (m *MockDashboardCommands) TriggerAutomation(ctx context.Context, actor domain.Actor, id string) error {
	if m.TriggerAutomationFunc == nil {
		panic("called not defined TriggerAutomationFunc")
	}
	return m.TriggerAutomationFunc(ctx, actor, id)
}

// DashboardQueriesContractTesting verifies that any DashboardQueries implementation adheres to query contracts.
func DashboardQueriesContractTesting(t *testing.T, queries dashboards.DashboardQueries) {
	ctx := context.Background()

	t.Run("Contract: ListDashboards returns without error", func(t *testing.T) {
		dashboards, err := queries.ListDashboards(ctx)
		require.NoError(t, err)
		assert.NotNil(t, dashboards)
	})

	t.Run("Contract: GetDashboard returns ErrDashboardNotFound for unknown id", func(t *testing.T) {
		_, err := queries.GetDashboard(ctx, "unknown-dashboard-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
	})

	t.Run("Contract: ListAutomations returns without error", func(t *testing.T) {
		automations, err := queries.ListAutomations(ctx)
		require.NoError(t, err)
		assert.NotNil(t, automations)
	})
}

// DashboardCommandsContractTesting verifies that any DashboardCommands implementation adheres to command contracts.
func DashboardCommandsContractTesting(t *testing.T, commands dashboards.DashboardCommands) {
	DashboardQueriesContractTesting(t, commands)

	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-test"}

	t.Run("Contract: CreateDashboard fails with invalid data", func(t *testing.T) {
		_, err := commands.CreateDashboard(ctx, actor, domain.Dashboard{Name: ""})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDashboard)
	})

	t.Run("Contract: UpdateDashboard fails with non-existent id", func(t *testing.T) {
		_, err := commands.UpdateDashboard(ctx, actor, domain.Dashboard{ID: "non-existent", Name: "Valid"})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
	})

	t.Run("Contract: DeleteDashboard fails with non-existent id", func(t *testing.T) {
		err := commands.DeleteDashboard(ctx, actor, "non-existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDashboardNotFound)
	})

	t.Run("Contract: CreateWidget fails with invalid data", func(t *testing.T) {
		_, err := commands.CreateWidget(ctx, actor, domain.Widget{Title: ""})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("Contract: DeleteWidget fails with non-existent widget", func(t *testing.T) {
		err := commands.DeleteWidget(ctx, actor, "dashboard-id", "non-existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrWidgetNotFound)
	})

	t.Run("Contract: TriggerAutomation fails with invalid id", func(t *testing.T) {
		err := commands.TriggerAutomation(ctx, actor, "invalid_id_not_automation")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAutomationNotFound)
	})
}
