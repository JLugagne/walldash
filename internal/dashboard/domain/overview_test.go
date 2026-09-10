package domain_test

import (
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatPtr(v float64) *float64 {
	return &v
}

func sensorNumberWidget() domain.Widget {
	now := time.Now().UTC()
	return domain.Widget{
		ID:          "widget-number",
		DashboardID: "ov-1",
		Type:        domain.WidgetTypeSensor,
		Title:       "Living Room Temperature",
		ColSpan:     1,
		RowSpan:     1,
		Config: domain.WidgetConfig{
			EntityIDs: []string{"sensor.temperature_salon"},
			Display:   domain.DisplayNumber,
			Unit:      "°C",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sensorArcWidget() domain.Widget {
	w := sensorNumberWidget()
	w.ID = "widget-arc"
	w.ColSpan = 2
	w.RowSpan = 2
	w.Config.Display = domain.DisplayArc
	w.Config.Min = floatPtr(0)
	w.Config.Max = floatPtr(40)
	return w
}

func sensorBarWidget() domain.Widget {
	w := sensorNumberWidget()
	w.ID = "widget-bar"
	w.ColSpan = 2
	w.RowSpan = 1
	w.Config.Display = domain.DisplayBar
	w.Config.Min = floatPtr(0)
	w.Config.Max = floatPtr(100)
	return w
}

func actuatorToggleWidget() domain.Widget {
	w := sensorNumberWidget()
	w.ID = "widget-toggle"
	w.Type = domain.WidgetTypeActuator
	w.Title = "Ceiling Light"
	w.Config.Display = domain.DisplayToggle
	w.Config.EntityIDs = []string{"light.salon_plafond"}
	w.Config.Unit = ""
	return w
}

func automationListWidget() domain.Widget {
	w := sensorNumberWidget()
	w.ID = "widget-list"
	w.Type = domain.WidgetTypeAutomationList
	w.Title = "Quick Automations"
	w.ColSpan = 2
	w.RowSpan = 2
	w.Config.Display = domain.DisplayList
	w.Config.EntityIDs = []string{"automation.eteindre_tout", "automation.depart"}
	w.Config.Unit = ""
	return w
}

func TestMinimumWidgetSizeMatrix(t *testing.T) {
	legal := []struct {
		widgetType string
		display    string
		want       domain.WidgetSize
	}{
		{domain.WidgetTypeSensor, domain.DisplayNumber, domain.WidgetSize{Cols: 1, Rows: 1}},
		{domain.WidgetTypeSensor, domain.DisplayBar, domain.WidgetSize{Cols: 2, Rows: 1}},
		{domain.WidgetTypeSensor, domain.DisplayArc, domain.WidgetSize{Cols: 2, Rows: 2}},
		{domain.WidgetTypeActuator, domain.DisplayToggle, domain.WidgetSize{Cols: 1, Rows: 1}},
		{domain.WidgetTypeAutomationList, domain.DisplayList, domain.WidgetSize{Cols: 2, Rows: 2}},
	}
	for _, tc := range legal {
		t.Run("legal "+tc.widgetType+"/"+tc.display, func(t *testing.T) {
			got, ok := domain.MinimumWidgetSize(tc.widgetType, tc.display)
			require.True(t, ok)
			assert.Equal(t, tc.want, got)
		})
	}

	illegal := []struct {
		widgetType string
		display    string
	}{
		{domain.WidgetTypeSensor, domain.DisplayToggle},
		{domain.WidgetTypeSensor, domain.DisplayList},
		{domain.WidgetTypeActuator, domain.DisplayNumber},
		{domain.WidgetTypeActuator, domain.DisplayArc},
		{domain.WidgetTypeActuator, domain.DisplayBar},
		{domain.WidgetTypeActuator, domain.DisplayList},
		{domain.WidgetTypeAutomationList, domain.DisplayNumber},
		{domain.WidgetTypeAutomationList, domain.DisplayToggle},
		{domain.WidgetTypeSensor, ""},
		{domain.WidgetTypeSensor, "tile"},
		{"unknown_type", domain.DisplayNumber},
	}
	for _, tc := range illegal {
		t.Run("illegal "+tc.widgetType+"/"+tc.display, func(t *testing.T) {
			_, ok := domain.MinimumWidgetSize(tc.widgetType, tc.display)
			assert.False(t, ok)
		})
	}
}

func TestIsSupportedWidgetType(t *testing.T) {
	for _, wt := range []string{domain.WidgetTypeSensor, domain.WidgetTypeActuator, domain.WidgetTypeAutomationList} {
		t.Run("supported "+wt, func(t *testing.T) {
			assert.True(t, domain.IsSupportedWidgetType(wt))
		})
	}
	for _, wt := range []string{"", "sensor_gauge", "automation", "toggle"} {
		t.Run("unsupported "+wt, func(t *testing.T) {
			assert.False(t, domain.IsSupportedWidgetType(wt))
		})
	}
}

func TestWidgetValidateAccepts(t *testing.T) {
	tests := map[string]domain.Widget{
		"sensor number":   sensorNumberWidget(),
		"sensor bar":      sensorBarWidget(),
		"sensor arc":      sensorArcWidget(),
		"actuator toggle": actuatorToggleWidget(),
		"automation list": automationListWidget(),
		"empty title is now accepted": func() domain.Widget {
			w := sensorNumberWidget()
			w.Title = "   "
			return w
		}(),
		"span larger than the display minimum": func() domain.Widget {
			w := sensorNumberWidget()
			w.ColSpan = 4
			w.RowSpan = 3
			return w
		}(),
		"negative bounds on an arc": func() domain.Widget {
			w := sensorArcWidget()
			w.Config.Min = floatPtr(-20)
			w.Config.Max = floatPtr(0)
			return w
		}(),
		"actuator on media_player": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.EntityIDs = []string{"media_player.enceinte_salon"}
			return w
		}(),
		"bounds are ignored by toggle": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.Min = nil
			w.Config.Max = nil
			return w
		}(),
	}

	for name, w := range tests {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, w.Validate())
		})
	}
}

