package aijson

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

const unitsPerMeter = 40

var zoneColorPalette = []string{"#3b82f6", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899", "#64748b", "#059669"}

var idSeq int

func nextID(prefix string) string {
	idSeq++
	return fmt.Sprintf("%s-import-%d", prefix, idSeq)
}

func metersToUnits(m float64) float64 {
	return math.Round(m * unitsPerMeter)
}

func cmToUnits(cm float64) float64 {
	return math.Round(cm * unitsPerMeter / 100)
}

type rawPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type rawOpening struct {
	Type      string  `json:"type"`
	OffsetM   float64 `json:"offset_m"`
	WidthCm   float64 `json:"width_cm"`
	FlipSide  bool    `json:"flip_side"`
	FlipHinge bool    `json:"flip_hinge"`
	HideDoor  bool    `json:"hide_door"`
}

type rawWall struct {
	Start       rawPoint     `json:"start"`
	End         rawPoint     `json:"end"`
	ThicknessCm float64      `json:"thickness_cm"`
	Openings    []rawOpening `json:"openings"`
}

type rawZone struct {
	Name   string     `json:"name"`
	Color  string     `json:"color"`
	Points []rawPoint `json:"points"`
}

type rawPlan struct {
	Walls []rawWall `json:"walls"`
	Zones []rawZone `json:"zones"`
}

func convertOpening(raw rawOpening, wallLenUnits float64) (domain.WallOpening, error) {
	if raw.Type != "door" && raw.Type != "window" {
		return domain.WallOpening{}, fmt.Errorf("opening type must be 'door' or 'window', got %q", raw.Type)
	}

	defaultWidthCm := 90.0
	if raw.Type == "window" {
		defaultWidthCm = 120.0
	}
	widthCm := raw.WidthCm
	if widthCm <= 0 {
		widthCm = defaultWidthCm
	}

	offset := metersToUnits(raw.OffsetM)
	width := cmToUnits(widthCm)

	if wallLenUnits > 0 {
		halfW := width / 2
		if offset < halfW {
			offset = halfW
		}
		maxOffset := wallLenUnits - halfW
		if offset > maxOffset {
			offset = maxOffset
		}
		if maxOffset < halfW {
			offset = wallLenUnits / 2
		}
	}

	return domain.WallOpening{
		ID:        nextID(raw.Type),
		Type:      raw.Type,
		Offset:    math.Round(offset),
		Width:     math.Round(width),
		FlipSide:  raw.FlipSide,
		FlipHinge: raw.FlipHinge,
		HideDoor:  raw.Type == "door" && raw.HideDoor,
	}, nil
}

func convertWall(raw rawWall) (domain.WallSegment, error) {
	x1 := metersToUnits(raw.Start.X)
	y1 := metersToUnits(raw.Start.Y)
	x2 := metersToUnits(raw.End.X)
	y2 := metersToUnits(raw.End.Y)

	thicknessCm := raw.ThicknessCm
	if thicknessCm <= 0 {
		thicknessCm = 20
	}

	wallLen := math.Hypot(x2-x1, y2-y1)
	openings := make([]domain.WallOpening, 0, len(raw.Openings))
	for _, o := range raw.Openings {
		opening, err := convertOpening(o, wallLen)
		if err != nil {
			return domain.WallSegment{}, err
		}
		openings = append(openings, opening)
	}

	return domain.WallSegment{
		ID:        nextID("wall"),
		X1:        math.Round(x1),
		Y1:        math.Round(y1),
		X2:        math.Round(x2),
		Y2:        math.Round(y2),
		Thickness: cmToUnits(thicknessCm),
		Openings:  openings,
	}, nil
}

func convertZone(raw rawZone, index int) (domain.Zone, error) {
	if len(raw.Points) < 3 {
		return domain.Zone{}, fmt.Errorf("zone %d must have at least 3 points", index+1)
	}

	points := make([]domain.Point2D, len(raw.Points))
	for i, p := range raw.Points {
		points[i] = domain.Point2D{
			X: metersToUnits(p.X),
			Y: metersToUnits(p.Y),
		}
	}

	name := raw.Name
	if name == "" {
		name = fmt.Sprintf("Zone %d", index+1)
	}

	color := raw.Color
	if color == "" {
		color = zoneColorPalette[index%len(zoneColorPalette)]
	}

	return domain.Zone{
		ID:     nextID("zone"),
		Name:   name,
		Color:  color,
		Points: points,
	}, nil
}

func FromJSON(data []byte, levelID string) (domain.Plan, error) {
	idSeq = 0
	var raw rawPlan
	if err := json.Unmarshal(data, &raw); err != nil {
		return domain.Plan{}, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(raw.Walls) == 0 {
		return domain.Plan{}, errors.New("plan must contain at least one wall")
	}

	walls := make([]domain.WallSegment, 0, len(raw.Walls))
	for i, w := range raw.Walls {
		wall, err := convertWall(w)
		if err != nil {
			return domain.Plan{}, fmt.Errorf("walls[%d]: %w", i, err)
		}
		walls = append(walls, wall)
	}

	zones := make([]domain.Zone, 0, len(raw.Zones))
	for i, z := range raw.Zones {
		zone, err := convertZone(z, i)
		if err != nil {
			return domain.Plan{}, fmt.Errorf("zones[%d]: %w", i, err)
		}
		zones = append(zones, zone)
	}

	return domain.Plan{
		LevelID: levelID,
		Walls:   walls,
		Zones:   zones,
	}, nil
}
