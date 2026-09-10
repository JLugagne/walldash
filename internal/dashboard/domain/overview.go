package domain

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Widget types name what a widget is bound to and what a tap on it does.
// They never encode how the widget is drawn: that is the Display Mode axis.
const (
	WidgetTypeSensor         = "sensor"
	WidgetTypeActuator       = "actuator"
	WidgetTypeAutomationList = "automation_list"
	WidgetTypeWeather        = "weather"
)

// Display modes name how a widget draws the data it is bound to.
// DisplayArc is a dial open at the bottom, never a full circle.
const (
	DisplayNumber  = "number"
	DisplayArc     = "arc"
	DisplayBar     = "bar"
	DisplayToggle  = "toggle"
	DisplayList    = "list"
	DisplayWeather = "weather"
)

// Grid dimensions applied to an Overview Dashboard created without explicit ones.
// A dashboard occupies exactly the viewport, so these are cell counts, not pixels.
const (
	DefaultGridCols = 12
	DefaultGridRows = 8
)

// WidgetSize is a footprint in Widget Grid cells.
type WidgetSize struct {
	Cols int
	Rows int
}

var widgetDisplayMatrix = map[string]map[string]WidgetSize{
	WidgetTypeWeather: {
		DisplayWeather: {Cols: 2, Rows: 2},
	},
	WidgetTypeSensor: {
		DisplayNumber: {Cols: 1, Rows: 1},
		DisplayBar:    {Cols: 2, Rows: 1},
		DisplayArc:    {Cols: 2, Rows: 2},
	},
	WidgetTypeActuator: {
		DisplayToggle: {Cols: 1, Rows: 1},
	},
	WidgetTypeAutomationList: {
		DisplayList: {Cols: 2, Rows: 2},
	},
}

// MinimumWidgetSize returns the smallest footprint a widget of that type and display
// may occupy, and reports whether the pair is legal at all. It is the single source of
// truth for both questions: a false second result means the display mode is not offered
// for that widget type, not that the pair has no minimum.
func MinimumWidgetSize(widgetType, display string) (WidgetSize, bool) {
	displays, ok := widgetDisplayMatrix[widgetType]
	if !ok {
		return WidgetSize{}, false
	}
	size, ok := displays[display]
	return size, ok
}

// IsSupportedWidgetType reports whether the type is one the domain knows how to validate.
func IsSupportedWidgetType(widgetType string) bool {
	_, ok := widgetDisplayMatrix[widgetType]
	return ok
}

// WidgetConfig holds configuration specific to a widget instance.
// Min and Max are pointers so that an absent bound stays distinguishable from a bound of
// zero, which a temperature arc needs. Labels renames the bound devices for this widget
// only; it never touches DevicePlacement.CustomName.
type WidgetConfig struct {
	EntityIDs    []string          `json:"entity_ids,omitempty"`
	Display      string            `json:"display"`
	Labels       map[string]string `json:"labels,omitempty"`
	Min          *float64          `json:"min,omitempty"`
	Max          *float64          `json:"max,omitempty"`
	Unit         string            `json:"unit,omitempty"`
	WeatherMode  string            `json:"weather_mode,omitempty"`
	WeatherDays  int               `json:"weather_days,omitempty"`
	Latitude     *float64          `json:"latitude,omitempty"`
	Longitude    *float64          `json:"longitude,omitempty"`
	LocationName string            `json:"location_name,omitempty"`
	Units        string            `json:"units,omitempty"`
}

