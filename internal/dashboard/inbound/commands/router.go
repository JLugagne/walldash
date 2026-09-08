package commands

import (
	svcactions "github.com/JLugagne/walldash/internal/dashboard/domain/service/actions"
	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	svcoverviews "github.com/JLugagne/walldash/internal/dashboard/domain/service/overviews"
	restore "github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
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

	if overviewCommands, ok := commands.(svcoverviews.OverviewCommands); ok {
		SetupOverviewRoutes(r, controller, overviewCommands)
	}
	if restoreCommands, ok := commands.(restore.RestoreCommands); ok {
		SetupRestoreRoutes(r, controller, restoreCommands)
	}
}