func TestWidgetValidateRejects(t *testing.T) {
	tests := map[string]domain.Widget{
		"empty id": func() domain.Widget {
			w := sensorNumberWidget()
			w.ID = "  "
			return w
		}(),
		"empty dashboard id": func() domain.Widget {
			w := sensorNumberWidget()
			w.DashboardID = ""
			return w
		}(),
		"unsupported widget type": func() domain.Widget {
			w := sensorNumberWidget()
			w.Type = "sensor_gauge"
			return w
		}(),
		"illegal pair: sensor with toggle": func() domain.Widget {
			w := sensorNumberWidget()
			w.Config.Display = domain.DisplayToggle
			return w
		}(),
		"illegal pair: actuator with arc": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.Display = domain.DisplayArc
			w.ColSpan = 2
			w.RowSpan = 2
			w.Config.Min = floatPtr(0)
			w.Config.Max = floatPtr(1)
			return w
		}(),
		"illegal pair: automation list with number": func() domain.Widget {
			w := automationListWidget()
			w.Config.Display = domain.DisplayNumber
			return w
		}(),
		"missing display": func() domain.Widget {
			w := sensorNumberWidget()
			w.Config.Display = ""
			return w
		}(),
		"below minimum: arc narrower than 2 cols": func() domain.Widget {
			w := sensorArcWidget()
			w.ColSpan = 1
			return w
		}(),
		"below minimum: arc shorter than 2 rows": func() domain.Widget {
			w := sensorArcWidget()
			w.RowSpan = 1
			return w
		}(),
		"below minimum: bar narrower than 2 cols": func() domain.Widget {
			w := sensorBarWidget()
			w.ColSpan = 1
			return w
		}(),
		"below minimum: list smaller than 2x2": func() domain.Widget {
			w := automationListWidget()
			w.RowSpan = 1
			return w
		}(),
		"zero col span": func() domain.Widget {
			w := sensorNumberWidget()
			w.ColSpan = 0
			return w
		}(),
		"negative row span": func() domain.Widget {
			w := sensorNumberWidget()
			w.RowSpan = -1
			return w
		}(),
		"negative col": func() domain.Widget {
			w := sensorNumberWidget()
			w.Col = -1
			return w
		}(),
		"negative row": func() domain.Widget {
			w := sensorNumberWidget()
			w.Row = -2
			return w
		}(),
		"sensor with no entity": func() domain.Widget {
			w := sensorNumberWidget()
			w.Config.EntityIDs = nil
			return w
		}(),
		"sensor with two entities": func() domain.Widget {
			w := sensorNumberWidget()
			w.Config.EntityIDs = []string{"sensor.a", "sensor.b"}
			return w
		}(),
		"actuator with two entities": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.EntityIDs = []string{"light.a", "light.b"}
			return w
		}(),
		"automation list with no entity": func() domain.Widget {
			w := automationListWidget()
			w.Config.EntityIDs = []string{}
			return w
		}(),
		"blank entity id": func() domain.Widget {
			w := sensorNumberWidget()
			w.Config.EntityIDs = []string{"   "}
			return w
		}(),
		"actuator on a sensor domain": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.EntityIDs = []string{"sensor.temperature_salon"}
			return w
		}(),
		"actuator on a climate domain": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.EntityIDs = []string{"climate.thermostat_salon"}
			return w
		}(),
		"actuator on a malformed entity id": func() domain.Widget {
			w := actuatorToggleWidget()
			w.Config.EntityIDs = []string{"light_salon"}
			return w
		}(),
		"arc without bounds": func() domain.Widget {
			w := sensorArcWidget()
			w.Config.Min = nil
			w.Config.Max = nil
			return w
		}(),
		"arc without max": func() domain.Widget {
			w := sensorArcWidget()
			w.Config.Max = nil
			return w
		}(),
		"arc with inverted bounds": func() domain.Widget {
			w := sensorArcWidget()
			w.Config.Min = floatPtr(40)
			w.Config.Max = floatPtr(0)
			return w
		}(),
		"arc with equal bounds": func() domain.Widget {
			w := sensorArcWidget()
			w.Config.Min = floatPtr(20)
			w.Config.Max = floatPtr(20)
			return w
		}(),
		"bar without min": func() domain.Widget {
			w := sensorBarWidget()
			w.Config.Min = nil
			return w
		}(),
		"bar with inverted bounds": func() domain.Widget {
			w := sensorBarWidget()
			w.Config.Min = floatPtr(100)
			w.Config.Max = floatPtr(0)
			return w
		}(),
	}

	for name, w := range tests {
		t.Run(name, func(t *testing.T) {
			err := w.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidWidget)
			assert.True(t, domain.IsDomainError(err))
		})
	}
}

