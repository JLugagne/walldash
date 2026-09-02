package ha

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// HomeAssistantRepository defines operations for fetching states from Home Assistant.
type HomeAssistantRepository interface {
	GetStates(ctx context.Context) ([]domain.Device, error)
	GetState(ctx context.Context, entityID string) (domain.Device, error)
}
