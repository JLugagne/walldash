package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/overviews/overviewstest"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/commands"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverviewsCommandRoutes(t *testing.T) {
	controller := inbound.NewController()

	t.Run("POST /api/overviews creates overview", func(t *testing.T) {
		mockCommands := &overviewstest.MockOverviewCommands{
			CreateOverviewFunc: func(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
				overview.ID = "ov-created-1"
				return overview, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.CreateOverviewRequest{Name: "Salon", Order: 1})
		req := httptest.NewRequest(http.MethodPost, "/api/overviews", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                        `json:"status"`
			Data   pkgdashboard.OverviewResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ov-created-1", resp.Data.ID)
		assert.Equal(t, "Salon", resp.Data.Name)
	})

	t.Run("PUT /api/overviews/{id} updates overview", func(t *testing.T) {
		mockCommands := &overviewstest.MockOverviewCommands{
			UpdateOverviewFunc: func(ctx context.Context, actor domain.Actor, overview domain.OverviewDashboard) (domain.OverviewDashboard, error) {
				return overview, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.UpdateOverviewRequest{Name: "Salon Rénové", Order: 2})
		req := httptest.NewRequest(http.MethodPut, "/api/overviews/ov-1", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                        `json:"status"`
			Data   pkgdashboard.OverviewResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ov-1", resp.Data.ID)
		assert.Equal(t, "Salon Rénové", resp.Data.Name)
	})

	t.Run("DELETE /api/overviews/{id} deletes overview", func(t *testing.T) {
		mockCommands := &overviewstest.MockOverviewCommands{
			DeleteOverviewFunc: func(ctx context.Context, actor domain.Actor, id string) error {
				if id == "ov-1" {
					return nil
				}
				return domain.ErrOverviewNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/overviews/ov-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		req404 := httptest.NewRequest(http.MethodDelete, "/api/overviews/unknown", nil)
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code)
	})

	t.Run("POST /api/overviews/{id}/widgets creates widget", func(t *testing.T) {
		mockCommands := &overviewstest.MockOverviewCommands{
			CreateWidgetFunc: func(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error) {
				widget.ID = "w-new-1"
				return widget, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.CreateWidgetRequest{
			Type:  "automation_list",
			Title: "Scénarios",
			Order: 0,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.cinema"},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/overviews/ov-1/widgets", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                      `json:"status"`
			Data   pkgdashboard.WidgetResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "w-new-1", resp.Data.ID)
		assert.Equal(t, "ov-1", resp.Data.DashboardID)
		assert.Equal(t, "Scénarios", resp.Data.Title)
	})

	t.Run("DELETE /api/overviews/{id}/widgets/{widgetId} deletes widget", func(t *testing.T) {
		mockCommands := &overviewstest.MockOverviewCommands{
			DeleteWidgetFunc: func(ctx context.Context, actor domain.Actor, dashboardID, widgetID string) error {
				if widgetID == "w-1" {
					return nil
				}
				return domain.ErrWidgetNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/overviews/ov-1/widgets/w-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/automations/{id}/trigger triggers automation", func(t *testing.T) {
		var triggeredID string
		mockCommands := &overviewstest.MockOverviewCommands{
			TriggerAutomationFunc: func(ctx context.Context, actor domain.Actor, id string) error {
				if id == "automation.cinema" {
					triggeredID = id
					return nil
				}
				return domain.ErrAutomationNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupOverviewRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodPost, "/api/automations/automation.cinema/trigger", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "automation.cinema", triggeredID)

		var resp struct {
			Status string                                 `json:"status"`
			Data   pkgdashboard.TriggerAutomationResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "triggered", resp.Data.Status)
		assert.Equal(t, "automation.cinema", resp.Data.ID)
	})
}
