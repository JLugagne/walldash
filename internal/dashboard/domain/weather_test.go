package domain_test

import (
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func weatherWidget() domain.Widget {
	return domain.Widget{
		ID:          "widget-weather",
		DashboardID: "ov-1",
		Type:        domain.WidgetTypeWeather,
		ColSpan:     2,
		RowSpan:     2,
		Config: domain.WidgetConfig{
			EntityIDs:   []string{},
			Display:     domain.DisplayWeather,
			WeatherMode: domain.WeatherModeCurrent,
		},
	}
}

func TestWidgetValidateWeather(t *testing.T) {
	t.Run("accepts current", func(t *testing.T) {
		require.NoError(t, weatherWidget().Validate())
	})

	t.Run("accepts ndays within range", func(t *testing.T) {
		w := weatherWidget()
		w.Config.WeatherMode = domain.WeatherModeNDays
		w.Config.WeatherDays = 5
		require.NoError(t, w.Validate())
	})

	t.Run("accepts imperial units", func(t *testing.T) {
		w := weatherWidget()
		w.Config.Units = domain.WeatherUnitsImperial
		require.NoError(t, w.Validate())
	})

	rejections := map[string]func() domain.Widget{
		"invalid mode": func() domain.Widget {
			w := weatherWidget()
			w.Config.WeatherMode = "hourly"
			return w
		},
		"ndays without days": func() domain.Widget {
			w := weatherWidget()
			w.Config.WeatherMode = domain.WeatherModeNDays
			return w
		},
		"ndays above range": func() domain.Widget {
			w := weatherWidget()
			w.Config.WeatherMode = domain.WeatherModeNDays
			w.Config.WeatherDays = 15
			return w
		},
		"invalid units": func() domain.Widget {
			w := weatherWidget()
			w.Config.Units = "kelvin"
			return w
		},
		"with entity id": func() domain.Widget {
			w := weatherWidget()
			w.Config.EntityIDs = []string{"sensor.temp"}
			return w
		},
	}

	for name, build := range rejections {
		t.Run(name+" returns ErrInvalidWidget", func(t *testing.T) {
			err := build().Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidWidget)
		})
	}
}
