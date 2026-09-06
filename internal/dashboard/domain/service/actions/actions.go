package actions

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// ActionCommands defines operations for executing permitted actions against devices.
type ActionCommands interface {
	ExecuteAction(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error
}
