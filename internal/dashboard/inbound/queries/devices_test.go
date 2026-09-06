package queries_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices/devicestest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevicesQueryHandler(t *testing.T) {
	controller := inbound.NewController()

	t.Run("GET /api/devices returns available devices", func(t *testing.T) {
		mockQueries := &devicestest.MockDeviceQueries{
			ListAvailableDevicesFunc: func(ctx context.Context) ([]domain.Device, error) {
				return []domain.Device{
					{
						ID:     "light.salon",
						Name:   "Plafonnier Salon",
						Domain: "light",
						State:  "on",
					},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupDeviceRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Status string                        `json:"status"`
			Data   []pkgdashboard.DeviceResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "light.salon", resp.Data[0].ID)
		assert.Equal(t, "Plafonnier Salon", resp.Data[0].Name)
	})

	t.Run("GET /api/devices returns error when service fails", func(t *testing.T) {
		mockQueries := &devicestest.MockDeviceQueries{
			ListAvailableDevicesFunc: func(ctx context.Context) ([]domain.Device, error) {
				return nil, errors.New("backend error")
			},
		}

		router := mux.NewRouter()
		queries.SetupDeviceRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("GET /api/levels/{id}/placements returns level placements", func(t *testing.T) {
		mockQueries := &devicestest.MockDeviceQueries{
			ListPlacementsFunc: func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
				assert.Equal(t, "lvl-1", levelID)
				return []domain.DevicePlacement{
					{
						ID:       "p-1",
						LevelID:  "lvl-1",
						DeviceID: "light.salon",
						X:        100,
						Y:        200,
					},
				}, nil
			},
		}

		router := mux.NewRouter()
		queries.SetupDeviceRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels/lvl-1/placements", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Status string                                 `json:"status"`
			Data   []pkgdashboard.DevicePlacementResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		require.Len(t, resp.Data, 1)
		assert.Equal(t, "p-1", resp.Data[0].ID)
		assert.Equal(t, 100.0, resp.Data[0].X)
	})

	t.Run("GET /api/levels/{id}/placements returns 400/fail on ErrLevelNotFound", func(t *testing.T) {
		mockQueries := &devicestest.MockDeviceQueries{
			ListPlacementsFunc: func(ctx context.Context, levelID string) ([]domain.DevicePlacement, error) {
				return nil, domain.ErrLevelNotFound
			},
		}

		router := mux.NewRouter()
		queries.SetupDeviceRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/levels/missing/placements", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
