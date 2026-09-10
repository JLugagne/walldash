package converters_test

import (
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeatherWidgetConversion(t *testing.T) {
	lat, lon := 48.8566, 2.3522
	req := pkgdashboard.CreateWidgetRequest{
		Type:    "weather",
		ColSpan: 2,
		RowSpan: 2,
		Config: pkgdashboard.WidgetConfigDTO{
			Display:      "weather",
			WeatherMode:  "ndays",
			WeatherDays:  5,
			Latitude:     &lat,
			Longitude:    &lon,
			LocationName: "Paris",
			Units:        "metric",
		},
	}
	w := converters.ToDomainWidget("ov-1", req)
	assert.Equal(t, domain.WeatherModeNDays, w.Config.WeatherMode)
	assert.Equal(t, 5, w.Config.WeatherDays)
	require.NotNil(t, w.Config.Latitude)
	assert.Equal(t, lat, *w.Config.Latitude)
	assert.Equal(t, "Paris", w.Config.LocationName)
	assert.Equal(t, "metric", w.Config.Units)

	pub := converters.ToPublicWidget(w).Config
	assert.Equal(t, "ndays", pub.WeatherMode)
	assert.Equal(t, 5, pub.WeatherDays)
	assert.Equal(t, "Paris", pub.LocationName)

	t.Run("defaults missing weather mode to current", func(t *testing.T) {
		dom := converters.ToDomainWidget("ov-1", pkgdashboard.CreateWidgetRequest{
			Type: "weather", ColSpan: 2, RowSpan: 2,
			Config: pkgdashboard.WidgetConfigDTO{Display: "weather"},
		})
		assert.Equal(t, domain.WeatherModeCurrent, dom.Config.WeatherMode)
	})

	t.Run("does not pollute non-weather configs", func(t *testing.T) {
		dom := converters.ToDomainWidget("ov-1", pkgdashboard.CreateWidgetRequest{
			Type: "sensor", ColSpan: 1, RowSpan: 1,
			Config: pkgdashboard.WidgetConfigDTO{Display: "number", EntityIDs: []string{"sensor.temp"}},
		})
		assert.Empty(t, dom.Config.WeatherMode)
	})
}

func TestToPublicWeatherConfig(t *testing.T) {
	cases := map[string]struct {
		cfg      domain.HomeConfig
		tempUnit string
		windUnit string
	}{
		"metric celsius":      {domain.HomeConfig{TemperatureUnit: "°C", LengthUnit: "km"}, "celsius", "kmh"},
		"imperial fahrenheit": {domain.HomeConfig{TemperatureUnit: "°F", LengthUnit: "mi"}, "fahrenheit", "mph"},
		"empty defaults":      {domain.HomeConfig{}, "celsius", "kmh"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := converters.ToPublicWeatherConfig(tc.cfg)
			assert.Equal(t, tc.tempUnit, got.TemperatureUnit)
			assert.Equal(t, tc.windUnit, got.WindUnit)
		})
	}
}
