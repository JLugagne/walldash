package queries

import (
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcdashboards "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	"github.com/gorilla/mux"
)

// DashboardsHandler handles query requests for dashboards and automations.
type DashboardsHandler struct {
	controller *inbound.Controller
	queries    svcdashboards.DashboardQueries
}

// NewDashboardsHandler creates a new DashboardsHandler instance.
func NewDashboardsHandler(controller *inbound.Controller, queries svcdashboards.DashboardQueries) *DashboardsHandler {
	return &DashboardsHandler{
		controller: controller,
		queries:    queries,
	}
}

// ListDashboards handles GET /api/dashboards.
func (h *DashboardsHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	dashboards, err := h.queries.ListDashboards(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicDashboards(dashboards)
	h.controller.SendSuccess(w, r, response)
}

// GetDashboard handles GET /api/dashboards/{id}.
func (h *DashboardsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dashboard, err := h.queries.GetDashboard(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDashboardNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicDashboard(dashboard)
	h.controller.SendSuccess(w, r, response)
}

// ListAutomations handles GET /api/automations.
func (h *DashboardsHandler) ListAutomations(w http.ResponseWriter, r *http.Request) {
	automations, err := h.queries.ListAutomations(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicAutomations(automations)
	h.controller.SendSuccess(w, r, response)
}

// SetupDashboardRoutes registers dashboard query routes on the router.
func SetupDashboardRoutes(r *mux.Router, controller *inbound.Controller, queries svcdashboards.DashboardQueries) {
	handler := NewDashboardsHandler(controller, queries)
	r.HandleFunc("/api/dashboards", handler.ListDashboards).Methods(http.MethodGet)
	r.HandleFunc("/api/dashboards/{id}", handler.GetDashboard).Methods(http.MethodGet)
	r.HandleFunc("/api/automations", handler.ListAutomations).Methods(http.MethodGet)
}
