package devices

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// DeviceQueries defines read-only operations for Home Assistant devices and level placements.
type DeviceQueries interface {
	ListAvailableDevices(ctx context.Context) ([]domain.Device, error)
	ListPlacements(ctx context.Context, levelID string) ([]domain.DevicePlacement, error)
	GetPlacement(ctx context.Context, id string) (domain.DevicePlacement, error)
}

// DeviceCommands defines mutating operations for device placements, extending DeviceQueries.
type DeviceCommands interface {
	DeviceQueries
	SavePlacement(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error)
	DeletePlacement(ctx context.Context, actor domain.Actor, levelID string, placementID string) error
}
