package queries

import (
	"net/http"

	svcdevices "github.com/JLugagne/walldash/internal/dashboard/domain/service/devices"
	svclevels "github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	svcoverviews "github.com/JLugagne/walldash/internal/dashboard/domain/service/overviews"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ExportHandler bundles the three query interfaces required to produce a
// full data export.
type ExportHandler struct {
	controller      *inbound.Controller
	levelQueries    svclevels.LevelQueries
	deviceQueries   svcdevices.DeviceQueries
	overviewQueries svcoverviews.OverviewQueries
}

// NewExportHandler constructs an ExportHandler.
func NewExportHandler(controller *inbound.Controller, levelQueries svclevels.LevelQueries, deviceQueries svcdevices.DeviceQueries, overviewQueries svcoverviews.OverviewQueries) *ExportHandler {
	return &ExportHandler{
		controller:      controller,
		levelQueries:    levelQueries,
		deviceQueries:   deviceQueries,
		overviewQueries: overviewQueries,
	}
}

// Export handles GET /api/export.
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	levels, err := h.levelQueries.ListLevels(r.Context())
	if err != nil {
		h.controller.SendError(w, r, err)
		return
	}

	exportedLevels := make([]pkgdashboard.ExportedLevel, 0, len(levels))
	for _, lvl := range levels {
		exported := pkgdashboard.ExportedLevel{
			Level:      converters.ToPublicLevel(lvl),
			Placements: []pkgdashboard.DevicePlacementResponse{},
		}

		plan, planErr := h.levelQueries.GetPlan(r.Context(), lvl.ID)
		if planErr == nil {
			pub := converters.ToPublicPlan(plan)
			exported.Plan = &pub
		}

		placements, placementErr := h.deviceQueries.ListPlacements(r.Context(), lvl.ID)
		if placementErr == nil {
			exported.Placements = converters.ToPublicPlacements(placements)
		}

		exportedLevels = append(exportedLevels, exported)
	}

	overviewsPublic := []pkgdashboard.OverviewResponse{}
	if overviews, err := h.overviewQueries.ListOverviews(r.Context()); err == nil {
		overviewsPublic = converters.ToPublicOverviews(overviews)
	}

	response := pkgdashboard.ExportResponse{
		Version:   "1",
		Levels:    exportedLevels,
		Overviews: overviewsPublic,
	}

	h.controller.SendSuccess(w, r, response)
}