func TestWidgetValidateInGridBounds(t *testing.T) {
	t.Run("widget flush against the far corner fits", func(t *testing.T) {
		w := sensorArcWidget()
		w.Col = 10
		w.Row = 6
		require.NoError(t, w.ValidateIn(domain.DefaultGridCols, domain.DefaultGridRows))
	})

	t.Run("widget overflowing the width is rejected", func(t *testing.T) {
		w := sensorArcWidget()
		w.Col = 11
		w.Row = 0
		err := w.ValidateIn(domain.DefaultGridCols, domain.DefaultGridRows)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("widget overflowing the height is rejected", func(t *testing.T) {
		w := sensorArcWidget()
		w.Col = 0
		w.Row = 7
		err := w.ValidateIn(domain.DefaultGridCols, domain.DefaultGridRows)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("ValidateIn still applies the content rules", func(t *testing.T) {
		w := sensorArcWidget()
		w.Config.Min = nil
		err := w.ValidateIn(domain.DefaultGridCols, domain.DefaultGridRows)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})
}

func TestWidgetOverlaps(t *testing.T) {
	rect := func(col, row, colSpan, rowSpan int) domain.Widget {
		return domain.Widget{Col: col, Row: row, ColSpan: colSpan, RowSpan: rowSpan}
	}

	tests := []struct {
		name string
		a    domain.Widget
		b    domain.Widget
		want bool
	}{
		{"identical rectangles overlap", rect(0, 0, 2, 2), rect(0, 0, 2, 2), true},
		{"partial corner overlap", rect(0, 0, 2, 2), rect(1, 1, 2, 2), true},
		{"contained rectangle overlaps", rect(0, 0, 4, 4), rect(1, 1, 1, 1), true},
		{"side by side does not overlap", rect(0, 0, 2, 2), rect(2, 0, 2, 2), false},
		{"stacked does not overlap", rect(0, 0, 2, 2), rect(0, 2, 2, 2), false},
		{"diagonal neighbours do not overlap", rect(0, 0, 2, 2), rect(2, 2, 2, 2), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.a.Overlaps(tc.b))
			assert.Equal(t, tc.want, tc.b.Overlaps(tc.a))
		})
	}
}

func TestOverviewDashboardValidate(t *testing.T) {
	now := time.Now().UTC()
	valid := func() domain.OverviewDashboard {
		first := sensorArcWidget()
		first.Col, first.Row = 0, 0
		second := actuatorToggleWidget()
		second.Col, second.Row = 2, 0
		third := automationListWidget()
		third.Col, third.Row = 0, 2
		return domain.OverviewDashboard{
			ID:        "ov-1",
			Name:      "Living Room & Den",
			Order:     0,
			Cols:      domain.DefaultGridCols,
			Rows:      domain.DefaultGridRows,
			CreatedAt: now,
			UpdatedAt: now,
			Widgets:   []domain.Widget{first, second, third},
		}
	}

	t.Run("valid dashboard passes", func(t *testing.T) {
		require.NoError(t, valid().Validate())
	})

	t.Run("dashboard without widgets passes", func(t *testing.T) {
		o := valid()
		o.Widgets = nil
		require.NoError(t, o.Validate())
	})

	overviewRejections := map[string]domain.OverviewDashboard{
		"empty id": func() domain.OverviewDashboard {
			o := valid()
			o.ID = "  "
			return o
		}(),
		"empty name": func() domain.OverviewDashboard {
			o := valid()
			o.Name = ""
			return o
		}(),
		"zero cols": func() domain.OverviewDashboard {
			o := valid()
			o.Cols = 0
			return o
		}(),
		"negative rows": func() domain.OverviewDashboard {
			o := valid()
			o.Rows = -1
			return o
		}(),
		"two widgets sharing a cell": func() domain.OverviewDashboard {
			o := valid()
			o.Widgets[1].Col = 1
			o.Widgets[1].Row = 1
			return o
		}(),
	}

	for name, o := range overviewRejections {
		t.Run(name+" returns ErrInvalidOverview", func(t *testing.T) {
			err := o.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidOverview)
			assert.True(t, domain.IsDomainError(err))
		})
	}

	widgetRejections := map[string]domain.OverviewDashboard{
		"a widget leaving the grid": func() domain.OverviewDashboard {
			o := valid()
			o.Widgets[2].Row = 7
			return o
		}(),
		"a widget larger than a narrow grid": func() domain.OverviewDashboard {
			o := valid()
			o.Cols = 2
			o.Rows = 2
			o.Widgets = o.Widgets[:1]
			o.Widgets[0].ColSpan = 3
			return o
		}(),
		"an invalid child widget": func() domain.OverviewDashboard {
			o := valid()
			o.Widgets[0].Config.Display = domain.DisplayList
			return o
		}(),
	}

	for name, o := range widgetRejections {
		t.Run(name+" returns ErrInvalidWidget", func(t *testing.T) {
			err := o.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidWidget)
		})
	}
}

