package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/overviews"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

var (
	_ overviews.OverviewRepository = (*overviewRepo)(nil)
	_ overviews.OverviewRepository = (*Adapter)(nil)
	_ overviews.WidgetRepository   = (*widgetRepo)(nil)
	_ overviews.WidgetRepository   = (*Adapter)(nil)
)

type overviewRepo struct {
	db dbExecutor
}

func (r *overviewRepo) CreateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	log := logger.LoggerFromContext(ctx)
	if err := overview.Validate(); err != nil {
		return domain.OverviewDashboard{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if overview.CreatedAt.IsZero() {
		overview.CreatedAt = now
	}
	if overview.UpdatedAt.IsZero() {
		overview.UpdatedAt = now
	}

	query := `INSERT INTO overview_dashboards (id, name, "order", cols, rows, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, overview.ID, overview.Name, overview.Order, overview.Cols, overview.Rows, overview.CreatedAt, overview.UpdatedAt)
	if err != nil {
		log.WithError(err).WithField("overview_id", overview.ID).Error("failed to insert overview dashboard")
		return domain.OverviewDashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	if overview.Widgets == nil {
		overview.Widgets = []domain.Widget{}
	}
	return overview, nil
}

func (r *overviewRepo) FindOverviewByID(ctx context.Context, id string) (domain.OverviewDashboard, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", cols, rows, created_at, updated_at FROM overview_dashboards WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var ov domain.OverviewDashboard
	var createdAt, updatedAt time.Time
	err := row.Scan(&ov.ID, &ov.Name, &ov.Order, &ov.Cols, &ov.Rows, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OverviewDashboard{}, errors.Join(domain.ErrOverviewNotFound, err)
		}
		log.WithError(err).WithField("overview_id", id).Error("failed to query overview dashboard by id")
		return domain.OverviewDashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	ov.CreatedAt = createdAt
	ov.UpdatedAt = updatedAt

	wRepo := &widgetRepo{db: r.db}
	widgets, err := wRepo.FindWidgetsByDashboardID(ctx, id)
	if err != nil {
		return domain.OverviewDashboard{}, err
	}
	ov.Widgets = widgets

	return ov, nil
}

func (r *overviewRepo) FindAllOverviews(ctx context.Context) ([]domain.OverviewDashboard, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", cols, rows, created_at, updated_at FROM overview_dashboards ORDER BY "order" ASC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to query all overview dashboards")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	var result []domain.OverviewDashboard
	wRepo := &widgetRepo{db: r.db}

	for rows.Next() {
		var ov domain.OverviewDashboard
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&ov.ID, &ov.Name, &ov.Order, &ov.Cols, &ov.Rows, &createdAt, &updatedAt); err != nil {
			log.WithError(err).Error("failed to scan overview dashboard row")
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
		result = []domain.OverviewDashboard{}
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

func (r *overviewRepo) UpdateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	log := logger.LoggerFromContext(ctx)
	if err := overview.Validate(); err != nil {
		return domain.OverviewDashboard{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	overview.UpdatedAt = now

	query := `UPDATE overview_dashboards SET name = ?, "order" = ?, cols = ?, rows = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, overview.Name, overview.Order, overview.Cols, overview.Rows, overview.UpdatedAt, overview.ID)
	if err != nil {
		log.WithError(err).WithField("overview_id", overview.ID).Error("failed to update overview dashboard")
		return domain.OverviewDashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return domain.OverviewDashboard{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return domain.OverviewDashboard{}, errors.Join(domain.ErrOverviewNotFound, errors.New("no overview dashboard found to update"))
	}

	wRepo := &widgetRepo{db: r.db}
	widgets, err := wRepo.FindWidgetsByDashboardID(ctx, overview.ID)
	if err == nil {
		overview.Widgets = widgets
	} else {
		overview.Widgets = []domain.Widget{}
	}

	return overview, nil
}

func (r *overviewRepo) DeleteOverview(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM overview_dashboards WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).WithField("overview_id", id).Error("failed to delete overview dashboard")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return errors.Join(domain.ErrOverviewNotFound, errors.New("no overview dashboard found to delete"))
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

// Adapter delegation methods for OverviewRepository
func (a *Adapter) CreateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	return (&overviewRepo{db: a.db}).CreateOverview(ctx, overview)
}

func (a *Adapter) FindOverviewByID(ctx context.Context, id string) (domain.OverviewDashboard, error) {
	return (&overviewRepo{db: a.db}).FindOverviewByID(ctx, id)
}

func (a *Adapter) FindAllOverviews(ctx context.Context) ([]domain.OverviewDashboard, error) {
	return (&overviewRepo{db: a.db}).FindAllOverviews(ctx)
}

func (a *Adapter) UpdateOverview(ctx context.Context, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
	return (&overviewRepo{db: a.db}).UpdateOverview(ctx, overview)
}

func (a *Adapter) DeleteOverview(ctx context.Context, id string) error {
	return (&overviewRepo{db: a.db}).DeleteOverview(ctx, id)
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
