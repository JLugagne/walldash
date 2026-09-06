package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// LevelsHandler processes command HTTP requests for levels and plans.
type LevelsHandler struct {
	controller *inbound.Controller
	commands   svclevels.LevelCommands
	validate   *validator.Validate
}

// NewLevelsHandler constructs a new LevelsHandler.
func NewLevelsHandler(controller *inbound.Controller, commands svclevels.LevelCommands) *LevelsHandler {
	return &LevelsHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// CreateLevel handles POST /api/levels.
func (h *LevelsHandler) CreateLevel(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.CreateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateLevelRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidCreateLevelRequest, err))
		return
	}

	domainLevel := converters.ToDomainCreateLevel(req)
	actor := domain.ActorFromContext(r.Context())

	created, err := h.commands.CreateLevel(r.Context(), actor, domainLevel)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLevel) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicLevel(created)
	h.controller.SendSuccess(w, r, response)
}

// UpdateLevel handles PUT /api/levels/{id}.
func (h *LevelsHandler) UpdateLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pkgdashboard.UpdateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateLevelRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidUpdateLevelRequest, err))
		return
	}

	domainLevel := converters.ToDomainUpdateLevel(id, req)
	actor := domain.ActorFromContext(r.Context())

	updated, err := h.commands.UpdateLevel(r.Context(), actor, domainLevel)
	if err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrInvalidLevel) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicLevel(updated)
	h.controller.SendSuccess(w, r, response)
}

// DeleteLevel handles DELETE /api/levels/{id}.
func (h *LevelsHandler) DeleteLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.DeleteLevel(r.Context(), actor, id); err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrInvalidLevel) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]string{"deleted_id": id})
}

// ReorderLevels handles POST /api/levels/reorder.
func (h *LevelsHandler) ReorderLevels(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.ReorderLevelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidReorderLevelsRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidReorderLevelsRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	if err := h.commands.ReorderLevels(r.Context(), actor, req.LevelIDs); err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrInvalidLevel) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]bool{"reordered": true})
}

// SavePlan handles PUT /api/levels/{id}/plan.
func (h *LevelsHandler) SavePlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pkgdashboard.SavePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, err))
		return
	}

	domainPlan := converters.ToDomainPlan(id, req)
	actor := domain.ActorFromContext(r.Context())

	saved, err := h.commands.SavePlan(r.Context(), actor, domainPlan)
	if err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrInvalidPlan) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicPlan(saved)
	h.controller.SendSuccess(w, r, response)
}

// SetupLevelRoutes registers level command routes on the router.
func SetupLevelRoutes(r *mux.Router, controller *inbound.Controller, commands svclevels.LevelCommands) {
	handler := NewLevelsHandler(controller, commands)
	r.HandleFunc("/api/levels", handler.CreateLevel).Methods(http.MethodPost)
	r.HandleFunc("/api/levels/reorder", handler.ReorderLevels).Methods(http.MethodPost)
	r.HandleFunc("/api/levels/{id}/plan", handler.SavePlan).Methods(http.MethodPut)
	r.HandleFunc("/api/levels/{id}", handler.UpdateLevel).Methods(http.MethodPut)
	r.HandleFunc("/api/levels/{id}", handler.DeleteLevel).Methods(http.MethodDelete)
}
