package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestDeviceConverters(t *testing.T) {
	t.Run("ToPublicDevice converts domain.Device correctly", func(t *testing.T) {
		dev := domain.Device{
			ID:     "light.kitchen",
			Name:   "Cuisine",
			Domain: "light",
			State:  "on",
			Attributes: map[string]any{
				"brightness": 200,
			},
		}

		pub := converters.ToPublicDevice(dev)
		assert.Equal(t, dev.ID, pub.ID)
		assert.Equal(t, dev.Name, pub.Name)
		assert.Equal(t, dev.Domain, pub.Domain)
		assert.Equal(t, dev.State, pub.State)
		assert.Equal(t, dev.Attributes, pub.Attributes)
	})

	t.Run("ToPublicDevices converts slice and handles empty slice", func(t *testing.T) {
		devs := []domain.Device{
			{ID: "light.1", Name: "L1", Domain: "light"},
			{ID: "switch.1", Name: "S1", Domain: "switch"},
		}
		res := converters.ToPublicDevices(devs)
		assert.Len(t, res, 2)
		assert.Equal(t, "light.1", res[0].ID)

		emptyRes := converters.ToPublicDevices(nil)
		assert.NotNil(t, emptyRes)
		assert.Empty(t, emptyRes)
	})

	t.Run("ToDomainSavePlacement converts public request to domain.DevicePlacement", func(t *testing.T) {
		req := pkgdashboard.SavePlacementRequest{
			ID:         "p-1",
			DeviceID:   "light.kitchen",
			X:          100.5,
			Y:          200.5,
			Icon:       "lightbulb",
			CustomName: "Lumière Cuisine",
		}

		dom := converters.ToDomainSavePlacement("lvl-123", req)
		assert.Equal(t, "p-1", dom.ID)
		assert.Equal(t, "lvl-123", dom.LevelID)
		assert.Equal(t, "light.kitchen", dom.DeviceID)
		assert.Equal(t, 100.5, dom.X)
		assert.Equal(t, 200.5, dom.Y)
		assert.Equal(t, "lightbulb", dom.Icon)
		assert.Equal(t, "Lumière Cuisine", dom.CustomName)
	})

	t.Run("ToPublicPlacement converts domain.DevicePlacement correctly", func(t *testing.T) {
		now := time.Now().UTC()
		dom := domain.DevicePlacement{
			ID:         "p-1",
			LevelID:    "lvl-1",
			DeviceID:   "sensor.temperature",
			X:          50.0,
			Y:          60.0,
			Icon:       "thermometer",
			CustomName: "Température",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		pub := converters.ToPublicPlacement(dom)
		assert.Equal(t, dom.ID, pub.ID)
		assert.Equal(t, dom.LevelID, pub.LevelID)
		assert.Equal(t, dom.DeviceID, pub.DeviceID)
		assert.Equal(t, dom.X, pub.X)
		assert.Equal(t, dom.Y, pub.Y)
		assert.Equal(t, dom.Icon, pub.Icon)
		assert.Equal(t, dom.CustomName, pub.CustomName)
		assert.Equal(t, dom.CreatedAt, pub.CreatedAt)
		assert.Equal(t, dom.UpdatedAt, pub.UpdatedAt)
	})

	t.Run("ToPublicPlacements converts slice and handles empty slice", func(t *testing.T) {
		doms := []domain.DevicePlacement{
			{ID: "p-1", LevelID: "lvl-1", DeviceID: "light.1", X: 10, Y: 20},
		}
		res := converters.ToPublicPlacements(doms)
		assert.Len(t, res, 1)

		emptyRes := converters.ToPublicPlacements(nil)
		assert.NotNil(t, emptyRes)
		assert.Empty(t, emptyRes)
	})
}
