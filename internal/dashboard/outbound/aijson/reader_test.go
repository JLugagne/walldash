package aijson_test

import (
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/outbound/aijson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromJSON(t *testing.T) {
	t.Run("parses walls with coordinates in meters", func(t *testing.T) {
		json := []byte(`{
			"walls": [
				{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 20},
				{"start": {"x": 5.2, "y": 0}, "end": {"x": 5.2, "y": 4.3}, "thickness_cm": 20}
			],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls, 2)
		assert.Equal(t, "test-level", plan.LevelID)

		assert.Equal(t, float64(0), plan.Walls[0].X1)
		assert.Equal(t, float64(0), plan.Walls[0].Y1)
		assert.Equal(t, float64(208), plan.Walls[0].X2)
		assert.Equal(t, float64(0), plan.Walls[0].Y2)
		assert.Equal(t, float64(8), plan.Walls[0].Thickness)

		assert.Equal(t, float64(208), plan.Walls[1].X1)
		assert.Equal(t, float64(208), plan.Walls[1].X2)
		assert.Equal(t, float64(172), plan.Walls[1].Y2)
	})

	t.Run("parses openings on walls", func(t *testing.T) {
		json := []byte(`{
			"walls": [{
				"start": {"x": 0, "y": 0},
				"end": {"x": 5.2, "y": 0},
				"thickness_cm": 20,
				"openings": [
					{"type": "door", "offset_m": 1.1, "width_cm": 90, "flip_side": true},
					{"type": "window", "offset_m": 3.5, "width_cm": 120}
				]
			}],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls, 1)
		require.Len(t, plan.Walls[0].Openings, 2)

		o1 := plan.Walls[0].Openings[0]
		assert.Equal(t, "door", o1.Type)
		assert.Equal(t, float64(44), o1.Offset)
		assert.Equal(t, float64(36), o1.Width)
		assert.True(t, o1.FlipSide)

		o2 := plan.Walls[0].Openings[1]
		assert.Equal(t, "window", o2.Type)
		assert.False(t, o2.FlipSide)
	})

	t.Run("parses zones", func(t *testing.T) {
		json := []byte(`{
			"walls": [
				{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 20}
			],
			"zones": [{
				"name": "Living Room",
				"color": "#3b82f6",
				"points": [
					{"x": 0, "y": 0},
					{"x": 5.2, "y": 0},
					{"x": 5.2, "y": 4.3},
					{"x": 0, "y": 4.3}
				]
			}]
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Zones, 1)
		assert.Equal(t, "Living Room", plan.Zones[0].Name)
		assert.Equal(t, "#3b82f6", plan.Zones[0].Color)
		require.Len(t, plan.Zones[0].Points, 4)
	})

	t.Run("returns error for no walls", func(t *testing.T) {
		json := []byte(`{"walls": [], "zones": []}`)
		_, err := aijson.FromJSON(json, "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one wall")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, err := aijson.FromJSON([]byte(`not json`), "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})

	t.Run("returns error for opening with invalid type", func(t *testing.T) {
		json := []byte(`{
			"walls": [{
				"start": {"x": 0, "y": 0},
				"end": {"x": 5.2, "y": 0},
				"thickness_cm": 20,
				"openings": [{"type": "hatch", "offset_m": 1.0, "width_cm": 80}]
			}],
			"zones": []
		}`)
		_, err := aijson.FromJSON(json, "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "door' or 'window'")
	})

	t.Run("uses default width for door when width_cm is zero", func(t *testing.T) {
		json := []byte(`{
			"walls": [{
				"start": {"x": 0, "y": 0},
				"end": {"x": 5.2, "y": 0},
				"thickness_cm": 20,
				"openings": [{"type": "door", "offset_m": 1.0, "width_cm": 0}]
			}],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls[0].Openings, 1)
		assert.Equal(t, float64(36), plan.Walls[0].Openings[0].Width)
	})

	t.Run("uses default width for window when width_cm is zero", func(t *testing.T) {
		json := []byte(`{
			"walls": [{
				"start": {"x": 0, "y": 0},
				"end": {"x": 5.2, "y": 0},
				"thickness_cm": 20,
				"openings": [{"type": "window", "offset_m": 1.0, "width_cm": 0}]
			}],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls[0].Openings, 1)
		assert.Equal(t, float64(48), plan.Walls[0].Openings[0].Width)
	})

	t.Run("uses default thickness when thickness_cm is zero", func(t *testing.T) {
		json := []byte(`{
			"walls": [{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 0}],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		assert.Equal(t, float64(8), plan.Walls[0].Thickness)
	})

	t.Run("generates default name and color for nameless zones", func(t *testing.T) {
		json := []byte(`{
			"walls": [{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 20}],
			"zones": [{
				"points": [{"x": 0, "y": 0}, {"x": 1, "y": 0}, {"x": 1, "y": 1}]
			}]
		}`)
		plan, err := aijson.FromJSON(json, "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Zones, 1)
		assert.Equal(t, "Zone 1", plan.Zones[0].Name)
		assert.NotEmpty(t, plan.Zones[0].Color)
	})

	t.Run("returns domain Plan with LevelID", func(t *testing.T) {
		json := []byte(`{
			"walls": [{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 20}],
			"zones": []
		}`)
		plan, err := aijson.FromJSON(json, "custom-level-id")
		require.NoError(t, err)
		assert.Equal(t, "custom-level-id", plan.LevelID)
	})

	t.Run("rejects zone with less than 3 points", func(t *testing.T) {
		json := []byte(`{
			"walls": [{"start": {"x": 0, "y": 0}, "end": {"x": 5.2, "y": 0}, "thickness_cm": 20}],
			"zones": [{"points": [{"x": 0, "y": 0}, {"x": 1, "y": 0}]}]
		}`)
		_, err := aijson.FromJSON(json, "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 3 points")
	})
}
