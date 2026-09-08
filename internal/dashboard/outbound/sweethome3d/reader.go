package sweethome3d

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

const unitsPerCm = 0.4

var idSeq int

func nextID(prefix string) string {
	idSeq++
	return fmt.Sprintf("%s-sh3d-%d", prefix, idSeq)
}

func cmToUnits(cm float64) float64 {
	return math.Round(cm * unitsPerCm)
}

type sh3dHome struct {
	XMLName  xml.Name      `xml:"home"`
	Walls    []sh3dWall    `xml:"wall"`
	Rooms    []sh3dRoom    `xml:"room"`
	Openings []sh3dOpening `xml:"doorOrWindow"`
	Levels   []sh3dLevel   `xml:"level"`
}

type sh3dWall struct {
	ID        string  `xml:"id,attr"`
	XStart    float64 `xml:"xStart,attr"`
	YStart    float64 `xml:"yStart,attr"`
	XEnd      float64 `xml:"xEnd,attr"`
	YEnd      float64 `xml:"yEnd,attr"`
	Thickness float64 `xml:"thickness,attr"`
	Level     string  `xml:"level,attr"`
}

type sh3dPoint struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}

type sh3dRoom struct {
	Name  string      `xml:"name,attr"`
	Point []sh3dPoint `xml:"point"`
	Level string      `xml:"level,attr"`
}

type sh3dOpening struct {
	Wall  string  `xml:"wall,attr"`
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Width float64 `xml:"width,attr"`
	Name  string  `xml:"name,attr"`
	Level string  `xml:"level,attr"`
}

func openingType(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "window") ||
		strings.Contains(lower, "fenêtre") ||
		strings.Contains(lower, "fenetre") ||
		strings.Contains(lower, "coulissante") ||
		strings.Contains(lower, "américaine") {
		return "window"
	}
	return "door"
}

func computeOpeningOffset(wall sh3dWall, wx, wy float64) float64 {
	dx := wall.XEnd - wall.XStart
	dy := wall.YEnd - wall.YStart
	wallLen := math.Hypot(dx, dy)
	if wallLen < 1e-6 {
		return 0
	}
	return ((wx-wall.XStart)*dx + (wy-wall.YStart)*dy) / wallLen
}

func parseXML(data []byte, levelID string) (domain.Plan, error) {
	var home sh3dHome
	if err := xml.Unmarshal(data, &home); err != nil {
		return domain.Plan{}, fmt.Errorf("invalid Home.xml: %w", err)
	}
	if len(home.Walls) == 0 {
		return domain.Plan{}, errors.New("Home.xml contains no walls")
	}
	return buildPlan(home.Walls, home.Rooms, home.Openings, levelID), nil
}

func FromReader(r io.Reader, levelID string) (domain.Plan, error) {
	idSeq = 0
	homeXML, err := readHomeXML(r)
	if err != nil {
		return domain.Plan{}, err
	}
	return parseXML(homeXML, levelID)
}

