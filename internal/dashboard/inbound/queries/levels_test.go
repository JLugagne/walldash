package queries_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svclevelstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels/levelstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelsQueryHandler(t *testing.T) {
	controller := inbound.NewController()

	t.Run("GET /api/levels returns list of levels", func(t *testing.T) {
		mockQueries := &svclevelstest.MockLevelQueries{
			ListLevelsFunc: func(ctx context.Context) ([]domain.Level, error) {
				return []domain.Level{
					{
						ID:        "lvl-1",
						Name:      "RDC",
						Order:     1,
						IsOutdoor: false,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupLevelRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Status string                       `json:"status"`
			Data   []pkgdashboard.LevelResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "lvl-1", resp.Data[0].ID)
		assert.Equal(t, "RDC", resp.Data[0].Name)
	})

	t.Run("GET /api/levels/{id} returns level when found", func(t *testing.T) {
		mockQueries := &svclevelstest.MockLevelQueries{
			GetLevelFunc: func(ctx context.Context, id string) (domain.Level, error) {
				assert.Equal(t, "lvl-1", id)
				return domain.Level{
					ID:        "lvl-1",
					Name:      "RDC",
					Order:     1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupLevelRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels/lvl-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                     `json:"status"`
			Data   pkgdashboard.LevelResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "lvl-1", resp.Data.ID)
	})

	t.Run("GET /api/levels/{id} returns error when not found", func(t *testing.T) {
		mockQueries := &svclevelstest.MockLevelQueries{
			GetLevelFunc: func(ctx context.Context, id string) (domain.Level, error) {
				return domain.Level{}, domain.ErrLevelNotFound
			},
		}

		router := mux.NewRouter()
		queries.SetupLevelRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels/lvl-unknown", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp inbound.ResponseFail
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "fail", resp.Status)
		assert.Equal(t, "LEVEL_NOT_FOUND", resp.Error.Code)
	})

	t.Run("GET /api/levels/{id}/plan returns plan when found", func(t *testing.T) {
		mockQueries := &svclevelstest.MockLevelQueries{
			GetPlanFunc: func(ctx context.Context, levelID string) (domain.Plan, error) {
				return domain.Plan{
					LevelID: levelID,
					Walls: []domain.WallSegment{
						{ID: "w1", X1: 0, Y1: 0, X2: 10, Y2: 10, Thickness: 5},
					},
					Zones: []domain.Zone{},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupLevelRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels/lvl-1/plan", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                    `json:"status"`
			Data   pkgdashboard.PlanResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "lvl-1", resp.Data.LevelID)
		assert.Len(t, resp.Data.Walls, 1)
	})
}
