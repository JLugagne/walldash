package sweethome3d

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
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
}

type sh3dWall struct {
	ID        string  `xml:"id,attr"`
	XStart    float64 `xml:"xStart,attr"`
	YStart    float64 `xml:"yStart,attr"`
	XEnd      float64 `xml:"xEnd,attr"`
	YEnd      float64 `xml:"yEnd,attr"`
	Thickness float64 `xml:"thickness,attr"`
}

type sh3dPoint struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}

type sh3dRoom struct {
	Name  string      `xml:"name,attr"`
	Point []sh3dPoint `xml:"point"`
}

type sh3dOpening struct {
	Wall  string  `xml:"wall,attr"`
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Width float64 `xml:"width,attr"`
	Name  string  `xml:"name,attr"`
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

	walls := make([]domain.WallSegment, 0, len(home.Walls))
	keptWallIDs := make(map[string]bool, len(home.Walls))
	wallSourceIDs := make([]string, 0, len(home.Walls))
	wallsByID := make(map[string]sh3dWall, len(home.Walls))
	for _, w := range home.Walls {
		wallsByID[w.ID] = w
	}

	for _, w := range home.Walls {
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
		walls = append(walls, seg)
	}

	for i := range walls {
		var openings []domain.WallOpening
		for _, o := range home.Openings {
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
			openings = append(openings, opening)
		}
		walls[i].Openings = openings
	}

	zones := make([]domain.Zone, 0, len(home.Rooms))
	zoneColorPalette := []string{"#3b82f6", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899", "#64748b", "#059669"}
	for i, room := range home.Rooms {
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
		Walls:   walls,
		Zones:   zones,
	}, nil
}

func FromReader(r io.Reader, levelID string) (domain.Plan, error) {
	idSeq = 0
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.Plan{}, fmt.Errorf("reading sh3d file: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return domain.Plan{}, fmt.Errorf("invalid sh3d archive: %w", err)
	}

	var homeXML []byte
	for _, f := range zipReader.File {
		if f.Name == "Home.xml" {
			rc, err := f.Open()
			if err != nil {
				return domain.Plan{}, fmt.Errorf("opening Home.xml: %w", err)
			}
			homeXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return domain.Plan{}, fmt.Errorf("reading Home.xml: %w", err)
			}
			break
		}
	}

	if homeXML == nil {
		return domain.Plan{}, errors.New("Home.xml not found in .sh3d archive")
	}

	return parseXML(homeXML, levelID)
}
