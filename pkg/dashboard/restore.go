package dashboard

import "errors"

var ErrInvalidRestoreRequest = errors.New("invalid restore request")

// RestoreRequest replays a previously exported snapshot (see ExportResponse).
// Device placements are only restored when IncludeDevices is set.
type RestoreRequest struct {
	Version        string             `json:"version" validate:"required"`
	Levels         []ExportedLevel    `json:"levels" validate:"dive"`
	Overviews      []OverviewResponse `json:"overviews" validate:"dive"`
	IncludeDevices bool               `json:"include_devices"`
}

// RestoreResponse summarizes what a restore replayed.
type RestoreResponse struct {
	Levels     int `json:"levels"`
	Plans      int `json:"plans"`
	Placements int `json:"placements"`
	Overviews  int `json:"overviews"`
	Widgets    int `json:"widgets"`
}
