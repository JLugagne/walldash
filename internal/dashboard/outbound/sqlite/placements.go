package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/placements"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

var (
	_ placements.DevicePlacementRepository = (*placementRepo)(nil)
	_ placements.DevicePlacementRepository = (*Adapter)(nil)
)

type placementRepo struct {
	db dbExecutor
}

func (r *placementRepo) SavePlacement(ctx context.Context, p domain.DevicePlacement) (domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)
	if p.Layer == "" {
		p.Layer = domain.DefaultPlacementLayer
	}
	if err := p.Validate(); err != nil {
		return domain.DevicePlacement{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	query := `
		INSERT INTO device_placements (id, level_id, device_id, x, y, icon, render_domain, custom_name, layer, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			level_id = excluded.level_id,
			device_id = excluded.device_id,
			x = excluded.x,
			y = excluded.y,
			icon = excluded.icon,
			render_domain = excluded.render_domain,
			custom_name = excluded.custom_name,
			layer = excluded.layer,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query, p.ID, p.LevelID, p.DeviceID, p.X, p.Y, p.Icon, p.RenderDomain, p.CustomName, p.Layer, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		log.WithError(err).WithField("placement_id", p.ID).Error("failed to save device placement")
		return domain.DevicePlacement{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return p, nil
}

func (r *placementRepo) FindPlacementByID(ctx context.Context, id string) (domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)
	query := `
		SELECT id, level_id, device_id, x, y, icon, render_domain, custom_name, layer, created_at, updated_at
		FROM device_placements
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var p domain.DevicePlacement
	var icon, renderDomain, customName, layer sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(&p.ID, &p.LevelID, &p.DeviceID, &p.X, &p.Y, &icon, &renderDomain, &customName, &layer, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DevicePlacement{}, errors.Join(domain.ErrPlacementNotFound, err)
		}
		log.WithError(err).WithField("placement_id", id).Error("failed to query device placement by id")
		return domain.DevicePlacement{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	p.Icon = icon.String
	p.RenderDomain = renderDomain.String
	p.CustomName = customName.String
	p.Layer = layer.String
	if p.Layer == "" {
		p.Layer = domain.DefaultPlacementLayer
	}
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return p, nil
}

func (r *placementRepo) FindPlacementsByLevelID(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)
	query := `
		SELECT id, level_id, device_id, x, y, icon, render_domain, custom_name, layer, created_at, updated_at
		FROM device_placements
		WHERE level_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, levelID)
	if err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("failed to query device placements by level_id")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	var result []domain.DevicePlacement
	for rows.Next() {
		var p domain.DevicePlacement
		var icon, renderDomain, customName, layer sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&p.ID, &p.LevelID, &p.DeviceID, &p.X, &p.Y, &icon, &renderDomain, &customName, &layer, &createdAt, &updatedAt); err != nil {
			log.WithError(err).Error("failed to scan device placement row")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}

		p.Icon = icon.String
		p.RenderDomain = renderDomain.String
		p.CustomName = customName.String
		p.Layer = layer.String
		if p.Layer == "" {
			p.Layer = domain.DefaultPlacementLayer
		}
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if result == nil {
		result = []domain.DevicePlacement{}
	}
	return result, nil
}

func (r *placementRepo) DeletePlacement(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM device_placements WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).WithField("placement_id", id).Error("failed to delete device placement")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return errors.Join(domain.ErrPlacementNotFound, fmt.Errorf("placement %s not found", id))
	}

	return nil
}

func (r *placementRepo) ReassignLayer(ctx context.Context, levelID string, oldLayer string, newLayer string) error {
	log := logger.LoggerFromContext(ctx)
	now := time.Now().UTC().Truncate(time.Second)
	query := `UPDATE device_placements SET layer = ?, updated_at = ? WHERE level_id = ? AND layer = ?`
	_, err := r.db.ExecContext(ctx, query, newLayer, now, levelID, oldLayer)
	if err != nil {
		log.WithError(err).WithField("level_id", levelID).WithField("old_layer", oldLayer).WithField("new_layer", newLayer).Error("failed to reassign layer")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

// Delegate methods for Adapter
func (a *Adapter) SavePlacement(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
	return (&placementRepo{db: a.db}).SavePlacement(ctx, placement)
}

func (a *Adapter) FindPlacementByID(ctx context.Context, id string) (domain.DevicePlacement, error) {
	return (&placementRepo{db: a.db}).FindPlacementByID(ctx, id)
}

func (a *Adapter) FindPlacementsByLevelID(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	return (&placementRepo{db: a.db}).FindPlacementsByLevelID(ctx, levelID)
}

func (a *Adapter) DeletePlacement(ctx context.Context, id string) error {
	return (&placementRepo{db: a.db}).DeletePlacement(ctx, id)
}

func (a *Adapter) ReassignLayer(ctx context.Context, levelID string, oldLayer string, newLayer string) error {
	return (&placementRepo{db: a.db}).ReassignLayer(ctx, levelID, oldLayer, newLayer)
}
