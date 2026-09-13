package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/stretchr/testify/require"
)

// TestSameOriginIsSchemeAware is the regression test for the audit finding "CSRF origin gate is
// scheme-blind". The gate used to compare hosts only, so a page served over plain HTTP from the
// same host:port was treated as same-origin on an HTTPS deployment. Cross-site request forgery
// protection has to compare the whole origin: scheme, host and port.
func TestSameOriginIsSchemeAware(t *testing.T) {
	handler := middleware.SameOrigin(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name    string
		target  string
		origin  string
		referer string
		want    int
	}{
		{name: "same https origin", target: "https://walldash.lan:8080/api/levels", origin: "https://walldash.lan:8080", want: http.StatusNoContent},
		{name: "same http origin", target: "http://walldash.lan:8080/api/levels", origin: "http://walldash.lan:8080", want: http.StatusNoContent},
		{name: "default port elided on https", target: "https://walldash.lan/api/levels", origin: "https://walldash.lan:443", want: http.StatusNoContent},

		{name: "plain http origin against https", target: "https://walldash.lan:8080/api/levels", origin: "http://walldash.lan:8080", want: http.StatusForbidden},
		{name: "https origin against plain http", target: "http://walldash.lan:8080/api/levels", origin: "https://walldash.lan:8080", want: http.StatusForbidden},
		{name: "other port", target: "https://walldash.lan:8080/api/levels", origin: "https://walldash.lan:9090", want: http.StatusForbidden},
		{name: "other host", target: "https://walldash.lan/api/levels", origin: "https://evil.example", want: http.StatusForbidden},
		{name: "suffix lookalike", target: "https://walldash.lan/api/levels", origin: "https://walldash.lan.evil.example", want: http.StatusForbidden},
		{name: "prefix lookalike", target: "https://walldash.lan/api/levels", origin: "https://evilwalldash.lan", want: http.StatusForbidden},
		{name: "subdomain", target: "https://walldash.lan/api/levels", origin: "https://evil.walldash.lan", want: http.StatusForbidden},
		{name: "userinfo trick", target: "https://walldash.lan/api/levels", origin: "https://walldash.lan@evil.example", want: http.StatusForbidden},
		{name: "opaque null origin", target: "https://walldash.lan/api/levels", origin: "null", want: http.StatusForbidden},
		{name: "file origin", target: "https://walldash.lan/api/levels", origin: "file:///tmp/evil.html", want: http.StatusForbidden},
		{name: "hostile origin, hostile referer", target: "https://walldash.lan/api/levels", origin: "https://evil.example", referer: "https://evil.example/page", want: http.StatusForbidden},

		{name: "no origin and no referer", target: "https://walldash.lan/api/levels", want: http.StatusForbidden},
		{name: "no origin, same-origin referer", target: "https://walldash.lan/api/levels", referer: "https://walldash.lan/setup", want: http.StatusNoContent},
		{name: "no origin, cross-scheme referer", target: "https://walldash.lan/api/levels", referer: "http://walldash.lan/setup", want: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.target, nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.referer != "" {
				req.Header.Set("Referer", tc.referer)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, tc.want, rec.Code)
		})
	}
}

// TestSameOriginTrustedOriginsAreSchemeExact checks the allow-list: entries carrying a scheme must
// match the whole canonical origin, while bare-host entries stay convenient but only on the scheme
// the request itself arrived on.
func TestSameOriginTrustedOriginsAreSchemeExact(t *testing.T) {
	trusted := middleware.SameOrigin([]string{"https://walldash.example.com", "legacy.example.com"})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))

	tests := []struct {
		name   string
		target string
		origin string
		want   int
	}{
		{name: "trusted origin, same scheme", target: "https://internal.local/api/levels", origin: "https://walldash.example.com", want: http.StatusNoContent},
		{name: "trusted origin, other scheme", target: "http://internal.local/api/levels", origin: "http://walldash.example.com", want: http.StatusForbidden},
		{name: "trusted origin, wrong port", target: "https://internal.local/api/levels", origin: "https://walldash.example.com:8443", want: http.StatusForbidden},
		{name: "bare host entry, same scheme", target: "https://internal.local/api/levels", origin: "https://legacy.example.com", want: http.StatusNoContent},
		{name: "bare host entry, other scheme", target: "http://internal.local/api/levels", origin: "https://legacy.example.com", want: http.StatusForbidden},
		{name: "untrusted host", target: "https://internal.local/api/levels", origin: "https://evil.example", want: http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.target, nil)
			req.Header.Set("Origin", tc.origin)
			rec := httptest.NewRecorder()
			trusted.ServeHTTP(rec, req)
			require.Equal(t, tc.want, rec.Code)
		})
	}
}

// TestSameOriginHonoursForwardedProto covers the reverse-proxy deployment: TLS terminates upstream,
// so the scheme the browser used is only visible in X-Forwarded-Proto.
func TestSameOriginHonoursForwardedProto(t *testing.T) {
	handler := middleware.SameOrigin(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://walldash.lan/api/levels", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Origin", "https://walldash.lan")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// The same request claiming a plain-http origin is still refused.
	req = httptest.NewRequest(http.MethodPost, "http://walldash.lan/api/levels", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Origin", "http://walldash.lan")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

// TestSameOriginLetsSafeMethodsThrough keeps reads working: the gate only guards state changes.
func TestSameOriginLetsSafeMethodsThrough(t *testing.T) {
	handler := middleware.SameOrigin(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		req := httptest.NewRequest(method, "https://walldash.lan/api/levels", nil)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, "%s must not be blocked", method)
	}
}
