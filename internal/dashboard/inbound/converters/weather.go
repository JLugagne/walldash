package converters

import (
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ToPublicWeatherConfig converts a domain HomeConfig into a public WeatherConfigResponse.
func ToPublicWeatherConfig(c domain.HomeConfig) pkgdashboard.WeatherConfigResponse {
	return pkgdashboard.WeatherConfigResponse{
		Configured:      c.Configured,
		Latitude:        c.Latitude,
		Longitude:       c.Longitude,
		Name:            c.LocationName,
		TemperatureUnit: temperatureUnitName(c.TemperatureUnit),
		WindUnit:        windUnitName(c.LengthUnit),
	}
}

func temperatureUnitName(raw string) string {
	if strings.Contains(strings.ToUpper(raw), "F") {
		return "fahrenheit"
	}
	return "celsius"
}

func windUnitName(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mi", "mile", "miles":
		return "mph"
	default:
		return "kmh"
	}
}
