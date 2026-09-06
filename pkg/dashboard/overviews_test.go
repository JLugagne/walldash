package dashboard_test

import (
	"testing"

	"github.com/JLugagne/walldash/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOverviewRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid CreateOverviewRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateOverviewRequest{
			Name:  "Living Room Dashboard",
			Order: 1,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("valid CreateOverviewRequest with explicit grid size passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateOverviewRequest{
			Name: "Living Room Dashboard",
			Cols: 12,
			Rows: 8,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing Name fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateOverviewRequest{
			Name: "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateOverviewRequest))
	})
}

func TestUpdateOverviewRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid UpdateOverviewRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.UpdateOverviewRequest{
			Name:  "Nouveau Nom",
			Order: 2,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("empty Name fails validation", func(t *testing.T) {
		req := pkgdashboard.UpdateOverviewRequest{
			Name: "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidUpdateOverviewRequest))
	})
}

func TestCreateWidgetRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid CreateWidgetRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "automation_list",
			Title:   "My Scenes",
			Order:   0,
			Col:     0,
			Row:     0,
			ColSpan: 2,
			RowSpan: 2,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.eteindre_tout"},
				Display:   "list",
			},
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("empty Title passes validation (title is optional)", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "automation_list",
			Title:   "",
			ColSpan: 2,
			RowSpan: 2,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.eteindre_tout"},
				Display:   "list",
			},
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing Type fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "",
			Title:   "Titre",
			ColSpan: 1,
			RowSpan: 1,
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateWidgetRequest))
	})

	t.Run("missing ColSpan fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "sensor",
			Title:   "Temperature",
			ColSpan: 0,
			RowSpan: 1,
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})

	t.Run("missing Display fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:    "sensor",
			Title:   "Temperature",
			ColSpan: 1,
			RowSpan: 1,
			Config:  pkgdashboard.WidgetConfigDTO{EntityIDs: []string{"sensor.temp"}},
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})
}

func TestUpdateWidgetRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid UpdateWidgetRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.UpdateWidgetRequest{
			Title: "Nouveau Titre",
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"sensor.temp"},
				Display:   "number",
			},
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing Display fails validation", func(t *testing.T) {
		req := pkgdashboard.UpdateWidgetRequest{
			Title:  "Nouveau Titre",
			Config: pkgdashboard.WidgetConfigDTO{EntityIDs: []string{"sensor.temp"}},
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidUpdateWidgetRequest))
	})
}

func TestUpdateLayoutRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid UpdateLayoutRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.UpdateLayoutRequest{
			Positions: []pkgdashboard.WidgetPositionDTO{
				{ID: "w-1", Col: 0, Row: 0, ColSpan: 1, RowSpan: 1},
			},
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("empty positions fails validation", func(t *testing.T) {
		req := pkgdashboard.UpdateLayoutRequest{Positions: nil}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidUpdateLayoutRequest))
	})

	t.Run("position missing ID fails validation", func(t *testing.T) {
		req := pkgdashboard.UpdateLayoutRequest{
			Positions: []pkgdashboard.WidgetPositionDTO{
				{ID: "", Col: 0, Row: 0, ColSpan: 1, RowSpan: 1},
			},
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})
}
