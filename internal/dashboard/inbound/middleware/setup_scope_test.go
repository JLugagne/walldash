package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JLugagne/egauth/tokens"
	"github.com/JLugagne/egauth/tokens/basic"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRequireSetupScopeOnWritesExemptsDeviceSafeRoutes(t *testing.T) {
	issuer := basic.NewIssuer(basic.Config{
		Store:      basic.NewMemoryStore(),
		Issuer:     "walldash",
		SecretKey:  "0123456789abcdef0123456789abcdef",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		ClaimsProvider: basic.ClaimsProviderFunc(func(ctx context.Context, uid uuid.UUID, tid string) (basic.Claims, error) {
			return basic.Claims{Subject: uid, TenantID: tid}, nil
		}),
	})
	guard := RequireSetupScopeOnWrites(issuer, tokens.DefaultCookies(), nil, []string{"/api/actions"})

	t.Run("a named device-safe route with a dynamic segment is exempt", func(t *testing.T) {
		router := mux.NewRouter()
		router.Use(guard)
		reached := false
		router.HandleFunc("/api/automations/{id}/trigger", func(w http.ResponseWriter, _ *http.Request) {
			reached = true
			w.WriteHeader(http.StatusNoContent)
		}).Methods(http.MethodPost).Name(DeviceSafeRoutePrefix + "automation-trigger")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/automations/automation.cinema/trigger", nil))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.True(t, reached, "the device-safe handler must be reached without the setup:manage scope")
	})

	t.Run("a dynamic route without the device-safe name stays guarded", func(t *testing.T) {
		router := mux.NewRouter()
		router.Use(guard)
		router.HandleFunc("/api/dashboards/{id}", func(http.ResponseWriter, *http.Request) {
			t.Fatal("a guarded handler must never be reached without the setup:manage scope")
		}).Methods(http.MethodPut)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/dashboards/abc", nil))

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
