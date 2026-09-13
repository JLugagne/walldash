package app

import (
	"context"
	"errors"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
)

// RestoreBackup replaces all levels, plans, placements, dashboards and widgets
// with a validated backup snapshot, inside a single transaction: either the
// whole snapshot is replayed or nothing changes. Device placements are only
// replayed when input.IncludeDevices is set, so restores can skip entity
// bindings that may not exist on the target instance.
func (a *App) RestoreBackup(ctx context.Context, actor domain.Actor, input restore.RestoreInput) (restore.RestoreSummary, error) {
	var summary restore.RestoreSummary

	// A snapshot is a graph, so every reference inside it must resolve against the snapshot
	// itself. Validating the items one by one is not enough: a plan could name a level that is
	// absent (and then be served by GET /api/levels/{id}/plan for a level that does not exist),
	// and a widget could be listed inside one dashboard while naming another, escaping the grid
	// validation that the ordinary widget routes apply to the dashboard it actually lands on.
	//
	// Identifiers and timestamps are deliberately replayed as sent: a backup round-trips the
	// export, and the snapshot's own ids are what hold the graph together.
	levelIDs := make(map[string]bool, len(input.Levels))
	for _, lvl := range input.Levels {
		if err := lvl.Validate(); err != nil {
			return summary, err
		}
		if levelIDs[lvl.ID] {
			return summary, errors.Join(domain.ErrInvalidLevel, errors.New("duplicate level id in snapshot: "+lvl.ID))
		}
		levelIDs[lvl.ID] = true
	}

	for _, plan := range input.Plans {
		if err := plan.Validate(); err != nil {
			return summary, err
		}
		if !levelIDs[plan.LevelID] {
			return summary, errors.Join(domain.ErrInvalidPlan,
				errors.New("plan references a level absent from the snapshot: "+plan.LevelID))
		}
	}

	if input.IncludeDevices {
		for _, placement := range input.Placements {
			if err := placement.Validate(); err != nil {
				return summary, err
			}
			if !levelIDs[placement.LevelID] {
				return summary, errors.Join(domain.ErrInvalidPlacement,
					errors.New("placement references a level absent from the snapshot: "+placement.LevelID))
			}
		}
	}

	dashboardIDs := make(map[string]bool, len(input.Dashboards))
	for _, dashboard := range input.Dashboards {
		if err := dashboard.Validate(); err != nil {
			return summary, err
		}
		if dashboardIDs[dashboard.ID] {
			return summary, errors.Join(domain.ErrInvalidDashboard,
				errors.New("duplicate dashboard id in snapshot: "+dashboard.ID))
		}
		dashboardIDs[dashboard.ID] = true

		// Dashboard.Validate checks each widget against the dashboard it is nested in, so the
		// nesting has to be the truth before that check means anything.
		for _, widget := range dashboard.Widgets {
			if widget.DashboardID != dashboard.ID {
				return summary, errors.Join(domain.ErrInvalidWidget,
					errors.New("widget "+widget.ID+" is listed under dashboard "+dashboard.ID+
						" but names dashboard "+widget.DashboardID))
			}
		}
	}
	if err := a.uow.Do(ctx, func(repos uow.Repositories) error {
		existingLevels, err := repos.Levels.FindAll(ctx)
		if err != nil {
			return err
		}
		for _, lvl := range existingLevels {
			placements, err := repos.Placements.FindPlacementsByLevelID(ctx, lvl.ID)
			if err != nil {
				return err
			}
			for _, placement := range placements {
				if err := repos.Placements.DeletePlacement(ctx, placement.ID); err != nil {
					return err
				}
			}
			if err := repos.Levels.Delete(ctx, lvl.ID); err != nil {
				return err
			}
		}
		existingDashboards, err := repos.Dashboards.FindAllDashboards(ctx)
		if err != nil {
			return err
		}
		for _, dashboard := range existingDashboards {
			if err := repos.Widgets.DeleteWidgetsByDashboardID(ctx, dashboard.ID); err != nil {
				return err
			}
			if err := repos.Dashboards.DeleteDashboard(ctx, dashboard.ID); err != nil {
				return err
			}
		}
		for _, lvl := range input.Levels {
			if _, err := repos.Levels.Create(ctx, lvl); err != nil {
				return err
			}
			summary.Levels++
		}
		for _, plan := range input.Plans {
			if _, err := repos.Plans.Save(ctx, plan); err != nil {
				return err
			}
			summary.Plans++
		}
		if input.IncludeDevices {
			for _, placement := range input.Placements {
				if _, err := repos.Placements.SavePlacement(ctx, placement); err != nil {
					return err
				}
				summary.Placements++
			}
		}
		for _, dashboard := range input.Dashboards {
			parent := dashboard
			parent.Widgets = nil
			if _, err := repos.Dashboards.CreateDashboard(ctx, parent); err != nil {
				return err
			}
			summary.Dashboards++
			for _, widget := range dashboard.Widgets {
				if _, err := repos.Widgets.CreateWidget(ctx, widget); err != nil {
					return err
				}
				summary.Widgets++
			}
		}
		return nil
	}); err != nil {
		return restore.RestoreSummary{}, err
	}
	return summary, nil
}
