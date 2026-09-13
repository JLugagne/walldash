package dashboard

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPendingApprovalFlow(t *testing.T) {
	_, _, server, owner := setupAuthServer(t)

	resp := owner.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, readBody(t, resp), `"role":"owner"`)

	joiner := newAuthTestClient(t, server)
	joiner.fetchCSRF()

	resp = joiner.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, readBody(t, resp), `"status":"pending"`)

	resp = joiner.do(http.MethodPost, "/api/auth/redeem", nil, nil)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodGet, "/api/setup/auth/pending", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var pendingPayload struct {
		Data []struct {
			PendingID string `json:"pending_id"`
			Label     string `json:"label"`
			Approved  bool   `json:"approved"`
			Code      string `json:"code"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pendingPayload))
	resp.Body.Close()
	require.Len(t, pendingPayload.Data, 1)
	require.Empty(t, pendingPayload.Data[0].Code)
	require.False(t, pendingPayload.Data[0].Approved)
	pendingID := pendingPayload.Data[0].PendingID
	require.NotEmpty(t, pendingID)

	resp = owner.do(http.MethodPost, "/api/setup/auth/pending/"+pendingID+"/approve", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = joiner.do(http.MethodPost, "/api/auth/redeem", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	redeemBody := readBody(t, resp)
	require.Contains(t, redeemBody, `"status":"authenticated"`)
	require.Contains(t, redeemBody, `"role":"device"`)

	resp = joiner.do(http.MethodPost, "/api/auth/redeem", nil, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

}

func TestInviteEnrollmentFlow(t *testing.T) {
	_, _, server, owner := setupAuthServer(t)

	resp := owner.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/setup/auth/invites", map[string]string{"role": "admin"}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var invitePayload struct {
		Data struct {
			Token string `json:"token"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&invitePayload))
	resp.Body.Close()
	require.NotEmpty(t, invitePayload.Data.Token)
	require.Equal(t, "admin", invitePayload.Data.Role)

	joiner := newAuthTestClient(t, server)
	joiner.fetchCSRF()
	resp = joiner.do(http.MethodPost, "/api/auth/invite/redeem", map[string]string{"token": invitePayload.Data.Token}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, readBody(t, resp), `"role":"admin"`)

	reuse := newAuthTestClient(t, server)
	reuse.fetchCSRF()
	resp = reuse.do(http.MethodPost, "/api/auth/invite/redeem", map[string]string{"token": invitePayload.Data.Token}, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	requireJSendCode(t, resp, "invalid_invite")

	resp = reuse.do(http.MethodPost, "/api/auth/invite/redeem", map[string]string{"token": "not-a-token"}, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	requireJSendCode(t, resp, "invalid_invite")
}

func TestInviteRejectsOwnerRole(t *testing.T) {
	_, _, _, owner := setupAuthServer(t)

	resp := owner.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/setup/auth/invites", map[string]string{"role": "owner"}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	requireJSendCode(t, resp, "INVALID_ROLE")
}

func TestRedeemRejectsMissingEnrollment(t *testing.T) {
	_, _, _, client := setupAuthServer(t)

	resp := client.do(http.MethodPost, "/api/auth/redeem", nil, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	requireJSendCode(t, resp, "invalid_enrollment")
}
