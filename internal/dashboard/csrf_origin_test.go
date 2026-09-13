package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// postAsHost sends a JSON request with an explicit Host and Origin, carrying the caller's cookies.
// It exists because the cross-scheme case can only be expressed by controlling both headers: the
// deployment origin is what the operator allow-listed, not the address the test server listens on.
func postAsHost(t *testing.T, server *httptest.Server, client *authTestClient, host, origin, path string, body any) (int, string) {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, server.URL+path, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Host = host
	req.Header.Set("Origin", origin)
	req.Header.Set("Content-Type", "application/json")
	// The deployment terminates TLS in a reverse proxy (see docs/reverse-proxy.md), which is what
	// makes the browser-facing scheme https while this test server speaks plain HTTP.
	req.Header.Set("X-Forwarded-Proto", "https")

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	for _, cookie := range client.client.Jar.Cookies(serverURL) {
		req.AddCookie(cookie)
	}

	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(raw)
}

// TestCrossSchemeOriginIsRejectedEndToEnd is the end-to-end regression test for the audit finding
// "CSRF origin gate is scheme-blind". With the deployment origin allow-listed as
// https://walldash.lan:8443, a request that presents the same host over plain http used to be
// accepted (host-only comparison) and must now be refused, while the genuine https origin keeps
// working.
func TestCrossSchemeOriginIsRejectedEndToEnd(t *testing.T) {
	const (
		deploymentHost   = "walldash.lan:8443"
		deploymentOrigin = "https://walldash.lan:8443"
	)

	ctx := context.Background()
	router := mux.NewRouter()
	dash, err := New(ctx, Config{
		DBPath:         filepath.Join(t.TempDir(), "csrf.db"),
		Version:        "test",
		TokenSecret:    restartTokenSecret,
		AllowedOrigins: []string{deploymentOrigin},
	}, router)
	require.NoError(t, err)
	defer func() { _ = dash.Close() }()

	server := httptest.NewServer(router)
	defer server.Close()

	owner := newAuthTestClient(t, server)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	code, body := postAsHost(t, server, owner, deploymentHost, deploymentOrigin, "/api/levels",
		map[string]any{"name": "Ground floor"})
	require.Equal(t, http.StatusOK, code, "the allow-listed origin must work: %s", body)

	code, body = postAsHost(t, server, owner, deploymentHost, "http://"+deploymentHost, "/api/levels",
		map[string]any{"name": "Stolen"})
	require.Equal(t, http.StatusForbidden, code,
		"a plain-http origin of the same host must not be accepted, got %d: %s", code, body)

	// The refused write did not land.
	resp := owner.do(http.MethodGet, "/api/levels", nil, nil)
	levels := readBody(t, resp)
	require.NotContains(t, levels, "Stolen")
	require.Contains(t, levels, "Ground floor")
}
