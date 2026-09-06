package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcoverviews "github.com/JLugagne/walldash/internal/dashboard/domain/service/overviews"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// OverviewsHandler handles mutating operations for overview dashboards, widgets, and automations.
type OverviewsHandler struct {
	controller *inbound.Controller
	commands   svcoverviews.OverviewCommands
	validate   *validator.Validate
}

// NewOverviewsHandler creates a new OverviewsHandler instance.
func NewOverviewsHandler(controller *inbound.Controller, commands svcoverviews.OverviewCommands) *OverviewsHandler {
	return &OverviewsHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// SetupOverviewRoutes registers all overview command routes onto the router.
func SetupOverviewRoutes(r *mux.Router, controller *inbound.Controller, commands svcoverviews.OverviewCommands) {
	handler := NewOverviewsHandler(controller, commands)
	r.HandleFunc("/api/overviews", handler.CreateOverview).Methods(http.MethodPost)
	r.HandleFunc("/api/overviews/{id}", handler.UpdateOverview).Methods(http.MethodPut)
	r.HandleFunc("/api/overviews/{id}", handler.DeleteOverview).Methods(http.MethodDelete)
	r.HandleFunc("/api/overviews/{id}/widgets", handler.CreateWidget).Methods(http.MethodPost)
	r.HandleFunc("/api/overviews/{id}/widgets/{widgetId}", handler.UpdateWidget).Methods(http.MethodPut)
	r.HandleFunc("/api/overviews/{id}/widgets/{widgetId}", handler.DeleteWidget).Methods(http.MethodDelete)
	r.HandleFunc("/api/overviews/{id}/layout", handler.UpdateLayout).Methods(http.MethodPut)
	r.HandleFunc("/api/automations/{id}/trigger", handler.TriggerAutomation).Methods(http.MethodPost)
}

// CreateOverview handles POST /api/overviews.
func (h *OverviewsHandler) CreateOverview(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.CreateOverviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateOverviewRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateOverviewRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainOverview := converters.ToDomainOverview(req)

	created, err := h.commands.CreateOverview(r.Context(), actor, domainOverview)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOverview) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicOverview(created)
	h.controller.SendSuccess(w, r, response)
}

// UpdateOverview handles PUT /api/overviews/{id}.
func (h *OverviewsHandler) UpdateOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pkgdashboard.UpdateOverviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateOverviewRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateOverviewRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	domainOverview := converters.ToDomainUpdateOverview(id, req)

	updated, err := h.commands.UpdateOverview(r.Context(), actor, domainOverview)
	if err != nil {
		if errors.Is(err, domain.ErrOverviewNotFound) || errors.Is(err, domain.ErrInvalidOverview) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicOverview(updated)
	h.controller.SendSuccess(w, r, response)
}

// DeleteOverview handles DELETE /api/overviews/{id}.
func (h *OverviewsHandler) DeleteOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.DeleteOverview(r.Context(), actor, id); err != nil {
		if errors.Is(err, domain.ErrOverviewNotFound) {
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

// CreateWidget handles POST /api/overviews/{id}/widgets.
func (h *OverviewsHandler) CreateWidget(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, domain.ErrOverviewNotFound) || errors.Is(err, domain.ErrInvalidWidget) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicWidget(created)
	h.controller.SendSuccess(w, r, response)
}

// UpdateWidget handles PUT /api/overviews/{id}/widgets/{widgetId}.
func (h *OverviewsHandler) UpdateWidget(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, domain.ErrWidgetNotFound) || errors.Is(err, domain.ErrOverviewNotFound) || errors.Is(err, domain.ErrInvalidWidget) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicWidget(updated)
	h.controller.SendSuccess(w, r, response)
}

// UpdateLayout handles PUT /api/overviews/{id}/layout.
func (h *OverviewsHandler) UpdateLayout(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, domain.ErrOverviewNotFound) || errors.Is(err, domain.ErrInvalidOverview) ||
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

// DeleteWidget handles DELETE /api/overviews/{id}/widgets/{widgetId}.
func (h *OverviewsHandler) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]
	widgetID := vars["widgetId"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.DeleteWidget(r.Context(), actor, dashboardID, widgetID); err != nil {
		if errors.Is(err, domain.ErrWidgetNotFound) || errors.Is(err, domain.ErrOverviewNotFound) {
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
func (h *OverviewsHandler) TriggerAutomation(w http.ResponseWriter, r *http.Request) {
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
