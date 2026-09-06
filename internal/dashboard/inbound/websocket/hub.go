package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, r.Host)
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
	allowedOrigins []string
	upgrader       websocket.Upgrader
	mu             sync.Mutex
	running        bool
}

// NewHub creates a new WebSocket Hub instance with optional allowed origins for CORS/CSWSH protection.
func NewHub(actionCommands svcactions.ActionCommands, allowedOrigins ...string) *Hub {
	h := &Hub{
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan []byte, 256),
		stop:           make(chan struct{}),
		actionCommands: actionCommands,
		allowedOrigins: allowedOrigins,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.CheckOrigin,
	}
	return h
}

// CheckOrigin verifies that the WebSocket request origin matches the server host or allowed origins.
func (h *Hub) CheckOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if strings.EqualFold(u.Host, r.Host) {
		return true
	}

	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || strings.EqualFold(u.Host, allowed) || strings.EqualFold(origin, allowed) {
			return true
		}
		if allowedURL, err := url.Parse(allowed); err == nil && strings.EqualFold(u.Host, allowedURL.Host) {
			return true
		}
	}

	return false
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
				client.closeSend()
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
				client.closeSend()
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				client.sendMsg(message)
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
	conn, err := h.upgrader.Upgrade(w, r, nil)
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
	client.sendMsg(initMsg)

	// Detach from HTTP request cancellation so readPump has a valid session context
	sessionCtx := context.WithoutCancel(r.Context())
	go client.writePump()
	go client.readPump(sessionCtx)
}

// Client represents a single WebSocket client session.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
	closed bool
}

func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.send)
	}
}

func (c *Client) sendMsg(msg []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
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
			c.sendMsg(pongMsg)

		case "action":
			var actionMsg ClientActionMessage
			if err := json.Unmarshal(message, &actionMsg); err != nil {
				errMsg, _ := json.Marshal(map[string]string{
					"type":  "error",
					"error": "invalid action message format",
				})
				c.sendMsg(errMsg)
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
				c.sendMsg(errMsg)
				continue
			}

			actor := domain.ActorFromContext(ctx)
			actionCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
			err := c.hub.actionCommands.ExecuteAction(actionCtx, actor, cmd)
			cancel()
			if err != nil {
				errMsg, _ := json.Marshal(map[string]string{
					"type":      "error",
					"error":     err.Error(),
					"entity_id": actionMsg.EntityID,
				})
				c.sendMsg(errMsg)
				continue
			}

			successMsg, _ := json.Marshal(map[string]string{
				"type":      "action_success",
				"entity_id": actionMsg.EntityID,
				"action":    actionMsg.Action,
			})
			c.sendMsg(successMsg)
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
