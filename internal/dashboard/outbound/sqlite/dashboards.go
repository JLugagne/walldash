package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/dashboards"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

var (
	_ dashboards.DashboardRepository = (*dashboardRepo)(nil)
	_ dashboards.DashboardRepository = (*Adapter)(nil)
	_ dashboards.WidgetRepository    = (*widgetRepo)(nil)
	_ dashboards.WidgetRepository    = (*Adapter)(nil)
)

type dashboardRepo struct {
	db dbExecutor
}

func (r *dashboardRepo) CreateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	if err := dashboard.Validate(); err != nil {
		return domain.Dashboard{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if dashboard.CreatedAt.IsZero() {
		dashboard.CreatedAt = now
	}
	if dashboard.UpdatedAt.IsZero() {
		dashboard.UpdatedAt = now
	}

	query := `INSERT INTO dashboards (id, name, "order", cols, rows, bg_image, bg_opacity, bg_blur, bg_dim, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, dashboard.ID, dashboard.Name, dashboard.Order, dashboard.Cols, dashboard.Rows, dashboard.BackgroundImage, dashboard.BackgroundOpacity, dashboard.BackgroundBlur, dashboard.BackgroundDim, dashboard.CreatedAt, dashboard.UpdatedAt)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to insert dashboard")
		return domain.Dashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	if dashboard.Widgets == nil {
		dashboard.Widgets = []domain.Widget{}
	}
	return dashboard, nil
}

func (r *dashboardRepo) FindDashboardByID(ctx context.Context, id string) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", cols, rows, bg_image, bg_opacity, bg_blur, bg_dim, created_at, updated_at FROM dashboards WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var ov domain.Dashboard
	var createdAt, updatedAt time.Time
	err := row.Scan(&ov.ID, &ov.Name, &ov.Order, &ov.Cols, &ov.Rows, &ov.BackgroundImage, &ov.BackgroundOpacity, &ov.BackgroundBlur, &ov.BackgroundDim, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Dashboard{}, errors.Join(domain.ErrDashboardNotFound, err)
		}
		log.WithError(err).WithField("dashboard_id", id).Error("failed to query dashboard by id")
		return domain.Dashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	ov.CreatedAt = createdAt
	ov.UpdatedAt = updatedAt

	wRepo := &widgetRepo{db: r.db}
	widgets, err := wRepo.FindWidgetsByDashboardID(ctx, id)
	if err != nil {
		return domain.Dashboard{}, err
	}
	ov.Widgets = widgets

	return ov, nil
}

func (r *dashboardRepo) FindAllDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", cols, rows, bg_image, bg_opacity, bg_blur, bg_dim, created_at, updated_at FROM dashboards ORDER BY "order" ASC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to query all dashboards")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	var result []domain.Dashboard
	wRepo := &widgetRepo{db: r.db}

	for rows.Next() {
		var ov domain.Dashboard
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&ov.ID, &ov.Name, &ov.Order, &ov.Cols, &ov.Rows, &ov.BackgroundImage, &ov.BackgroundOpacity, &ov.BackgroundBlur, &ov.BackgroundDim, &createdAt, &updatedAt); err != nil {
			log.WithError(err).Error("failed to scan dashboard row")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		ov.CreatedAt = createdAt
		ov.UpdatedAt = updatedAt
		result = append(result, ov)
	}
	_ = rows.Close()

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if result == nil {
		result = []domain.Dashboard{}
	}

	for i := range result {
		widgets, err := wRepo.FindWidgetsByDashboardID(ctx, result[i].ID)
		if err != nil {
			return nil, err
		}
		result[i].Widgets = widgets
	}
	return result, nil
}

func (r *dashboardRepo) UpdateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	log := logger.LoggerFromContext(ctx)
	if err := dashboard.Validate(); err != nil {
		return domain.Dashboard{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	dashboard.UpdatedAt = now

	query := `UPDATE dashboards SET name = ?, "order" = ?, cols = ?, rows = ?, bg_image = ?, bg_opacity = ?, bg_blur = ?, bg_dim = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, dashboard.Name, dashboard.Order, dashboard.Cols, dashboard.Rows, dashboard.BackgroundImage, dashboard.BackgroundOpacity, dashboard.BackgroundBlur, dashboard.BackgroundDim, dashboard.UpdatedAt, dashboard.ID)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboard.ID).Error("failed to update dashboard")
		return domain.Dashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return domain.Dashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return domain.Dashboard{}, errors.Join(domain.ErrDashboardNotFound, errors.New("no dashboard found to update"))
	}

	wRepo := &widgetRepo{db: r.db}
	widgets, err := wRepo.FindWidgetsByDashboardID(ctx, dashboard.ID)
	if err == nil {
		dashboard.Widgets = widgets
	} else {
		dashboard.Widgets = []domain.Widget{}
	}

	return dashboard, nil
}

