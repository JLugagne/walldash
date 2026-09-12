package app

import (
	"context"
	"errors"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/google/uuid"
)

// ListDashboards retrieves all dashboards ordered by order ascending.
func (a *App) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	dashboards, err := a.dashboardsRepo.FindAllDashboards(ctx)
	if err != nil {
		log.WithError(err).Error("failed to list dashboards")
		return nil, err
	}
	return dashboards, nil
}

// GetDashboard retrieves a dashboard by its identifier.
func (a *App) GetDashboard(ctx context.Context, id string) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	dashboard, err := a.dashboardsRepo.FindDashboardByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", id).Error("failed to get dashboard")
		return domain.Dashboard{}, err
	}
	return dashboard, nil
}

// ListAutomations returns all automations discovered from Home Assistant.
func (a *App) ListAutomations(ctx context.Context) ([]domain.Automation, error) {
	log := logger.LoggerFromContext(ctx)
	automations, err := a.haRepo.GetAutomations(ctx)
	if err != nil {
		log.WithError(err).Error("failed to list automations from HA")
		return nil, err
	}
	return automations, nil
}

// CreateDashboard validates and creates a new dashboard.
func (a *App) CreateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(dashboard.Name) == "" {
		return domain.Dashboard{}, errors.Join(domain.ErrInvalidDashboard, errors.New("dashboard name is required"))
	}
	if strings.TrimSpace(dashboard.ID) == "" {
		dashboard.ID = uuid.NewString()
	}

	existing, err := a.dashboardsRepo.FindAllDashboards(ctx)
	if err == nil && len(existing) > 0 && dashboard.Order == 0 {
		maxOrder := 0
		for _, o := range existing {
			if o.Order >= maxOrder {
				maxOrder = o.Order + 1
			}
		}
		dashboard.Order = maxOrder
	}

	created, err := a.dashboardsRepo.CreateDashboard(ctx, dashboard)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to create dashboard")
		return domain.Dashboard{}, err
	}

	log.WithField("dashboard_id", created.ID).WithField("actor", actor.UserID).Info("dashboard created successfully")
	return created, nil
}

// UpdateDashboard renames and reorders an existing dashboard, and may resize its Widget
// Grid. A zero Cols or Rows preserves the stored dimension. Resizing is refused when it would
// push an existing widget out of the viewport or make two widgets overlap, so the grid invariant
// cannot be broken from this route. Returns domain.ErrDashboardNotFound, or an error wrapping
// domain.ErrInvalidDashboard or domain.ErrInvalidWidget.
func (a *App) UpdateDashboard(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(dashboard.ID) == "" {
		return domain.Dashboard{}, errors.Join(domain.ErrInvalidDashboard, errors.New("dashboard id is required"))
	}
	if strings.TrimSpace(dashboard.Name) == "" {
		return domain.Dashboard{}, errors.Join(domain.ErrInvalidDashboard, errors.New("dashboard name is required"))
	}

	current, err := a.dashboardsRepo.FindDashboardByID(ctx, dashboard.ID)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to load dashboard for grid size preservation")
		return domain.Dashboard{}, err
	}
	if dashboard.Cols == 0 {
		dashboard.Cols = current.Cols
	}
	if dashboard.Rows == 0 {
		dashboard.Rows = current.Rows
	}

	if dashboard.Cols != current.Cols || dashboard.Rows != current.Rows {
		widgets, err := a.widgetsRepo.FindWidgetsByDashboardID(ctx, dashboard.ID)
		if err != nil {
			log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to load widgets for grid resize validation")
			return domain.Dashboard{}, err
		}
		resized := dashboard
		resized.Widgets = widgets
		if err := resized.Validate(); err != nil {
			return domain.Dashboard{}, err
		}
	}

	updated, err := a.dashboardsRepo.UpdateDashboard(ctx, dashboard)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to update dashboard")
		return domain.Dashboard{}, err
	}

	log.WithField("dashboard_id", updated.ID).WithField("actor", actor.UserID).Info("dashboard updated successfully")
	return updated, nil
}

