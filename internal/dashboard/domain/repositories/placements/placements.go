package placements

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// DevicePlacementRepository defines operations for storing and retrieving device placements on level plans.
type DevicePlacementRepository interface {
	SavePlacement(ctx context.Context, placement domain.DevicePlacement) (domain.DevicePlacement, error)
	FindPlacementByID(ctx context.Context, id string) (domain.DevicePlacement, error)
	FindPlacementsByLevelID(ctx context.Context, levelID string) ([]domain.DevicePlacement, error)
	DeletePlacement(ctx context.Context, id string) error
}
