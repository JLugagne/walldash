package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svclevelstest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels/levelstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/commands"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelsCommandHandler(t *testing.T) {
	controller := inbound.NewController()

	t.Run("POST /api/levels creates level successfully", func(t *testing.T) {
		mockCommands := &svclevelstest.MockLevelCommands{
			CreateLevelFunc: func(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
				assert.Equal(t, "Étage 1", level.Name)
				assert.True(t, level.IsOutdoor)
				return domain.Level{
					ID:        "lvl-created",
					Name:      level.Name,
					Order:     1,
					IsOutdoor: level.IsOutdoor,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		reqBody := pkgdashboard.CreateLevelRequest{
			Name:      "Étage 1",
			IsOutdoor: true,
		}
		data, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/levels", bytes.NewReader(data))
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
		assert.Equal(t, "lvl-created", resp.Data.ID)
		assert.Equal(t, "Étage 1", resp.Data.Name)
		assert.Equal(t, []string{"controls", "sensors"}, resp.Data.Layers)
	})

	t.Run("POST /api/levels with empty name returns 400 validation error", func(t *testing.T) {
		mockCommands := &svclevelstest.MockLevelCommands{}
		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		reqBody := pkgdashboard.CreateLevelRequest{
			Name: "",
		}
		data, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/levels", bytes.NewReader(data))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp inbound.ResponseFail
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "fail", resp.Status)
		assert.Equal(t, "INVALID_CREATE_LEVEL_REQUEST", resp.Error.Code)
	})

	t.Run("PUT /api/levels/{id} updates level successfully", func(t *testing.T) {
		mockCommands := &svclevelstest.MockLevelCommands{
			UpdateLevelFunc: func(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
				assert.Equal(t, "lvl-1", level.ID)
				assert.Equal(t, "Nouveau Nom", level.Name)
				return level, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		reqBody := pkgdashboard.UpdateLevelRequest{
			Name:      "Nouveau Nom",
			IsOutdoor: false,
		}
		data, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/levels/lvl-1", bytes.NewReader(data))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("DELETE /api/levels/{id} deletes level successfully", func(t *testing.T) {
		deleted := false
		mockCommands := &svclevelstest.MockLevelCommands{
			DeleteLevelFunc: func(ctx context.Context, actor domain.Actor, id string) error {
				assert.Equal(t, "lvl-1", id)
				deleted = true
				return nil
			},
		}

		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/levels/lvl-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, deleted)
	})

	t.Run("POST /api/levels/reorder reorders levels successfully", func(t *testing.T) {
		reordered := false
		mockCommands := &svclevelstest.MockLevelCommands{
			ReorderLevelsFunc: func(ctx context.Context, actor domain.Actor, orderedIDs []string) error {
				assert.Equal(t, []string{"lvl-2", "lvl-1"}, orderedIDs)
				reordered = true
				return nil
			},
		}

		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		reqBody := pkgdashboard.ReorderLevelsRequest{
			LevelIDs: []string{"lvl-2", "lvl-1"},
		}
		data, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/levels/reorder", bytes.NewReader(data))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, reordered)
	})

	t.Run("PUT /api/levels/{id}/plan saves plan successfully", func(t *testing.T) {
		mockCommands := &svclevelstest.MockLevelCommands{
			SavePlanFunc: func(ctx context.Context, actor domain.Actor, plan domain.Plan) (domain.Plan, error) {
				assert.Equal(t, "lvl-1", plan.LevelID)
				assert.Len(t, plan.Walls, 1)
				return plan, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupLevelRoutes(router, controller, mockCommands)

		reqBody := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "w1",
					X1:        0,
					Y1:        0,
					X2:        100,
					Y2:        0,
					Thickness: 10,
				},
			},
			Zones: []pkgdashboard.ZoneDTO{},
		}
		data, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/levels/lvl-1/plan", bytes.NewReader(data))
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
	})
}
