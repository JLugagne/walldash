package converters

import (
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
)

// ToDomainOverview converts a public CreateOverviewRequest to a domain OverviewDashboard.
func ToDomainOverview(req pkgdashboard.CreateOverviewRequest) domain.OverviewDashboard {
	return domain.OverviewDashboard{
		Name:    req.Name,
		Order:   req.Order,
		Widgets: []domain.Widget{},
	}
}

// ToDomainUpdateOverview converts a public UpdateOverviewRequest to a domain OverviewDashboard.
func ToDomainUpdateOverview(id string, req pkgdashboard.UpdateOverviewRequest) domain.OverviewDashboard {
	return domain.OverviewDashboard{
		ID:      id,
		Name:    req.Name,
		Order:   req.Order,
		Widgets: []domain.Widget{},
	}
}

// ToPublicWidget converts a domain Widget to a public WidgetResponse.
func ToPublicWidget(w domain.Widget) pkgdashboard.WidgetResponse {
	entityIDs := w.Config.EntityIDs
	if entityIDs == nil {
		entityIDs = []string{}
	}
	return pkgdashboard.WidgetResponse{
		ID:          w.ID,
		DashboardID: w.DashboardID,
		Type:        w.Type,
		Title:       w.Title,
		Order:       w.Order,
		Config: pkgdashboard.WidgetConfigDTO{
			EntityIDs: entityIDs,
		},
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
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
	entityIDs := req.Config.EntityIDs
	if entityIDs == nil {
		entityIDs = []string{}
	}
	return domain.Widget{
		DashboardID: dashboardID,
		Type:        req.Type,
		Title:       req.Title,
		Order:       req.Order,
		Config: domain.WidgetConfig{
			EntityIDs: entityIDs,
		},
	}
}

// ToPublicOverview converts a domain OverviewDashboard to a public OverviewResponse.
func ToPublicOverview(ov domain.OverviewDashboard) pkgdashboard.OverviewResponse {
	return pkgdashboard.OverviewResponse{
		ID:        ov.ID,
		Name:      ov.Name,
		Order:     ov.Order,
		CreatedAt: ov.CreatedAt,
		UpdatedAt: ov.UpdatedAt,
		Widgets:   ToPublicWidgets(ov.Widgets),
	}
}

// ToPublicOverviews converts a slice of domain OverviewDashboards to public OverviewResponses.
func ToPublicOverviews(overviews []domain.OverviewDashboard) []pkgdashboard.OverviewResponse {
	if overviews == nil {
		return []pkgdashboard.OverviewResponse{}
	}
	res := make([]pkgdashboard.OverviewResponse, len(overviews))
	for i, o := range overviews {
		res[i] = ToPublicOverview(o)
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
