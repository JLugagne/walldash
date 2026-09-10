package queries

import (
	"net/http"

	svcweather "github.com/JLugagne/walldash/internal/dashboard/domain/service/weather"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	"github.com/gorilla/mux"
)

// WeatherHandler handles weather configuration query endpoints.
type WeatherHandler struct {
	controller *inbound.Controller
	queries    svcweather.WeatherQueries
}

// NewWeatherHandler creates a new WeatherHandler.
func NewWeatherHandler(controller *inbound.Controller, queries svcweather.WeatherQueries) *WeatherHandler {
	return &WeatherHandler{
		controller: controller,
		queries:    queries,
	}
}

// GetWeatherConfig handles GET /api/weather/config.
func (h *WeatherHandler) GetWeatherConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.queries.GetWeatherConfig(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}
	h.controller.SendSuccess(w, r, converters.ToPublicWeatherConfig(config))
}

// SetupWeatherRoutes registers weather query routes on the router.
func SetupWeatherRoutes(r *mux.Router, controller *inbound.Controller, queries svcweather.WeatherQueries) {
	handler := NewWeatherHandler(controller, queries)
	r.HandleFunc("/api/weather/config", handler.GetWeatherConfig).Methods(http.MethodGet)
}
