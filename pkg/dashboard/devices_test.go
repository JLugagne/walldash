package dashboard_test

import (
	"testing"

	"github.com/JLugagne/walldash/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSavePlacementRequestValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid SavePlacementRequest passes validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			ID:         "placement-1",
			DeviceID:   "light.living_room",
			X:          120.0,
			Y:          250.0,
			Icon:       "lightbulb",
			CustomName: "Living Room Ceiling Light",
			Layer:      "controls",
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("valid SavePlacementRequest with custom layer passes validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID: "light.living_room",
			X:        100.0,
			Y:        100.0,
			Layer:    "security",
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("valid SavePlacementRequest without ID passes validation (auto-id)", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID: "switch.coffee_maker",
			X:        0.0,
			Y:        0.0,
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("missing DeviceID fails validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID: "",
			X:        50.0,
			Y:        50.0,
		}
		err := validate.Struct(req)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidSavePlacementRequest))
	})

	t.Run("negative X coordinate fails validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID: "sensor.temp",
			X:        -1.0,
			Y:        10.0,
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})

	t.Run("negative Y coordinate fails validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID: "sensor.temp",
			X:        10.0,
			Y:        -0.5,
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})

	t.Run("supported RenderDomain override passes validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID:     "switch.lamp",
			X:            10.0,
			Y:            10.0,
			RenderDomain: "light",
		}
		require.NoError(t, validate.Struct(req))
	})

	t.Run("unsupported RenderDomain fails validation", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			DeviceID:     "switch.lamp",
			X:            10.0,
			Y:            10.0,
			RenderDomain: "camera",
		}
		err := validate.Struct(req)
		require.Error(t, err)
	})
}
