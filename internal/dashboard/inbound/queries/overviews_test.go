package queries_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/overviews/overviewstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverviewsQueryRoutes(t *testing.T) {
	controller := inbound.NewController()

	t.Run("GET /api/overviews returns list of overviews", func(t *testing.T) {
		mockQueries := &overviewstest.MockOverviewQueries{
			ListOverviewsFunc: func(ctx context.Context) ([]domain.OverviewDashboard, error) {
				return []domain.OverviewDashboard{
					{ID: "ov-1", Name: "Tableau 1", Order: 0},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupOverviewRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/overviews", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                          `json:"status"`
			Data   []pkgdashboard.OverviewResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "ov-1", resp.Data[0].ID)
	})

	t.Run("GET /api/overviews/{id} returns single overview", func(t *testing.T) {
		mockQueries := &overviewstest.MockOverviewQueries{
			GetOverviewFunc: func(ctx context.Context, id string) (domain.OverviewDashboard, error) {
				if id == "ov-1" {
					return domain.OverviewDashboard{ID: "ov-1", Name: "Tableau 1"}, nil
				}
				return domain.OverviewDashboard{}, domain.ErrOverviewNotFound
			},
		}

		router := mux.NewRouter()
		queries.SetupOverviewRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/overviews/ov-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// 404 for missing
		req404 := httptest.NewRequest(http.MethodGet, "/api/overviews/unknown", nil)
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code) // SendFail returns 400
	})

	t.Run("GET /api/automations returns list of automations", func(t *testing.T) {
		mockQueries := &overviewstest.MockOverviewQueries{
			ListAutomationsFunc: func(ctx context.Context) ([]domain.Automation, error) {
				return []domain.Automation{
					{ID: "automation.cinema", Name: "Cinéma", State: "on", Current: 0},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupOverviewRoutes(router, controller, mockQueries)

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
