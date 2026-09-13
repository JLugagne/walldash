package dashboard

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func inviteClientRole(t *testing.T, owner *authTestClient, server *httptest.Server, role string) (*authTestClient, string) {
	t.Helper()
	resp := owner.do(http.MethodPost, "/api/setup/auth/invites", map[string]string{"role": role}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var created struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	require.NotEmpty(t, created.Data.Token)

	client := newAuthTestClient(t, server)
	resp = client.do(http.MethodPost, "/api/auth/invite/redeem", map[string]string{"token": created.Data.Token}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var redeemed struct {
		Data struct {
			Device struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"device"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&redeemed))
	resp.Body.Close()
	require.Equal(t, role, redeemed.Data.Device.Role)
	require.NotEmpty(t, redeemed.Data.Device.ID)
	return client, redeemed.Data.Device.ID
}

func TestAdminCannotRevokeOwner(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	ownerID, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	admin, _ := inviteClientRole(t, owner, server, "admin")

	resp := admin.do(http.MethodPost, "/api/setup/auth/devices/"+ownerID+"/revoke", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode,
		"an admin must never be allowed to revoke an owner account")
	requireJSendCode(t, resp, "FORBIDDEN")

	acct, err := dash.Auth.GetAccount(ctx, ownerID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusActive, acct.Status)

	resp = owner.do(http.MethodPost, "/api/auth/refresh", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestAdminCannotRelabelOwner(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	ownerID, _ := enrollClient(t, ctx, dash, owner)
	before, err := dash.Auth.GetAccount(ctx, ownerID)
	require.NoError(t, err)

	admin, _ := inviteClientRole(t, owner, server, "admin")

	resp := admin.do(http.MethodPost, "/api/setup/auth/devices/"+ownerID+"/label", map[string]string{"label": "pwned"}, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode,
		"an admin must never be allowed to relabel an owner account")
	requireJSendCode(t, resp, "FORBIDDEN")

	after, err := dash.Auth.GetAccount(ctx, ownerID)
	require.NoError(t, err)
	require.Equal(t, before.Label, after.Label)
}

func TestOwnerCanRevokeAdmin(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	_, adminID := inviteClientRole(t, owner, server, "admin")

	resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/revoke", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	acct, err := dash.Auth.GetAccount(ctx, adminID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusRevoked, acct.Status)
}

func TestLastOwnerCannotRevokeOrDemoteItself(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	ownerID, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+ownerID+"/revoke", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode,
		"the last active owner must not be able to revoke itself")
	requireJSendCode(t, resp, "FORBIDDEN")

	resp = owner.do(http.MethodPost, "/api/setup/auth/devices/"+ownerID+"/role", map[string]string{"role": "device"}, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode,
		"the last active owner must not be able to demote itself")
	requireJSendCode(t, resp, "FORBIDDEN")

	acct, err := dash.Auth.GetAccount(ctx, ownerID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusActive, acct.Status)
	require.Equal(t, domain.RoleOwner, acct.Role)
}

func TestDemotedAdminCannotRestoreAdminRole(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	admin, adminID := inviteClientRole(t, owner, server, "admin")

	resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/role", map[string]string{"role": "device"}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	acct, err := dash.Auth.GetAccount(ctx, adminID)
	require.NoError(t, err)
	require.Equal(t, domain.RoleDevice, acct.Role)

	resp = admin.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/role", map[string]string{"role": "admin"}, nil)
	body := readBody(t, resp)
	acct, err = dash.Auth.GetAccount(ctx, adminID)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode,
		"a demoted admin must not be able to restore its admin role (got HTTP %d, %s)", resp.StatusCode, body)
	require.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, resp.StatusCode)
	require.Equal(t, domain.RoleDevice, acct.Role)
}

func TestDemotedAdminStaleCookieGetsForbidden(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	admin, adminID := inviteClientRole(t, owner, server, "admin")

	_, err := dash.Accounts.SetRole(ctx, adminID, domain.RoleDevice)
	require.NoError(t, err)

	resp := admin.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/role", map[string]string{"role": "admin"}, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode,
		"a stale admin session must be rejected by the DB-fresh role check")
	requireJSendCode(t, resp, "FORBIDDEN")

	acct, err := dash.Auth.GetAccount(ctx, adminID)
	require.NoError(t, err)
	require.Equal(t, domain.RoleDevice, acct.Role)
}

func TestRevokedDeviceWebSocketIsTerminated(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	admin, adminID := inviteClientRole(t, owner, server, "admin")

	wsURL := "wss" + strings.TrimPrefix(server.URL, "https") + "/api/ws"
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	header := http.Header{"Cookie": []string{dash.Cookies.AccessName + "=" + admin.cookie(dash.Cookies.AccessName)}}
	conn, wsResp, err := dialer.Dial(wsURL, header)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, wsResp.StatusCode)
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "connected")

	resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/revoke", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	deadline := time.Now().Add(3 * time.Second)
	terminated := false
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		_, frame, readErr := conn.ReadMessage()
		if readErr != nil {
			terminated = true
			break
		}
		require.NotContains(t, string(frame), "action_success")
	}
	require.True(t, terminated, "a revoked device's websocket must be terminated, not left open")
}
