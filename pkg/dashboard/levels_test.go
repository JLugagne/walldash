package dashboard_test

import (
	"testing"

	"github.com/JLugagne/walldash/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelRequestsValidation(t *testing.T) {
	validate := validator.New()

	t.Run("CreateLevelRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.CreateLevelRequest{
			Name:      "Ground Floor",
			IsOutdoor: false,
		}
		require.NoError(t, validate.Struct(validReq))

		validWithLayers := pkgdashboard.CreateLevelRequest{
			Name:   "Ground Floor",
			Layers: []pkgdashboard.LayerRequest{{Name: "controls", HideGauges: false}, {Name: "sensors", HideGauges: false}, {Name: "lighting", HideGauges: false}},
		}
		require.NoError(t, validate.Struct(validWithLayers))

		invalidReq := pkgdashboard.CreateLevelRequest{
			Name: "",
		}
		err := validate.Struct(invalidReq)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidCreateLevelRequest))
	})

	t.Run("UpdateLevelRequest validation", func(t *testing.T) {
		validReq := pkgdashboard.UpdateLevelRequest{
			Name:      "1st Floor",
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
					Name:  "Living Room",
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

		validWithOpenings := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "w1",
					X1:        0,
					Y1:        0,
					X2:        100,
					Y2:        0,
					Thickness: 10,
					Openings: []pkgdashboard.WallOpeningDTO{
						{
							ID:     "win-1",
							Type:   "window",
							Offset: 30,
							Width:  20,
						},
						{
							ID:        "door-1",
							Type:      "door",
							Offset:    70,
							Width:     25,
							FlipSide:  true,
							FlipHinge: true,
							HideDoor:  true,
						},
					},
				},
			},
			Zones: []pkgdashboard.ZoneDTO{},
		}
		require.NoError(t, validate.Struct(validWithOpenings))
		assert.True(t, validWithOpenings.Walls[0].Openings[1].FlipSide)
		assert.True(t, validWithOpenings.Walls[0].Openings[1].FlipHinge)
		assert.True(t, validWithOpenings.Walls[0].Openings[1].HideDoor)

		invalidOpeningType := pkgdashboard.SavePlanRequest{
			Walls: []pkgdashboard.WallSegmentDTO{
				{
					ID:        "w1",
					Thickness: 10,
					Openings: []pkgdashboard.WallOpeningDTO{
						{
							ID:     "bad-1",
							Type:   "invalid",
							Offset: 10,
							Width:  20,
						},
					},
				},
			},
		}
		err = validate.Struct(invalidOpeningType)
		require.Error(t, err)
	})
}
