package commands

import (
	svcactions "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/actions"
	svcdevices "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/levels"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/gorilla/mux"
)

// SetupRoutes registers command HTTP handlers onto the provided router.
func SetupRoutes(r *mux.Router, controller *inbound.Controller, commands svchealth.HealthCommands) {
	if levelCommands, ok := commands.(svclevels.LevelCommands); ok {
		SetupLevelRoutes(r, controller, levelCommands)
	}

	if deviceCommands, ok := commands.(svcdevices.DeviceCommands); ok {
		SetupDeviceRoutes(r, controller, deviceCommands)
	}

	if actionCommands, ok := commands.(svcactions.ActionCommands); ok {
		SetupActionRoutes(r, controller, actionCommands)
	}
}
