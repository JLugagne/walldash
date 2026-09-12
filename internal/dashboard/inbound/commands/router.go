package commands

import (
	svcactions "github.com/JLugagne/walldash/internal/dashboard/domain/service/actions"
	svcdashboards "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	restore "github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/gorilla/mux"
)

// SetupRoutes registers command HTTP handlers onto the provided router.
func SetupRoutes(r *mux.Router, controller *inbound.Controller, commands svchealth.HealthCommands) {
	SetupAdminRoutes(r, controller, commands)

	if actionCommands, ok := commands.(svcactions.ActionCommands); ok {
		SetupActionRoutes(r, controller, actionCommands)
	}
}

// SetupAdminRoutes registers the owner-only management command routes (levels, placements,
// dashboards, widgets, restore). The action route is intentionally excluded: a device
// account is allowed to trigger whitelisted actions.
func SetupAdminRoutes(r *mux.Router, controller *inbound.Controller, commands svchealth.HealthCommands) {
	if levelCommands, ok := commands.(svclevels.LevelCommands); ok {
		SetupLevelRoutes(r, controller, levelCommands)
	}

	if deviceCommands, ok := commands.(svcdevices.DeviceCommands); ok {
		SetupDeviceRoutes(r, controller, deviceCommands)
	}

	if dashboardCommands, ok := commands.(svcdashboards.DashboardCommands); ok {
		SetupDashboardRoutes(r, controller, dashboardCommands)
	}
	if restoreCommands, ok := commands.(restore.RestoreCommands); ok {
		SetupRestoreRoutes(r, controller, restoreCommands)
	}
}
