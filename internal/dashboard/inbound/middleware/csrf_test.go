package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRF_SafeMethods(t *testing.T) {
	manager := middleware.NewCSRFTokenManager()
	csrfMiddleware := middleware.CSRF(manager)

	handlerCalled := false
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest(method, "/api/levels", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.True(t, handlerCalled)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestCSRF_StateChangingMethods_MissingHeaders(t *testing.T) {
	manager := middleware.NewCSRFTokenManager()
	csrfMiddleware := middleware.CSRF(manager)

	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/levels", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.Contains(t, rec.Body.String(), "CSRF")
		})
	}
}

func TestCSRF_ValidCustomHeader(t *testing.T) {
	manager := middleware.NewCSRFTokenManager()
	csrfMiddleware := middleware.CSRF(manager)

	handlerCalled := false
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
	req.Host = "localhost:8080"
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_ValidToken(t *testing.T) {
	manager := middleware.NewCSRFTokenManager()
	token := manager.GenerateToken()
	require.NotEmpty(t, token)

	csrfMiddleware := middleware.CSRF(manager)

	handlerCalled := false
	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("valid token succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("X-CSRF-Token", token)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid token fails with 403", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("X-CSRF-Token", "invalid-fake-token-1234")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestCSRF_OriginAndRefererCheck(t *testing.T) {
	manager := middleware.NewCSRFTokenManager()
	csrfMiddleware := middleware.CSRF(manager)

	handler := csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("matching origin succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Origin", "http://localhost:8080")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("mismatched origin rejected with 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Origin", "http://attacker.com")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "origin mismatch")
	})

	t.Run("mismatched referer rejected with 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Referer", "http://malicious-site.com/exploit.html")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "referer mismatch")
	})

	t.Run("matching referer succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/levels", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Referer", "http://localhost:8080/dashboard")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
