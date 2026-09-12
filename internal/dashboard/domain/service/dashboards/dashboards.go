package dashboards

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// DashboardQueries defines read-only operations for dashboards and automations.
type DashboardQueries interface {
	ListDashboards(ctx context.Context) ([]domain.Dashboard, error)
	GetDashboard(ctx context.Context, id string) (domain.Dashboard, error)
	ListAutomations(ctx context.Context) ([]domain.Automation, error)
}

// DashboardCommands defines mutating operations for dashboards and widgets, extending DashboardQueries.
type DashboardCommands interface {
	DashboardQueries
	CreateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error)
	UpdateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error)
	DeleteDashboard(ctx context.Context, actor domain.Actor, id string) error
	CreateWidget(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error)
	UpdateWidget(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error)
	DeleteWidget(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error
	UpdateLayout(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error
	TriggerAutomation(ctx context.Context, actor domain.Actor, id string) error
}
