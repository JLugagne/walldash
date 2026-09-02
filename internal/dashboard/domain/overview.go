package domain

import (
	"errors"
	"strings"
	"time"
)

// Supported widget types for overview dashboards
const (
	WidgetTypeAutomationList = "automation_list"
)

// SupportedWidgetTypes lists all valid widget types
var SupportedWidgetTypes = []string{
	WidgetTypeAutomationList,
}

// IsSupportedWidgetType verifies if the widget type is supported.
func IsSupportedWidgetType(widgetType string) bool {
	for _, t := range SupportedWidgetTypes {
		if t == widgetType {
			return true
		}
	}
	return false
}

// WidgetConfig holds configuration specific to a widget instance.
type WidgetConfig struct {
	EntityIDs []string `json:"entity_ids,omitempty"`
}

// Widget represents an interactive visual component embedded inside an Overview Dashboard.
type Widget struct {
	ID          string
	DashboardID string
	Type        string
	Title       string
	Order       int
	Config      WidgetConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks whether the widget entity conforms to domain rules.
func (w Widget) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.Join(ErrInvalidWidget, errors.New("widget id cannot be empty"))
	}
	if strings.TrimSpace(w.DashboardID) == "" {
		return errors.Join(ErrInvalidWidget, errors.New("widget dashboard_id cannot be empty"))
	}
	if strings.TrimSpace(w.Title) == "" {
		return errors.Join(ErrInvalidWidget, errors.New("widget title cannot be empty"))
	}
	if !IsSupportedWidgetType(w.Type) {
		return errors.Join(ErrInvalidWidget, errors.New("unsupported widget type: "+w.Type))
	}
	return nil
}

// OverviewDashboard represents a customizable high-level overview view.
type OverviewDashboard struct {
	ID        string
	Name      string
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
	Widgets   []Widget
}

// Validate checks whether the overview dashboard conforms to domain rules.
func (o OverviewDashboard) Validate() error {
	if strings.TrimSpace(o.ID) == "" {
		return errors.Join(ErrInvalidOverview, errors.New("overview id cannot be empty"))
	}
	if strings.TrimSpace(o.Name) == "" {
		return errors.Join(ErrInvalidOverview, errors.New("overview name cannot be empty"))
	}
	for _, w := range o.Widgets {
		if err := w.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Automation represents an automation discovered from Home Assistant.
type Automation struct {
	ID            string
	Name          string
	State         string // "on" / "off"
	Current       int    // Number of currently running instances (> 0 means active)
	LastTriggered *time.Time
}

// Validate checks whether the automation entity is valid.
func (a Automation) Validate() error {
	trimmed := strings.TrimSpace(a.ID)
	if trimmed == "" {
		return errors.Join(ErrAutomationNotFound, errors.New("automation id cannot be empty"))
	}
	if !strings.HasPrefix(trimmed, "automation.") {
		return errors.Join(ErrAutomationNotFound, errors.New("automation id must start with 'automation.' prefix"))
	}
	return nil
}
