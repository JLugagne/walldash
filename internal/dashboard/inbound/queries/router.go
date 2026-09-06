package queries

import (
	"net/http"

	svcdevices "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/gorilla/mux"
)

// SetupRoutes registers all query HTTP handlers onto the provided router.
func SetupRoutes(r *mux.Router, controller *inbound.Controller, queries svchealth.HealthQueries) {
	healthHandler := NewHealthHandler(controller, queries)
	r.HandleFunc("/api/health", healthHandler.GetHealth).Methods(http.MethodGet)

	if levelQueries, ok := queries.(svclevels.LevelQueries); ok {
		SetupLevelRoutes(r, controller, levelQueries)
	}

	if deviceQueries, ok := queries.(svcdevices.DeviceQueries); ok {
		SetupDeviceRoutes(r, controller, deviceQueries)
	}
}
