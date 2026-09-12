package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards/dashboardstest"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/commands"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardsCommandRoutes(t *testing.T) {
	controller := inbound.NewController()

	t.Run("POST /api/dashboards creates dashboard", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			CreateDashboardFunc: func(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
				dashboard.ID = "ov-created-1"
				return dashboard, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.CreateDashboardRequest{Name: "Living Room", Order: 1})
		req := httptest.NewRequest(http.MethodPost, "/api/dashboards", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                         `json:"status"`
			Data   pkgdashboard.DashboardResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ov-created-1", resp.Data.ID)
		assert.Equal(t, "Living Room", resp.Data.Name)
	})

	t.Run("PUT /api/dashboards/{id} updates dashboard", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			UpdateDashboardFunc: func(ctx context.Context, actor domain.Actor, dashboard domain.Dashboard) (domain.Dashboard, error) {
				return dashboard, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.UpdateDashboardRequest{Name: "Renovated Living Room", Order: 2})
		req := httptest.NewRequest(http.MethodPut, "/api/dashboards/ov-1", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                         `json:"status"`
			Data   pkgdashboard.DashboardResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "ov-1", resp.Data.ID)
		assert.Equal(t, "Renovated Living Room", resp.Data.Name)
	})

	t.Run("DELETE /api/dashboards/{id} deletes dashboard", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			DeleteDashboardFunc: func(ctx context.Context, actor domain.Actor, id string) error {
				if id == "ov-1" {
					return nil
				}
				return domain.ErrDashboardNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/dashboards/ov-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		req404 := httptest.NewRequest(http.MethodDelete, "/api/dashboards/unknown", nil)
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code)
	})

	t.Run("POST /api/dashboards/{id}/widgets creates widget", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			CreateWidgetFunc: func(ctx context.Context, actor domain.Actor, widget domain.Widget) (domain.Widget, error) {
				widget.ID = "w-new-1"
				return widget, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.CreateWidgetRequest{
			Type:    "automation_list",
			Title:   "Scenes",
			Order:   0,
			Col:     0,
			Row:     0,
			ColSpan: 2,
			RowSpan: 2,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.cinema"},
				Display:   "list",
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/dashboards/ov-1/widgets", bytes.NewReader(payload))
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
		assert.Equal(t, "Scenes", resp.Data.Title)
	})

	t.Run("DELETE /api/dashboards/{id}/widgets/{widgetId} deletes widget", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			DeleteWidgetFunc: func(ctx context.Context, actor domain.Actor, dashboardID, widgetID string) error {
				if widgetID == "w-1" {
					return nil
				}
				return domain.ErrWidgetNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		req := httptest.NewRequest(http.MethodDelete, "/api/dashboards/ov-1/widgets/w-1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("PUT /api/dashboards/{id}/widgets/{widgetId} updates widget", func(t *testing.T) {
		mockCommands := &dashboardstest.MockDashboardCommands{
			UpdateWidgetFunc: func(ctx context.Context, actor domain.Actor, dashboardID string, widget domain.Widget) (domain.Widget, error) {
				if widget.ID != "w-1" {
					return domain.Widget{}, domain.ErrWidgetNotFound
				}
				widget.DashboardID = dashboardID
				return widget, nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.UpdateWidgetRequest{
			Title: "Nouveau Titre",
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"sensor.temp"},
				Display:   "number",
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/dashboards/ov-1/widgets/w-1", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                      `json:"status"`
			Data   pkgdashboard.WidgetResponse `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "w-1", resp.Data.ID)
		assert.Equal(t, "ov-1", resp.Data.DashboardID)
		assert.Equal(t, "Nouveau Titre", resp.Data.Title)

		req404 := httptest.NewRequest(http.MethodPut, "/api/dashboards/ov-1/widgets/unknown", bytes.NewReader(payload))
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code)
	})

	t.Run("PUT /api/dashboards/{id}/layout updates layout", func(t *testing.T) {
		var receivedPositions []domain.WidgetPosition
		mockCommands := &dashboardstest.MockDashboardCommands{
			UpdateLayoutFunc: func(ctx context.Context, actor domain.Actor, dashboardID string, positions []domain.WidgetPosition) error {
				if dashboardID == "unknown" {
					return domain.ErrDashboardNotFound
				}
				receivedPositions = positions
				return nil
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

		payload, _ := json.Marshal(pkgdashboard.UpdateLayoutRequest{
			Positions: []pkgdashboard.WidgetPositionDTO{
				{ID: "w-1", Col: 1, Row: 1, ColSpan: 2, RowSpan: 2},
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/dashboards/ov-1/layout", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.Len(t, receivedPositions, 1)
		assert.Equal(t, "w-1", receivedPositions[0].ID)
		assert.Equal(t, 1, receivedPositions[0].Col)

		req404 := httptest.NewRequest(http.MethodPut, "/api/dashboards/unknown/layout", bytes.NewReader(payload))
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		assert.Equal(t, http.StatusBadRequest, rec404.Code)
	})

	t.Run("POST /api/automations/{id}/trigger triggers automation", func(t *testing.T) {
		var triggeredID string
		mockCommands := &dashboardstest.MockDashboardCommands{
			TriggerAutomationFunc: func(ctx context.Context, actor domain.Actor, id string) error {
				if id == "automation.cinema" {
					triggeredID = id
					return nil
				}
				return domain.ErrAutomationNotFound
			},
		}

		router := mux.NewRouter()
		commands.SetupDashboardRoutes(router, controller, mockCommands)

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
