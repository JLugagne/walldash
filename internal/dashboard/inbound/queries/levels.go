package queries

import (
	"errors"
	"net/http"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svclevels "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	"github.com/gorilla/mux"
)

// LevelsHandler handles query requests for levels and plans.
type LevelsHandler struct {
	controller *inbound.Controller
	queries    svclevels.LevelQueries
}

// NewLevelsHandler creates a new LevelsHandler instance.
func NewLevelsHandler(controller *inbound.Controller, queries svclevels.LevelQueries) *LevelsHandler {
	return &LevelsHandler{
		controller: controller,
		queries:    queries,
	}
}

// ListLevels handles GET /api/levels.
func (h *LevelsHandler) ListLevels(w http.ResponseWriter, r *http.Request) {
	levels, err := h.queries.ListLevels(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicLevels(levels)
	h.controller.SendSuccess(w, r, response)
}

// GetLevel handles GET /api/levels/{id}.
func (h *LevelsHandler) GetLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	level, err := h.queries.GetLevel(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicLevel(level)
	h.controller.SendSuccess(w, r, response)
}

// GetPlan handles GET /api/levels/{id}/plan.
func (h *LevelsHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	plan, err := h.queries.GetPlan(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) || errors.Is(err, domain.ErrLevelNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	response := converters.ToPublicPlan(plan)
	h.controller.SendSuccess(w, r, response)
}

// SetupLevelRoutes registers level query routes on the router.
func SetupLevelRoutes(r *mux.Router, controller *inbound.Controller, queries svclevels.LevelQueries) {
	handler := NewLevelsHandler(controller, queries)
	r.HandleFunc("/api/levels", handler.ListLevels).Methods(http.MethodGet)
	r.HandleFunc("/api/levels/{id}/plan", handler.GetPlan).Methods(http.MethodGet)
	r.HandleFunc("/api/levels/{id}", handler.GetLevel).Methods(http.MethodGet)
}
