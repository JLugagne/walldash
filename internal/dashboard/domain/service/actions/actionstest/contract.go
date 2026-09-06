package actionstest

import (
	"context"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/service/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockActionCommands is a function-based mock implementation of actions.ActionCommands.
type MockActionCommands struct {
	ExecuteActionFunc func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error
}

func (m *MockActionCommands) ExecuteAction(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
	if m.ExecuteActionFunc == nil {
		panic("called not defined ExecuteActionFunc")
	}
	return m.ExecuteActionFunc(ctx, actor, cmd)
}

// ActionCommandsContractTesting runs all contract tests for an ActionCommands implementation.
func ActionCommandsContractTesting(t *testing.T, svc actions.ActionCommands, validEntityID string) {
	ctx := context.Background()
	actor := domain.Actor{UserID: "test-user"}

	t.Run("Contract: ExecuteAction succeeds for valid whitelisted command", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: validEntityID,
			Action:   domain.ActionToggle,
		}
		err := svc.ExecuteAction(ctx, actor, cmd)
		require.NoError(t, err)
	})

	t.Run("Contract: ExecuteAction fails with ErrActionNotAllowed for disallowed action", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: validEntityID,
			Action:   "dimmer_up",
		}
		err := svc.ExecuteAction(ctx, actor, cmd)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: ExecuteAction fails with ErrActionNotAllowed for disallowed domain", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "sensor.temperature_salon",
			Action:   domain.ActionToggle,
		}
		err := svc.ExecuteAction(ctx, actor, cmd)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.True(t, domain.IsDomainError(err))
	})
}