// DeleteDashboard removes a dashboard and its child widgets.
func (a *App) DeleteDashboard(ctx context.Context, actor domain.Actor, id string) error {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(id) == "" {
		return errors.Join(domain.ErrInvalidDashboard, errors.New("dashboard id is required"))
	}

	err := a.uow.Do(ctx, func(repos uow.Repositories) error {
		if err := repos.Widgets.DeleteWidgetsByDashboardID(ctx, id); err != nil {
			return err
		}
		return repos.Dashboards.DeleteDashboard(ctx, id)
	})
	if err != nil {
		log.WithError(err).WithField("dashboard_id", id).Error("failed to delete dashboard")
		return err
	}

	log.WithField("dashboard_id", id).WithField("actor", actor.UserID).Info("dashboard deleted successfully")
	return nil
}

// CreateWidget anchors a new widget in the Widget Grid of its dashboard. The candidate is
// checked against the dashboard as a whole inside the same transaction that writes it, so a
// widget whose type and display pair is illegal, whose footprint is below the display minimum,
// which leaves the grid, or which overlaps an existing widget is refused and nothing is written.
// An empty title is accepted: the UI falls back to the config label then to the Home Assistant
// device name. Returns domain.ErrDashboardNotFound when the parent dashboard is unknown, and an
// error wrapping domain.ErrInvalidWidget or domain.ErrInvalidDashboard when the resulting
// dashboard would break the grid invariant.
func (a *App) CreateWidget(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(widget.DashboardID) == "" {
		return domain.Widget{}, errors.Join(domain.ErrInvalidWidget, errors.New("widget dashboard_id is required"))
	}
	if strings.TrimSpace(widget.ID) == "" {
		widget.ID = uuid.NewString()
	}

	var created domain.Widget
	err := a.uow.Do(ctx, func(repos uow.Repositories) error {
		dashboard, err := repos.Dashboards.FindDashboardByID(ctx, widget.DashboardID)
		if err != nil {
			return err
		}

		siblings, err := repos.Widgets.FindWidgetsByDashboardID(ctx, widget.DashboardID)
		if err != nil {
			return err
		}

		candidate := make([]domain.Widget, 0, len(siblings)+1)
		candidate = append(candidate, siblings...)
		candidate = append(candidate, widget)
		dashboard.Widgets = candidate

		if err := dashboard.Validate(); err != nil {
			return err
		}

		created, err = repos.Widgets.CreateWidget(ctx, widget)
		return err
	})
	if err != nil {
		log.WithError(err).WithField("widget_id", widget.ID).Error("failed to create widget")
		return domain.Widget{}, err
	}

	log.WithField("widget_id", created.ID).WithField("actor", actor.UserID).Info("widget created successfully")
	return created, nil
}

// DeleteWidget removes a widget from a dashboard.
func (a *App) DeleteWidget(ctx context.Context, actor domain.Actor, dashboardID string, widgetID string) error {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(widgetID) == "" {
		return errors.Join(domain.ErrInvalidWidget, errors.New("widget id is required"))
	}

	widget, err := a.widgetsRepo.FindWidgetByID(ctx, widgetID)
	if err != nil {
		log.WithError(err).WithField("widget_id", widgetID).Error("widget not found")
		return err
	}

	if dashboardID != "" && widget.DashboardID != dashboardID {
		return errors.Join(domain.ErrWidgetNotFound, errors.New("widget does not belong to specified dashboard"))
	}

	if err := a.widgetsRepo.DeleteWidget(ctx, widgetID); err != nil {
		log.WithError(err).WithField("widget_id", widgetID).Error("failed to delete widget")
		return err
	}

	log.WithField("widget_id", widgetID).WithField("actor", actor.UserID).Info("widget deleted successfully")
	return nil
}

