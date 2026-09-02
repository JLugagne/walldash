package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices/devicestest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/commands"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevicesCommandHandler(t *testing.T) {
	controller := inbound.NewController()

	t.Run("POST /api/levels/{id}/placements saves placement", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{
			SavePlacementFunc: func(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
				assert.Equal(t, "lvl-1", placement.LevelID)
				assert.Equal(t, "light.salon", placement.DeviceID)
				assert.Equal(t, 150.0, placement.X)
				assert.Equal(t, 250.0, placement.Y)
				placement.ID = "saved-p-1"
				return placement, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		body := map[string]any{
			"device_id":   "light.salon",
			"x":           150.0,
			"y":           250.0,
			"icon":        "lightbulb",
			"custom_name": "Plafond",
		}
		raw, err := json.Marshal(body)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/levels/lvl-1/placements", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Status string                               `json:"status"`
			Data   pkgdashboard.DevicePlacementResponse `json:"data"`
		}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "saved-p-1", resp.Data.ID)
		assert.Equal(t, "light.salon", resp.Data.DeviceID)
	})

	t.Run("POST /api/levels/{id}/placements rejects invalid body", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodPost, "/api/levels/lvl-1/placements", bytes.NewReader([]byte("{invalid-json")))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("POST /api/levels/{id}/placements rejects validation failure (missing device_id)", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		body := map[string]any{
			"device_id": "",
			"x":         10,
			"y":         10,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/levels/lvl-1/placements", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("POST /api/levels/{id}/placements returns 400 on ErrLevelNotFound", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{
			SavePlacementFunc: func(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
				return domain.DevicePlacement{}, domain.ErrLevelNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		body := map[string]any{
			"device_id": "light.salon",
			"x":         10,
			"y":         10,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/levels/lvl-missing/placements", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("POST /api/levels/{id}/placements returns 500 on unexpected error", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{
			SavePlacementFunc: func(ctx context.Context, actor domain.Actor, placement domain.DevicePlacement) (domain.DevicePlacement, error) {
				return domain.DevicePlacement{}, errors.New("db error")
			},
		}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		body := map[string]any{
			"device_id": "light.salon",
			"x":         10,
			"y":         10,
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/levels/lvl-1/placements", bytes.NewReader(raw))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("DELETE /api/levels/{id}/placements/{placementId} deletes placement", func(t *testing.T) {
		deleted := false
		mockCommands := &devicestest.MockDeviceCommands{
			DeletePlacementFunc: func(ctx context.Context, actor domain.Actor, levelID string, placementID string) error {
				assert.Equal(t, "lvl-1", levelID)
				assert.Equal(t, "p-1", placementID)
				deleted = true
				return nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/levels/lvl-1/placements/p-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, deleted)
	})

	t.Run("DELETE /api/levels/{id}/placements/{placementId} returns 400 on ErrPlacementNotFound", func(t *testing.T) {
		mockCommands := &devicestest.MockDeviceCommands{
			DeletePlacementFunc: func(ctx context.Context, actor domain.Actor, levelID string, placementID string) error {
				return domain.ErrPlacementNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupDeviceRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/levels/lvl-1/placements/p-missing", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
