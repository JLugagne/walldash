package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	svcdevices "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// DevicesHandler processes command HTTP requests for device placements.
type DevicesHandler struct {
	controller *inbound.Controller
	commands   svcdevices.DeviceCommands
	validate   *validator.Validate
}

// NewDevicesHandler constructs a new DevicesHandler.
func NewDevicesHandler(controller *inbound.Controller, commands svcdevices.DeviceCommands) *DevicesHandler {
	return &DevicesHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// SetupDeviceRoutes registers device command endpoints onto the router.
func SetupDeviceRoutes(r *mux.Router, controller *inbound.Controller, commands svcdevices.DeviceCommands) {
	handler := NewDevicesHandler(controller, commands)
	r.HandleFunc("/api/levels/{id}/placements", handler.SavePlacement).Methods(http.MethodPost)
	r.HandleFunc("/api/levels/{id}/placements/{placementId}", handler.DeletePlacement).Methods(http.MethodDelete)
}

// SavePlacement handles POST /api/levels/{id}/placements.
func (h *DevicesHandler) SavePlacement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	levelID := vars["id"]

	var req pkgdashboard.SavePlacementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlacementRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidSavePlacementRequest, err))
		return
	}

	domainPlacement := converters.ToDomainSavePlacement(levelID, req)
	actor := domain.ActorFromContext(r.Context())

	saved, err := h.commands.SavePlacement(r.Context(), actor, domainPlacement)
	if err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrInvalidPlacement) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicPlacement(saved)
	h.controller.SendSuccess(w, r, response)
}

// DeletePlacement handles DELETE /api/levels/{id}/placements/{placementId}.
func (h *DevicesHandler) DeletePlacement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	levelID := vars["id"]
	placementID := vars["placementId"]

	actor := domain.ActorFromContext(r.Context())

	if err := h.commands.DeletePlacement(r.Context(), actor, levelID, placementID); err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) || errors.Is(err, domain.ErrPlacementNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	h.controller.SendSuccess(w, r, map[string]string{
		"status": "deleted",
		"id":     placementID,
	})
}
