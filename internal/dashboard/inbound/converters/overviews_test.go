package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestOverviewConverters(t *testing.T) {
	now := time.Now().UTC()

	t.Run("ToDomainOverview converts CreateOverviewRequest", func(t *testing.T) {
		req := pkgdashboard.CreateOverviewRequest{
			Name:  "Tableau Salon",
			Order: 2,
		}
		res := converters.ToDomainOverview(req)
		assert.Equal(t, "Tableau Salon", res.Name)
		assert.Equal(t, 2, res.Order)
	})

	t.Run("ToDomainUpdateOverview converts UpdateOverviewRequest", func(t *testing.T) {
		req := pkgdashboard.UpdateOverviewRequest{
			Name:  "Tableau Mis à Jour",
			Order: 3,
		}
		res := converters.ToDomainUpdateOverview("ov-1", req)
		assert.Equal(t, "ov-1", res.ID)
		assert.Equal(t, "Tableau Mis à Jour", res.Name)
		assert.Equal(t, 3, res.Order)
	})

	t.Run("ToDomainWidget converts CreateWidgetRequest", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:  "automation_list",
			Title: "Scénarios Rapides",
			Order: 1,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.cinema", "automation.depart"},
			},
		}
		res := converters.ToDomainWidget("ov-1", req)
		assert.Equal(t, "ov-1", res.DashboardID)
		assert.Equal(t, "automation_list", res.Type)
		assert.Equal(t, "Scénarios Rapides", res.Title)
		assert.Equal(t, []string{"automation.cinema", "automation.depart"}, res.Config.EntityIDs)
	})

	t.Run("ToPublicOverview and ToPublicWidgets converts domain entities", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:        "ov-1",
			Name:      "Overview 1",
			Order:     0,
			CreatedAt: now,
			UpdatedAt: now,
			Widgets: []domain.Widget{
				{
					ID:          "w-1",
					DashboardID: "ov-1",
					Type:        "automation_list",
					Title:       "Favoris",
					Order:       0,
					Config: domain.WidgetConfig{
						EntityIDs: []string{"automation.eteindre_tout"},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}

		pub := converters.ToPublicOverview(ov)
		assert.Equal(t, "ov-1", pub.ID)
		assert.Equal(t, "Overview 1", pub.Name)
		assert.Len(t, pub.Widgets, 1)
		assert.Equal(t, "w-1", pub.Widgets[0].ID)
		assert.Equal(t, "Favoris", pub.Widgets[0].Title)
		assert.Equal(t, []string{"automation.eteindre_tout"}, pub.Widgets[0].Config.EntityIDs)

		pubList := converters.ToPublicOverviews([]domain.OverviewDashboard{ov})
		assert.Len(t, pubList, 1)
	})

	t.Run("ToPublicAutomation converts domain Automation", func(t *testing.T) {
		auto := domain.Automation{
			ID:            "automation.cinema",
			Name:          "Mode Cinéma",
			State:         "on",
			Current:       1,
			LastTriggered: &now,
		}
		pub := converters.ToPublicAutomation(auto)
		assert.Equal(t, "automation.cinema", pub.ID)
		assert.Equal(t, "Mode Cinéma", pub.Name)
		assert.Equal(t, "on", pub.State)
		assert.Equal(t, 1, pub.Current)
		assert.Equal(t, &now, pub.LastTriggered)

		pubList := converters.ToPublicAutomations([]domain.Automation{auto})
		assert.Len(t, pubList, 1)
	})
}
