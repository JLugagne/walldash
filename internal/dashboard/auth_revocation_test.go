package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// restartTokenSecret keeps the JWT signing key stable across the simulated restarts below. In the
// shipped configuration the key is TOKEN_SECRET or an auto-generated secret sealed with the KEK
// file next to the database, so it survives a restart either way.
const restartTokenSecret = "audit-regression-token-secret-0123456789"

// bootDashboardAt starts a dashboard over an existing database file, so a test can stop one
// instance and boot another against the same state to simulate a process restart.
func bootDashboardAt(t *testing.T, ctx context.Context, dbPath string) (*Dashboard, *httptest.Server) {
	t.Helper()
	router := mux.NewRouter()
	dash, err := New(ctx, Config{DBPath: dbPath, Version: "test", TokenSecret: restartTokenSecret}, router)
	require.NoError(t, err)
	return dash, httptest.NewTLSServer(router)
}

// rawRequest sends a request carrying exactly the supplied Cookie header and nothing else, so no
// cookie jar can rotate or clear credentials between attempts.
func rawRequest(t *testing.T, server *httptest.Server, method, path, cookieHeader string, body any) (int, string) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, server.URL+path, reader)
	require.NoError(t, err)
	req.Header.Set("Origin", server.URL)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(raw)
}

// TestAccessTokenRevocationSurvivesRestart is the regression test for the audit finding
// "Access-token revocation does not survive a restart". Revocation cutoffs lived only in the
// in-process tracker built by dashboard.New, so after any restart the still-unexpired access JWT
// of a revoked, demoted or logged-out account was accepted again — and an ex-admin's stale token
// still carried setup:manage.
func TestAccessTokenRevocationSurvivesRestart(t *testing.T) {
	ctx := context.Background()

	t.Run("a revoked device stays revoked", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "revoked.db")
		dash1, server1 := bootDashboardAt(t, ctx, dbPath)

		owner := newAuthTestClient(t, server1)
		_, role := enrollClient(t, ctx, dash1, owner)
		require.Equal(t, "owner", role)

		device := newAuthTestClient(t, server1)
		deviceID, deviceRole := enrollClient(t, ctx, dash1, device)
		require.Equal(t, "device", deviceRole)

		staleCookie := dash1.Cookies.AccessName + "=" + device.cookie(dash1.Cookies.AccessName)

		resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+deviceID+"/revoke", nil, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		// Past the sub-second cutoff boundary, so the rejection below cannot be an accident of
		// the JWT `iat` being truncated to whole seconds.
		time.Sleep(1200 * time.Millisecond)

		code, _ := rawRequest(t, server1, http.MethodGet, "/api/levels", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code, "precondition: the token is rejected before the restart")

		server1.Close()
		require.NoError(t, dash1.Close())

		dash2, server2 := bootDashboardAt(t, ctx, dbPath)
		defer func() { server2.Close(); _ = dash2.Close() }()

		code, body := rawRequest(t, server2, http.MethodGet, "/api/levels", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code,
			"a revoked access token must stay revoked after a restart, got %d: %s", code, body)

		code, _ = rawRequest(t, server2, http.MethodPost, "/api/actions", staleCookie,
			map[string]string{"entity_id": "light.kitchen", "action": "toggle"})
		require.Equal(t, http.StatusUnauthorized, code,
			"a revoked device must not control Home Assistant after a restart")
	})

	t.Run("a demoted admin cannot use its stale setup scope", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "demoted.db")
		dash1, server1 := bootDashboardAt(t, ctx, dbPath)

		owner := newAuthTestClient(t, server1)
		_, role := enrollClient(t, ctx, dash1, owner)
		require.Equal(t, "owner", role)

		admin, adminID := inviteClientRole(t, owner, server1, "admin")
		staleCookie := dash1.Cookies.AccessName + "=" + admin.cookie(dash1.Cookies.AccessName)

		resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+adminID+"/role", map[string]string{"role": "device"}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		time.Sleep(1200 * time.Millisecond)

		code, _ := rawRequest(t, server1, http.MethodGet, "/api/export", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code, "precondition: the demoted token lost its scope before the restart")

		server1.Close()
		require.NoError(t, dash1.Close())

		dash2, server2 := bootDashboardAt(t, ctx, dbPath)
		defer func() { server2.Close(); _ = dash2.Close() }()

		code, body := rawRequest(t, server2, http.MethodGet, "/api/export", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code,
			"a demoted admin must not regain setup:manage after a restart, got %d: %s", code, body)

		code, _ = rawRequest(t, server2, http.MethodPost, "/api/levels", staleCookie,
			map[string]string{"name": "restart-write"})
		require.Equal(t, http.StatusUnauthorized, code, "a demoted admin must not write configuration after a restart")

		code, _ = rawRequest(t, server2, http.MethodPost, "/api/setup/auth/invites", staleCookie,
			map[string]string{"role": "admin"})
		require.NotEqual(t, http.StatusOK, code,
			"a demoted admin must never mint an invitation from a stale token (got %d)", code)
	})

	t.Run("a logged-out owner stays logged out", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "logout.db")
		dash1, server1 := bootDashboardAt(t, ctx, dbPath)

		owner := newAuthTestClient(t, server1)
		_, role := enrollClient(t, ctx, dash1, owner)
		require.Equal(t, "owner", role)

		staleCookie := dash1.Cookies.AccessName + "=" + owner.cookie(dash1.Cookies.AccessName)

		resp := owner.do(http.MethodPost, "/api/auth/logout", nil, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

		time.Sleep(1200 * time.Millisecond)

		code, _ := rawRequest(t, server1, http.MethodGet, "/api/levels", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code, "precondition: logout rejects the access token immediately")

		server1.Close()
		require.NoError(t, dash1.Close())

		dash2, server2 := bootDashboardAt(t, ctx, dbPath)
		defer func() { server2.Close(); _ = dash2.Close() }()

		code, body := rawRequest(t, server2, http.MethodGet, "/api/levels", staleCookie, nil)
		require.Equal(t, http.StatusUnauthorized, code,
			"a logged-out access token must stay rejected after a restart, got %d: %s", code, body)
	})
}

