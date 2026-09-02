package dashboard_test

import (
	"testing"

	"github.com/JLugagne/ha-dash/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOverviewRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid CreateOverviewRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.CreateOverviewRequest{
			Name:  "Tableau Salon",
			Order: 1,
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
			Type:  "automation_list",
			Title: "Mes Scénarios",
			Order: 0,
			Config: pkgdashboard.WidgetConfigDTO{
				EntityIDs: []string{"automation.eteindre_tout"},
			},
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing Type fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:  "",
			Title: "Titre",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateWidgetRequest))
	})

	t.Run("missing Title fails validation", func(t *testing.T) {
		req := pkgdashboard.CreateWidgetRequest{
			Type:  "automation_list",
			Title: "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})
}