type sh3dLevel struct {
	ID        string  `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	Elevation float64 `xml:"elevation,attr"`
}

// buildPlan assembles a domain Plan from pre-filtered Sweet Home 3D subsets.
// Degenerate geometry is tolerated: zero-length wall segments (often slivers
// that collapse after unit rounding) are skipped with their openings, and
// sub-unit thickness/width values are clamped to one unit so the plan validates.
func buildPlan(walls []sh3dWall, rooms []sh3dRoom, openings []sh3dOpening, levelID string) domain.Plan {
	segments := make([]domain.WallSegment, 0, len(walls))
	keptWallIDs := make(map[string]bool, len(walls))
	wallSourceIDs := make([]string, 0, len(walls))
	wallsByID := make(map[string]sh3dWall, len(walls))
	for _, w := range walls {
		wallsByID[w.ID] = w
	}
	for _, w := range walls {
		thickness := w.Thickness
		if thickness <= 0 {
			thickness = 10
		}
		thicknessUnits := cmToUnits(thickness)
		if thicknessUnits <= 0 {
			thicknessUnits = 1
		}
		seg := domain.WallSegment{
			ID:        nextID("wall"),
			X1:        cmToUnits(w.XStart),
			Y1:        cmToUnits(w.YStart),
			X2:        cmToUnits(w.XEnd),
			Y2:        cmToUnits(w.YEnd),
			Thickness: thicknessUnits,
			Openings:  nil,
		}
		if seg.X1 == seg.X2 && seg.Y1 == seg.Y2 {
			continue
		}
		keptWallIDs[w.ID] = true
		wallSourceIDs = append(wallSourceIDs, w.ID)
		segments = append(segments, seg)
	}
	for i := range segments {
		var segmentOpenings []domain.WallOpening
		for _, o := range openings {
			if o.Wall != wallSourceIDs[i] || !keptWallIDs[o.Wall] {
				continue
			}
			width := o.Width
			if width <= 0 {
				width = 80
			}
			widthUnits := cmToUnits(width)
			if widthUnits <= 0 {
				widthUnits = 1
			}
			srcWall := wallsByID[wallSourceIDs[i]]
			offsetCm := computeOpeningOffset(srcWall, o.X, o.Y)
			wallLenCm := math.Hypot(srcWall.XEnd-srcWall.XStart, srcWall.YEnd-srcWall.YStart)
			halfW := width / 2
			if offsetCm < halfW {
				offsetCm = halfW
			}
			if offsetCm > wallLenCm-halfW {
				offsetCm = wallLenCm - halfW
			}
			if offsetCm < 0 {
				offsetCm = halfW
			}
			opening := domain.WallOpening{
				ID:        nextID(openingType(o.Name)),
				Type:      openingType(o.Name),
				Offset:    cmToUnits(offsetCm),
				Width:     widthUnits,
				FlipSide:  false,
				FlipHinge: false,
				HideDoor:  false,
			}
			segmentOpenings = append(segmentOpenings, opening)
		}
		segments[i].Openings = segmentOpenings
	}
	zones := make([]domain.Zone, 0, len(rooms))
	zoneColorPalette := []string{"#3b82f6", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899", "#64748b", "#059669"}
	for i, room := range rooms {
		if len(room.Point) < 3 {
			continue
		}
		points := make([]domain.Point2D, len(room.Point))
		for j, p := range room.Point {
			points[j] = domain.Point2D{
				X: cmToUnits(p.X),
				Y: cmToUnits(p.Y),
			}
		}
		name := strings.TrimSpace(room.Name)
		if name == "" {
			name = fmt.Sprintf("Room %d", i+1)
		}
		color := zoneColorPalette[i%len(zoneColorPalette)]
		zones = append(zones, domain.Zone{
			ID:     nextID("zone"),
			Name:   name,
			Color:  color,
			Points: points,
		})
	}
	return domain.Plan{
		LevelID: levelID,
		Walls:   segments,
		Zones:   zones,
	}
}

// ImportedLevel is one Sweet Home 3D level with its pre-built plan.
// Plan.LevelID carries the source level id as a provisional value; callers
// replace it with the real Level id before saving.
type ImportedLevel struct {
	Name      string
	Elevation float64
	Plan      domain.Plan
}

// readHomeXML extracts and returns the Home.xml document from a .sh3d archive.
func readHomeXML(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading sh3d file: %w", err)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid sh3d archive: %w", err)
	}
	for _, f := range zipReader.File {
		if f.Name == "Home.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening Home.xml: %w", err)
			}
			homeXML, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("reading Home.xml: %w", err)
			}
			return homeXML, nil
		}
	}
	return nil, errors.New("Home.xml not found in .sh3d archive")
}

// FromReaderLevels parses every level of a .sh3d file, ordered by elevation
// ascending, each with its own plan. Levels without walls are still returned
// so callers can create them; items referencing an unknown level fall back to
// the first level. Files declaring no levels import as a single unnamed level.
func FromReaderLevels(r io.Reader) ([]ImportedLevel, error) {
	idSeq = 0
	homeXML, err := readHomeXML(r)
	if err != nil {
		return nil, err
	}
	var home sh3dHome
	if err := xml.Unmarshal(homeXML, &home); err != nil {
		return nil, fmt.Errorf("invalid Home.xml: %w", err)
	}
	if len(home.Walls) == 0 {
		return nil, errors.New("Home.xml contains no walls")
	}
	levels := append([]sh3dLevel(nil), home.Levels...)
	sort.Slice(levels, func(i, j int) bool {
		if levels[i].Elevation == levels[j].Elevation {
			return levels[i].ID < levels[j].ID
		}
		return levels[i].Elevation < levels[j].Elevation
	})
	if len(levels) == 0 {
		levels = []sh3dLevel{{ID: "", Name: "", Elevation: 0}}
	}
	known := make(map[string]bool, len(levels))
	for _, l := range levels {
		known[l.ID] = true
	}
	firstID := levels[0].ID
	assign := func(levelID string) string {
		if levelID == "" || !known[levelID] {
			return firstID
		}
		return levelID
	}
	groupedWalls := make(map[string][]sh3dWall)
	for _, w := range home.Walls {
		id := assign(w.Level)
		groupedWalls[id] = append(groupedWalls[id], w)
	}
	groupedRooms := make(map[string][]sh3dRoom)
	for _, room := range home.Rooms {
		id := assign(room.Level)
		groupedRooms[id] = append(groupedRooms[id], room)
	}
	groupedOpenings := make(map[string][]sh3dOpening)
	for _, o := range home.Openings {
		id := assign(o.Level)
		groupedOpenings[id] = append(groupedOpenings[id], o)
	}
	imported := make([]ImportedLevel, 0, len(levels))
	for i, l := range levels {
		name := strings.TrimSpace(l.Name)
		if name == "" {
			name = fmt.Sprintf("Level %d", i+1)
		}
		plan := buildPlan(groupedWalls[l.ID], groupedRooms[l.ID], groupedOpenings[l.ID], l.ID)
		imported = append(imported, ImportedLevel{
			Name:      name,
			Elevation: l.Elevation,
			Plan:      plan,
		})
	}
	return imported, nil
}
