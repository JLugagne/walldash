package queries_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svchealthtest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health/healthtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealthHandler(t *testing.T) {
	c := inbound.NewController()

	t.Run("returns 200 and success response when health is ok", func(t *testing.T) {
		mockQueries := &svchealthtest.MockHealthQueries{
			GetHealthFunc: func(ctx context.Context) (domain.Health, error) {
				return domain.Health{
					Status:    domain.HealthStatusOK,
					Database:  "ok",
					Version:   "0.1.0",
					Timestamp: time.Date(2026, 9, 2, 23, 0, 0, 0, time.UTC),
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupRoutes(router, c, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res struct {
			Status string `json:"status"`
			Data   struct {
				Status   string `json:"status"`
				Database string `json:"database"`
				Version  string `json:"version"`
			} `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, "success", res.Status)
		assert.Equal(t, "ok", res.Data.Status)
		assert.Equal(t, "ok", res.Data.Database)
		assert.Equal(t, "0.1.0", res.Data.Version)
	})

	t.Run("returns 500 when health check fails", func(t *testing.T) {
		mockQueries := &svchealthtest.MockHealthQueries{
			GetHealthFunc: func(ctx context.Context) (domain.Health, error) {
				return domain.Health{}, errors.Join(domain.ErrHealthCheckFailed, errors.New("db error"))
			},
		}

		router := mux.NewRouter()
		queries.SetupRoutes(router, c, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var res struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, "error", res.Status)
		assert.Equal(t, "HEALTH_CHECK_FAILED", res.Code)
	})
}
