package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// RestoreHandler replays backup snapshots.
type RestoreHandler struct {
	controller *inbound.Controller
	commands   restore.RestoreCommands
	validate   *validator.Validate
}

// NewRestoreHandler creates a RestoreHandler.
func NewRestoreHandler(controller *inbound.Controller, commands restore.RestoreCommands) *RestoreHandler {
	return &RestoreHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// SetupRestoreRoutes registers restore HTTP handlers onto the provided router.
func SetupRestoreRoutes(r *mux.Router, controller *inbound.Controller, commands restore.RestoreCommands) {
	handler := NewRestoreHandler(controller, commands)
	r.HandleFunc("/api/restore", handler.RestoreBackup).Methods(http.MethodPost)
}

// RestoreBackup replaces current data with the snapshot in the request body.
// Device placements are only replayed when include_devices is set.
func (h *RestoreHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidRestoreRequest, err))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidRestoreRequest, err))
		return
	}
	if strings.TrimSpace(req.Version) != "1" {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidRestoreRequest, errors.New("unsupported restore version: "+req.Version)))
		return
	}
	actor := domain.ActorFromContext(r.Context())
	input := converters.ToDomainRestoreInput(req)
	summary, err := h.commands.RestoreBackup(r.Context(), actor, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLevel) || errors.Is(err, domain.ErrInvalidPlan) || errors.Is(err, domain.ErrInvalidPlacement) || errors.Is(err, domain.ErrInvalidOverview) || errors.Is(err, domain.ErrInvalidWidget) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}
	h.controller.SendSuccess(w, r, converters.ToPublicRestore(summary))
}
