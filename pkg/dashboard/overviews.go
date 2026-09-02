package dashboard

import (
	"time"

	"github.com/JLugagne/ha-dash/domain"
)

// CreateOverviewRequest contains payload for creating a new overview dashboard.
type CreateOverviewRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Order int    `json:"order"`
}

// UpdateOverviewRequest contains payload for updating an overview dashboard.
type UpdateOverviewRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Order int    `json:"order"`
}

// WidgetConfigDTO contains configuration data for a widget in API payloads.
type WidgetConfigDTO struct {
	EntityIDs []string `json:"entity_ids"`
}

// CreateWidgetRequest contains payload for adding a new widget to an overview dashboard.
type CreateWidgetRequest struct {
	Type   string          `json:"type" validate:"required"`
	Title  string          `json:"title" validate:"required,min=1,max=100"`
	Order  int             `json:"order"`
	Config WidgetConfigDTO `json:"config"`
}

// WidgetResponse represents a widget in API responses.
type WidgetResponse struct {
	ID          string          `json:"id"`
	DashboardID string          `json:"dashboard_id"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Order       int             `json:"order"`
	Config      WidgetConfigDTO `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// OverviewResponse represents an overview dashboard in API responses.
type OverviewResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Order     int              `json:"order"`
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
	ErrInvalidTriggerAutomationRequest = &domain.Error{
		Code:    "INVALID_TRIGGER_AUTOMATION_REQUEST",
		Message: "invalid trigger automation request data",
	}
)
