package app

import (
	"context"
	"strings"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

// ExecuteAction validates that an incoming device action conforms to the Action Whitelist (ADR 0002)
// and delegates execution to the Home Assistant adapter, notifying any registered state broadcasters.
func (a *App) ExecuteAction(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
	log := logger.LoggerFromContext(ctx)

	// 1. Enforce strict Action Whitelist
	if err := cmd.Validate(); err != nil {
		log.WithError(err).
			WithField("entity_id", cmd.EntityID).
			WithField("action", cmd.Action).
			WithField("actor", actor.UserID).
			Warn("action rejected by whitelist")
		return err
	}

	parts := strings.Split(cmd.EntityID, ".")
	domainStr := parts[0]

	// 2. Delegate service call to Home Assistant repository
	if err := a.haRepo.CallService(ctx, domainStr, cmd.Action, cmd.EntityID); err != nil {
		log.WithError(err).
			WithField("entity_id", cmd.EntityID).
			WithField("action", cmd.Action).
			WithField("actor", actor.UserID).
			Error("failed to execute action on Home Assistant")
		return err
	}

	log.WithField("entity_id", cmd.EntityID).
		WithField("action", cmd.Action).
		WithField("actor", actor.UserID).
		Info("action executed successfully")

	// 3. If broadcaster is registered, fetch updated state and broadcast real-time update
	if a.broadcaster != nil {
		if updated, err := a.haRepo.GetState(ctx, cmd.EntityID); err == nil {
			a.broadcaster.BroadcastDevice(updated)
		} else {
			log.WithError(err).WithField("entity_id", cmd.EntityID).Warn("failed to fetch updated state for broadcast")
		}
	}

	return nil
}
