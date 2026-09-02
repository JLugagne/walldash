package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/plans"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

var _ plans.PlanRepository = (*Adapter)(nil)

type planRepo struct {
	db dbExecutor
}

func (r *planRepo) Save(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	log := logger.LoggerFromContext(ctx)
	if err := plan.Validate(); err != nil {
		return domain.Plan{}, err
	}

	if plan.Walls == nil {
		plan.Walls = []domain.WallSegment{}
	}
	if plan.Zones == nil {
		plan.Zones = []domain.Zone{}
	}

	wallsJSON, err := json.Marshal(plan.Walls)
	if err != nil {
		return domain.Plan{}, errors.Join(domain.ErrInvalidPlan, err)
	}

	zonesJSON, err := json.Marshal(plan.Zones)
	if err != nil {
		return domain.Plan{}, errors.Join(domain.ErrInvalidPlan, err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	query := `
		INSERT INTO plans (level_id, walls_json, zones_json, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(level_id) DO UPDATE SET
			walls_json = excluded.walls_json,
			zones_json = excluded.zones_json,
			updated_at = excluded.updated_at
	`
	_, err = r.db.ExecContext(ctx, query, plan.LevelID, string(wallsJSON), string(zonesJSON), now)
	if err != nil {
		log.WithError(err).WithField("level_id", plan.LevelID).Error("failed to save plan")
		return domain.Plan{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	return plan, nil
}

func (r *planRepo) FindByLevelID(ctx context.Context, levelID string) (domain.Plan, error) {
	log := logger.LoggerFromContext(ctx)
	query := `SELECT walls_json, zones_json FROM plans WHERE level_id = ?`
	row := r.db.QueryRowContext(ctx, query, levelID)

	var wallsJSON, zonesJSON string
	err := row.Scan(&wallsJSON, &zonesJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Plan{}, errors.Join(domain.ErrPlanNotFound, err)
		}
		log.WithError(err).WithField("level_id", levelID).Error("failed to find plan by level_id")
		return domain.Plan{}, errors.Join(domain.ErrDatabaseUnavailable, err)
	}

	var walls []domain.WallSegment
	if err := json.Unmarshal([]byte(wallsJSON), &walls); err != nil {
		log.WithError(err).Error("failed to unmarshal walls json")
		return domain.Plan{}, errors.Join(domain.ErrInvalidPlan, err)
	}
	if walls == nil {
		walls = []domain.WallSegment{}
	}

	var zones []domain.Zone
	if err := json.Unmarshal([]byte(zonesJSON), &zones); err != nil {
		log.WithError(err).Error("failed to unmarshal zones json")
		return domain.Plan{}, errors.Join(domain.ErrInvalidPlan, err)
	}
	if zones == nil {
		zones = []domain.Zone{}
	}

	return domain.Plan{
		LevelID: levelID,
		Walls:   walls,
		Zones:   zones,
	}, nil
}

func (r *planRepo) DeleteByLevelID(ctx context.Context, levelID string) error {
	log := logger.LoggerFromContext(ctx)
	query := `DELETE FROM plans WHERE level_id = ?`
	_, err := r.db.ExecContext(ctx, query, levelID)
	if err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("failed to delete plan by level_id")
		return errors.Join(domain.ErrDatabaseUnavailable, err)
	}
	return nil
}

func (a *Adapter) Save(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	return (&planRepo{db: a.db}).Save(ctx, plan)
}

func (a *Adapter) FindByLevelID(ctx context.Context, levelID string) (domain.Plan, error) {
	return (&planRepo{db: a.db}).FindByLevelID(ctx, levelID)
}

func (a *Adapter) DeleteByLevelID(ctx context.Context, levelID string) error {
	return (&planRepo{db: a.db}).DeleteByLevelID(ctx, levelID)
}
