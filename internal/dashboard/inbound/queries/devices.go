package queries

import (
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	"github.com/gorilla/mux"
)

// DevicesHandler processes read-only HTTP requests for devices and device placements.
type DevicesHandler struct {
	controller *inbound.Controller
	queries    svcdevices.DeviceQueries
}

// NewDevicesHandler constructs a new DevicesHandler.
func NewDevicesHandler(controller *inbound.Controller, queries svcdevices.DeviceQueries) *DevicesHandler {
	return &DevicesHandler{
		controller: controller,
		queries:    queries,
	}
}

// SetupDeviceRoutes registers device query endpoints onto the provided router.
func SetupDeviceRoutes(r *mux.Router, controller *inbound.Controller, queries svcdevices.DeviceQueries) {
	handler := NewDevicesHandler(controller, queries)
	r.HandleFunc("/api/devices", handler.ListDevices).Methods(http.MethodGet)
	r.HandleFunc("/api/levels/{id}/placements", handler.ListPlacements).Methods(http.MethodGet)
}

// ListDevices handles GET /api/devices.
func (h *DevicesHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.queries.ListAvailableDevices(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicDevices(devices)
	h.controller.SendSuccess(w, r, response)
}

// ListPlacements handles GET /api/levels/{id}/placements.
func (h *DevicesHandler) ListPlacements(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	levelID := vars["id"]

	placements, err := h.queries.ListPlacements(r.Context(), levelID)
	if err != nil {
		if errors.Is(err, domain.ErrLevelNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicPlacements(placements)
	h.controller.SendSuccess(w, r, response)
}
