package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcdashboards "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// DashboardsHandler handles mutating operations for dashboards, widgets, and automations.
type DashboardsHandler struct {
	controller *inbound.Controller
	commands   svcdashboards.DashboardCommands
	validate   *validator.Validate
}

// NewDashboardsHandler creates a new DashboardsHandler instance.
func NewDashboardsHandler(controller *inbound.Controller, commands svcdashboards.DashboardCommands) *DashboardsHandler {
	return &DashboardsHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// SetupDashboardRoutes registers all dashboard command routes onto the router.
func SetupDashboardRoutes(r *mux.Router, controller *inbound.Controller, commands svcdashboards.DashboardCommands) {
	handler := NewDashboardsHandler(controller, commands)
	r.HandleFunc("/api/dashboards", handler.CreateDashboard).Methods(http.MethodPost)
	r.HandleFunc("/api/dashboards/{id}", handler.UpdateDashboard).Methods(http.MethodPut)
	r.HandleFunc("/api/dashboards/{id}", handler.DeleteDashboard).Methods(http.MethodDelete)
	r.HandleFunc("/api/dashboards/{id}/widgets", handler.CreateWidget).Methods(http.MethodPost)
	r.HandleFunc("/api/dashboards/{id}/widgets/{widgetId}", handler.UpdateWidget).Methods(http.MethodPut)
	r.HandleFunc("/api/dashboards/{id}/widgets/{widgetId}", handler.DeleteWidget).Methods(http.MethodDelete)
	r.HandleFunc("/api/dashboards/{id}/layout", handler.UpdateLayout).Methods(http.MethodPut)
	r.HandleFunc("/api/automations/{id}/trigger", handler.TriggerAutomation).Methods(http.MethodPost)
}

// CreateDashboard handles POST /api/dashboards.
func (h *DashboardsHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.CreateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateDashboardRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateDashboardRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainDashboard := converters.ToDomainDashboard(req)

	created, err := h.commands.CreateDashboard(r.Context(), actor, domainDashboard)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDashboard) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicDashboard(created)
	h.controller.SendSuccess(w, r, response)
}

// UpdateDashboard handles PUT /api/dashboards/{id}.
func (h *DashboardsHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pkgdashboard.UpdateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateDashboardRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateDashboardRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainDashboard := converters.ToDomainUpdateDashboard(id, req)

	updated, err := h.commands.UpdateDashboard(r.Context(), actor, domainDashboard)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) || errors.Is(err, domain.ErrInvalidDashboard) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicDashboard(updated)
	h.controller.SendSuccess(w, r, response)
}

// DeleteDashboard handles DELETE /api/dashboards/{id}.
func (h *DashboardsHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.DeleteDashboard(r.Context(), actor, id); err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]string{
		"id":     id,
		"status": "deleted",
	})
}

// CreateWidget handles POST /api/dashboards/{id}/widgets.
func (h *DashboardsHandler) CreateWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	var req pkgdashboard.CreateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateWidgetRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateWidgetRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainWidget := converters.ToDomainWidget(dashboardID, req)

	created, err := h.commands.CreateWidget(r.Context(), actor, domainWidget)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) || errors.Is(err, domain.ErrInvalidWidget) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicWidget(created)
	h.controller.SendSuccess(w, r, response)
}

// UpdateWidget handles PUT /api/dashboards/{id}/widgets/{widgetId}.
func (h *DashboardsHandler) UpdateWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]
	widgetID := vars["widgetId"]

	var req pkgdashboard.UpdateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateWidgetRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateWidgetRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainWidget := converters.ToDomainUpdateWidget(dashboardID, widgetID, req)

	updated, err := h.commands.UpdateWidget(r.Context(), actor, dashboardID, domainWidget)
	if err != nil {
		if errors.Is(err, domain.ErrWidgetNotFound) || errors.Is(err, domain.ErrDashboardNotFound) || errors.Is(err, domain.ErrInvalidWidget) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicWidget(updated)
	h.controller.SendSuccess(w, r, response)
}

// UpdateLayout handles PUT /api/dashboards/{id}/layout.
func (h *DashboardsHandler) UpdateLayout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	var req pkgdashboard.UpdateLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateLayoutRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateLayoutRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	positions := converters.ToDomainWidgetPositions(req)

	if err := h.commands.UpdateLayout(r.Context(), actor, dashboardID, positions); err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) || errors.Is(err, domain.ErrInvalidDashboard) ||
			errors.Is(err, domain.ErrInvalidWidget) || errors.Is(err, domain.ErrWidgetNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]string{
		"dashboard_id": dashboardID,
		"status":       "updated",
	})
}

// DeleteWidget handles DELETE /api/dashboards/{id}/widgets/{widgetId}.
func (h *DashboardsHandler) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]
	widgetID := vars["widgetId"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.DeleteWidget(r.Context(), actor, dashboardID, widgetID); err != nil {
		if errors.Is(err, domain.ErrWidgetNotFound) || errors.Is(err, domain.ErrDashboardNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]string{
		"id":           widgetID,
		"dashboard_id": dashboardID,
		"status":       "deleted",
	})
}

// TriggerAutomation handles POST /api/automations/{id}/trigger.
func (h *DashboardsHandler) TriggerAutomation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.TriggerAutomation(r.Context(), actor, id); err != nil {
		if errors.Is(err, domain.ErrAutomationNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, pkgdashboard.TriggerAutomationResponse{
		Status: "triggered",
		ID:     id,
	})
}
