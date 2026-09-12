package restore

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// RestoreInput is a validated backup snapshot ready to replay.
// Plans, Placements and Widgets reference their parent ids.
type RestoreInput struct {
	Levels         []domain.Level
	Plans          []domain.Plan
	Placements     []domain.DevicePlacement
	Dashboards     []domain.Dashboard
	IncludeDevices bool
}

// RestoreSummary counts replayed entities.
type RestoreSummary struct {
	Levels     int
	Plans      int
	Placements int
	Dashboards int
	Widgets    int
}

// RestoreCommands replays a backup snapshot, replacing current data.
type RestoreCommands interface {
	RestoreBackup(ctx context.Context, actor domain.Actor, input RestoreInput) (RestoreSummary, error)
}
