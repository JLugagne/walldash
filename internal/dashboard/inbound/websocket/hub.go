package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svcactions "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/actions"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 5 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow local dashboard connections
	},
}

var _ domain.DeviceBroadcaster = (*Hub)(nil)

// Hub maintains active client connections and broadcasts device state events.
type Hub struct {
	clients        map[*Client]bool
	register       chan *Client
	unregister     chan *Client
	broadcast      chan []byte
	stop           chan struct{}
	actionCommands svcactions.ActionCommands
	mu             sync.Mutex
	running        bool
}

// NewHub creates a new WebSocket Hub instance.
func NewHub(actionCommands svcactions.ActionCommands) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte, 256),
		stop:           make(chan struct{}),
		actionCommands: actionCommands,
	}
}

// Run listens for register, unregister and broadcast events until Stop is called.
func (h *Hub) Run() {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	for {
		select {
		case <-h.stop:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.running = false
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Stop stops the Hub event loop.
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		close(h.stop)
		h.running = false
	}
}

// BroadcastDevice serializes a device update and distributes it to all connected clients.
func (h *Hub) BroadcastDevice(device domain.Device) {
	event := StateChangedEvent{
		Type:     "state_changed",
		EntityID: device.ID,
		State:    device.State,
		Device:   device,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.Lock()
	running := h.running
	h.mu.Unlock()

	if running {
		select {
		case h.broadcast <- payload:
		default:
		}
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket connection.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.LoggerFromContext(r.Context()).WithError(err).Warn("failed to upgrade websocket connection")
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Send initial connection confirmation
	initMsg, _ := json.Marshal(map[string]string{
		"type":    "connected",
		"message": "Connected to ha-dash real-time WebSocket",
	})
	client.send <- initMsg

	go client.writePump()
	go client.readPump(r.Context())
}

// Client represents a single WebSocket client session.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var raw map[string]any
		if err := json.Unmarshal(message, &raw); err != nil {
			continue
		}

		msgType, _ := raw["type"].(string)
		switch msgType {
		case "ping":
			pongMsg, _ := json.Marshal(map[string]string{"type": "pong"})
			select {
			case c.send <- pongMsg:
			default:
			}

		case "action":
			var actionMsg ClientActionMessage
			if err := json.Unmarshal(message, &actionMsg); err != nil {
				errMsg, _ := json.Marshal(map[string]string{
					"type":  "error",
					"error": "invalid action message format",
				})
				c.send <- errMsg
				continue
			}

			cmd := domain.ActionCommand{
				EntityID: actionMsg.EntityID,
				Action:   actionMsg.Action,
			}

			// Validate whitelist before execution
			if err := cmd.Validate(); err != nil {
				errMsg, _ := json.Marshal(map[string]string{
					"type":      "error",
					"code":      "ACTION_NOT_ALLOWED",
					"error":     err.Error(),
					"entity_id": actionMsg.EntityID,
				})
				c.send <- errMsg
				continue
			}

			actor := domain.ActorFromContext(ctx)
			if err := c.hub.actionCommands.ExecuteAction(ctx, actor, cmd); err != nil {
				errMsg, _ := json.Marshal(map[string]string{
					"type":      "error",
					"error":     err.Error(),
					"entity_id": actionMsg.EntityID,
				})
				c.send <- errMsg
				continue
			}

			successMsg, _ := json.Marshal(map[string]string{
				"type":      "action_success",
				"entity_id": actionMsg.EntityID,
				"action":    actionMsg.Action,
			})
			c.send <- successMsg
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
