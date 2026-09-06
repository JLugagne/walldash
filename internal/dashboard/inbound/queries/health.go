package queries

import (
	"net/http"

	svchealth "github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
)

// HealthHandler handles health check query endpoints.
type HealthHandler struct {
	controller *inbound.Controller
	queries    svchealth.HealthQueries
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(controller *inbound.Controller, queries svchealth.HealthQueries) *HealthHandler {
	return &HealthHandler{
		controller: controller,
		queries:    queries,
	}
}

// GetHealth handles GET /api/health requests.
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.queries.GetHealth(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}

	response := converters.ToPublicHealth(health)
	h.controller.SendSuccess(w, r, response)
}
