package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	svcactions "github.com/JLugagne/walldash/internal/dashboard/domain/service/actions"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// ActionsHandler processes command HTTP requests for executing device actions.
type ActionsHandler struct {
	controller *inbound.Controller
	commands   svcactions.ActionCommands
	validate   *validator.Validate
}

// NewActionsHandler constructs a new ActionsHandler.
func NewActionsHandler(controller *inbound.Controller, commands svcactions.ActionCommands) *ActionsHandler {
	return &ActionsHandler{
		controller: controller,
		commands:   commands,
		validate:   validator.New(),
	}
}

// SetupActionRoutes registers action execution routes onto the router.
func SetupActionRoutes(r *mux.Router, controller *inbound.Controller, commands svcactions.ActionCommands) {
	handler := NewActionsHandler(controller, commands)
	r.HandleFunc("/api/actions", handler.ExecuteAction).Methods(http.MethodPost).Name(middleware.DeviceSafeRoutePrefix + "actions")
}

// ExecuteAction handles POST /api/actions.
func (h *ActionsHandler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	var req pkgdashboard.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidActionRequest, err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.controller.SendFail(w, r, nil, errors.Join(pkgdashboard.ErrInvalidActionRequest, err))
		return
	}

	domainCmd := converters.ToDomainAction(req)
	actor := domain.ActorFromContext(r.Context())

	if err := h.commands.ExecuteAction(r.Context(), actor, domainCmd); err != nil {
		if errors.Is(err, domain.ErrActionNotAllowed) || errors.Is(err, domain.ErrDeviceNotFound) {
			h.controller.SendFail(w, r, nil, err)
			return
		}
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicAction(domainCmd)
	h.controller.SendSuccess(w, r, response)
}
