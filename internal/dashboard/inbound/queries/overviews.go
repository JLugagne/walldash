package queries

import (
	"errors"
	"net/http"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svcoverviews "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/overviews"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	"github.com/gorilla/mux"
)

// OverviewsHandler handles query requests for overview dashboards and automations.
type OverviewsHandler struct {
	controller *inbound.Controller
	queries    svcoverviews.OverviewQueries
}

// NewOverviewsHandler creates a new OverviewsHandler instance.
func NewOverviewsHandler(controller *inbound.Controller, queries svcoverviews.OverviewQueries) *OverviewsHandler {
	return &OverviewsHandler{
		controller: controller,
		queries:    queries,
	}
}

// ListOverviews handles GET /api/overviews.
func (h *OverviewsHandler) ListOverviews(w http.ResponseWriter, r *http.Request) {
	overviews, err := h.queries.ListOverviews(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicOverviews(overviews)
	h.controller.SendSuccess(w, r, response)
}

// GetOverview handles GET /api/overviews/{id}.
func (h *OverviewsHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	overview, err := h.queries.GetOverview(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOverviewNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicOverview(overview)
	h.controller.SendSuccess(w, r, response)
}

// ListAutomations handles GET /api/automations.
func (h *OverviewsHandler) ListAutomations(w http.ResponseWriter, r *http.Request) {
	automations, err := h.queries.ListAutomations(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicAutomations(automations)
	h.controller.SendSuccess(w, r, response)
}

// SetupOverviewRoutes registers overview query routes on the router.
func SetupOverviewRoutes(r *mux.Router, controller *inbound.Controller, queries svcoverviews.OverviewQueries) {
	handler := NewOverviewsHandler(controller, queries)
	r.HandleFunc("/api/overviews", handler.ListOverviews).Methods(http.MethodGet)
	r.HandleFunc("/api/overviews/{id}", handler.GetOverview).Methods(http.MethodGet)
	r.HandleFunc("/api/automations", handler.ListAutomations).Methods(http.MethodGet)
}
