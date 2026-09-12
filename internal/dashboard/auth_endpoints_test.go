package dashboard

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authTestClient struct {
	t      *testing.T
	base   string
	client *http.Client
	csrf   string
}

func setupAuthServer(t *testing.T) (context.Context, *Dashboard, *httptest.Server, *authTestClient) {
	t.Helper()
	ctx := context.Background()
	router := mux.NewRouter()
	dash, err := New(ctx, Config{
		DBPath:  filepath.Join(t.TempDir(), "auth.db"),
		Version: "test",
	}, router)
	require.NoError(t, err)

	server := httptest.NewTLSServer(router)
	t.Cleanup(func() {
		server.Close()
		_ = dash.Close()
	})

	client := newAuthTestClient(t, server)
	client.fetchCSRF()
	return ctx, dash, server, client
}

func newAuthTestClient(t *testing.T, server *httptest.Server) *authTestClient {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := *server.Client()
	client.Jar = jar
	return &authTestClient{t: t, base: server.URL, client: &client}
}

func (c *authTestClient) do(method, path string, body any, mutate func(*http.Request)) *http.Response {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(c.t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	require.NoError(c.t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	req.Header.Set("Origin", c.base)
	if mutate != nil {
		mutate(req)
	}
	resp, err := c.client.Do(req)
	require.NoError(c.t, err)
	return resp
}

func (c *authTestClient) fetchCSRF() {
	c.t.Helper()
	resp := c.do(http.MethodGet, "/api/csrf-token", nil, nil)
	defer resp.Body.Close()
	require.Equal(c.t, http.StatusOK, resp.StatusCode)

	var out struct {
		Data struct {
			CSRFToken string `json:"csrf_token"`
		} `json:"data"`
	}
	require.NoError(c.t, json.NewDecoder(resp.Body).Decode(&out))
	require.NotEmpty(c.t, out.Data.CSRFToken)
	c.csrf = out.Data.CSRFToken
}

func (c *authTestClient) cookie(name string) string {
	c.t.Helper()
	u, err := url.Parse(c.base)
	require.NoError(c.t, err)
	for _, cookie := range c.client.Jar.Cookies(u) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(raw)
}

func decodeAuthenticated(t *testing.T, resp *http.Response) bool {
	t.Helper()
	defer resp.Body.Close()
	var out struct {
		Data struct {
			Authenticated bool `json:"authenticated"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out.Data.Authenticated
}

func requireJSendCode(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	defer resp.Body.Close()
	var out struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.Equal(t, "error", out.Status)
	assert.Equal(t, want, out.Code)
}

func TestAuthEndpointsFlow(t *testing.T) {
	ctx, dash, server, client := setupAuthServer(t)

	resp := client.do(http.MethodGet, "/api/health", nil, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	requireJSendCode(t, resp, "UNAUTHORIZED")

	resp = client.do(http.MethodGet, "/api/auth/status", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, decodeAuthenticated(t, resp))

	resp = client.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	connectBody := readBody(t, resp)
	require.Contains(t, connectBody, `"status":"pending"`)
	require.NotContains(t, connectBody, `"code"`)

	pending := dash.Auth.ListPending(ctx)
	require.Len(t, pending, 1)
	code := pending[0].Code
	require.Len(t, code, 6)

	resp = client.do(http.MethodGet, "/api/health", nil, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp = client.do(http.MethodPost, "/api/auth/verify", map[string]string{"code": code}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	verifyBody := readBody(t, resp)
	require.Contains(t, verifyBody, `"role":"owner"`)
	require.Empty(t, dash.Auth.ListPending(ctx))

	resp = client.do(http.MethodGet, "/api/health", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp = client.do(http.MethodGet, "/api/auth/me", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	meBody := readBody(t, resp)
	require.Contains(t, meBody, `"role":"owner"`)

	resp = client.do(http.MethodGet, "/api/auth/status", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, decodeAuthenticated(t, resp))

	beforeRefresh := client.cookie(dash.Cookies.RefreshName)
	require.NotEmpty(t, beforeRefresh)

	resp = client.do(http.MethodPost, "/api/auth/refresh", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	afterRefresh := client.cookie(dash.Cookies.RefreshName)
	require.NotEmpty(t, afterRefresh)
	require.NotEqual(t, beforeRefresh, afterRefresh)

	resp = client.do(http.MethodPost, "/api/auth/logout", nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	resp = client.do(http.MethodPost, "/api/auth/refresh", nil, func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: dash.Cookies.RefreshName, Value: afterRefresh})
	})
	require.NotEqual(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	wsURL := "wss" + strings.TrimPrefix(server.URL, "https") + "/api/ws"
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	conn, wsResp, err := dialer.Dial(wsURL, nil)
	if conn != nil {
		_ = conn.Close()
	}
	require.Error(t, err)
	require.NotNil(t, wsResp)
	defer wsResp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, wsResp.StatusCode)
}

func TestAuthVerifyRejectsBadCode(t *testing.T) {
	ctx, dash, _, client := setupAuthServer(t)

	resp := client.do(http.MethodPost, "/api/auth/connect", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	pending := dash.Auth.ListPending(ctx)
	require.Len(t, pending, 1)
	bad := "000000"
	if bad == pending[0].Code {
		bad = "111111"
	}

	resp = client.do(http.MethodPost, "/api/auth/verify", map[string]string{"code": bad}, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	requireJSendCode(t, resp, "invalid_code")
}

func TestWebSocketRequiresValidAccessCookie(t *testing.T) {
	ctx, dash, server, client := setupAuthServer(t)

	wsURL := "wss" + strings.TrimPrefix(server.URL, "https") + "/api/ws"
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	t.Run("without access cookie is rejected", func(t *testing.T) {
		conn, resp, err := dialer.Dial(wsURL, nil)
		if conn != nil {
			_ = conn.Close()
		}
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("with invalid access cookie is rejected", func(t *testing.T) {
		header := http.Header{"Cookie": []string{dash.Cookies.AccessName + "=not-a-valid-token"}}
		conn, resp, err := dialer.Dial(wsURL, header)
		if conn != nil {
			_ = conn.Close()
		}
		require.Error(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("with valid access cookie upgrades", func(t *testing.T) {
		resp := client.do(http.MethodPost, "/api/auth/connect", nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		pending := dash.Auth.ListPending(ctx)
		require.Len(t, pending, 1)

		resp = client.do(http.MethodPost, "/api/auth/verify", map[string]string{"code": pending[0].Code}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		access := client.cookie(dash.Cookies.AccessName)
		require.NotEmpty(t, access)

		header := http.Header{"Cookie": []string{dash.Cookies.AccessName + "=" + access}}
		conn, wsResp, err := dialer.Dial(wsURL, header)
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, wsResp.StatusCode)
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		require.NoError(t, err)
		require.Contains(t, string(msg), "connected")
	})
}

func TestRefreshTrustsAllowedOrigins(t *testing.T) {
	router := mux.NewRouter()
	dash, err := New(context.Background(), Config{
		DBPath:         filepath.Join(t.TempDir(), "auth-origins.db"),
		Version:        "test",
		AllowedOrigins: []string{"https://walldash.example.com"},
	}, router)
	require.NoError(t, err)
	server := httptest.NewTLSServer(router)
	t.Cleanup(func() {
		server.Close()
		_ = dash.Close()
	})

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/auth/refresh", nil)
	require.NoError(t, err)
	req.Host = "walldash-internal:8080"
	req.Header.Set("Origin", "https://walldash.example.com")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotEqual(t, http.StatusForbidden, resp.StatusCode, "an allowed origin must not be rejected by the egauth same-origin check")
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
