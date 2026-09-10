package domain

// Weather modes select which slice of the forecast a Weather Widget renders.
const (
	WeatherModeCurrent  = "current"
	WeatherModeToday    = "today"
	WeatherModeTomorrow = "tomorrow"
	WeatherModeNDays    = "ndays"

	WeatherUnitsMetric   = "metric"
	WeatherUnitsImperial = "imperial"
)

// HomeConfig holds the subset of the Home Assistant instance configuration that
// Weather Widgets need: the instance location and its preferred unit system.
type HomeConfig struct {
	Configured      bool
	Latitude        float64
	Longitude       float64
	LocationName    string
	TemperatureUnit string
	LengthUnit      string
}
