package dashboard

import (
	"time"

	"github.com/JLugagne/walldash/domain"
)

// CreateLevelRequest contains payload for creating a new level.
type CreateLevelRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	IsOutdoor bool     `json:"is_outdoor"`
	Layers    []string `json:"layers,omitempty"`
}

// UpdateLevelRequest contains payload for updating an existing level.
type UpdateLevelRequest struct {
	Name      string   `json:"name" validate:"required,min=1,max=100"`
	IsOutdoor bool     `json:"is_outdoor"`
	Layers    []string `json:"layers,omitempty"`
}

// ReorderLevelsRequest specifies the desired order of level IDs.
type ReorderLevelsRequest struct {
	LevelIDs []string `json:"level_ids" validate:"required,min=1,dive,required"`
}

// LevelResponse represents a level entity in API responses.
type LevelResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Order     int       `json:"order"`
	IsOutdoor bool      `json:"is_outdoor"`
	Layers    []string  `json:"layers"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Point2DDTO represents 2D coordinates in API requests and responses.
type Point2DDTO struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WallOpeningDTO represents a door or window attached to a wall segment in plan operations.
type WallOpeningDTO struct {
	ID        string  `json:"id" validate:"required"`
	Type      string  `json:"type" validate:"required,oneof=door window"`
	Offset    float64 `json:"offset" validate:"gte=0"`
	Width     float64 `json:"width" validate:"gt=0"`
	FlipSide  bool    `json:"flip_side"`
	FlipHinge bool    `json:"flip_hinge"`
	HideDoor  bool    `json:"hide_door"`
}

// WallSegmentDTO represents a wall segment in plan operations.
type WallSegmentDTO struct {
	ID        string           `json:"id" validate:"required"`
	X1        float64          `json:"x1"`
	Y1        float64          `json:"y1"`
	X2        float64          `json:"x2"`
	Y2        float64          `json:"y2"`
	Thickness float64          `json:"thickness" validate:"gt=0"`
	Openings  []WallOpeningDTO `json:"openings" validate:"dive"`
}

// ZoneDTO represents a 2D closed polygon zone in plan operations.
type ZoneDTO struct {
	ID             string       `json:"id" validate:"required"`
	Name           string       `json:"name" validate:"required"`
	Color          string       `json:"color" validate:"required"`
	Points         []Point2DDTO `json:"points" validate:"required,min=3,dive"`
	TempSensor     string       `json:"temp_sensor,omitempty"`
	TempMin        *float64     `json:"temp_min,omitempty"`
	TempMax        *float64     `json:"temp_max,omitempty"`
	HumiditySensor string       `json:"humidity_sensor,omitempty"`
	HumidityMin    *float64     `json:"humidity_min,omitempty"`
	HumidityMax    *float64     `json:"humidity_max,omitempty"`
}

// SavePlanRequest contains the walls and zones to persist for a level's plan.
type SavePlanRequest struct {
	Walls []WallSegmentDTO `json:"walls" validate:"dive"`
	Zones []ZoneDTO        `json:"zones" validate:"dive"`
}

// PlanResponse represents the 2D architectural blueprint in API responses.
type PlanResponse struct {
	LevelID string           `json:"level_id"`
	Walls   []WallSegmentDTO `json:"walls"`
	Zones   []ZoneDTO        `json:"zones"`
}

// Validation errors for dashboard public level types
var (
	ErrInvalidCreateLevelRequest = &domain.Error{
		Code:    "INVALID_CREATE_LEVEL_REQUEST",
		Message: "invalid create level request data",
	}
	ErrInvalidUpdateLevelRequest = &domain.Error{
		Code:    "INVALID_UPDATE_LEVEL_REQUEST",
		Message: "invalid update level request data",
	}
	ErrInvalidReorderLevelsRequest = &domain.Error{
		Code:    "INVALID_REORDER_LEVELS_REQUEST",
		Message: "invalid reorder levels request data",
	}
	ErrInvalidSavePlanRequest = &domain.Error{
		Code:    "INVALID_SAVE_PLAN_REQUEST",
		Message: "invalid save plan request data",
	}
)
