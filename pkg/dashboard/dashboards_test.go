package dashboard_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/JLugagne/walldash/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDashboardRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid CreateDashboardRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateDashboardRequest{
			Name:  "Living Room Dashboard",
			Order: 1,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("valid CreateDashboardRequest with explicit grid size passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateDashboardRequest{
			Name: "Living Room Dashboard",
			Cols: 12,
			Rows: 8,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing Name fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateDashboardRequest{
			Name: "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateDashboardRequest))
	})
}

func TestUpdateDashboardRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid UpdateDashboardRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.UpdateDashboardRequest{
			Name:  "Nouveau Nom",
			Order: 2,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("empty Name fails validation", func(t *testing.T) {
		req := pkgdashboard.UpdateDashboardRequest{
			Name: "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidUpdateDashboardRequest))
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

func TestCreateWidgetRequestLayoutOverflowRejected(t *testing.T) {
	validate := validator.New()
	base := pkgdashboard.CreateWidgetRequest{
		Type:    "sensor",
		ColSpan: 1,
		RowSpan: 1,
		Config: pkgdashboard.WidgetConfigDTO{
			EntityIDs: []string{"sensor.temperature_salon"},
			Display:   "number",
		},
	}

	t.Run("audit payload with col=MaxInt64 is rejected", func(t *testing.T) {
		var req pkgdashboard.CreateWidgetRequest
		raw := `{"type":"sensor","title":"overflow","col":9223372036854775807,"row":0,"col_span":1,"row_span":1,"config":{"display":"number","entity_ids":["sensor.temperature_salon"]}}`
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		require.Error(t, validate.Struct(req))
	})

	t.Run("row MaxInt64 is rejected", func(t *testing.T) {
		req := base
		req.Row = math.MaxInt64
		require.Error(t, validate.Struct(req))
	})

	t.Run("col_span MaxInt64 is rejected", func(t *testing.T) {
		req := base
		req.ColSpan = math.MaxInt64
		require.Error(t, validate.Struct(req))
	})

	t.Run("row_span MaxInt64 is rejected", func(t *testing.T) {
		req := base
		req.RowSpan = math.MaxInt64
		require.Error(t, validate.Struct(req))
	})

	t.Run("coordinate above the 1024 bound is rejected", func(t *testing.T) {
		req := base
		req.Col = 1025
		require.Error(t, validate.Struct(req))
	})

	t.Run("span above the 1024 bound is rejected", func(t *testing.T) {
		req := base
		req.RowSpan = 1025
		require.Error(t, validate.Struct(req))
	})

	t.Run("layout at the 1024 bound passes", func(t *testing.T) {
		req := base
		req.Col = 1024
		req.Row = 1024
		req.ColSpan = 1024
		req.RowSpan = 1024
		require.NoError(t, validate.Struct(req))
	})
}

func TestUpdateLayoutRequestLayoutOverflowRejected(t *testing.T) {
	validate := validator.New()
	base := func() pkgdashboard.UpdateLayoutRequest {
		return pkgdashboard.UpdateLayoutRequest{
			Positions: []pkgdashboard.WidgetPositionDTO{
				{ID: "w-1", Col: 0, Row: 0, ColSpan: 1, RowSpan: 1},
			},
		}
	}

	t.Run("col MaxInt64 is rejected", func(t *testing.T) {
		req := base()
		req.Positions[0].Col = math.MaxInt64
		require.Error(t, validate.Struct(req))
	})

	t.Run("col_span MaxInt64 is rejected", func(t *testing.T) {
		req := base()
		req.Positions[0].ColSpan = math.MaxInt64
		require.Error(t, validate.Struct(req))
	})

	t.Run("position above the 1024 bound is rejected", func(t *testing.T) {
		req := base()
		req.Positions[0].Row = 1025
		require.Error(t, validate.Struct(req))
	})

	t.Run("layout at the 1024 bound passes", func(t *testing.T) {
		req := base()
		req.Positions[0].Col = 1024
		req.Positions[0].Row = 1024
		req.Positions[0].ColSpan = 1024
		req.Positions[0].RowSpan = 1024
		require.NoError(t, validate.Struct(req))
	})
}