// Widget represents an interactive visual component anchored in the Widget Grid of an
// Overview Dashboard. Col and Row are the zero-based top-left cell, ColSpan and RowSpan
// the footprint in cells.
type Widget struct {
	ID          string
	DashboardID string
	Type        string
	Title       string
	Order       int
	Col         int
	Row         int
	ColSpan     int
	RowSpan     int
	Config      WidgetConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks every rule that does not need the enclosing dashboard: identity, the
// type and display pair, the display minimum footprint, non-negative anchoring, entity
// cardinality, the bounds required by arc and bar, and the action whitelist for an
// actuator. Title is deliberately not checked: it is optional and the UI falls back to
// the config label then to the Home Assistant device name. Returns an error wrapping
// ErrInvalidWidget.
func (w Widget) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.Join(ErrInvalidWidget, errors.New("widget id cannot be empty"))
	}
	if strings.TrimSpace(w.DashboardID) == "" {
		return errors.Join(ErrInvalidWidget, errors.New("widget dashboard_id cannot be empty"))
	}
	if !IsSupportedWidgetType(w.Type) {
		return errors.Join(ErrInvalidWidget, errors.New("unsupported widget type: "+w.Type))
	}

	minimum, ok := MinimumWidgetSize(w.Type, w.Config.Display)
	if !ok {
		return errors.Join(ErrInvalidWidget, errors.New("widget type "+w.Type+" does not accept display mode "+quoteDisplay(w.Config.Display)))
	}

	if w.Col < 0 || w.Row < 0 {
		return errors.Join(ErrInvalidWidget, errors.New("widget col and row must be non-negative"))
	}
	if w.ColSpan <= 0 || w.RowSpan <= 0 {
		return errors.Join(ErrInvalidWidget, errors.New("widget col_span and row_span must be strictly positive"))
	}
	if w.ColSpan < minimum.Cols || w.RowSpan < minimum.Rows {
		return errors.Join(ErrInvalidWidget, errors.New("display mode "+w.Config.Display+" requires at least "+
			strconv.Itoa(minimum.Cols)+"x"+strconv.Itoa(minimum.Rows)+" cells"))
	}

	if err := w.validateEntities(); err != nil {
		return err
	}
	if err := w.validateWeatherConfig(); err != nil {
		return err
	}
	return w.validateBounds()
}

// ValidateIn is Validate plus the viewport invariant col+col_span <= cols and
// row+row_span <= rows, so that no widget ever leaves a dashboard that never scrolls.
// Returns an error wrapping ErrInvalidWidget.
func (w Widget) ValidateIn(cols, rows int) error {
	if err := w.Validate(); err != nil {
		return err
	}
	if w.Col+w.ColSpan > cols {
		return errors.Join(ErrInvalidWidget, errors.New("widget "+w.ID+" overflows the grid width"))
	}
	if w.Row+w.RowSpan > rows {
		return errors.Join(ErrInvalidWidget, errors.New("widget "+w.ID+" overflows the grid height"))
	}
	return nil
}

// Overlaps reports whether the two widget rectangles share at least one grid cell.
// Rectangles that only touch by an edge do not overlap.
func (w Widget) Overlaps(other Widget) bool {
	return w.Col < other.Col+other.ColSpan &&
		other.Col < w.Col+w.ColSpan &&
		w.Row < other.Row+other.RowSpan &&
		other.Row < w.Row+w.RowSpan
}

func (w Widget) validateEntities() error {
	switch w.Type {
	case WidgetTypeSensor, WidgetTypeActuator:
		if len(w.Config.EntityIDs) != 1 {
			return errors.Join(ErrInvalidWidget, errors.New("widget type "+w.Type+" requires exactly one entity"))
		}
	case WidgetTypeAutomationList:
		if len(w.Config.EntityIDs) == 0 {
			return errors.Join(ErrInvalidWidget, errors.New("widget type "+w.Type+" requires at least one entity"))
		}
	case WidgetTypeWeather:
		if len(w.Config.EntityIDs) != 0 {
			return errors.Join(ErrInvalidWidget, errors.New("widget type "+w.Type+" requires no entities"))
		}
	}

	for _, entityID := range w.Config.EntityIDs {
		if strings.TrimSpace(entityID) == "" {
			return errors.Join(ErrInvalidWidget, errors.New("widget entity id cannot be empty"))
		}
	}

	if w.Type != WidgetTypeActuator {
		return nil
	}

	entityDomain, _, found := strings.Cut(w.Config.EntityIDs[0], ".")
	if !found || !IsAllowedActionDomain(entityDomain) {
		return errors.Join(ErrInvalidWidget, errors.New("actuator widget entity "+w.Config.EntityIDs[0]+" is not in an actionable domain"))
	}
	return nil
}