func (r *dashboardRepo) DeleteDashboard(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM dashboards WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", id).Error("failed to delete dashboard")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return errors.Join(domain.ErrDashboardNotFound, errors.New("no dashboard found to delete"))
	}

	return nil
}

type widgetRepo struct {
	db dbExecutor
}

func (r *widgetRepo) CreateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)
	if err := widget.Validate(); err != nil {
		return domain.Widget{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if widget.CreatedAt.IsZero() {
		widget.CreatedAt = now
	}
	if widget.UpdatedAt.IsZero() {
		widget.UpdatedAt = now
	}

	if widget.Config.EntityIDs == nil {
		widget.Config.EntityIDs = []string{}
	}

	configBytes, err := json.Marshal(widget.Config)
	if err != nil {
		return domain.Widget{}, errors.Join(domain.ErrInvalidWidget, err)
	}

	query := `
		INSERT INTO widgets (id, dashboard_id, type, title, "order", col, row, col_span, row_span, config_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query, widget.ID, widget.DashboardID, widget.Type, widget.Title, widget.Order, widget.Col, widget.Row, widget.ColSpan, widget.RowSpan, string(configBytes), widget.CreatedAt, widget.UpdatedAt)
	if err != nil {
		log.WithError(err).WithField("widget_id", widget.ID).Error("failed to insert widget")
		return domain.Widget{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return widget, nil
}

func (r *widgetRepo) FindWidgetByID(ctx context.Context, id string) (domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)
	query := `
		SELECT id, dashboard_id, type, title, "order", col, row, col_span, row_span, config_json, created_at, updated_at
		FROM widgets
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var w domain.Widget
	var configJSON string
	var createdAt, updatedAt time.Time
	err := row.Scan(&w.ID, &w.DashboardID, &w.Type, &w.Title, &w.Order, &w.Col, &w.Row, &w.ColSpan, &w.RowSpan, &configJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Widget{}, errors.Join(domain.ErrWidgetNotFound, err)
		}
		log.WithError(err).WithField("widget_id", id).Error("failed to query widget by id")
		return domain.Widget{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	if configJSON != "" {
		_ = json.Unmarshal([]byte(configJSON), &w.Config)
	}
	if w.Config.EntityIDs == nil {
		w.Config.EntityIDs = []string{}
	}

	w.CreatedAt = createdAt
	w.UpdatedAt = updatedAt
	return w, nil
}

func (r *widgetRepo) FindWidgetsByDashboardID(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)
	query := `
		SELECT id, dashboard_id, type, title, "order", col, row, col_span, row_span, config_json, created_at, updated_at
		FROM widgets
		WHERE dashboard_id = ?
		ORDER BY "order" ASC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, dashboardID)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboardID).Error("failed to query widgets by dashboard id")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	var result []domain.Widget
	for rows.Next() {
		var w domain.Widget
		var configJSON string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&w.ID, &w.DashboardID, &w.Type, &w.Title, &w.Order, &w.Col, &w.Row, &w.ColSpan, &w.RowSpan, &configJSON, &createdAt, &updatedAt); err != nil {
			log.WithError(err).Error("failed to scan widget row")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		if configJSON != "" {
			_ = json.Unmarshal([]byte(configJSON), &w.Config)
		}
		if w.Config.EntityIDs == nil {
			w.Config.EntityIDs = []string{}
		}

		w.CreatedAt = createdAt
		w.UpdatedAt = updatedAt
		result = append(result, w)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if result == nil {
		result = []domain.Widget{}
	}
	return result, nil
}

func (r *widgetRepo) UpdateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	log := logger.LoggerFromContext(ctx)
	if err := widget.Validate(); err != nil {
		return domain.Widget{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	widget.UpdatedAt = now

	if widget.Config.EntityIDs == nil {
		widget.Config.EntityIDs = []string{}
	}

	configBytes, err := json.Marshal(widget.Config)
	if err != nil {
		return domain.Widget{}, errors.Join(domain.ErrInvalidWidget, err)
	}

	query := `UPDATE widgets SET title = ?, "order" = ?, col = ?, row = ?, col_span = ?, row_span = ?, config_json = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, widget.Title, widget.Order, widget.Col, widget.Row, widget.ColSpan, widget.RowSpan, string(configBytes), widget.UpdatedAt, widget.ID)
	if err != nil {
		log.WithError(err).WithField("widget_id", widget.ID).Error("failed to update widget")
		return domain.Widget{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return domain.Widget{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return domain.Widget{}, errors.Join(domain.ErrWidgetNotFound, errors.New("no widget found to update"))
	}

	return widget, nil
}

func (r *widgetRepo) DeleteWidget(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM widgets WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).WithField("widget_id", id).Error("failed to delete widget")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return errors.Join(domain.ErrWidgetNotFound, fmt.Errorf("widget %s not found to delete", id))
	}

	return nil
}

