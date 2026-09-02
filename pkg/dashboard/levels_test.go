package dashboard_test

import (
	"testing"

	"github.com/JLugagne/ha-dash/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelRequestsValidation(t *testing.T) {
	validate := validator.New()

	t.Run("CreateLevelRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.CreateLevelRequest{
			Name:      "Rez-de-chaussée",
			IsOutdoor: false,
		}
		require.NoError(t, validate.Struct(validReq))

		invalidReq := pkgdashboard.CreateLevelRequest{
			Name: "",
		}
		err := validate.Struct(invalidReq)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateLevelRequest))
	})

	t.Run("UpdateLevelRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.UpdateLevelRequest{
			Name:      "Étage 1",
			IsOutdoor: false,
		}
		require.NoError(t, validate.Struct(validReq))

		invalidReq := pkgdashboard.UpdateLevelRequest{
			Name: "",
		}
		err := validate.Struct(invalidReq)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidUpdateLevelRequest))
	})

	t.Run("ReorderLevelsRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.ReorderLevelsRequest{
			LevelIDs: []string{"id-1", "id-2"},
		}
		require.NoError(t, validate.Struct(validReq))

		invalidReqEmpty := pkgdashboard.ReorderLevelsRequest{
			LevelIDs: []string{},
		}
		err := validate.Struct(invalidReqEmpty)
		require.Error(t, err)

		invalidReqBlankElement := pkgdashboard.ReorderLevelsRequest{
			LevelIDs: []string{""},
		}
		err = validate.Struct(invalidReqBlankElement)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidReorderLevelsRequest))
	})

	t.Run("SavePlanRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "w1",
					X1:        0,
					Y1:        0,
					X2:        100,
					Y2:        0,
					Thickness: 10,
				},
			},
			Zones: []pkgdashboard.ZoneDTO{
				{
					ID:    "z1",
					Name:  "Salon",
					Color: "#3b82f6",
					Points: []pkgdashboard.Point2DDTO{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
					},
				},
			},
		}
		require.NoError(t, validate.Struct(validReq))

		invalidReq := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "", // missing required ID
					Thickness: 0,  // thickness must be gt=0
				},
			},
		}
		err := validate.Struct(invalidReq)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidSavePlanRequest))
	})
}
