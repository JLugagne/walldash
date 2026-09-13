package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/JLugagne/egauth/tokens/issuertest"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/actions/actionstest"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSubjectA = "11111111-1111-1111-1111-111111111111"
	testSubjectB = "22222222-2222-2222-2222-222222222222"
)

func newAuthenticatedWSServer(t *testing.T, hub *Hub) *httptest.Server {
	t.Helper()
	verifier := &issuertest.MockVerifier[struct{}]{
		VerifyAccessTokenForTenantFunc: func(ctx context.Context, tenantID string, token string) (*tokens.Claims[struct{}], error) {
			switch token {
			case "token-a":
				return &tokens.Claims[struct{}]{Subject: uuid.MustParse(testSubjectA)}, nil
			case "token-b":
				return &tokens.Claims[struct{}]{Subject: uuid.MustParse(testSubjectB)}, nil
			default:
				return nil, errors.New("unknown access token")
			}
		},
	}
	cookies := tokens.DefaultCookies()
	handler := basic.ContextMiddleware(verifier, http.HandlerFunc(hub.ServeWS), tokens.WithCookieAuth[struct{}](cookies))
	return httptest.NewServer(handler)
}

func dialAuthenticatedWS(t *testing.T, serverURL string, token string) *websocket.Conn {
	t.Helper()
	cookies := tokens.DefaultCookies()
	header := http.Header{"Cookie": []string{cookies.AccessName + "=" + token}}
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial("ws"+strings.TrimPrefix(serverURL, "http"), header)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	return conn
}

func readConnectedFrame(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)
	var frame map[string]any
	require.NoError(t, json.Unmarshal(message, &frame))
	require.Equal(t, "connected", frame["type"])
}

func expectSessionRevokedDisconnect(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, response, err := conn.ReadMessage()
	require.NoError(t, err)
	var frame map[string]any
	require.NoError(t, json.Unmarshal(response, &frame))
	assert.Equal(t, "error", frame["type"])
	assert.Equal(t, "SESSION_REVOKED", frame["code"])

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	require.Error(t, err, "hub must close the connection")
}

func TestSessionValidatorRejectsRevokedSubject(t *testing.T) {
	var mu sync.Mutex
	var executed []domain.ActionCommand
	mockCommands := &actionstest.MockActionCommands{
		ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
			mu.Lock()
			executed = append(executed, cmd)
			mu.Unlock()
			return nil
		},
	}

	validatorSubjects := make(chan string, 1)
	hub := NewHub(mockCommands)
	hub.SetSessionValidator(func(ctx context.Context, subject string) error {
		validatorSubjects <- subject
		return errors.New("account is missing or revoked")
	})
	go hub.Run()
	defer hub.Stop()

	server := newAuthenticatedWSServer(t, hub)
	defer server.Close()

	conn := dialAuthenticatedWS(t, server.URL, "token-a")
	defer conn.Close()
	readConnectedFrame(t, conn)

	require.NoError(t, conn.WriteJSON(ClientActionMessage{
		Type:     "action",
		EntityID: "light.salon_plafond",
		Action:   "toggle",
	}))
	expectSessionRevokedDisconnect(t, conn)

	select {
	case subject := <-validatorSubjects:
		assert.Equal(t, testSubjectA, subject)
	default:
		t.Fatal("session validator was not called with the client subject")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, executed, "action must not execute for a revoked session")
}

func TestSessionValidatorAllowsActiveSubject(t *testing.T) {
	var mu sync.Mutex
	var executed []domain.ActionCommand
	var validated []string
	mockCommands := &actionstest.MockActionCommands{
		ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
			mu.Lock()
			executed = append(executed, cmd)
			mu.Unlock()
			return nil
		},
	}

	hub := NewHub(mockCommands)
	hub.SetSessionValidator(func(ctx context.Context, subject string) error {
		mu.Lock()
		validated = append(validated, subject)
		mu.Unlock()
		return nil
	})
	go hub.Run()
	defer hub.Stop()

	server := newAuthenticatedWSServer(t, hub)
	defer server.Close()

	conn := dialAuthenticatedWS(t, server.URL, "token-a")
	defer conn.Close()
	readConnectedFrame(t, conn)

	require.NoError(t, conn.WriteJSON(ClientActionMessage{
		Type:     "action",
		EntityID: "light.salon_plafond",
		Action:   "toggle",
	}))

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, response, err := conn.ReadMessage()
	require.NoError(t, err)
	var frame map[string]any
	require.NoError(t, json.Unmarshal(response, &frame))
	assert.Equal(t, "action_success", frame["type"])
	assert.Equal(t, "light.salon_plafond", frame["entity_id"])
	assert.Equal(t, "toggle", frame["action"])

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, executed, 1)
	assert.Equal(t, "light.salon_plafond", executed[0].EntityID)
	assert.Equal(t, "toggle", executed[0].Action)
	assert.Equal(t, []string{testSubjectA}, validated)
}

func TestSessionValidatorRejectsUnboundSubject(t *testing.T) {
	var executed atomic.Bool
	mockCommands := &actionstest.MockActionCommands{
		ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
			executed.Store(true)
			return nil
		},
	}

	hub := NewHub(mockCommands)
	hub.SetSessionValidator(func(ctx context.Context, subject string) error {
		return nil
	})
	go hub.Run()
	defer hub.Stop()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close()
	readConnectedFrame(t, conn)

	require.NoError(t, conn.WriteJSON(ClientActionMessage{
		Type:     "action",
		EntityID: "light.salon_plafond",
		Action:   "toggle",
	}))
	expectSessionRevokedDisconnect(t, conn)

	assert.False(t, executed.Load(), "unbound subject must be rejected")
}

func TestCloseUserClosesOnlyMatchingSubject(t *testing.T) {
	mockCommands := &actionstest.MockActionCommands{
		ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
			return nil
		},
	}
	hub := NewHub(mockCommands)
	go hub.Run()
	defer hub.Stop()

	server := newAuthenticatedWSServer(t, hub)
	defer server.Close()

	connA := dialAuthenticatedWS(t, server.URL, "token-a")
	defer connA.Close()
	readConnectedFrame(t, connA)

	connB := dialAuthenticatedWS(t, server.URL, "token-b")
	defer connB.Close()
	readConnectedFrame(t, connB)

	hub.CloseUser(testSubjectA)

	expectSessionRevokedDisconnect(t, connA)

	require.NoError(t, connB.WriteJSON(map[string]string{"type": "ping"}))
	_ = connB.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, messageB, err := connB.ReadMessage()
	require.NoError(t, err)
	var frameB map[string]any
	require.NoError(t, json.Unmarshal(messageB, &frameB))
	assert.Equal(t, "pong", frameB["type"], "client B must stay connected")
}