func (r *widgetRepo) DeleteWidgetsByDashboardID(ctx context.Context, dashboardID string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM widgets WHERE dashboard_id = ?`
	_, err := r.db.ExecContext(ctx, query, dashboardID)
	if err != nil {
		log.WithError(err).WithField("dashboard_id", dashboardID).Error("failed to delete widgets by dashboard id")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (r *widgetRepo) ReplaceWidgetPositions(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error {
	log := logger.LoggerFromContext(ctx)
	query := `UPDATE widgets SET col = ?, row = ?, col_span = ?, row_span = ? WHERE id = ? AND dashboard_id = ?`
	for _, p := range positions {
		res, err := r.db.ExecContext(ctx, query, p.Col, p.Row, p.ColSpan, p.RowSpan, p.ID, dashboardID)
		if err != nil {
			log.WithError(err).WithField("widget_id", p.ID).Error("failed to update widget position")
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		if rowsAffected == 0 {
			return errors.Join(domain.ErrWidgetNotFound, fmt.Errorf("widget %s not found in dashboard %s", p.ID, dashboardID))
		}
	}
	return nil
}

// Adapter delegation methods for DashboardRepository
func (a *Adapter) CreateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	return (&dashboardRepo{db: a.db}).CreateDashboard(ctx, dashboard)
}

func (a *Adapter) FindDashboardByID(ctx context.Context, id string) (domain.Dashboard, error) {
	return (&dashboardRepo{db: a.db}).FindDashboardByID(ctx, id)
}

func (a *Adapter) FindAllDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	return (&dashboardRepo{db: a.db}).FindAllDashboards(ctx)
}

func (a *Adapter) UpdateDashboard(ctx context.Context, dashboard domain.Dashboard) (domain.Dashboard, error) {
	return (&dashboardRepo{db: a.db}).UpdateDashboard(ctx, dashboard)
}

func (a *Adapter) DeleteDashboard(ctx context.Context, id string) error {
	return (&dashboardRepo{db: a.db}).DeleteDashboard(ctx, id)
}

// Adapter delegation methods for WidgetRepository
func (a *Adapter) CreateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	return (&widgetRepo{db: a.db}).CreateWidget(ctx, widget)
}

func (a *Adapter) FindWidgetByID(ctx context.Context, id string) (domain.Widget, error) {
	return (&widgetRepo{db: a.db}).FindWidgetByID(ctx, id)
}

func (a *Adapter) FindWidgetsByDashboardID(ctx context.Context, dashboardID string) ([]domain.Widget, error) {
	return (&widgetRepo{db: a.db}).FindWidgetsByDashboardID(ctx, dashboardID)
}

func (a *Adapter) UpdateWidget(ctx context.Context, widget domain.Widget) (domain.Widget, error) {
	return (&widgetRepo{db: a.db}).UpdateWidget(ctx, widget)
}

func (a *Adapter) DeleteWidget(ctx context.Context, id string) error {
	return (&widgetRepo{db: a.db}).DeleteWidget(ctx, id)
}

func (a *Adapter) DeleteWidgetsByDashboardID(ctx context.Context, dashboardID string) error {
	return (&widgetRepo{db: a.db}).DeleteWidgetsByDashboardID(ctx, dashboardID)
}

func (a *Adapter) ReplaceWidgetPositions(ctx context.Context, dashboardID string, positions []domain.WidgetPosition) error {
	return (&widgetRepo{db: a.db}).ReplaceWidgetPositions(ctx, dashboardID, positions)
}
