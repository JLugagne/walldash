package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func enrollClient(t *testing.T, ctx context.Context, dash *Dashboard, client *authTestClient) (string, string) {
	t.Helper()
	resp := client.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	pending := dash.Auth.ListPending(ctx)
	require.NotEmpty(t, pending)

	resp = client.do(http.MethodPost, "/api/auth/verify", map[string]string{"code": pending[len(pending)-1].Code}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Status string `json:"status"`
		Data   struct {
			Device struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"device"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	resp.Body.Close()
	require.Equal(t, "success", out.Status)
	return out.Data.Device.ID, out.Data.Device.Role
}

func TestSetupEndpointsOwnerView(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)

	ownerID, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	resp := owner.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodGet, "/api/setup/auth/pending", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var pendingResp struct {
		Status string `json:"status"`
		Data   []struct {
			DeviceID  string `json:"device_id"`
			Label     string `json:"label"`
			Code      string `json:"code"`
			ExpiresAt string `json:"expires_at"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pendingResp))
	resp.Body.Close()
	require.Equal(t, "success", pendingResp.Status)
	require.Len(t, pendingResp.Data, 1)
	require.NotEmpty(t, pendingResp.Data[0].DeviceID)
	require.Len(t, pendingResp.Data[0].Code, 6)
	require.NotEmpty(t, pendingResp.Data[0].ExpiresAt)

	resp = owner.do(http.MethodGet, "/api/setup/auth/devices", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var devicesResp struct {
		Status string `json:"status"`
		Data   []struct {
			ID        string  `json:"id"`
			Label     string  `json:"label"`
			Role      string  `json:"role"`
			Status    string  `json:"status"`
			CreatedAt string  `json:"created_at"`
			LastSeen  *string `json:"last_seen"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&devicesResp))
	resp.Body.Close()
	require.Equal(t, "success", devicesResp.Status)
	require.Len(t, devicesResp.Data, 1)
	require.Equal(t, ownerID, devicesResp.Data[0].ID)
	require.Equal(t, "owner", devicesResp.Data[0].Role)
	require.Equal(t, "active", devicesResp.Data[0].Status)
	require.NotEmpty(t, devicesResp.Data[0].CreatedAt)
}

func TestSetupEndpointsDeviceForbidden(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, ownerRole := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", ownerRole)

	device := newAuthTestClient(t, server)
	device.fetchCSRF()
	_, deviceRole := enrollClient(t, ctx, dash, device)
	require.Equal(t, "device", deviceRole)

	resp := device.do(http.MethodGet, "/api/setup/auth/devices", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	requireJSendCode(t, resp, "FORBIDDEN")

	resp = device.do(http.MethodGet, "/api/setup/auth/pending", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	requireJSendCode(t, resp, "FORBIDDEN")
}

func TestSetupEndpointsRevokeKillsRefresh(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	ownerID, _ := enrollClient(t, ctx, dash, owner)

	refresh := owner.cookie(dash.Cookies.RefreshName)
	require.NotEmpty(t, refresh)

	resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+ownerID+"/revoke", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/auth/refresh", nil, func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: dash.Cookies.RefreshName, Value: refresh})
	})
	require.NotEqual(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	acct, err := dash.Auth.GetAccount(ctx, ownerID)
	require.NoError(t, err)
	require.Equal(t, "revoked", string(acct.Status))
}

func TestSetupEndpointsRateLimited(t *testing.T) {
	_, _, _, client := setupAuthServer(t)

	for i := 0; i < 5; i++ {
		resp := client.do(http.MethodPost, "/api/auth/connect", nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	resp := client.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	requireJSendCode(t, resp, "RATE_LIMITED")
}
