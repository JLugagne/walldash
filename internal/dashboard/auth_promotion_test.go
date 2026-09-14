package dashboard

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// spaAuthFetch performs the same request the browser SPA does through authFetch: the request is
// sent once, and on a 401 (expired access cookie) or a 403 (route gated on a scope the token does
// not carry yet) it is retried exactly once after a silent POST /api/auth/refresh. It returns the
// status the SPA ends up acting on.
func (c *authTestClient) spaAuthFetch(method, path string, body any) int {
	c.t.Helper()
	resp := c.do(method, path, body, nil)
	status := resp.StatusCode
	resp.Body.Close()
	if status != http.StatusUnauthorized && status != http.StatusForbidden {
		return status
	}

	refresh := c.do(http.MethodPost, "/api/auth/refresh", nil, nil)
	refresh.Body.Close()
	if refresh.StatusCode != http.StatusNoContent {
		return status
	}

	resp = c.do(method, path, body, nil)
	status = resp.StatusCode
	resp.Body.Close()
	return status
}

// alignToSecondBoundary sleeps until the start of the next whole second, so a promotion and the
// requests that follow it share one wall-clock second.
func alignToSecondBoundary(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Nanosecond() > 50_000_000 {
		require.True(t, time.Now().Before(deadline), "could not align to a second boundary")
		time.Sleep(2 * time.Millisecond)
	}
}

// TestPromotedDeviceStaysSignedIn is the regression test for "admin devices do not stay signed
// in". Promoting a device used to publish an "access-token-only" revocation cutoff; egauth rejects
// every token whose `iat` is at or before that cutoff, and since a JWT `iat` is truncated to whole
// seconds while the cutoff keeps sub-second precision, the token minted by the promoted device's
// own refresh was rejected too. The SPA retries a request only once, so the promoted device landed
// on the login screen — which enrolls it again as a brand-new pending `device`, losing the role it
// had just been granted. Only the promoted (admin) devices were affected: the owner is never
// promoted and a plain device account never is either.
//
// The promotion is aligned with a second boundary because that is the worst case for a kiosk that
// polls on socket events and can therefore issue its next request in the promotion's second. A run
// whose flow crosses into the next second proves nothing, so it is retried on a fresh device.
func TestPromotedDeviceStaysSignedIn(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	for attempt := 1; ; attempt++ {
		require.LessOrEqual(t, attempt, 5, "could not keep the whole flow inside one second")

		device, deviceID := inviteClientRole(t, owner, server, "device")
		require.Equal(t, http.StatusOK, device.spaAuthFetch(http.MethodGet, "/api/auth/me", nil),
			"precondition: the device session works before the promotion")

		alignToSecondBoundary(t)
		promotedAt := time.Now()

		resp := owner.do(http.MethodPost, "/api/setup/auth/devices/"+deviceID+"/role",
			map[string]string{"role": "admin"}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		status := device.spaAuthFetch(http.MethodGet, "/api/auth/me", nil)
		if time.Now().Unix() != promotedAt.Unix() {
			continue
		}

		require.Equal(t, http.StatusOK, status,
			"a device promoted to admin must stay signed in, got HTTP %d (attempt %d)", status, attempt)
		return
	}
}

// TestPromotedDevicePicksUpSetupScope documents what replaces the promotion revocation: nothing is
// invalidated, the promoted device keeps its session with the older, smaller scope set, and a
// setup-scoped route answers 403 until the next rotation re-issues a token derived from the
// account's current role. That 403 is what the SPA refreshes and retries on.
func TestPromotedDevicePicksUpSetupScope(t *testing.T) {
	ctx, dash, server, owner := setupAuthServer(t)
	enrollClient(t, ctx, dash, owner)

	device, deviceID := inviteClientRole(t, owner, server, "device")

	resp := device.do(http.MethodGet, "/api/export", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "precondition: a device has no setup scope")
	resp.Body.Close()

	resp = owner.do(http.MethodPost, "/api/setup/auth/devices/"+deviceID+"/role",
		map[string]string{"role": "admin"}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// The session survives the promotion: the access token is neither revoked (401) nor cleared,
	// and /api/auth/me already reports the account's new role.
	resp = device.do(http.MethodGet, "/api/auth/me", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, readBody(t, resp), `"role":"admin"`)

	// The token issued before the promotion still carries the smaller scope set, so the gated
	// route answers 403 — insufficient scope, not "sign in again".
	resp = device.do(http.MethodGet, "/api/export", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// One silent refresh re-derives the scopes from the account, which is what the SPA does when
	// it sees the 403. The rejected request never reached a handler, so retrying the write is safe.
	require.Equal(t, http.StatusOK, device.spaAuthFetch(http.MethodGet, "/api/export", nil),
		"refreshing must pick up the setup scope the promotion granted")

	write := map[string]string{"name": "Promoted write"}
	require.Equal(t, http.StatusOK, device.spaAuthFetch(http.MethodPost, "/api/levels", write),
		"a promoted device must be able to write through the setup-scope gate")
}
