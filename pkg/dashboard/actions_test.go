package dashboard_test

import (
	"testing"

	"github.com/JLugagne/walldash/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid ActionRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.ActionRequest{
			EntityID: "light.living_room",
			Action:   "toggle",
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("empty entity_id fails validation", func(t *testing.T) {
		req := pkgdashboard.ActionRequest{
			EntityID: "",
			Action:   "toggle",
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidActionRequest))
	})

	t.Run("empty action fails validation", func(t *testing.T) {
		req := pkgdashboard.ActionRequest{
			EntityID: "light.living_room",
			Action:   "",
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})
}
