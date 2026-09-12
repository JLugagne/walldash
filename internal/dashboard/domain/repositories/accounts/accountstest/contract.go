package accountstest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAccountRepository is a function-based mock implementation of accounts.AccountRepository.
type MockAccountRepository struct {
	CreateFunc   func(ctx context.Context, account domain.Account) (domain.Account, error)
	FindByIDFunc func(ctx context.Context, id string) (domain.Account, error)
	FindAllFunc  func(ctx context.Context) ([]domain.Account, error)
	SetRoleFunc  func(ctx context.Context, id string, role domain.Role) (domain.Account, error)
	RevokeFunc   func(ctx context.Context, id string) (domain.Account, error)
	TouchFunc    func(ctx context.Context, id string) error
}

func (m *MockAccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	if m.CreateFunc == nil {
		panic("called not defined CreateFunc")
	}
	return m.CreateFunc(ctx, account)
}

func (m *MockAccountRepository) FindByID(ctx context.Context, id string) (domain.Account, error) {
	if m.FindByIDFunc == nil {
		panic("called not defined FindByIDFunc")
	}
	return m.FindByIDFunc(ctx, id)
}

func (m *MockAccountRepository) FindAll(ctx context.Context) ([]domain.Account, error) {
	if m.FindAllFunc == nil {
		panic("called not defined FindAllFunc")
	}
	return m.FindAllFunc(ctx)
}

func (m *MockAccountRepository) SetRole(ctx context.Context, id string, role domain.Role) (domain.Account, error) {
	if m.SetRoleFunc == nil {
		panic("called not defined SetRoleFunc")
	}
	return m.SetRoleFunc(ctx, id, role)
}

func (m *MockAccountRepository) Revoke(ctx context.Context, id string) (domain.Account, error) {
	if m.RevokeFunc == nil {
		panic("called not defined RevokeFunc")
	}
	return m.RevokeFunc(ctx, id)
}

func (m *MockAccountRepository) Touch(ctx context.Context, id string) error {
	if m.TouchFunc == nil {
		panic("called not defined TouchFunc")
	}
	return m.TouchFunc(ctx, id)
}

// AccountRepositoryContractTesting runs the shared conformance suite for an AccountRepository.
func AccountRepositoryContractTesting(t *testing.T, repo accounts.AccountRepository) {
	ctx := context.Background()

	t.Run("Contract: first account is created as owner", func(t *testing.T) {
		created, err := repo.Create(ctx, domain.Account{
			ID:    "acct-contract-owner",
			Label: "Living room tablet",
			Role:  domain.RoleDevice,
		})
		require.NoError(t, err)
		assert.Equal(t, "acct-contract-owner", created.ID)
		assert.Equal(t, domain.RoleOwner, created.Role)
		assert.Equal(t, domain.StatusActive, created.Status)
		assert.Equal(t, "Living room tablet", created.Label)
		assert.False(t, created.CreatedAt.IsZero())
		assert.Nil(t, created.LastSeen)
		assert.Nil(t, created.RevokedAt)
	})

	t.Run("Contract: subsequent account keeps the requested role", func(t *testing.T) {
		created, err := repo.Create(ctx, domain.Account{
			ID:    "acct-contract-device",
			Label: "Phone",
			Role:  domain.RoleDevice,
		})
		require.NoError(t, err)
		assert.Equal(t, domain.RoleDevice, created.Role)
		assert.Equal(t, domain.StatusActive, created.Status)
	})

	t.Run("Contract: Create defaults an empty role to device", func(t *testing.T) {
		created, err := repo.Create(ctx, domain.Account{
			ID:    "acct-contract-default",
			Label: "Kitchen display",
		})
		require.NoError(t, err)
		assert.Equal(t, domain.RoleDevice, created.Role)
	})

	t.Run("Contract: FindByID returns the stored account", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "acct-contract-owner")
		require.NoError(t, err)
		assert.Equal(t, "acct-contract-owner", found.ID)
		assert.Equal(t, domain.RoleOwner, found.Role)
		assert.Equal(t, domain.StatusActive, found.Status)
		assert.Equal(t, "Living room tablet", found.Label)
	})

	t.Run("Contract: FindByID reports ErrAccountNotFound for a missing account", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "acct-missing")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAccountNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindAll returns every account", func(t *testing.T) {
		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		ids := make(map[string]bool, len(all))
		for _, account := range all {
			ids[account.ID] = true
		}
		assert.True(t, ids["acct-contract-owner"])
		assert.True(t, ids["acct-contract-device"])
		assert.True(t, ids["acct-contract-default"])
	})

	t.Run("Contract: SetRole updates the role", func(t *testing.T) {
		updated, err := repo.SetRole(ctx, "acct-contract-device", domain.RoleAdmin)
		require.NoError(t, err)
		assert.Equal(t, domain.RoleAdmin, updated.Role)

		found, err := repo.FindByID(ctx, "acct-contract-device")
		require.NoError(t, err)
		assert.Equal(t, domain.RoleAdmin, found.Role)
	})

	t.Run("Contract: SetRole rejects an unknown role", func(t *testing.T) {
		_, err := repo.SetRole(ctx, "acct-contract-device", domain.Role("wizard"))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidRole)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: SetRole reports ErrAccountNotFound for a missing account", func(t *testing.T) {
		_, err := repo.SetRole(ctx, "acct-missing", domain.RoleDevice)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	})

	t.Run("Contract: Revoke marks the account revoked and is idempotent", func(t *testing.T) {
		revoked, err := repo.Revoke(ctx, "acct-contract-default")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusRevoked, revoked.Status)
		require.NotNil(t, revoked.RevokedAt)

		firstRevokedAt := *revoked.RevokedAt
		again, err := repo.Revoke(ctx, "acct-contract-default")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusRevoked, again.Status)
		require.NotNil(t, again.RevokedAt)
		assert.WithinDuration(t, firstRevokedAt, *again.RevokedAt, time.Second)
	})

	t.Run("Contract: Revoke reports ErrAccountNotFound for a missing account", func(t *testing.T) {
		_, err := repo.Revoke(ctx, "acct-missing")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	})

	t.Run("Contract: Touch updates the last seen timestamp", func(t *testing.T) {
		require.NoError(t, repo.Touch(ctx, "acct-contract-owner"))

		found, err := repo.FindByID(ctx, "acct-contract-owner")
		require.NoError(t, err)
		require.NotNil(t, found.LastSeen)
	})

	t.Run("Contract: Touch reports ErrAccountNotFound for a missing account", func(t *testing.T) {
		err := repo.Touch(ctx, "acct-missing")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	})
}
