package sweethome3d_test

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/outbound/sweethome3d"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSh3dZip(homeXML string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, _ := w.Create("Home.xml")
	f.Write([]byte(homeXML))
	w.Close()
	return buf.Bytes()
}

func TestFromReader(t *testing.T) {
	t.Run("parses walls from hoxml", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <wall id="w2" xStart="500.0" yStart="0.0" xEnd="500.0" yEnd="400.0" thickness="10.0"/>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls, 2)
		assert.Equal(t, "test-level", plan.LevelID)

		w1 := plan.Walls[0]
		w2 := plan.Walls[1]
		assert.Equal(t, float64(0), w1.X1)
		assert.Equal(t, float64(0), w1.Y1)
		assert.Equal(t, float64(200), w1.X2)
		assert.Equal(t, float64(0), w1.Y2)
		assert.Equal(t, float64(4), w1.Thickness)

		assert.Equal(t, float64(200), w2.X1)
		assert.Equal(t, float64(0), w2.Y1)
		assert.Equal(t, float64(200), w2.X2)
		assert.Equal(t, float64(160), w2.Y2)
	})

	t.Run("parses rooms as zones", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <room name="Living Room">
    <point x="10.0" y="10.0"/>
    <point x="490.0" y="10.0"/>
    <point x="490.0" y="390.0"/>
    <point x="10.0" y="390.0"/>
  </room>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Zones, 1)
		assert.Equal(t, "Living Room", plan.Zones[0].Name)
		assert.NotEmpty(t, plan.Zones[0].Color)
		require.Len(t, plan.Zones[0].Points, 4)
		assert.Equal(t, float64(4), plan.Zones[0].Points[0].X)
		assert.Equal(t, float64(4), plan.Zones[0].Points[0].Y)
	})

	t.Run("parses doors and windows as openings", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <doorOrWindow wall="w1" x="100.0" y="0.0" width="80.0" name="Porte"/>
  <doorOrWindow wall="w1" x="350.0" y="0.0" width="120.0" name="Window"/>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls, 1)
		require.Len(t, plan.Walls[0].Openings, 2)

		o1 := plan.Walls[0].Openings[0]
		assert.Equal(t, "door", o1.Type)
		assert.Equal(t, float64(40), o1.Offset)
		assert.Equal(t, float64(32), o1.Width)

		o2 := plan.Walls[0].Openings[1]
		assert.Equal(t, "window", o2.Type)
	})

	t.Run("returns error for missing Hoxml", func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		w.Close()
		_, err := sweethome3d.FromReader(bytes.NewReader(buf.Bytes()), "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Home.xml not found")
	})

	t.Run("returns error for no walls", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <room name="Room">
    <point x="0.0" y="0.0"/>
    <point x="100.0" y="0.0"/>
    <point x="100.0" y="100.0"/>
  </room>
</home>`
		zipData := buildSh3dZip(xml)
		_, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no walls")
	})

	t.Run("omits invalid rooms with less than 3 points", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="100.0" yEnd="0.0" thickness="10.0"/>
  <room name="Bad Room">
    <point x="0.0" y="0.0"/>
    <point x="100.0" y="0.0"/>
  </room>
  <room name="Good Room">
    <point x="0.0" y="50.0"/>
    <point x="100.0" y="50.0"/>
    <point x="100.0" y="150.0"/>
  </room>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Zones, 1)
		assert.Equal(t, "Good Room", plan.Zones[0].Name)
	})

	t.Run("parses wall with default thickness", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="100.0" yEnd="0.0" thickness="0.0"/>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls, 1)
		assert.Equal(t, float64(4), plan.Walls[0].Thickness)
	})

	t.Run("returns domain Plan with LevelID set", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="100.0" yEnd="0.0" thickness="10.0"/>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "my-level-42")
		require.NoError(t, err)
		assert.Equal(t, "my-level-42", plan.LevelID)
	})

	t.Run("handles empty ZIP gracefully", func(t *testing.T) {
		_, err := sweethome3d.FromReader(bytes.NewReader([]byte("not-a-zip")), "test-level")
		require.Error(t, err)
	})

	t.Run("detects coulissante as window", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <doorOrWindow wall="w1" x="200.0" y="0.0" width="120.0" name="Porte-fenêtre coulissante"/>
</home>`
		zipData := buildSh3dZip(xml)
		plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
		require.NoError(t, err)
		require.Len(t, plan.Walls[0].Openings, 1)
		assert.Equal(t, "window", plan.Walls[0].Openings[0].Type)
	})
}

