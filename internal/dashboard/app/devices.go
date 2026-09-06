package app

import (
	"context"
	"errors"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/pkg/logger"
	"github.com/google/uuid"
)

// ListAvailableDevices retrieves all supported devices from Home Assistant.
func (a *App) ListAvailableDevices(ctx context.Context) ([]domain.Device, error) {
	log := logger.LoggerFromContext(ctx)
	devices, err := a.haRepo.GetStates(ctx)
	if err != nil {
		log.WithError(err).Error("failed to list available devices from Home Assistant")
		return nil, err
	}
	return devices, nil
}

// ListPlacements retrieves all device placements for a specific level.
func (a *App) ListPlacements(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)

	// Ensure level exists
	if _, err := a.levelsRepo.FindByID(ctx, levelID); err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("level not found when listing placements")
		return nil, err
	}

	placements, err := a.placementsRepo.FindPlacementsByLevelID(ctx, levelID)
	if err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("failed to list device placements")
		return nil, err
	}
	return placements, nil
}

// GetPlacement retrieves a device placement by its identifier.
func (a *App) GetPlacement(ctx context.Context, id string) (domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)
	p, err := a.placementsRepo.FindPlacementByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("placement_id", id).Error("failed to get device placement")
		return domain.DevicePlacement{}, err
	}
	return p, nil
}

// SavePlacement persists or updates a device placement on a level plan.
func (a *App) SavePlacement(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
	log := logger.LoggerFromContext(ctx)

	// Ensure target level exists
	level, err := a.levelsRepo.FindByID(ctx, placement.LevelID)
	if err != nil {
		log.WithError(err).WithField("level_id", placement.LevelID).Error("level not found for placement")
		return domain.DevicePlacement{}, err
	}

	if strings.TrimSpace(placement.ID) == "" {
		placement.ID = uuid.NewString()
	}

	if strings.TrimSpace(placement.Layer) == "" {
		placement.Layer = domain.DefaultPlacementLayer
	}

	if err := placement.Validate(); err != nil {
		return domain.DevicePlacement{}, err
	}

	// Ensure level layers include the placement layer
	containsLayer := false
	for _, l := range level.Layers {
		if l == placement.Layer {
			containsLayer = true
			break
		}
	}
	if !containsLayer && len(level.Layers) > 0 {
		level.Layers = append(level.Layers, placement.Layer)
		if _, err := a.levelsRepo.Update(ctx, level); err != nil {
			log.WithError(err).WithField("level_id", level.ID).WithField("layer", placement.Layer).Warn("failed to update level with new layer")
		}
	}

	saved, err := a.placementsRepo.SavePlacement(ctx, placement)
	if err != nil {
		log.WithError(err).WithField("placement_id", placement.ID).Error("failed to save placement")
		return domain.DevicePlacement{}, err
	}

	log.WithField("placement_id", saved.ID).WithField("actor", actor.UserID).Info("device placement saved successfully")
	return saved, nil
}

// DeletePlacement removes a device placement from a level plan.
func (a *App) DeletePlacement(ctx context.Context, actor domain.Actor, levelID string, placementID string) error {
	log := logger.LoggerFromContext(ctx)

	// Ensure target level exists
	if _, err := a.levelsRepo.FindByID(ctx, levelID); err != nil {
		log.WithError(err).WithField("level_id", levelID).Error("level not found for placement deletion")
		return err
	}

	existing, err := a.placementsRepo.FindPlacementByID(ctx, placementID)
	if err != nil {
		log.WithError(err).WithField("placement_id", placementID).Error("placement not found for deletion")
		return err
	}

	if existing.LevelID != levelID {
		return errors.Join(domain.ErrPlacementNotFound, errors.New("placement does not belong to specified level"))
	}

	if err := a.placementsRepo.DeletePlacement(ctx, placementID); err != nil {
		log.WithError(err).WithField("placement_id", placementID).Error("failed to delete placement")
		return err
	}

	log.WithField("placement_id", placementID).WithField("actor", actor.UserID).Info("device placement deleted successfully")
	return nil
}
