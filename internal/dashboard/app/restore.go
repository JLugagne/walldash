package app

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/restore"
)

// RestoreBackup replaces all levels, plans, placements, overviews and widgets
// with a validated backup snapshot, inside a single transaction: either the
// whole snapshot is replayed or nothing changes. Device placements are only
// replayed when input.IncludeDevices is set, so restores can skip entity
// bindings that may not exist on the target instance.
func (a *App) RestoreBackup(ctx context.Context, actor domain.Actor, input restore.RestoreInput) (restore.RestoreSummary, error) {
	var summary restore.RestoreSummary
	for _, lvl := range input.Levels {
		if err := lvl.Validate(); err != nil {
			return summary, err
		}
	}
	for _, plan := range input.Plans {
		if err := plan.Validate(); err != nil {
			return summary, err
		}
	}
	if input.IncludeDevices {
		for _, placement := range input.Placements {
			if err := placement.Validate(); err != nil {
				return summary, err
			}
		}
	}
	for _, overview := range input.Overviews {
		if err := overview.Validate(); err != nil {
			return summary, err
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
		existingOverviews, err := repos.Overviews.FindAllOverviews(ctx)
		if err != nil {
			return err
		}
		for _, overview := range existingOverviews {
			if err := repos.Widgets.DeleteWidgetsByDashboardID(ctx, overview.ID); err != nil {
				return err
			}
			if err := repos.Overviews.DeleteOverview(ctx, overview.ID); err != nil {
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
		for _, overview := range input.Overviews {
			dashboard := overview
			dashboard.Widgets = nil
			if _, err := repos.Overviews.CreateOverview(ctx, dashboard); err != nil {
				return err
			}
			summary.Overviews++
			for _, widget := range overview.Widgets {
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