// UpdateWidget overlays the title and configuration of an existing widget, leaving its
// grid position untouched: position only ever changes through UpdateLayout. Returns
// domain.ErrWidgetNotFound when widgetID is unknown or does not belong to dashboardID,
// domain.ErrInvalidDashboard when the parent dashboard cannot be loaded, and
// domain.ErrInvalidWidget when the resulting widget violates the dashboard grid.
func (a *App) UpdateWidget(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(widget.ID) == "" {
		return domain.Widget{}, errors.Join(domain.ErrInvalidWidget, errors.New("widget id is required"))
	}

	existing, err := a.widgetsRepo.FindWidgetByID(ctx, widget.ID)
	if err != nil {
		log.WithError(err).WithField("widget_id", widget.ID).Error("widget not found")
		return domain.Widget{}, err
	}

	if dashboardID != "" && existing.DashboardID != dashboardID {
		return domain.Widget{}, errors.Join(domain.ErrWidgetNotFound, errors.New("widget does not belong to specified dashboard"))
	}

	existing.Title = widget.Title
	existing.Config = widget.Config

	dashboard, err := a.dashboardsRepo.FindDashboardByID(ctx, existing.DashboardID)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", existing.DashboardID).Error("parent dashboard not found")
		return domain.Widget{}, err
	}

	if err := existing.ValidateIn(dashboard.Cols, dashboard.Rows); err != nil {
		return domain.Widget{}, err
	}

	updated, err := a.widgetsRepo.UpdateWidget(ctx, existing)
	if err != nil {
		log.WithError(err).WithField("widget_id", existing.ID).Error("failed to update widget")
		return domain.Widget{}, err
	}

	log.WithField("widget_id", updated.ID).WithField("actor", actor.UserID).Info("widget updated successfully")
	return updated, nil
}

// UpdateLayout replaces the grid positions of the widgets named in positions, all inside
// a single transaction: the candidate layout is validated against the dashboard grid
// invariant before ReplaceWidgetPositions ever writes, so a rejected layout leaves the
// stored positions untouched. Returns domain.ErrWidgetNotFound when a position names a
// widget outside dashboardID, and an error wrapping domain.ErrInvalidDashboard or
// domain.ErrInvalidWidget when the resulting layout would violate the grid invariant.
func (a *App) UpdateLayout(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(dashboardID) == "" {
		return errors.Join(domain.ErrInvalidDashboard, errors.New("dashboard id is required"))
	}

	err := a.uow.Do(ctx, func(repos uow.Repositories) error {
		dashboard, err := repos.Dashboards.FindDashboardByID(ctx, dashboardID)
		if err != nil {
			return err
		}

		widgets, err := repos.Widgets.FindWidgetsByDashboardID(ctx, dashboardID)
		if err != nil {
			return err
		}

		widgetIndexByID := make(map[string]int, len(widgets))
		for i, w := range widgets {
			widgetIndexByID[w.ID] = i
		}

		for _, p := range positions {
			idx, ok := widgetIndexByID[p.ID]
			if !ok {
				return errors.Join(domain.ErrWidgetNotFound, errors.New("widget "+p.ID+" does not belong to dashboard "+dashboardID))
			}
			widgets[idx].Col = p.Col
			widgets[idx].Row = p.Row
			widgets[idx].ColSpan = p.ColSpan
			widgets[idx].RowSpan = p.RowSpan
		}
		dashboard.Widgets = widgets

		if err := dashboard.Validate(); err != nil {
			return err
		}

		return repos.Widgets.ReplaceWidgetPositions(ctx, dashboardID, positions)
	})
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboardID).Error("failed to update widget layout")
		return err
	}

	log.WithField("dashboard_id", dashboardID).WithField("actor", actor.UserID).Info("widget layout updated successfully")
	return nil
}

// TriggerAutomation triggers a Home Assistant automation.
func (a *App) TriggerAutomation(ctx context.Context, actor domain.Actor, id string) error {
	log := logger.LoggerFromContext(ctx)

	trimmedID := strings.TrimSpace(id)
	if !strings.HasPrefix(trimmedID, "automation.") {
		return errors.Join(domain.ErrAutomationNotFound, errors.New("invalid automation id: "+id))
	}

	if err := a.haRepo.TriggerAutomation(ctx, trimmedID); err != nil {
		log.WithError(err).WithField("automation_id", id).Error("failed to trigger automation")
		return err
	}

	log.WithField("automation_id", id).WithField("actor", actor.UserID).Info("automation triggered successfully")
	return nil
}
