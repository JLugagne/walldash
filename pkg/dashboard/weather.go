package dashboard

// WeatherConfigResponse represents the Home Assistant location and unit
// configuration used to seed Weather Widget defaults.
type WeatherConfigResponse struct {
	Configured      bool    `json:"configured"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Name            string  `json:"name"`
	TemperatureUnit string  `json:"temperature_unit"`
	WindUnit        string  `json:"wind_unit"`
}
