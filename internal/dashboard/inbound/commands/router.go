package commands

import (
	svchealth "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/gorilla/mux"
)

// SetupRoutes registers command HTTP handlers onto the provided router.
func SetupRoutes(r *mux.Router, controller *inbound.Controller, commands svchealth.HealthCommands) {
	// Future command handlers (e.g. POST /api/actions, PUT /api/plans) will be registered here
	_ = r
	_ = controller
	_ = commands
}
