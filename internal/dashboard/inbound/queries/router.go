package queries

import (
	"net/http"

	svcdashboards "github.com/JLugagne/walldash/internal/dashboard/domain/service/dashboards"
	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	svchealth "github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	svcweather "github.com/JLugagne/walldash/internal/dashboard/domain/service/weather"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/middleware"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router, controller *inbound.Controller, queries svchealth.HealthQueries, tokenManager ...middleware.TokenManager) {
	healthHandler := NewHealthHandler(controller, queries)
	r.HandleFunc("/api/health", healthHandler.GetHealth).Methods(http.MethodGet)

	if len(tokenManager) > 0 && tokenManager[0] != nil {
		SetupCSRFRoutes(r, controller, tokenManager[0])
	}

	if levelQueries, ok := queries.(svclevels.LevelQueries); ok {
		SetupLevelRoutes(r, controller, levelQueries)
	}

	if deviceQueries, ok := queries.(svcdevices.DeviceQueries); ok {
		SetupDeviceRoutes(r, controller, deviceQueries)
	}

	if dashboardQueries, ok := queries.(svcdashboards.DashboardQueries); ok {
		SetupDashboardRoutes(r, controller, dashboardQueries)
	}

	if weatherQueries, ok := queries.(svcweather.WeatherQueries); ok {
		SetupWeatherRoutes(r, controller, weatherQueries)
	}
	levelQueries, lOk := queries.(svclevels.LevelQueries)
	deviceQueries, dOk := queries.(svcdevices.DeviceQueries)
	dashboardQueries, oOk := queries.(svcdashboards.DashboardQueries)
	if lOk && dOk && oOk {
		exportHandler := NewExportHandler(controller, levelQueries, deviceQueries, dashboardQueries)
		r.HandleFunc("/api/export", exportHandler.Export).Methods(http.MethodGet)
	}
}
