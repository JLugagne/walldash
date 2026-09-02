package overviews

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// OverviewRepository defines repository operations for OverviewDashboard domain entities.
type OverviewRepository interface {
	CreateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	FindOverviewByID(ctx context.Context, id string) (domain.OverviewDashboard, error)
	FindAllOverviews(ctx context.Context) ([]domain.OverviewDashboard, error)
	UpdateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error)
	DeleteOverview(ctx context.Context, id string) error
}

// WidgetRepository defines repository operations for Widget domain entities.
type WidgetRepository interface {
	CreateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	FindWidgetByID(ctx context.Context, id string) (domain.Widget, error)
	FindWidgetsByDashboardID(ctx context.Context, dashboardID string) ([]domain.Widget, error)
	UpdateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	DeleteWidget(ctx context.Context, id string) error
	DeleteWidgetsByDashboardID(ctx context.Context, dashboardID string) error
}
