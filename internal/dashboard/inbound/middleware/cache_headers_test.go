package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/stretchr/testify/require"
)

// TestAPIResponsesAreNotCacheable is the regression test for the audit finding "Authenticated API
// responses carry no Cache-Control". The API returns configuration, device and account data, so a
// shared intermediary cache must not be free to store it.
func TestAPIResponsesAreNotCacheable(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://walldash.lan/api/levels", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

// TestStaticAssetsKeepTheirOwnCaching makes sure the API rule does not leak onto the SPA assets,
// whose caching is what keeps the dashboard fast.
func TestStaticAssetsKeepTheirOwnCaching(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://walldash.lan/assets/index-abc123.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Empty(t, rec.Header().Get("Cache-Control"))
}
