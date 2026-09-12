package queries_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards/dashboardstest"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardsQueryRoutes(t *testing.T) {
	controller := inbound.NewController()

	t.Run("GET /api/dashboards returns list of dashboards", func(t *testing.T) {
		mockQueries := &dashboardstest.MockDashboardQueries{
			ListDashboardsFunc: func(ctx context.Context) ([]domain.Dashboard, error) {
				return []domain.Dashboard{
					{ID: "ov-1", Name: "Dashboard 1", Order: 0},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupDashboardRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/dashboards", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                           `json:"status"`
			Data   []pkgdashboard.DashboardResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "ov-1", resp.Data[0].ID)
	})

	t.Run("GET /api/dashboards/{id} returns single dashboard", func(t *testing.T) {
		mockQueries := &dashboardstest.MockDashboardQueries{
			GetDashboardFunc: func(ctx context.Context, id string) (domain.Dashboard, error) {
				if id == "ov-1" {
					return domain.Dashboard{ID: "ov-1", Name: "Dashboard 1"}, nil
				}
				return domain.Dashboard{}, domain.ErrDashboardNotFound
			},
		}

		router := mux.NewRouter()
		queries.SetupDashboardRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/dashboards/ov-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// 404 for missing
		req404 := httptest.NewRequest(http.MethodGet, "/api/dashboards/unknown", nil)
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code) // SendFail returns 400
	})

	t.Run("GET /api/automations returns list of automations", func(t *testing.T) {
		mockQueries := &dashboardstest.MockDashboardQueries{
			ListAutomationsFunc: func(ctx context.Context) ([]domain.Automation, error) {
				return []domain.Automation{
					{ID: "automation.cinema", Name: "Cinema", State: "on", Current: 0},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupDashboardRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/automations", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                            `json:"status"`
			Data   []pkgdashboard.AutomationResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "automation.cinema", resp.Data[0].ID)
	})
}
