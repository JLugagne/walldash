package domain_test

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
)

func TestActorFromContext(t *testing.T) {
	t.Run("returns empty actor when context is nil", func(t *testing.T) {
		actor := domain.ActorFromContext(nil)
		assert.Equal(t, domain.Actor{}, actor)
	})

	t.Run("returns empty actor when context has no actor", func(t *testing.T) {
		actor := domain.ActorFromContext(context.Background())
		assert.Equal(t, domain.Actor{}, actor)
	})

	t.Run("returns actor when stored in context", func(t *testing.T) {
		expected := domain.Actor{
			UserID:   domain.UserID("usr_123"),
			TeamID:   domain.TeamID("team_456"),
			TenantID: domain.TenantID("tenant_789"),
		}
		ctx := domain.ContextWithActor(context.Background(), expected)
		actual := domain.ActorFromContext(ctx)
		assert.Equal(t, expected, actual)
	})
}
