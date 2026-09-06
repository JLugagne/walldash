package dashboard

import (
	"time"

	"github.com/JLugagne/walldash/domain"
)

// DeviceResponse represents a Home Assistant device/entity in API responses.
type DeviceResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Domain      string         `json:"domain"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastUpdated time.Time      `json:"last_updated"`
}

// DevicePlacementResponse represents a placed device on a level's plan in API responses.
type DevicePlacementResponse struct {
	ID           string    `json:"id"`
	LevelID      string    `json:"level_id"`
	DeviceID     string    `json:"device_id"`
	X            float64   `json:"x"`
	Y            float64   `json:"y"`
	Icon         string    `json:"icon,omitempty"`
	RenderDomain string    `json:"render_domain,omitempty"`
	CustomName   string    `json:"custom_name,omitempty"`
	Layer        string    `json:"layer,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SavePlacementRequest contains payload to place or move a device on a level plan.
type SavePlacementRequest struct {
	ID           string  `json:"id,omitempty"`
	DeviceID     string  `json:"device_id" validate:"required"`
	X            float64 `json:"x" validate:"gte=0"`
	Y            float64 `json:"y" validate:"gte=0"`
	Icon         string  `json:"icon,omitempty"`
	RenderDomain string  `json:"render_domain,omitempty" validate:"omitempty,oneof=light switch media_player sensor climate"`
	CustomName   string  `json:"custom_name,omitempty"`
	Layer        string  `json:"layer,omitempty"`
}

// Validation errors for dashboard public device types
var (
	ErrInvalidSavePlacementRequest = &domain.Error{
		Code:    "INVALID_SAVE_PLACEMENT_REQUEST",
		Message: "invalid save placement request data",
	}
)
