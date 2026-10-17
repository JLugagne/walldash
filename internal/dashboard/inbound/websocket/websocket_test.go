package websocket_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/actions/actionstest"
	ws "github.com/JLugagne/walldash/internal/dashboard/inbound/websocket"
	gorilla_ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketHub_ConnectionAndBroadcast(t *testing.T) {
	mockCommands := &actionstest.MockActionCommands{}
	hub := ws.NewHub(mockCommands)
	go hub.Run()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect first client
	dialer := gorilla_ws.Dialer{}
	conn1, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer conn1.Close()

	// Read initial connected message
	_, msgBytes, err := conn1.ReadMessage()
	require.NoError(t, err)
	var initialMsg map[string]any
	err = json.Unmarshal(msgBytes, &initialMsg)
	require.NoError(t, err)
	assert.Equal(t, "connected", initialMsg["type"])

	// Connect second client
	conn2, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn2.Close()

	_, _, err = conn2.ReadMessage() // initial connected message
	require.NoError(t, err)

	// Broadcast device update
	device := domain.Device{
		ID:     "light.salon_plafond",
		Name:   "Living Room Ceiling Light",
		Domain: domain.DomainLight,
		State:  "off",
		Attributes: map[string]any{
			"brightness": 128,
		},
	}
	hub.BroadcastDevice(device)

	// Verify both clients receive the broadcasted state change
	for idx, conn := range []*gorilla_ws.Conn{conn1, conn2} {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, broadcastBytes, err := conn.ReadMessage()
		require.NoError(t, err, "client %d failed to read broadcast", idx)

		var event ws.StateChangedEvent
		err = json.Unmarshal(broadcastBytes, &event)
		require.NoError(t, err)
		assert.Equal(t, "state_changed", event.Type)
		assert.Equal(t, "light.salon_plafond", event.EntityID)
		assert.Equal(t, "off", event.State)
		assert.Equal(t, "Living Room Ceiling Light", event.Device.Name)
	}
}

func TestWebSocketHub_ClientActionExecution(t *testing.T) {
	var mu sync.Mutex
	var executedCmds []domain.ActionCommand

	mockCommands := &actionstest.MockActionCommands{
		ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
			if err := cmd.Validate(); err != nil {
				return err
			}
			mu.Lock()
			executedCmds = append(executedCmds, cmd)
			mu.Unlock()
			return nil
		},
	}

	hub := ws.NewHub(mockCommands)
	go hub.Run()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := gorilla_ws.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Drain initial connected message
	_, _, _ = conn.ReadMessage()

	t.Run("valid action message triggers ExecuteAction and receives success ack", func(t *testing.T) {
		actionMsg := ws.ClientActionMessage{
			Type:     "action",
			EntityID: "light.salon_plafond",
			Action:   "toggle",
		}
		err := conn.WriteJSON(actionMsg)
		require.NoError(t, err)

		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, respBytes, err := conn.ReadMessage()
		require.NoError(t, err)

		var resp map[string]any
		err = json.Unmarshal(respBytes, &resp)
		require.NoError(t, err)
		assert.Equal(t, "action_success", resp["type"])
		assert.Equal(t, "light.salon_plafond", resp["entity_id"])

		mu.Lock()
		defer mu.Unlock()
		require.Len(t, executedCmds, 1)
		assert.Equal(t, "light.salon_plafond", executedCmds[0].EntityID)
		assert.Equal(t, "toggle", executedCmds[0].Action)
	})

	t.Run("disallowed action returns error message to client", func(t *testing.T) {
		badActionMsg := ws.ClientActionMessage{
			Type:     "action",
			EntityID: "light.salon_plafond",
			Action:   "unauthorized_cmd",
		}
		err := conn.WriteJSON(badActionMsg)
		require.NoError(t, err)

		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, respBytes, err := conn.ReadMessage()
		require.NoError(t, err)

		var resp map[string]any
		err = json.Unmarshal(respBytes, &resp)
		require.NoError(t, err)
		assert.Equal(t, "error", resp["type"])
		assert.Contains(t, resp["error"], "action unauthorized_cmd is not permitted by whitelist")
	})

	t.Run("disallowed domain returns error message to client", func(t *testing.T) {
		badDomainMsg := ws.ClientActionMessage{
			Type:     "action",
			EntityID: "sensor.temperature",
			Action:   "toggle",
		}
		err := conn.WriteJSON(badDomainMsg)
		require.NoError(t, err)

		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, respBytes, err := conn.ReadMessage()
		require.NoError(t, err)

		var resp map[string]any
		err = json.Unmarshal(respBytes, &resp)
		require.NoError(t, err)
		assert.Equal(t, "error", resp["type"])
		assert.Contains(t, resp["error"], "not permitted by action whitelist")
	})
}

func TestWebSocketHub_OriginCheck(t *testing.T) {
	mockCommands := &actionstest.MockActionCommands{}
	hub := ws.NewHub(mockCommands, "http://trusted.local")
	go hub.Run()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("connection without origin succeeds", func(t *testing.T) {
		dialer := gorilla_ws.Dialer{}
		conn, resp, err := dialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	})

	t.Run("connection with matching same-origin succeeds", func(t *testing.T) {
		dialer := gorilla_ws.Dialer{}
		header := http.Header{"Origin": []string{server.URL}}
		conn, resp, err := dialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn.Close()
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	})

	t.Run("connection with allowed origin succeeds", func(t *testing.T) {
		dialer := gorilla_ws.Dialer{}
		header := http.Header{"Origin": []string{"http://trusted.local"}}
		conn, resp, err := dialer.Dial(wsURL, header)
		require.NoError(t, err)
		defer conn.Close()
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	})

	t.Run("connection with unauthorized origin is rejected with 403", func(t *testing.T) {
		dialer := gorilla_ws.Dialer{}
		header := http.Header{"Origin": []string{"http://evil.com"}}
		conn, resp, err := dialer.Dial(wsURL, header)
		if conn != nil {
			conn.Close()
		}
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestBroadcastDevice_UsesSnakeCaseWireShape(t *testing.T) {
	hub := ws.NewHub(&actionstest.MockActionCommands{})
	go hub.Run()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := gorilla_ws.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	_, _, err = conn.ReadMessage() // initial connected frame
	require.NoError(t, err)

	updated := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	hub.BroadcastDevice(domain.Device{
		ID:          "light.kitchen",
		Name:        "Kitchen lamp",
		Domain:      domain.DomainLight,
		State:       "unavailable",
		Attributes:  map[string]any{"brightness": 128},
		LastUpdated: updated,
	})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, frame, err := conn.ReadMessage()
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(frame, &payload))
	device, ok := payload["device"].(map[string]any)
	require.True(t, ok, "frame must carry a device object")

	// The browser merges the frame into the same device map the REST fetch produced,
	// keyed by these snake_case field names.
	assert.Equal(t, "light.kitchen", device["id"])
	assert.Equal(t, "Kitchen lamp", device["name"])
	assert.Equal(t, "light", device["domain"])
	assert.Equal(t, "unavailable", device["state"])
	assert.NotNil(t, device["attributes"])
	assert.Equal(t, "2026-01-01T12:00:00Z", device["last_updated"])
	assert.NotContains(t, device, "ID")
	assert.NotContains(t, device, "LastUpdated")
}
