package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCORS_PreflightOptions(t *testing.T) {
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := middleware.CORS()
	handler := corsMiddleware(nextHandler)

	req := httptest.NewRequest(http.MethodOptions, "/api/levels", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, X-CSRF-Token")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, handlerCalled, "preflight should not call downstream handler")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowHeaders, "Content-Type")
	assert.Contains(t, allowHeaders, "Authorization")
	assert.Contains(t, allowHeaders, "X-Requested-With")
	assert.Contains(t, allowHeaders, "X-CSRF-Token")
}

func TestCORS_StandardRequest(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	corsMiddleware := middleware.CORS()
	handler := corsMiddleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("Origin", "http://localhost:8080")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:8080", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "ok", rec.Body.String())
}

func TestCORS_CustomAllowedOrigins(t *testing.T) {
	cfg := middleware.CORSConfig{
		AllowedOrigins: []string{"http://trusted.local", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	corsMiddleware := middleware.CORS(cfg)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(nextHandler)

	t.Run("allowed origin is reflected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", "http://trusted.local")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://trusted.local", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("disallowed origin is not set", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", "http://evil.com")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}
