package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestDashboardConverters(t *testing.T) {
	now := time.Now().UTC()

	t.Run("ToDomainDashboard converts CreateDashboardRequest and defaults grid size", func(t *testing.T) {
		req := pkgdashboard.CreateDashboardRequest{
			Name:  "Living Room Dashboard",
			Order: 2,
		}
		res := converters.ToDomainDashboard(req)
		assert.Equal(t, "Living Room Dashboard", res.Name)
		assert.Equal(t, 2, res.Order)
		assert.Equal(t, domain.DefaultGridCols, res.Cols)
		assert.Equal(t, domain.DefaultGridRows, res.Rows)
	})

	t.Run("ToDomainDashboard preserves an explicit grid size", func(t *testing.T) {
		req := pkgdashboard.CreateDashboardRequest{
			Name: "Living Room Dashboard",
			Cols: 6,
			Rows: 4,
		}
		res := converters.ToDomainDashboard(req)
		assert.Equal(t, 6, res.Cols)
		assert.Equal(t, 4, res.Rows)
	})

	t.Run("ToDomainUpdateDashboard converts UpdateDashboardRequest", func(t *testing.T) {
		req := pkgdashboard.UpdateDashboardRequest{
			Name:  "Updated Dashboard",
			Order: 3,
			Cols:  10,
			Rows:  6,
		}
		res := converters.ToDomainUpdateDashboard("ov-1", req)
		assert.Equal(t, "ov-1", res.ID)
		assert.Equal(t, "Updated Dashboard", res.Name)
		assert.Equal(t, 3, res.Order)
		assert.Equal(t, 10, res.Cols)
		assert.Equal(t, 6, res.Rows)
	})

	t.Run("ToDomainWidget converts CreateWidgetRequest", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "automation_list",
			Title:   "Quick Scenes",
			Order:   1,
			Col:     2,
			Row:     3,
			ColSpan: 2,
			RowSpan: 2,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.cinema", "automation.depart"},
				Display:   "list",
			},
		}
		res := converters.ToDomainWidget("ov-1", req)
		assert.Equal(t, "ov-1", res.DashboardID)
		assert.Equal(t, "automation_list", res.Type)
		assert.Equal(t, "Quick Scenes", res.Title)
		assert.Equal(t, 2, res.Col)
		assert.Equal(t, 3, res.Row)
		assert.Equal(t, 2, res.ColSpan)
		assert.Equal(t, 2, res.RowSpan)
		assert.Equal(t, []string{"automation.cinema", "automation.depart"}, res.Config.EntityIDs)
		assert.Equal(t, "list", res.Config.Display)
	})

	t.Run("ToDomainUpdateWidget converts UpdateWidgetRequest without touching position", func(t *testing.T) {
		min := 0.0
		max := 100.0
		req := pkgdashboard.UpdateWidgetRequest{
			Title: "Nouveau Titre",
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"sensor.temp"},
				Display:   "arc",
				Min:       &min,
				Max:       &max,
				Unit:      "°C",
			},
		}
		res := converters.ToDomainUpdateWidget("ov-1", "w-1", req)
		assert.Equal(t, "w-1", res.ID)
		assert.Equal(t, "ov-1", res.DashboardID)
		assert.Equal(t, "Nouveau Titre", res.Title)
		assert.Equal(t, "arc", res.Config.Display)
		assert.Equal(t, &min, res.Config.Min)
		assert.Equal(t, &max, res.Config.Max)
		assert.Equal(t, "°C", res.Config.Unit)
		assert.Zero(t, res.Col)
		assert.Zero(t, res.ColSpan)
	})

	t.Run("ToDomainWidgetPositions converts UpdateLayoutRequest", func(t *testing.T) {
		req := pkgdashboard.UpdateLayoutRequest{
			Positions: []pkgdashboard.WidgetPositionDTO{
				{ID: "w-1", Col: 0, Row: 0, ColSpan: 1, RowSpan: 1},
				{ID: "w-2", Col: 1, Row: 0, ColSpan: 2, RowSpan: 2},
			},
		}
		res := converters.ToDomainWidgetPositions(req)
		assert.Equal(t, []domain.WidgetPosition{
			{ID: "w-1", Col: 0, Row: 0, ColSpan: 1, RowSpan: 1},
			{ID: "w-2", Col: 1, Row: 0, ColSpan: 2, RowSpan: 2},
		}, res)
	})

	t.Run("ToPublicDashboard and ToPublicWidgets converts domain entities", func(t *testing.T) {
		ov := domain.Dashboard{
			ID:        "ov-1",
			Name:      "Dashboard 1",
			Order:     0,
			Cols:      12,
			Rows:      8,
			CreatedAt: now,
			UpdatedAt: now,
			Widgets: []domain.Widget{
				{
					ID:          "w-1",
					DashboardID: "ov-1",
					Type:        "automation_list",
					Title:       "Favoris",
					Order:       0,
					Col:         0,
					Row:         0,
					ColSpan:     2,
					RowSpan:     2,
					Config: domain.WidgetConfig{
						EntityIDs: []string{"automation.eteindre_tout"},
						Display:   "list",
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}

		pub := converters.ToPublicDashboard(ov)
		assert.Equal(t, "ov-1", pub.ID)
		assert.Equal(t, "Dashboard 1", pub.Name)
		assert.Equal(t, 12, pub.Cols)
		assert.Equal(t, 8, pub.Rows)
		assert.Len(t, pub.Widgets, 1)
		assert.Equal(t, "w-1", pub.Widgets[0].ID)
		assert.Equal(t, "Favoris", pub.Widgets[0].Title)
		assert.Equal(t, 2, pub.Widgets[0].ColSpan)
		assert.Equal(t, 2, pub.Widgets[0].RowSpan)
		assert.Equal(t, []string{"automation.eteindre_tout"}, pub.Widgets[0].Config.EntityIDs)
		assert.Equal(t, "list", pub.Widgets[0].Config.Display)

		pubList := converters.ToPublicDashboards([]domain.Dashboard{ov})
		assert.Len(t, pubList, 1)
	})

	t.Run("ToPublicAutomation converts domain Automation", func(t *testing.T) {
		auto := domain.Automation{
			ID:            "automation.cinema",
			Name:          "Cinema Mode",
			State:         "on",
			Current:       1,
			LastTriggered: &now,
		}
		pub := converters.ToPublicAutomation(auto)
		assert.Equal(t, "automation.cinema", pub.ID)
		assert.Equal(t, "Cinema Mode", pub.Name)
		assert.Equal(t, "on", pub.State)
		assert.Equal(t, 1, pub.Current)
		assert.Equal(t, &now, pub.LastTriggered)

		pubList := converters.ToPublicAutomations([]domain.Automation{auto})
		assert.Len(t, pubList, 1)
	})
}

func TestDashboardBackgroundConversion(t *testing.T) {
	t.Run("ToPublicDashboard carries background fields", func(t *testing.T) {
		ov := domain.Dashboard{
			ID:                "ov-1",
			Name:              "Home",
			Cols:              domain.DefaultGridCols,
			Rows:              domain.DefaultGridRows,
			BackgroundImage:   "/backgrounds/desert-night.jpg",
			BackgroundOpacity: 80,
			BackgroundBlur:    18,
			BackgroundDim:     40,
		}
		pub := converters.ToPublicDashboard(ov)
		assert.Equal(t, "/backgrounds/desert-night.jpg", pub.BackgroundImage)
		assert.Equal(t, 80, pub.BackgroundOpacity)
		assert.Equal(t, 18, pub.BackgroundBlur)
		assert.Equal(t, 40, pub.BackgroundDim)
	})

	t.Run("ToDomainDashboard applies background defaults when omitted", func(t *testing.T) {
		dom := converters.ToDomainDashboard(pkgdashboard.CreateDashboardRequest{Name: "Home"})
		assert.Equal(t, 95, dom.BackgroundOpacity)
		assert.Equal(t, 14, dom.BackgroundBlur)
		assert.Equal(t, 50, dom.BackgroundDim)
		assert.Equal(t, "/backgrounds/desert-night.jpg", dom.BackgroundImage)
	})

	t.Run("ToDomainUpdateDashboard passes background through", func(t *testing.T) {
		dom := converters.ToDomainUpdateDashboard("ov-1", pkgdashboard.UpdateDashboardRequest{
			Name:              "Home",
			BackgroundImage:   "/backgrounds/meadow.jpg",
			BackgroundOpacity: 60,
			BackgroundBlur:    0,
			BackgroundDim:     30,
		})
		assert.Equal(t, "/backgrounds/meadow.jpg", dom.BackgroundImage)
		assert.Equal(t, 60, dom.BackgroundOpacity)
		assert.Equal(t, 0, dom.BackgroundBlur)
		assert.Equal(t, 30, dom.BackgroundDim)
	})
}
