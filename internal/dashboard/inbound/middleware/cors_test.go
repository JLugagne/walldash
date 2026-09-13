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

	cfg := middleware.DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"http://localhost:5173"}
	corsMiddleware := middleware.CORS(cfg)
	handler := corsMiddleware(nextHandler)

	req := httptest.NewRequest(http.MethodOptions, "/api/levels", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

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
}

func TestCORS_StandardRequest(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	cfg := middleware.DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"http://localhost:8080"}
	corsMiddleware := middleware.CORS(cfg)
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

func TestCORS_BareHostAllowedOrigin(t *testing.T) {
	cfg := middleware.CORSConfig{
		AllowedOrigins: []string{"walldash.domain.tld"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}

	handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("https origin matches bare host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", "https://walldash.domain.tld")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "https://walldash.domain.tld", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("http origin with port matches bare host with port", func(t *testing.T) {
		portCfg := middleware.CORSConfig{
			AllowedOrigins: []string{"walldash.domain.tld:8443"},
			AllowedMethods: []string{"GET", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
		}
		portHandler := middleware.CORS(portCfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", "http://walldash.domain.tld:8443")
		rec := httptest.NewRecorder()
		portHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://walldash.domain.tld:8443", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("different host is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORS_SchemeSensitiveAllowedOrigins(t *testing.T) {
	handlerFor := func(origins ...string) http.Handler {
		cfg := middleware.DefaultCORSConfig()
		cfg.AllowedOrigins = origins
		return middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}

	request := func(handler http.Handler, origin string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	t.Run("explicit https entry does not reflect the http origin", func(t *testing.T) {
		rec := request(handlerFor("https://trusted.example"), "http://trusted.example")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("explicit https entry reflects the https origin", func(t *testing.T) {
		rec := request(handlerFor("https://trusted.example"), "https://trusted.example")
		assert.Equal(t, "https://trusted.example", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
	})

	t.Run("explicit default port is normalized", func(t *testing.T) {
		rec := request(handlerFor("https://trusted.example"), "https://trusted.example:443")
		assert.Equal(t, "https://trusted.example:443", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("explicit scheme match is case-insensitive", func(t *testing.T) {
		rec := request(handlerFor("HTTPS://Trusted.Example"), "https://trusted.example")
		assert.Equal(t, "https://trusted.example", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("different port is rejected", func(t *testing.T) {
		rec := request(handlerFor("https://trusted.example"), "https://trusted.example:8443")
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("bare host entry stays scheme-insensitive", func(t *testing.T) {
		handler := handlerFor("trusted.example")
		httpRec := request(handler, "http://trusted.example")
		httpsRec := request(handler, "https://trusted.example")
		assert.Equal(t, "http://trusted.example", httpRec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "https://trusted.example", httpsRec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("bare host entry with port still matches", func(t *testing.T) {
		rec := request(handlerFor("trusted.example:8443"), "http://trusted.example:8443")
		assert.Equal(t, "http://trusted.example:8443", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("null origin is rejected", func(t *testing.T) {
		rec := request(handlerFor("null"), "null")
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("unrelated origin is not reflected", func(t *testing.T) {
		rec := request(handlerFor("https://trusted.example"), "https://evil.example")
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}
