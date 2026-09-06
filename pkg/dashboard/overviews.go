package dashboard

import (
	"time"

	"github.com/JLugagne/ha-dash/domain"
)

// CreateOverviewRequest contains payload for creating a new overview dashboard.
// Cols and Rows are optional: a zero value lets the server apply the default Widget Grid size.
type CreateOverviewRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Order int    `json:"order"`
	Cols  int    `json:"cols,omitempty" validate:"gte=0"`
	Rows  int    `json:"rows,omitempty" validate:"gte=0"`
}

// UpdateOverviewRequest contains payload for updating an overview dashboard.
// Cols and Rows are optional: a zero value preserves the dashboard's current Widget Grid size.
type UpdateOverviewRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Order int    `json:"order"`
	Cols  int    `json:"cols,omitempty" validate:"gte=0"`
	Rows  int    `json:"rows,omitempty" validate:"gte=0"`
}

// WidgetConfigDTO contains configuration data for a widget in API payloads.
// Min and Max are pointers so an absent bound stays distinguishable from a bound of zero.
type WidgetConfigDTO struct {
	EntityIDs []string          `json:"entity_ids"`
	Display   string            `json:"display" validate:"required"`
	Labels    map[string]string `json:"labels,omitempty"`
	Min       *float64          `json:"min,omitempty"`
	Max       *float64          `json:"max,omitempty"`
	Unit      string            `json:"unit,omitempty"`
}

// CreateWidgetRequest contains payload for adding a new widget to an overview dashboard.
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

// OverviewResponse represents an overview dashboard in API responses.
type OverviewResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Order     int              `json:"order"`
	Cols      int              `json:"cols"`
	Rows      int              `json:"rows"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Widgets   []WidgetResponse `json:"widgets"`
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

// Validation errors for overview dashboard requests
var (
	ErrInvalidCreateOverviewRequest = &domain.Error{
		Code:    "INVALID_CREATE_OVERVIEW_REQUEST",
		Message: "invalid create overview request data",
	}
	ErrInvalidUpdateOverviewRequest = &domain.Error{
		Code:    "INVALID_UPDATE_OVERVIEW_REQUEST",
		Message: "invalid update overview request data",
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
