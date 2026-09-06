package dashboard

// ExportResponse contains a full snapshot of every level, its plan with
// walls/zones, device placements, and all overview dashboards.
type ExportResponse struct {
	Version   string             `json:"version"`
	Levels    []ExportedLevel    `json:"levels"`
	Overviews []OverviewResponse `json:"overviews"`
}

// ExportedLevel bundles a level, its 2D plan, and its device placements.
type ExportedLevel struct {
	Level      LevelResponse             `json:"level"`
	Plan       *PlanResponse             `json:"plan,omitempty"`
	Placements []DevicePlacementResponse `json:"placements"`
}
