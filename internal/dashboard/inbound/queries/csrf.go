package queries

import (
	"net/http"

	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/middleware"
	"github.com/gorilla/mux"
)

// CSRFHandler handles requests to obtain anti-CSRF tokens.
type CSRFHandler struct {
	controller   *inbound.Controller
	tokenManager middleware.TokenManager
}

// NewCSRFHandler creates a new CSRFHandler.
func NewCSRFHandler(controller *inbound.Controller, tokenManager middleware.TokenManager) *CSRFHandler {
	return &CSRFHandler{
		controller:   controller,
		tokenManager: tokenManager,
	}
}

// GetToken issues a new anti-CSRF token.
func (h *CSRFHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	token := h.tokenManager.GenerateToken()
	response := converters.ToPublicCSRFToken(token)
	h.controller.SendSuccess(w, r, response)
}

// SetupCSRFRoutes registers the CSRF token route on the provided router.
func SetupCSRFRoutes(r *mux.Router, controller *inbound.Controller, tokenManager middleware.TokenManager) {
	handler := NewCSRFHandler(controller, tokenManager)
	r.HandleFunc("/api/csrf-token", handler.GetToken).Methods(http.MethodGet)
}
