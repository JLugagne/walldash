package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/aijson"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/sweethome3d"
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
	r.HandleFunc("/api/levels/import/sh3d", handler.ImportSh3dLevels).Methods(http.MethodPost)
	r.HandleFunc("/api/levels/{id}/plan/import", handler.ImportPlan).Methods(http.MethodPost)
	r.HandleFunc("/api/levels/{id}/plan", handler.SavePlan).Methods(http.MethodPut)
	r.HandleFunc("/api/levels/{id}", handler.UpdateLevel).Methods(http.MethodPut)
	r.HandleFunc("/api/levels/{id}", handler.DeleteLevel).Methods(http.MethodDelete)
}

func (h *LevelsHandler) ImportPlan(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	vars := mux.Vars(r)
	id := vars["id"]

	var plan domain.Plan
	var err error

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		reader, parseErr := r.MultipartReader()
		if parseErr != nil {
			h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, parseErr))
			return
		}
		part, partErr := reader.NextPart()
		if partErr != nil {
			h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, partErr))
			return
		}
		plan, err = sweethome3d.FromReader(part, id)
	} else {
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, readErr))
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		plan, err = aijson.FromJSON(body, id)
	}

	if err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, err))
		return
	}

	actor := domain.ActorFromContext(r.Context())
	saved, err := h.commands.SavePlan(r.Context(), actor, plan)
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

// ImportSh3dLevels imports a Sweet Home 3D archive, creating one Level per
// declared level (ordered by elevation), each with its imported plan.
func (h *LevelsHandler) ImportSh3dLevels(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	reader, parseErr := r.MultipartReader()
	if parseErr != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, parseErr))
		return
	}
	part, partErr := reader.NextPart()
	if partErr != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, partErr))
		return
	}
	actor := domain.ActorFromContext(r.Context())
	created, err := h.commands.ImportSh3dLevels(r.Context(), actor, part)
	if err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlanRequest, err))
		return
	}
	response := converters.ToPublicLevels(created)
	h.controller.SendSuccess(w, r, response)
}
