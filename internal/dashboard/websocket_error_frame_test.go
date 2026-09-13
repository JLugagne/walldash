package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// dialDashboardWS opens an authenticated socket to /api/ws using the raw access cookie, so the
// test controls exactly which credential is presented.
func dialDashboardWS(t *testing.T, server *httptest.Server, dash *Dashboard, client *authTestClient) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws"
	header := http.Header{}
	header.Set("Cookie", dash.Cookies.AccessName+"="+client.cookie(dash.Cookies.AccessName))

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if resp != nil {
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "websocket handshake failed")
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// readUntilError reads frames until an "error" frame arrives and returns it.
func readUntilError(t *testing.T, conn *websocket.Conn) map[string]string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		var frame map[string]string
		require.NoError(t, json.Unmarshal(message, &frame))
		if frame["type"] == "error" {
			return frame
		}
	}
	t.Fatal("no error frame arrived over the websocket")
	return nil
}

// TestWebSocketActionErrorsDoNotLeakTheHomeAssistantURL is the regression test for the audit
// finding "WebSocket error frames echo the operator-configured Home Assistant URL and internal
// host to any device". The HTTP twin (POST /api/actions) answers with a domain code and a generic
// message; the socket must not be the one place where raw internal errors escape.
func TestWebSocketActionErrorsDoNotLeakTheHomeAssistantURL(t *testing.T) {
	ctx := context.Background()
	router := mux.NewRouter()

	// A closed loopback port makes the outbound call fail immediately, with the URL inside the
	// transport error.
	const haURL = "http://127.0.0.1:1"
	dash, err := New(ctx, Config{
		DBPath:      filepath.Join(t.TempDir(), "ws.db"),
		Version:     "test",
		HAUrl:       haURL,
		HAToken:     "ha-token-for-the-test",
		TokenSecret: restartTokenSecret,
	}, router)
	require.NoError(t, err)
	defer func() { _ = dash.Close() }()

	server := httptest.NewServer(router)
	defer server.Close()

	owner := newAuthTestClient(t, server)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	conn := dialDashboardWS(t, server, dash, owner)
	require.NoError(t, conn.WriteJSON(map[string]string{
		"type":      "action",
		"entity_id": "light.kitchen",
		"action":    "toggle",
	}))

	frame := readUntilError(t, conn)
	t.Logf("error frame: %v", frame)

	require.NotContains(t, frame["error"], "127.0.0.1", "the frame must not carry the HA host")
	require.NotContains(t, frame["error"], haURL, "the frame must not carry the HA URL")
	require.NotContains(t, frame["error"], "/api/services/", "the frame must not carry the HA request path")
	require.Equal(t, "HEALTH_CHECK_FAILED", frame["code"], "the frame should carry the domain code")
}