func (w Widget) validateBounds() error {
	if w.Config.Display != DisplayArc && w.Config.Display != DisplayBar {
		return nil
	}
	if w.Config.Min == nil || w.Config.Max == nil {
		return errors.Join(ErrInvalidWidget, errors.New("display mode "+w.Config.Display+" requires both min and max bounds"))
	}
	if *w.Config.Min >= *w.Config.Max {
		return errors.Join(ErrInvalidWidget, errors.New("display mode "+w.Config.Display+" requires min strictly lower than max"))
	}
	return nil
}

func quoteDisplay(display string) string {
	if strings.TrimSpace(display) == "" {
		return "an empty display mode"
	}
	return display
}

// OverviewDashboard represents a customizable high-level overview view.
// Cols and Rows are the Widget Grid dimensions, persisted per dashboard so that changing
// the default never rearranges an already composed dashboard.
type OverviewDashboard struct {
	ID    string
	Name  string
	Order int
	Cols  int
	Rows  int
	// BackgroundImage is a URL path rendered behind the Widget Grid, or empty for none.
	BackgroundImage string
	// BackgroundOpacity is the image opacity in percent (0-100).
	BackgroundOpacity int
	// BackgroundBlur is the image blur radius in pixels (0-32).
	BackgroundBlur int
	// BackgroundDim is the opacity in percent of the scrim drawn over the image (0-100).
	BackgroundDim int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Widgets       []Widget
}

// Validate checks identity, strictly positive grid dimensions, then every widget against
// the grid, then rejects any pair of overlapping widgets. Returns an error wrapping
// ErrInvalidOverview, or ErrInvalidWidget when a single widget is at fault.
func (o OverviewDashboard) Validate() error {
	if strings.TrimSpace(o.ID) == "" {
		return errors.Join(ErrInvalidOverview, errors.New("overview id cannot be empty"))
	}
	if strings.TrimSpace(o.Name) == "" {
		return errors.Join(ErrInvalidOverview, errors.New("overview name cannot be empty"))
	}
	if o.Cols <= 0 || o.Rows <= 0 {
		return errors.Join(ErrInvalidOverview, errors.New("overview grid dimensions must be strictly positive"))
	}

	if o.BackgroundOpacity < 0 || o.BackgroundOpacity > 100 {
		return errors.Join(ErrInvalidOverview, errors.New("overview background opacity must be between 0 and 100"))
	}
	if o.BackgroundBlur < 0 || o.BackgroundBlur > 32 {
		return errors.Join(ErrInvalidOverview, errors.New("overview background blur must be between 0 and 32"))
	}
	if o.BackgroundDim < 0 || o.BackgroundDim > 100 {
		return errors.Join(ErrInvalidOverview, errors.New("overview background dim must be between 0 and 100"))
	}
	for _, w := range o.Widgets {
		if err := w.ValidateIn(o.Cols, o.Rows); err != nil {
			return err
		}
	}

	for i := range o.Widgets {
		for j := i + 1; j < len(o.Widgets); j++ {
			if o.Widgets[i].Overlaps(o.Widgets[j]) {
				return errors.Join(ErrInvalidOverview, errors.New("widgets "+o.Widgets[i].ID+" and "+o.Widgets[j].ID+" overlap"))
			}
		}
	}
	return nil
}

// WidgetPosition is the payload of a layout write: where a widget sits, never what it shows.
type WidgetPosition struct {
	ID      string
	Col     int
	Row     int
	ColSpan int
	RowSpan int
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

func (w Widget) validateWeatherConfig() error {
	if w.Type != WidgetTypeWeather {
		return nil
	}

	switch w.Config.WeatherMode {
	case WeatherModeCurrent, WeatherModeToday, WeatherModeTomorrow:
	case WeatherModeNDays:
		if w.Config.WeatherDays < 1 || w.Config.WeatherDays > 14 {
			return errors.Join(ErrInvalidWidget, errors.New("weather widget with mode ndays requires weather_days between 1 and 14"))
		}
	default:
		return errors.Join(ErrInvalidWidget, errors.New("unsupported weather mode: "+w.Config.WeatherMode))
	}

	switch w.Config.Units {
	case "", WeatherUnitsMetric, WeatherUnitsImperial:
	default:
		return errors.Join(ErrInvalidWidget, errors.New("unsupported weather units: "+w.Config.Units))
	}

	return nil
}
