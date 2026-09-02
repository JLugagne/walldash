package domain_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverviewDashboardValidation(t *testing.T) {
	t.Run("valid overview dashboard passes validation", func(t *testing.T) {
		now := time.Now().UTC()
		ov := domain.OverviewDashboard{
			ID:        "ov-1",
			Name:      "Salon & Séjour",
			Order:     0,
			CreatedAt: now,
			UpdatedAt: now,
			Widgets: []domain.Widget{
				{
					ID:          "widget-1",
					DashboardID: "ov-1",
					Type:        domain.WidgetTypeAutomationList,
					Title:       "Automatisations Rapides",
					Order:       0,
					Config: domain.WidgetConfig{
						EntityIDs: []string{"automation.eteindre_tout", "automation.depart"},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}

		err := ov.Validate()
		require.NoError(t, err)
	})

	t.Run("empty ID returns ErrInvalidOverview", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:   "",
			Name: "Vue Principale",
		}
		err := ov.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidOverview)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("empty Name returns ErrInvalidOverview", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:   "ov-1",
			Name: "   ",
		}
		err := ov.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidOverview)
	})

	t.Run("invalid child widget returns ErrInvalidWidget", func(t *testing.T) {
		ov := domain.OverviewDashboard{
			ID:   "ov-1",
			Name: "Overview",
			Widgets: []domain.Widget{
				{
					ID:          "",
					DashboardID: "ov-1",
					Type:        domain.WidgetTypeAutomationList,
					Title:       "Bad Widget",
				},
			},
		}
		err := ov.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})
}

func TestWidgetValidation(t *testing.T) {
	t.Run("valid widget passes validation", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-1",
			DashboardID: "ov-1",
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Mes Favoris",
			Order:       1,
			Config: domain.WidgetConfig{
				EntityIDs: []string{"automation.matin"},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		require.NoError(t, w.Validate())
	})

	t.Run("empty ID returns ErrInvalidWidget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "",
			DashboardID: "ov-1",
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Favoris",
		}
		err := w.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("empty DashboardID returns ErrInvalidWidget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-1",
			DashboardID: "",
			Type:        domain.WidgetTypeAutomationList,
			Title:       "Favoris",
		}
		err := w.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("empty Title returns ErrInvalidWidget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-1",
			DashboardID: "ov-1",
			Type:        domain.WidgetTypeAutomationList,
			Title:       "  ",
		}
		err := w.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})

	t.Run("unsupported type returns ErrInvalidWidget", func(t *testing.T) {
		w := domain.Widget{
			ID:          "widget-1",
			DashboardID: "ov-1",
			Type:        "unknown_type",
			Title:       "Custom",
		}
		err := w.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidWidget)
	})
}

func TestAutomationValidation(t *testing.T) {
	t.Run("valid automation passes validation", func(t *testing.T) {
		now := time.Now().UTC()
		auto := domain.Automation{
			ID:            "automation.eteindre_tout",
			Name:          "Éteindre Tout",
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
