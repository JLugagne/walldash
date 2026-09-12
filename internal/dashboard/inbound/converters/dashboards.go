package converters

import (
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ToDomainDashboard converts a public CreateDashboardRequest to a domain Dashboard.
// A zero Cols or Rows is replaced by the default Widget Grid size.
const defaultBackgroundImage = "/backgrounds/desert-night.jpg"

func ToDomainDashboard(req pkgdashboard.CreateDashboardRequest) domain.Dashboard {
	cols := req.Cols
	if cols == 0 {
		cols = domain.DefaultGridCols
	}
	rows := req.Rows
	if rows == 0 {
		rows = domain.DefaultGridRows
	}
	bgImage := req.BackgroundImage
	if bgImage == "" {
		bgImage = defaultBackgroundImage
	}
	bgOpacity, bgBlur, bgDim := req.BackgroundOpacity, req.BackgroundBlur, req.BackgroundDim
	if bgOpacity == 0 && bgBlur == 0 && bgDim == 0 {
		bgOpacity, bgBlur, bgDim = 95, 14, 50
	}
	return domain.Dashboard{
		Name:              req.Name,
		Order:             req.Order,
		Cols:              cols,
		Rows:              rows,
		BackgroundImage:   bgImage,
		BackgroundOpacity: bgOpacity,
		BackgroundBlur:    bgBlur,
		BackgroundDim:     bgDim,
		Widgets:           []domain.Widget{},
	}
}

// ToDomainUpdateDashboard converts a public UpdateDashboardRequest to a domain Dashboard.
// Cols and Rows are passed through as-is, including zero: the app layer resolves a zero value
// against the dashboard's currently stored grid size.
func ToDomainUpdateDashboard(id string, req pkgdashboard.UpdateDashboardRequest) domain.Dashboard {
	return domain.Dashboard{
		ID:                id,
		Name:              req.Name,
		Order:             req.Order,
		Cols:              req.Cols,
		Rows:              req.Rows,
		BackgroundImage:   req.BackgroundImage,
		BackgroundOpacity: req.BackgroundOpacity,
		BackgroundBlur:    req.BackgroundBlur,
		BackgroundDim:     req.BackgroundDim,
		Widgets:           []domain.Widget{},
	}
}

func toPublicWidgetConfig(c domain.WidgetConfig) pkgdashboard.WidgetConfigDTO {
	entityIDs := c.EntityIDs
	if entityIDs == nil {
		entityIDs = []string{}
	}
	return pkgdashboard.WidgetConfigDTO{
		EntityIDs:     entityIDs,
		Display:       c.Display,
		Labels:        c.Labels,
		Min:           c.Min,
		Max:           c.Max,
		Unit:          c.Unit,
		WeatherMode:   c.WeatherMode,
		WeatherDays:   c.WeatherDays,
		Latitude:      c.Latitude,
		Longitude:     c.Longitude,
		LocationName:  c.LocationName,
		Units:         c.Units,
		OnAutomation:  c.OnAutomation,
		OffAutomation: c.OffAutomation,
	}
}

func toDomainWidgetConfig(dto pkgdashboard.WidgetConfigDTO) domain.WidgetConfig {
	entityIDs := dto.EntityIDs
	if entityIDs == nil {
		entityIDs = []string{}
	}
	weatherMode := dto.WeatherMode
	if weatherMode == "" && dto.Display == domain.DisplayWeather {
		weatherMode = domain.WeatherModeCurrent
	}
	return domain.WidgetConfig{
		EntityIDs:     entityIDs,
		Display:       dto.Display,
		Labels:        dto.Labels,
		Min:           dto.Min,
		Max:           dto.Max,
		Unit:          dto.Unit,
		WeatherMode:   weatherMode,
		WeatherDays:   dto.WeatherDays,
		Latitude:      dto.Latitude,
		Longitude:     dto.Longitude,
		LocationName:  dto.LocationName,
		Units:         dto.Units,
		OnAutomation:  dto.OnAutomation,
		OffAutomation: dto.OffAutomation,
	}
}

// ToPublicWidget converts a domain Widget to a public WidgetResponse.
func ToPublicWidget(w domain.Widget) pkgdashboard.WidgetResponse {
	return pkgdashboard.WidgetResponse{
		ID:          w.ID,
		DashboardID: w.DashboardID,
		Type:        w.Type,
		Title:       w.Title,
		Order:       w.Order,
		Col:         w.Col,
		Row:         w.Row,
		ColSpan:     w.ColSpan,
		RowSpan:     w.RowSpan,
		Config:      toPublicWidgetConfig(w.Config),
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

// ToPublicWidgets converts a slice of domain Widgets to public WidgetResponses.
func ToPublicWidgets(widgets []domain.Widget) []pkgdashboard.WidgetResponse {
	if widgets == nil {
		return []pkgdashboard.WidgetResponse{}
	}
	res := make([]pkgdashboard.WidgetResponse, len(widgets))
	for i, w := range widgets {
		res[i] = ToPublicWidget(w)
	}
	return res
}

// ToDomainWidget converts a public CreateWidgetRequest into a domain Widget.
func ToDomainWidget(dashboardID string, req pkgdashboard.CreateWidgetRequest) domain.Widget {
	return domain.Widget{
		DashboardID: dashboardID,
		Type:        req.Type,
		Title:       req.Title,
		Order:       req.Order,
		Col:         req.Col,
		Row:         req.Row,
		ColSpan:     req.ColSpan,
		RowSpan:     req.RowSpan,
		Config:      toDomainWidgetConfig(req.Config),
	}
}

// ToDomainUpdateWidget converts a public UpdateWidgetRequest into a domain Widget carrying only
// the fields an update may change: title and configuration. Position is deliberately absent,
// since it is only ever written through ToDomainWidgetPositions.
func ToDomainUpdateWidget(dashboardID, widgetID string, req pkgdashboard.UpdateWidgetRequest) domain.Widget {
	return domain.Widget{
		ID:          widgetID,
		DashboardID: dashboardID,
		Title:       req.Title,
		Config:      toDomainWidgetConfig(req.Config),
	}
}

// ToDomainWidgetPositions converts a public UpdateLayoutRequest into domain WidgetPositions.
func ToDomainWidgetPositions(req pkgdashboard.UpdateLayoutRequest) []domain.WidgetPosition {
	positions := make([]domain.WidgetPosition, len(req.Positions))
	for i, p := range req.Positions {
		positions[i] = domain.WidgetPosition{
			ID:      p.ID,
			Col:     p.Col,
			Row:     p.Row,
			ColSpan: p.ColSpan,
			RowSpan: p.RowSpan,
		}
	}
	return positions
}

// ToPublicDashboard converts a domain Dashboard to a public DashboardResponse.
func ToPublicDashboard(ov domain.Dashboard) pkgdashboard.DashboardResponse {
	return pkgdashboard.DashboardResponse{
		ID:                ov.ID,
		Name:              ov.Name,
		Order:             ov.Order,
		Cols:              ov.Cols,
		Rows:              ov.Rows,
		BackgroundImage:   ov.BackgroundImage,
		BackgroundOpacity: ov.BackgroundOpacity,
		BackgroundBlur:    ov.BackgroundBlur,
		BackgroundDim:     ov.BackgroundDim,
		CreatedAt:         ov.CreatedAt,
		UpdatedAt:         ov.UpdatedAt,
		Widgets:           ToPublicWidgets(ov.Widgets),
	}
}

// ToPublicDashboards converts a slice of domain Dashboards to public DashboardResponses.
func ToPublicDashboards(dashboards []domain.Dashboard) []pkgdashboard.DashboardResponse {
	if dashboards == nil {
		return []pkgdashboard.DashboardResponse{}
	}
	res := make([]pkgdashboard.DashboardResponse, len(dashboards))
	for i, o := range dashboards {
		res[i] = ToPublicDashboard(o)
	}
	return res
}

// ToPublicAutomation converts a domain Automation to a public AutomationResponse.
func ToPublicAutomation(a domain.Automation) pkgdashboard.AutomationResponse {
	return pkgdashboard.AutomationResponse{
		ID:            a.ID,
		Name:          a.Name,
		State:         a.State,
		Current:       a.Current,
		LastTriggered: a.LastTriggered,
	}
}

// ToPublicAutomations converts a slice of domain Automations to public AutomationResponses.
func ToPublicAutomations(automations []domain.Automation) []pkgdashboard.AutomationResponse {
	if automations == nil {
		return []pkgdashboard.AutomationResponse{}
	}
	res := make([]pkgdashboard.AutomationResponse, len(automations))
	for i, a := range automations {
		res[i] = ToPublicAutomation(a)
	}
	return res
}
