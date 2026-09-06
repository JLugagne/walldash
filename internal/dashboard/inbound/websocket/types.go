package websocket

import (
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
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
	Device   domain.Device `json:"device"`
}
