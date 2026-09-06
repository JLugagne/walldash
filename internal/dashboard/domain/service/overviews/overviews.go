package overviews

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// OverviewQueries defines read-only operations for overview dashboards and automations.
type OverviewQueries interface {
	ListOverviews(ctx context.Context) ([]domain.OverviewDashboard, error)
	GetOverview(ctx context.Context, id string) (domain.OverviewDashboard, error)
	ListAutomations(ctx context.Context) ([]domain.Automation, error)
}

// OverviewCommands defines mutating operations for overview dashboards and widgets, extending OverviewQueries.
type OverviewCommands interface {
	OverviewQueries
	CreateOverview(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	UpdateOverview(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	DeleteOverview(ctx context.Context, actor domain.Actor, id string) error
	CreateWidget(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error)
	UpdateWidget(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error)
	DeleteWidget(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error
	UpdateLayout(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error
	TriggerAutomation(ctx context.Context, actor domain.Actor, id string) error
}
