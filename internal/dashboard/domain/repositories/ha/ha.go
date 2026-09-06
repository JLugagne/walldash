package ha

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// HomeAssistantRepository defines operations for fetching states from Home Assistant and calling services.
type HomeAssistantRepository interface {
	GetStates(ctx context.Context) ([]domain.Device, error)
	GetState(ctx context.Context, entityID string) (domain.Device, error)
	CallService(ctx context.Context, domain string, service string, entityID string) error
}
