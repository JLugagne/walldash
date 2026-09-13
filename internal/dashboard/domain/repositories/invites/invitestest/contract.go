package invitestest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/invites"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockInviteRepository is a function-based mock implementation of invites.Repository.
type MockInviteRepository struct {
	CreateFunc         func(ctx context.Context, invite domain.Invite) (domain.Invite, error)
	FindBySelectorFunc func(ctx context.Context, selector string) (domain.Invite, error)
	FindAllFunc        func(ctx context.Context) ([]domain.Invite, error)
	ConsumeFunc        func(ctx context.Context, selector string, consumedBy string) (domain.Invite, error)
	RevokeFunc         func(ctx context.Context, selector string) (domain.Invite, error)
	DeleteExpiredFunc  func(ctx context.Context) (int64, error)
}

func (m *MockInviteRepository) Create(ctx context.Context, invite domain.Invite) (domain.Invite, error) {
	if m.CreateFunc == nil {
		panic("called not defined CreateFunc")
	}
	return m.CreateFunc(ctx, invite)
}

func (m *MockInviteRepository) FindBySelector(ctx context.Context, selector string) (domain.Invite, error) {
	if m.FindBySelectorFunc == nil {
		panic("called not defined FindBySelectorFunc")
	}
	return m.FindBySelectorFunc(ctx, selector)
}

func (m *MockInviteRepository) FindAll(ctx context.Context) ([]domain.Invite, error) {
	if m.FindAllFunc == nil {
		panic("called not defined FindAllFunc")
	}
	return m.FindAllFunc(ctx)
}

func (m *MockInviteRepository) Consume(ctx context.Context, selector string, consumedBy string) (domain.Invite, error) {
	if m.ConsumeFunc == nil {
		panic("called not defined ConsumeFunc")
	}
	return m.ConsumeFunc(ctx, selector, consumedBy)
}

func (m *MockInviteRepository) Revoke(ctx context.Context, selector string) (domain.Invite, error) {
	if m.RevokeFunc == nil {
		panic("called not defined RevokeFunc")
	}
	return m.RevokeFunc(ctx, selector)
}

func (m *MockInviteRepository) DeleteExpired(ctx context.Context) (int64, error) {
	if m.DeleteExpiredFunc == nil {
		panic("called not defined DeleteExpiredFunc")
	}
	return m.DeleteExpiredFunc(ctx)
}

// InviteRepositoryContractTesting runs the shared conformance suite for an invites.Repository.
func InviteRepositoryContractTesting(t *testing.T, repo invites.Repository) {
	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("Contract: Create stores an active invitation", func(t *testing.T) {
		created, err := repo.Create(ctx, domain.Invite{
			Selector:     "sel-active",
			VerifierHash: "hash-active",
			Role:         domain.RoleDevice,
			CreatedBy:    "owner-1",
			ExpiresAt:    now.Add(15 * time.Minute),
		})
		require.NoError(t, err)
		assert.Equal(t, "sel-active", created.Selector)
		assert.Equal(t, "hash-active", created.VerifierHash)
		assert.Equal(t, domain.RoleDevice, created.Role)
		assert.Equal(t, "owner-1", created.CreatedBy)
		assert.Nil(t, created.ConsumedAt)
		assert.Nil(t, created.RevokedAt)
	})

	t.Run("Contract: FindBySelector reports ErrInviteNotFound for a missing invite", func(t *testing.T) {
		_, err := repo.FindBySelector(ctx, "sel-missing")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInviteNotFound)
	})

	t.Run("Contract: Consume is single-use", func(t *testing.T) {
		_, err := repo.Create(ctx, domain.Invite{
			Selector:     "sel-once",
			VerifierHash: "hash-once",
			Role:         domain.RoleAdmin,
			CreatedBy:    "owner-1",
			ExpiresAt:    now.Add(15 * time.Minute),
		})
		require.NoError(t, err)

		consumed, err := repo.Consume(ctx, "sel-once", "device-1")
		require.NoError(t, err)
		require.NotNil(t, consumed.ConsumedAt)
		assert.Equal(t, "device-1", consumed.ConsumedBy)

		_, err = repo.Consume(ctx, "sel-once", "device-2")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInviteConsumed)
	})

	t.Run("Contract: Consume rejects an expired invitation", func(t *testing.T) {
		_, err := repo.Create(ctx, domain.Invite{
			Selector:     "sel-expired",
			VerifierHash: "hash-expired",
			Role:         domain.RoleDevice,
			CreatedBy:    "owner-1",
			ExpiresAt:    now.Add(-time.Minute),
		})
		require.NoError(t, err)
		_, err = repo.Consume(ctx, "sel-expired", "device-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInviteExpired)
	})

	t.Run("Contract: Revoke prevents consumption", func(t *testing.T) {
		_, err := repo.Create(ctx, domain.Invite{
			Selector:     "sel-revoked",
			VerifierHash: "hash-revoked",
			Role:         domain.RoleDevice,
			CreatedBy:    "owner-1",
			ExpiresAt:    now.Add(15 * time.Minute),
		})
		require.NoError(t, err)
		revoked, err := repo.Revoke(ctx, "sel-revoked")
		require.NoError(t, err)
		require.NotNil(t, revoked.RevokedAt)
		_, err = repo.Consume(ctx, "sel-revoked", "device-1")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInviteRevoked)
	})

	t.Run("Contract: FindAll returns every invitation", func(t *testing.T) {
		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		selectors := make(map[string]bool, len(all))
		for _, inv := range all {
			selectors[inv.Selector] = true
		}
		assert.True(t, selectors["sel-active"])
		assert.True(t, selectors["sel-once"])
	})

	t.Run("Contract: DeleteExpired removes only expired invitations", func(t *testing.T) {
		deleted, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
		_, err = repo.FindBySelector(ctx, "sel-expired")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInviteNotFound)
	})
}