func TestFromReaderDegenerateGeometry(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="7400">
  <wall id="w1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <wall id="w2" xStart="10.0" yStart="50.0" xEnd="11.0" yEnd="50.0" thickness="10.0"/>
  <wall id="w3" xStart="0.0" yStart="100.0" xEnd="500.0" yEnd="100.0" thickness="1.0"/>
  <doorOrWindow wall="w1" x="100.0" y="0.0" width="1.0" name="Porte"/>
  <doorOrWindow wall="w2" x="10.5" y="50.0" width="80.0" name="Porte"/>
</home>`
	zipData := buildSh3dZip(xml)
	plan, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
	require.NoError(t, err)
	require.Len(t, plan.Walls, 2)
	assert.Equal(t, float64(200), plan.Walls[0].X2)
	require.Len(t, plan.Walls[0].Openings, 1)
	assert.Greater(t, plan.Walls[0].Openings[0].Width, float64(0))
	assert.Greater(t, plan.Walls[1].Thickness, float64(0))
	require.NoError(t, plan.Validate())
}

func TestFromReaderLevels(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="7400">
  <level id="level0" name="Ground floor" elevation="0.0"/>
  <level id="level1" name="Upstairs" elevation="262.0"/>
  <level id="level2" name="Attic" elevation="524.0"/>
  <wall id="w1" level="level1" xStart="0.0" yStart="0.0" xEnd="500.0" yEnd="0.0" thickness="10.0"/>
  <wall id="w2" level="level0" xStart="0.0" yStart="0.0" xEnd="400.0" yEnd="0.0" thickness="10.0"/>
  <wall id="w3" xStart="0.0" yStart="50.0" xEnd="300.0" yEnd="50.0" thickness="10.0"/>
  <room level="level0" name="Kitchen">
    <point x="10.0" y="10.0"/>
    <point x="390.0" y="10.0"/>
    <point x="390.0" y="300.0"/>
  </room>
  <room level="level1" name="Bedroom">
    <point x="10.0" y="10.0"/>
    <point x="490.0" y="10.0"/>
    <point x="490.0" y="300.0"/>
  </room>
  <doorOrWindow level="level1" wall="w1" x="100.0" y="0.0" width="80.0" name="Door"/>
</home>`
	zipData := buildSh3dZip(xml)
	imported, err := sweethome3d.FromReaderLevels(bytes.NewReader(zipData))
	require.NoError(t, err)
	require.Len(t, imported, 3)
	assert.Equal(t, "Ground floor", imported[0].Name)
	assert.Equal(t, "Upstairs", imported[1].Name)
	assert.Equal(t, "Attic", imported[2].Name)
	require.Len(t, imported[0].Plan.Walls, 2)
	require.Len(t, imported[1].Plan.Walls, 1)
	assert.Empty(t, imported[2].Plan.Walls)
	require.Len(t, imported[0].Plan.Zones, 1)
	assert.Equal(t, "Kitchen", imported[0].Plan.Zones[0].Name)
	require.Len(t, imported[1].Plan.Zones, 1)
	require.Len(t, imported[1].Plan.Walls[0].Openings, 1)
	for _, lvl := range imported {
		require.NoError(t, lvl.Plan.Validate())
	}
}

func TestFromReaderRejectsOversizedHomeXML(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<home version="5300">` + strings.Repeat(" ", 40<<20) + `<wall id="w1" xStart="0.0" yStart="0.0" xEnd="100.0" yEnd="0.0" thickness="10.0"/></home>`
	zipData := buildSh3dZip(xml)
	_, err := sweethome3d.FromReader(bytes.NewReader(zipData), "test-level")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}