// TestRevokedAccountCannotRedeemPendingEnrollment covers the window between an owner approving a
// device enrollment and that device redeeming it: if the owner changes their mind and revokes the
// freshly created account, the waiting device must not be able to exchange its pending cookie for
// a live session.
//
// Regression test for the audit finding "A revoked account's pending enrollment can still be
// redeemed for a live session": Auth.Redeem looked the account up but never re-checked its status
// before issuing a token pair, and IssueTokenPair signs the supplied claims verbatim rather than
// going through the ClaimsProvider that rejects non-active accounts.
func TestRevokedAccountCannotRedeemPendingEnrollment(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	waiter := newAuthTestClient(t, server)
	resp := waiter.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, readBody(t, resp), `"status":"pending"`)

	pending := dash.Auth.ListPending(ctx)
	require.Len(t, pending, 1)
	pendingID := pending[0].PendingID
	deviceID := pending[0].DeviceID

	// The operator approves the enrollment, then revokes the account it created.
	resp = owner.do(http.MethodPost, "/api/setup/auth/pending/"+pendingID+"/approve", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/setup/auth/devices/"+deviceID+"/revoke", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	acct, err := dash.Auth.GetAccount(ctx, deviceID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusRevoked, acct.Status)

	// The revocation cutoff carries sub-second precision while a JWT `iat` is truncated to whole
	// seconds. Waiting past that boundary keeps this test honest: it must pass because the status
	// is checked, not because the freshly issued token looks revoked by accident.
	time.Sleep(1200 * time.Millisecond)

	resp = waiter.do(http.MethodPost, "/api/auth/redeem", nil, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"a revoked account must not redeem its pending enrollment, got %d: %s", resp.StatusCode, body)

	// No session may exist behind a rejected redeem either.
	resp = waiter.do(http.MethodGet, "/api/levels", nil, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

// TestRevokeDeviceDropsPendingEnrollment asserts the pending record itself is discarded when the
// account behind it is revoked, so a revoked device cannot keep polling /api/auth/redeem.
func TestRevokeDeviceDropsPendingEnrollment(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	waiter := newAuthTestClient(t, server)
	resp := waiter.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	pending := dash.Auth.ListPending(ctx)
	require.Len(t, pending, 1)
	deviceID := pending[0].DeviceID

	resp = owner.do(http.MethodPost, "/api/setup/auth/pending/"+pending[0].PendingID+"/approve", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/setup/auth/devices/"+deviceID+"/revoke", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	require.Empty(t, dash.Auth.ListPending(ctx),
		"revoking an account must drop its pending enrollment record")
}
