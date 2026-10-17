package websocket

import (
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// ClientActionMessage represents an incoming action message sent over WebSocket.
type ClientActionMessage struct {
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
	Action   string `json:"action"`
}

// StateChangedEvent represents an outbound real-time broadcast message.
type StateChangedEvent struct {
	Type     string        `json:"type"`
	EntityID string        `json:"entity_id"`
	State    string        `json:"state"`
	Device   DevicePayload `json:"device"`
}

// DevicePayload is the wire shape of a device inside a state_changed frame. It mirrors the
// REST device DTO (snake_case) so the browser merges the frame into the same device map the
// REST fetch produced; serializing the domain entity directly would emit Go field names.
type DevicePayload struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Domain      string         `json:"domain"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastUpdated time.Time      `json:"last_updated"`
}

// newDevicePayload maps a domain device to its snake_case wire shape.
func newDevicePayload(d domain.Device) DevicePayload {
	return DevicePayload{
		ID:          d.ID,
		Name:        d.Name,
		Domain:      d.Domain,
		State:       d.State,
		Attributes:  d.Attributes,
		LastUpdated: d.LastUpdated,
	}
}
