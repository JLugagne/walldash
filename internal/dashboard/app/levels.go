package app

import (
	"context"
	"errors"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/google/uuid"
)

// ListLevels retrieves all levels ordered by order ascending.
func (a *App) ListLevels(ctx context.Context) ([]domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	levels, err := a.levelsRepo.FindAll(ctx)
	if err != nil {
		log.WithError(err).Error("failed to list levels")
		return nil, err
	}
	return levels, nil
}

// GetLevel retrieves a level by its identifier.
func (a *App) GetLevel(ctx context.Context, id string) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)
	level, err := a.levelsRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("level_id", id).Error("failed to get level")
		return domain.Level{}, err
	}
	return level, nil
}

// GetPlan retrieves the 2D architectural blueprint for a level.
func (a *App) GetPlan(ctx context.Context, levelID string) (domain.Plan, error) {
	log := logger.LoggerFromContext(ctx)
	plan, err := a.plansRepo.FindByLevelID(ctx, levelID)
	if err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("failed to get plan")
		return domain.Plan{}, err
	}
	return plan, nil
}

// CreateLevel validates and persists a new level, generating an ID if omitted.
func (a *App) CreateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(level.Name) == "" {
		return domain.Level{}, errors.Join(domain.ErrInvalidLevel, errors.New("level name is required"))
	}
	if strings.TrimSpace(level.ID) == "" {
		level.ID = uuid.NewString()
	}
	if len(level.Layers) == 0 {
		level.Layers = append([]domain.Layer(nil), domain.DefaultLayers...)
	}

	// Auto-assign order if order is 0 and there are other levels
	existing, err := a.levelsRepo.FindAll(ctx)
	if err == nil && len(existing) > 0 && level.Order == 0 {
		maxOrder := 0
		for _, l := range existing {
			if l.Order >= maxOrder {
				maxOrder = l.Order + 1
			}
		}
		level.Order = maxOrder
	}

	created, err := a.levelsRepo.Create(ctx, level)
	if err != nil {
		log.WithError(err).WithField("level_id", level.ID).Error("failed to create level")
		return domain.Level{}, err
	}

	log.WithField("level_id", created.ID).WithField("actor", actor.UserID).Info("level created successfully")
	return created, nil
}

// UpdateLevel validates and updates an existing level entity.
func (a *App) UpdateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
	log := logger.LoggerFromContext(ctx)

	if err := level.Validate(); err != nil {
		return domain.Level{}, err
	}

	if len(level.Layers) == 0 {
		level.Layers = append([]domain.Layer(nil), domain.DefaultLayers...)
	}

	// Reassign placements to default layer if any layer was removed
	if a.placementsRepo != nil && a.levelsRepo != nil {
		existing, err := a.levelsRepo.FindByID(ctx, level.ID)
		if err == nil {
			newLayersMap := make(map[string]bool, len(level.Layers))
			for _, l := range level.Layers {
				newLayersMap[l.Name] = true
			}
			for _, oldL := range existing.Layers {
				if !newLayersMap[oldL.Name] {
					if err := a.placementsRepo.ReassignLayer(ctx, level.ID, oldL.Name, domain.DefaultPlacementLayer); err != nil {
						log.WithError(err).WithField("level_id", level.ID).WithField("old_layer", oldL.Name).Error("failed to reassign placements layer")
						return domain.Level{}, errors.Join(domain.ErrDatabaseUnavailable, err)
					}
				}
			}
		}
	}

	updated, err := a.levelsRepo.Update(ctx, level)
	if err != nil {
		log.WithError(err).WithField("level_id", level.ID).Error("failed to update level")
		return domain.Level{}, err
	}

	log.WithField("level_id", updated.ID).WithField("actor", actor.UserID).Info("level updated successfully")
	return updated, nil
}

// DeleteLevel removes a level and its associated plan.
func (a *App) DeleteLevel(ctx context.Context, actor domain.Actor, id string) error {
	log := logger.LoggerFromContext(ctx)

	if strings.TrimSpace(id) == "" {
		return errors.Join(domain.ErrInvalidLevel, errors.New("level id is required"))
	}

	if err := a.levelsRepo.Delete(ctx, id); err != nil {
		log.WithError(err).WithField("level_id", id).Error("failed to delete level")
		return err
	}

	log.WithField("level_id", id).WithField("actor", actor.UserID).Info("level deleted successfully")
	return nil
}

// ReorderLevels modifies the ordering position for a list of level IDs.
func (a *App) ReorderLevels(ctx context.Context, actor domain.Actor, orderedIDs []string) error {
	log := logger.LoggerFromContext(ctx)

	if len(orderedIDs) == 0 {
		return errors.Join(domain.ErrInvalidLevel, errors.New("ordered IDs cannot be empty"))
	}

	if err := a.levelsRepo.Reorder(ctx, orderedIDs); err != nil {
		log.WithError(err).Error("failed to reorder levels")
		return err
	}

	log.WithField("count", len(orderedIDs)).WithField("actor", actor.UserID).Info("levels reordered successfully")
	return nil
}

// SavePlan validates and stores the 2D architectural plan for an existing level.
func (a *App) SavePlan(ctx context.Context, actor domain.Actor, plan domain.Plan) (domain.Plan, error) {
	log := logger.LoggerFromContext(ctx)

	// Verify parent level existence
	if _, err := a.levelsRepo.FindByID(ctx, plan.LevelID); err != nil {
		log.WithError(err).WithField("level_id", plan.LevelID).Error("cannot save plan: parent level not found")
		return domain.Plan{}, err
	}

	if err := plan.Validate(); err != nil {
		return domain.Plan{}, err
	}

	saved, err := a.plansRepo.Save(ctx, plan)
	if err != nil {
		log.WithError(err).WithField("level_id", plan.LevelID).Error("failed to save plan")
		return domain.Plan{}, err
	}

	log.WithField("level_id", saved.LevelID).WithField("actor", actor.UserID).Info("plan saved successfully")
	return saved, nil
}
