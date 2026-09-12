package dashboard

import (
	"time"

	"github.com/JLugagne/walldash/domain"
)

// CreateDashboardRequest contains payload for creating a new dashboard.
// Cols and Rows are optional: a zero value lets the server apply the default Widget Grid size.
type CreateDashboardRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=100"`
	Order             int    `json:"order"`
	Cols              int    `json:"cols,omitempty" validate:"gte=0"`
	Rows              int    `json:"rows,omitempty" validate:"gte=0"`
	BackgroundImage   string `json:"background_image,omitempty"`
	BackgroundOpacity int    `json:"background_opacity,omitempty" validate:"omitempty,gte=0,lte=100"`
	BackgroundBlur    int    `json:"background_blur,omitempty" validate:"omitempty,gte=0,lte=32"`
	BackgroundDim     int    `json:"background_dim,omitempty" validate:"omitempty,gte=0,lte=100"`
}

// UpdateDashboardRequest contains payload for updating a dashboard.
// Cols and Rows are optional: a zero value preserves the dashboard's current Widget Grid size.
type UpdateDashboardRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=100"`
	Order             int    `json:"order"`
	Cols              int    `json:"cols,omitempty" validate:"gte=0"`
	Rows              int    `json:"rows,omitempty" validate:"gte=0"`
	BackgroundImage   string `json:"background_image,omitempty"`
	BackgroundOpacity int    `json:"background_opacity,omitempty" validate:"omitempty,gte=0,lte=100"`
	BackgroundBlur    int    `json:"background_blur,omitempty" validate:"omitempty,gte=0,lte=32"`
	BackgroundDim     int    `json:"background_dim,omitempty" validate:"omitempty,gte=0,lte=100"`
}

// WidgetConfigDTO contains configuration data for a widget in API payloads.
// Min and Max are pointers so an absent bound stays distinguishable from a bound of zero.
type WidgetConfigDTO struct {
	EntityIDs    []string          `json:"entity_ids"`
	Display      string            `json:"display" validate:"required"`
	Labels       map[string]string `json:"labels,omitempty"`
	Min          *float64          `json:"min,omitempty"`
	Max          *float64          `json:"max,omitempty"`
	Unit         string            `json:"unit,omitempty"`
	WeatherMode  string            `json:"weather_mode,omitempty" validate:"omitempty,oneof=current today tomorrow ndays"`
	WeatherDays  int               `json:"weather_days,omitempty" validate:"omitempty,gte=1,lte=14"`
	Latitude     *float64          `json:"latitude,omitempty" validate:"omitempty,gte=-90,lte=90"`
	Longitude    *float64          `json:"longitude,omitempty" validate:"omitempty,gte=-180,lte=180"`
	LocationName string            `json:"location_name,omitempty" validate:"max=100"`
	Units        string            `json:"units,omitempty" validate:"omitempty,oneof=metric imperial"`
}

// CreateWidgetRequest contains payload for adding a new widget to a dashboard.
// Title is optional: the UI falls back to the config label then to the Home Assistant device name.
type CreateWidgetRequest struct {
	Type    string          `json:"type" validate:"required"`
	Title   string          `json:"title" validate:"max=100"`
	Order   int             `json:"order"`
	Col     int             `json:"col" validate:"gte=0"`
	Row     int             `json:"row" validate:"gte=0"`
	ColSpan int             `json:"col_span" validate:"gt=0"`
	RowSpan int             `json:"row_span" validate:"gt=0"`
	Config  WidgetConfigDTO `json:"config"`
}

// UpdateWidgetRequest contains payload for updating a widget's title and configuration.
// Position is not part of this request: it is only ever changed through UpdateLayoutRequest.
type UpdateWidgetRequest struct {
	Title  string          `json:"title" validate:"max=100"`
	Config WidgetConfigDTO `json:"config"`
}

// WidgetPositionDTO is the payload of a single widget's placement in a layout write.
type WidgetPositionDTO struct {
	ID      string `json:"id" validate:"required"`
	Col     int    `json:"col" validate:"gte=0"`
	Row     int    `json:"row" validate:"gte=0"`
	ColSpan int    `json:"col_span" validate:"gt=0"`
	RowSpan int    `json:"row_span" validate:"gt=0"`
}

// UpdateLayoutRequest replaces the positions of some or all widgets of a dashboard in one write.
type UpdateLayoutRequest struct {
	Positions []WidgetPositionDTO `json:"positions" validate:"required,dive"`
}

// WidgetResponse represents a widget in API responses.
type WidgetResponse struct {
	ID          string          `json:"id"`
	DashboardID string          `json:"dashboard_id"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Order       int             `json:"order"`
	Col         int             `json:"col"`
	Row         int             `json:"row"`
	ColSpan     int             `json:"col_span"`
	RowSpan     int             `json:"row_span"`
	Config      WidgetConfigDTO `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// DashboardResponse represents a dashboard in API responses.
type DashboardResponse struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Order             int              `json:"order"`
	Cols              int              `json:"cols"`
	Rows              int              `json:"rows"`
	BackgroundImage   string           `json:"background_image"`
	BackgroundOpacity int              `json:"background_opacity"`
	BackgroundBlur    int              `json:"background_blur"`
	BackgroundDim     int              `json:"background_dim"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	Widgets           []WidgetResponse `json:"widgets"`
}

// AutomationResponse represents an automation in API responses.
type AutomationResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	State         string     `json:"state"`
	Current       int        `json:"current"`
	LastTriggered *time.Time `json:"last_triggered"`
}

// TriggerAutomationResponse represents response after triggering an automation.
type TriggerAutomationResponse struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

// Validation errors for dashboard requests
var (
	ErrInvalidCreateDashboardRequest = &domain.Error{
		Code:    "INVALID_CREATE_DASHBOARD_REQUEST",
		Message: "invalid create dashboard request data",
	}
	ErrInvalidUpdateDashboardRequest = &domain.Error{
		Code:    "INVALID_UPDATE_DASHBOARD_REQUEST",
		Message: "invalid update dashboard request data",
	}
	ErrInvalidCreateWidgetRequest = &domain.Error{
		Code:    "INVALID_CREATE_WIDGET_REQUEST",
		Message: "invalid create widget request data",
	}
	ErrInvalidUpdateWidgetRequest = &domain.Error{
		Code:    "INVALID_UPDATE_WIDGET_REQUEST",
		Message: "invalid update widget request data",
	}
	ErrInvalidUpdateLayoutRequest = &domain.Error{
		Code:    "INVALID_UPDATE_LAYOUT_REQUEST",
		Message: "invalid update layout request data",
	}
	ErrInvalidTriggerAutomationRequest = &domain.Error{
		Code:    "INVALID_TRIGGER_AUTOMATION_REQUEST",
		Message: "invalid trigger automation request data",
	}
)
