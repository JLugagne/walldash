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

// WallSegment represents a wall line between two points in 2D space.
type WallSegment struct {
	ID        string
	X1        float64
	Y1        float64
	X2        float64
	Y2        float64
	Thickness float64
}

// Validate ensures the wall segment is well-formed.
func (w WallSegment) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment id cannot be empty"))
	}
	if w.Thickness <= 0 {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment thickness must be positive"))
	}
	// A wall must have non-zero length
	if w.X1 == w.X2 && w.Y1 == w.Y2 {
		return errors.Join(ErrInvalidPlan, errors.New("wall segment start and end coordinates cannot be identical"))
	}
	return nil
}

// Zone represents a 2D closed polygon defining a room or outdoor area.
type Zone struct {
	ID     string
	Name   string
	Color  string
	Points []Point2D
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

// Level represents an indoor story or outdoor space of the building.
type Level struct {
	ID        string
	Name      string
	Order     int
	IsOutdoor bool
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
