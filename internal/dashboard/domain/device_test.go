package domain_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportedDomains(t *testing.T) {
	supported := []string{
		domain.DomainLight,
		domain.DomainSwitch,
		domain.DomainMediaPlayer,
		domain.DomainSensor,
		domain.DomainClimate,
	}

	for _, d := range supported {
		t.Run("supported domain: "+d, func(t *testing.T) {
			assert.True(t, domain.IsSupportedDomain(d))
		})
	}

	unsupported := []string{"", "camera", "vacuum", "automation", "fan"}
	for _, d := range unsupported {
		t.Run("unsupported domain: "+d, func(t *testing.T) {
			assert.False(t, domain.IsSupportedDomain(d))
		})
	}
}

func TestDeviceValidation(t *testing.T) {
	t.Run("valid device passes validation", func(t *testing.T) {
		dev := domain.Device{
			ID:     "light.living_room",
			Name:   "Plafonnier Salon",
			Domain: domain.DomainLight,
			State:  "on",
			Attributes: map[string]any{
				"brightness": 255,
			},
		}
		require.NoError(t, dev.Validate())
	})

	t.Run("empty device id fails validation", func(t *testing.T) {
		dev := domain.Device{
			ID:     "",
			Name:   "Plafonnier",
			Domain: domain.DomainLight,
		}
		err := dev.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDevice)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("unsupported domain fails validation", func(t *testing.T) {
		dev := domain.Device{
			ID:     "camera.front_door",
			Name:   "Camera Porte",
			Domain: "camera",
		}
		err := dev.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUnsupportedDomain)
		assert.True(t, domain.IsDomainError(err))
	})
}

func TestDevicePlacementValidation(t *testing.T) {
	now := time.Now().UTC()
	validPlacement := domain.DevicePlacement{
		ID:         "placement-1",
		LevelID:    "level-1",
		DeviceID:   "light.living_room",
		X:          120.5,
		Y:          250.0,
		Icon:       "lightbulb",
		CustomName: "Lumière principale",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	t.Run("valid placement passes validation", func(t *testing.T) {
		require.NoError(t, validPlacement.Validate())
	})

	t.Run("empty placement ID fails validation", func(t *testing.T) {
		p := validPlacement
		p.ID = "   "
		err := p.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlacement)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("empty level ID fails validation", func(t *testing.T) {
		p := validPlacement
		p.LevelID = ""
		err := p.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlacement)
	})

	t.Run("empty device ID fails validation", func(t *testing.T) {
		p := validPlacement
		p.DeviceID = ""
		err := p.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlacement)
	})

	t.Run("negative X coordinate fails validation", func(t *testing.T) {
		p := validPlacement
		p.X = -1.0
		err := p.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlacement)
	})

	t.Run("negative Y coordinate fails validation", func(t *testing.T) {
		p := validPlacement
		p.Y = -0.01
		err := p.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlacement)
	})
}