func TestAutomationValidation(t *testing.T) {
	t.Run("valid automation passes validation", func(t *testing.T) {
		now := time.Now().UTC()
		auto := domain.Automation{
			ID:            "automation.eteindre_tout",
			Name:          "Turn Off All",
			State:         "on",
			Current:       0,
			LastTriggered: &now,
		}
		require.NoError(t, auto.Validate())
	})

	t.Run("empty ID returns ErrAutomationNotFound", func(t *testing.T) {
		auto := domain.Automation{
			ID:   "",
			Name: "Invalide",
		}
		err := auto.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAutomationNotFound)
	})

	t.Run("malformed ID returns ErrAutomationNotFound", func(t *testing.T) {
		auto := domain.Automation{
			ID:   "light.wrong_domain",
			Name: "Mauvais domaine",
		}
		err := auto.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAutomationNotFound)
	})
}

func TestOverviewDashboardValidateBackground(t *testing.T) {
	valid := func() domain.OverviewDashboard {
		return domain.OverviewDashboard{
			ID:   "ov-1",
			Name: "Home",
			Cols: domain.DefaultGridCols,
			Rows: domain.DefaultGridRows,
		}
	}

	t.Run("accepts in-range background settings", func(t *testing.T) {
		o := valid()
		o.BackgroundImage = "/backgrounds/desert-night.jpg"
		o.BackgroundOpacity = 0
		o.BackgroundBlur = 32
		o.BackgroundDim = 100
		require.NoError(t, o.Validate())
	})

	rejections := map[string]domain.OverviewDashboard{}
	o := valid()
	o.BackgroundOpacity = 101
	rejections["opacity above range"] = o
	o = valid()
	o.BackgroundOpacity = -1
	rejections["opacity below range"] = o
	o = valid()
	o.BackgroundBlur = 33
	rejections["blur above range"] = o
	o = valid()
	o.BackgroundDim = -1
	rejections["dim below range"] = o

	for name, candidate := range rejections {
		t.Run(name+" returns ErrInvalidOverview", func(t *testing.T) {
			err := candidate.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidOverview)
			assert.True(t, domain.IsDomainError(err))
		})
	}
}
