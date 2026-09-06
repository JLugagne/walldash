package domain

import (
	"errors"
	"strings"
	"time"
)

// Point2D represents a 2D coordinate on a level's plan.
type Point2D struct {
	X float64
	Y float64
}

// WallOpening represents a door or window attached to a wall segment.
// HideDoor keeps the hole cut in the wall but suppresses the door leaf and
// frame in every rendering; it is only meaningful when Type is "door".
type WallOpening struct {
	ID        string
	Type      string // "door" or "window"
	Offset    float64
	Width     float64
	FlipSide  bool
	FlipHinge bool
	HideDoor  bool
}

// Validate ensures the wall opening is well-formed.
func (o WallOpening) Validate() error {
	if strings.TrimSpace(o.ID) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("wall opening id cannot be empty"))
	}
	if o.Type != "door" && o.Type != "window" {
		return errors.Join(ErrInvalidPlan, errors.New("wall opening type must be door or window"))
	}
	if o.Width <= 0 {
		return errors.Join(ErrInvalidPlan, errors.New("wall opening width must be positive"))
	}
	if o.Offset < 0 {
		return errors.Join(ErrInvalidPlan, errors.New("wall opening offset cannot be negative"))
	}
	return nil
}

// WallSegment represents a wall line between two points in 2D space.
type WallSegment struct {
	ID        string
	X1        float64
	Y1        float64
	X2        float64
	Y2        float64
	Thickness float64
	Openings  []WallOpening
}

// Validate ensures the wall segment is well-formed.
func (w WallSegment) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment id cannot be empty"))
	}
	if w.Thickness <= 0 {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment thickness must be positive"))
	}
	if w.X1 == w.X2 && w.Y1 == w.Y2 {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment start and end coordinates cannot be identical"))
	}
	for _, o := range w.Openings {
		if err := o.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Zone represents a 2D closed polygon defining a room or outdoor area.
type Zone struct {
	ID             string
	Name           string
	Color          string
	Points         []Point2D
	TempSensor     string
	TempMin        *float64
	TempMax        *float64
	HumiditySensor string
	HumidityMin    *float64
	HumidityMax    *float64
}

// Validate ensures the zone is well-formed.
func (z Zone) Validate() error {
	if strings.TrimSpace(z.ID) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("zone id cannot be empty"))
	}
	if strings.TrimSpace(z.Name) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("zone name cannot be empty"))
	}
	if strings.TrimSpace(z.Color) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("zone color cannot be empty"))
	}
	if len(z.Points) < 3 {
		return errors.Join(ErrInvalidPlan, errors.New("zone must contain at least 3 points to form a polygon"))
	}
	if z.TempMin != nil && z.TempMax != nil && *z.TempMin > *z.TempMax {
		return errors.Join(ErrInvalidPlan, errors.New("temp_min cannot be greater than temp_max"))
	}
	if z.HumidityMin != nil && z.HumidityMax != nil && *z.HumidityMin > *z.HumidityMax {
		return errors.Join(ErrInvalidPlan, errors.New("humidity_min cannot be greater than humidity_max"))
	}
	if z.TempSensor != "" {
		parts := strings.Split(z.TempSensor, ".")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return errors.Join(ErrInvalidPlan, errors.New("invalid temp_sensor entity_id: "+z.TempSensor))
		}
	}
	if z.HumiditySensor != "" {
		parts := strings.Split(z.HumiditySensor, ".")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return errors.Join(ErrInvalidPlan, errors.New("invalid humidity_sensor entity_id: "+z.HumiditySensor))
		}
	}
	return nil
}

// Plan represents the 2D architectural blueprint for a level.
type Plan struct {
	LevelID string
	Walls   []WallSegment
	Zones   []Zone
}

// Validate ensures the plan references a valid level and that sub-elements are valid.
func (p Plan) Validate() error {
	if strings.TrimSpace(p.LevelID) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("plan level_id cannot be empty"))
	}
	for _, w := range p.Walls {
		if err := w.Validate(); err != nil {
			return err
		}
	}
	for _, z := range p.Zones {
		if err := z.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultLayers defines the default display layers available for any level.
var DefaultLayers = []string{"controls", "sensors"}

// Level represents an indoor story or outdoor space of the building.
type Level struct {
	ID        string
	Name      string
	Order     int
	IsOutdoor bool
	Layers    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate ensures the level possesses required attributes.
func (l Level) Validate() error {
	if strings.TrimSpace(l.ID) == "" {
		return errors.Join(ErrInvalidLevel, errors.New("level id cannot be empty"))
	}
	if strings.TrimSpace(l.Name) == "" {
		return errors.Join(ErrInvalidLevel, errors.New("level name cannot be empty"))
	}
	return nil
}
