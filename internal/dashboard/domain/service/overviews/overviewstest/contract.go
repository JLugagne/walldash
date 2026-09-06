package overviewstest

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/overviews"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockOverviewQueries is a function-based mock implementation of overviews.OverviewQueries.
type MockOverviewQueries struct {
	ListOverviewsFunc   func(ctx context.Context) ([]domain.OverviewDashboard, error)
	GetOverviewFunc     func(ctx context.Context, id string) (domain.OverviewDashboard, error)
	ListAutomationsFunc func(ctx context.Context) ([]domain.Automation, error)
}

func (m *MockOverviewQueries) ListOverviews(ctx context.Context) ([]domain.OverviewDashboard, error) {
	if m.ListOverviewsFunc == nil {
		panic("called not defined ListOverviewsFunc")
	}
	return m.ListOverviewsFunc(ctx)
}

func (m *MockOverviewQueries) GetOverview(ctx context.Context, id string) (domain.OverviewDashboard, error) {
	if m.GetOverviewFunc == nil {
		panic("called not defined GetOverviewFunc")
	}
	return m.GetOverviewFunc(ctx, id)
}

func (m *MockOverviewQueries) ListAutomations(ctx context.Context) ([]domain.Automation, error) {
	if m.ListAutomationsFunc == nil {
		panic("called not defined ListAutomationsFunc")
	}
	return m.ListAutomationsFunc(ctx)
}

// MockOverviewCommands is a function-based mock implementation of overviews.OverviewCommands.
type MockOverviewCommands struct {
	ListOverviewsFunc     func(ctx context.Context) ([]domain.OverviewDashboard, error)
	GetOverviewFunc       func(ctx context.Context, id string) (domain.OverviewDashboard, error)
	ListAutomationsFunc   func(ctx context.Context) ([]domain.Automation, error)
	CreateOverviewFunc    func(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	UpdateOverviewFunc    func(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	DeleteOverviewFunc    func(ctx context.Context, actor domain.Actor, id string) error
	CreateWidgetFunc      func(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error)
	UpdateWidgetFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error)
	DeleteWidgetFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error
	UpdateLayoutFunc      func(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error
	TriggerAutomationFunc func(ctx context.Context, actor domain.Actor, id string) error
}

func (m *MockOverviewCommands) ListOverviews(ctx context.Context) ([]domain.OverviewDashboard, error) {
	if m.ListOverviewsFunc == nil {
		panic("called not defined ListOverviewsFunc")
	}
	return m.ListOverviewsFunc(ctx)
}

func (m *MockOverviewCommands) GetOverview(ctx context.Context, id string) (domain.OverviewDashboard, error) {
	if m.GetOverviewFunc == nil {
		panic("called not defined GetOverviewFunc")
	}
	return m.GetOverviewFunc(ctx, id)
}

func (m *MockOverviewCommands) ListAutomations(ctx context.Context) ([]domain.Automation, error) {
	if m.ListAutomationsFunc == nil {
		panic("called not defined ListAutomationsFunc")
	}
	return m.ListAutomationsFunc(ctx)
}

func (m *MockOverviewCommands) CreateOverview(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	if m.CreateOverviewFunc == nil {
		panic("called not defined CreateOverviewFunc")
	}
	return m.CreateOverviewFunc(ctx, actor, overview)
}

func (m *MockOverviewCommands) UpdateOverview(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	if m.UpdateOverviewFunc == nil {
		panic("called not defined UpdateOverviewFunc")
	}
	return m.UpdateOverviewFunc(ctx, actor, overview)
}

func (m *MockOverviewCommands) DeleteOverview(ctx context.Context, actor domain.Actor, id string) error {
	if m.DeleteOverviewFunc == nil {
		panic("called not defined DeleteOverviewFunc")
	}
	return m.DeleteOverviewFunc(ctx, actor, id)
}

func (m *MockOverviewCommands) CreateWidget(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error) {
	if m.CreateWidgetFunc == nil {
		panic("called not defined CreateWidgetFunc")
	}
	return m.CreateWidgetFunc(ctx, actor, widget)
}

func (m *MockOverviewCommands) UpdateWidget(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error) {
	if m.UpdateWidgetFunc == nil {
		panic("called not defined UpdateWidgetFunc")
	}
	return m.UpdateWidgetFunc(ctx, actor, dashboardID, widget)
}

func (m *MockOverviewCommands) DeleteWidget(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error {
	if m.DeleteWidgetFunc == nil {
		panic("called not defined DeleteWidgetFunc")
	}
	return m.DeleteWidgetFunc(ctx, actor, dashboardID, widgetID)
}

func (m *MockOverviewCommands) UpdateLayout(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error {
	if m.UpdateLayoutFunc == nil {
		panic("called not defined UpdateLayoutFunc")
	}
	return m.UpdateLayoutFunc(ctx, actor, dashboardID, positions)
}

func (m *MockOverviewCommands) TriggerAutomation(ctx context.Context, actor domain.Actor, id string) error {
	if m.TriggerAutomationFunc == nil {
		panic("called not defined TriggerAutomationFunc")
	}
	return m.TriggerAutomationFunc(ctx, actor, id)
}

// OverviewQueriesContractTesting verifies that any OverviewQueries implementation adheres to query contracts.
func OverviewQueriesContractTesting(t *testing.T, queries overviews.OverviewQueries) {
	ctx := context.Background()

	t.Run("Contract: ListOverviews returns without error", func(t *testing.T) {
		dashboards, err := queries.ListOverviews(ctx)
		require.NoError(t, err)
		assert.NotNil(t, dashboards)
	})

	t.Run("Contract: GetOverview returns ErrOverviewNotFound for unknown id", func(t *testing.T) {
		_, err := queries.GetOverview(ctx, "unknown-overview-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
	})

	t.Run("Contract: ListAutomations returns without error", func(t *testing.T) {
		automations, err := queries.ListAutomations(ctx)
		require.NoError(t, err)
		assert.NotNil(t, automations)
	})
}

// OverviewCommandsContractTesting verifies that any OverviewCommands implementation adheres to command contracts.
func OverviewCommandsContractTesting(t *testing.T, commands overviews.OverviewCommands) {
	OverviewQueriesContractTesting(t, commands)

	ctx := context.Background()
	actor := domain.Actor{UserID: "admin-test"}

	t.Run("Contract: CreateOverview fails with invalid data", func(t *testing.T) {
		_, err := commands.CreateOverview(ctx, actor, domain.OverviewDashboard{Name: ""})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidOverview)
	})

	t.Run("Contract: UpdateOverview fails with non-existent id", func(t *testing.T) {
		_, err := commands.UpdateOverview(ctx, actor, domain.OverviewDashboard{ID: "non-existent", Name: "Valid"})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
	})

	t.Run("Contract: DeleteOverview fails with non-existent id", func(t *testing.T) {
		err := commands.DeleteOverview(ctx, actor, "non-existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOverviewNotFound)
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
