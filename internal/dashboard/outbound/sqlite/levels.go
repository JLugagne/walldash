package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

func serializeLayers(layers []domain.Layer) string {
	if len(layers) == 0 {
		layers = domain.DefaultLayers
	}
	b, err := json.Marshal(layers)
	if err != nil {
		return `[{"name":"controls","hide_gauges":false},{"name":"sensors","hide_gauges":false}]`
	}
	return string(b)
}

func deserializeLayers(raw string) []domain.Layer {
	if raw == "" {
		return append([]domain.Layer(nil), domain.DefaultLayers...)
	}
	var layers []domain.Layer
	if err := json.Unmarshal([]byte(raw), &layers); err != nil || len(layers) == 0 {
		return append([]domain.Layer(nil), domain.DefaultLayers...)
	}
	return layers
}

var _ levels.LevelRepository = (*Adapter)(nil)

type levelRepo struct {
	db dbExecutor
}

func (r *levelRepo) Create(ctx context.Context, level domain.Level) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	if err := level.Validate(); err != nil {
		return domain.Level{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if level.CreatedAt.IsZero() {
		level.CreatedAt = now
	}
	if level.UpdatedAt.IsZero() {
		level.UpdatedAt = now
	}

	if len(level.Layers) == 0 {
		level.Layers = append([]domain.Layer(nil), domain.DefaultLayers...)
	}
	layersJSON := serializeLayers(level.Layers)

	query := `INSERT INTO levels (id, name, "order", is_outdoor, layers_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, level.ID, level.Name, level.Order, level.IsOutdoor, layersJSON, level.CreatedAt, level.UpdatedAt)
	if err != nil {
		log.WithError(err).WithField("level_id", level.ID).Error("failed to insert level")
		return domain.Level{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return level, nil
}

func (r *levelRepo) FindByID(ctx context.Context, id string) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", is_outdoor, layers_json, created_at, updated_at FROM levels WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var lvl domain.Level
	var isOutdoor int
	var layersJSON string
	var createdAt, updatedAt time.Time
	err := row.Scan(&lvl.ID, &lvl.Name, &lvl.Order, &isOutdoor, &layersJSON, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Level{}, errors.Join(domain.ErrLevelNotFound, err)
		}
		log.WithError(err).WithField("level_id", id).Error("failed to query level by id")
		return domain.Level{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	lvl.IsOutdoor = (isOutdoor != 0)
	lvl.Layers = deserializeLayers(layersJSON)
	lvl.CreatedAt = createdAt
	lvl.UpdatedAt = updatedAt
	return lvl, nil
}

func (r *levelRepo) FindAll(ctx context.Context) ([]domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT id, name, "order", is_outdoor, layers_json, created_at, updated_at FROM levels ORDER BY "order" ASC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to query all levels")
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	defer rows.Close()

	var result []domain.Level
	for rows.Next() {
		var lvl domain.Level
		var isOutdoor int
		var layersJSON string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&lvl.ID, &lvl.Name, &lvl.Order, &isOutdoor, &layersJSON, &createdAt, &updatedAt); err != nil {
			log.WithError(err).Error("failed to scan level row")
			return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		lvl.IsOutdoor = (isOutdoor != 0)
		lvl.Layers = deserializeLayers(layersJSON)
		lvl.CreatedAt = createdAt
		lvl.UpdatedAt = updatedAt
		result = append(result, lvl)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if result == nil {
		result = []domain.Level{}
	}
	return result, nil
}

func (r *levelRepo) Update(ctx context.Context, level domain.Level) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	if err := level.Validate(); err != nil {
		return domain.Level{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	level.UpdatedAt = now

	if len(level.Layers) == 0 {
		level.Layers = append([]domain.Layer(nil), domain.DefaultLayers...)
	}
	layersJSON := serializeLayers(level.Layers)

	query := `UPDATE levels SET name = ?, "order" = ?, is_outdoor = ?, layers_json = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, level.Name, level.Order, level.IsOutdoor, layersJSON, level.UpdatedAt, level.ID)
	if err != nil {
		log.WithError(err).WithField("level_id", level.ID).Error("failed to update level")
		return domain.Level{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return domain.Level{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return domain.Level{}, errors.Join(domain.ErrLevelNotFound, errors.New("no level found to update"))
	}

	return level, nil
}

func (r *levelRepo) Delete(ctx context.Context, id string) error {
	log := logger.LoggerFromContext(ctx)
	// Delete associated plan first (cascade)
	_, _ = r.db.ExecContext(ctx, `DELETE FROM plans WHERE level_id = ?`, id)

	query := `DELETE FROM levels WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.WithError(err).WithField("level_id", id).Error("failed to delete level")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	if rowsAffected == 0 {
		return errors.Join(domain.ErrLevelNotFound, errors.New("no level found to delete"))
	}

	return nil
}

func (r *levelRepo) Reorder(ctx context.Context, orderedIDs []string) error {
	log := logger.LoggerFromContext(ctx)
	for idx, id := range orderedIDs {
		query := `UPDATE levels SET "order" = ?, updated_at = ? WHERE id = ?`
		res, err := r.db.ExecContext(ctx, query, idx, time.Now().UTC().Truncate(time.Second), id)
		if err != nil {
			log.WithError(err).WithField("level_id", id).Error("failed to update order for level")
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Join(domain.ErrDatabaseUnavailable, err)
		}
		if rowsAffected == 0 {
			return errors.Join(domain.ErrLevelNotFound, fmt.Errorf("level %s not found during reorder", id))
		}
	}
	return nil
}

func (a *Adapter) Create(ctx context.Context, level domain.Level) (domain.Level, error) {
	return (&levelRepo{db: a.db}).Create(ctx, level)
}

func (a *Adapter) FindByID(ctx context.Context, id string) (domain.Level, error) {
	return (&levelRepo{db: a.db}).FindByID(ctx, id)
}

func (a *Adapter) FindAll(ctx context.Context) ([]domain.Level, error) {
	return (&levelRepo{db: a.db}).FindAll(ctx)
}

func (a *Adapter) Update(ctx context.Context, level domain.Level) (domain.Level, error) {
	return (&levelRepo{db: a.db}).Update(ctx, level)
}

func (a *Adapter) Delete(ctx context.Context, id string) error {
	return (&levelRepo{db: a.db}).Delete(ctx, id)
}

func (a *Adapter) Reorder(ctx context.Context, orderedIDs []string) error {
	return (&levelRepo{db: a.db}).Reorder(ctx, orderedIDs)
}