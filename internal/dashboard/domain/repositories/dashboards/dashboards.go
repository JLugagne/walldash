package dashboards

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// DashboardRepository defines repository operations for Dashboard domain entities.
type DashboardRepository interface {
	CreateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error)
	FindDashboardByID(ctx context.Context, id string) (domain.Dashboard, error)
	FindAllDashboards(ctx context.Context) ([]domain.Dashboard, error)
	UpdateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error)
	DeleteDashboard(ctx context.Context, id string) error
}

// WidgetRepository defines repository operations for Widget domain entities.
type WidgetRepository interface {
	CreateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	FindWidgetByID(ctx context.Context, id string) (domain.Widget, error)
	FindWidgetsByDashboardID(ctx context.Context, dashboardID string) ([]domain.Widget, error)
	UpdateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error)
	DeleteWidget(ctx context.Context, id string) error
	DeleteWidgetsByDashboardID(ctx context.Context, dashboardID string) error
	ReplaceWidgetPositions(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error
}
