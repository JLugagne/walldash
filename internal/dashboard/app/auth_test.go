package app

import (
	"context"
	"strings"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts/accountstest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type stubRefreshRevoker struct {
	called bool
	userID uuid.UUID
	err    error
}

func (s *stubRefreshRevoker) RevokeAllRefreshTokensForUser(_ context.Context, _ string, userID uuid.UUID) error {
	s.called = true
	s.userID = userID
	return s.err
}

func TestAuthListDevices(t *testing.T) {
	repo := &accountstest.MockAccountRepository{
		FindAllFunc: func(_ context.Context) ([]domain.Account, error) {
			return []domain.Account{{ID: "d1", Role: domain.RoleOwner, Status: domain.StatusActive}}, nil
		},
	}
	auth := NewAuth(nil, repo, nil, &stubRefreshRevoker{})

	devices, err := auth.ListDevices(context.Background())
	require.NoError(t, err)
	require.Len(t, devices, 1)
	require.Equal(t, "d1", devices[0].ID)
}

func TestAuthRevokeDevice(t *testing.T) {
	id := uuid.NewString()
	revoker := &stubRefreshRevoker{}
	var revoked string
	repo := &accountstest.MockAccountRepository{
		RevokeFunc: func(_ context.Context, accountID string) (domain.Account, error) {
			revoked = accountID
			return domain.Account{ID: accountID, Status: domain.StatusRevoked, Role: domain.RoleDevice}, nil
		},
	}
	auth := NewAuth(nil, repo, nil, revoker)

	acct, err := auth.RevokeDevice(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, domain.StatusRevoked, acct.Status)
	require.Equal(t, id, revoked)
	require.True(t, revoker.called)
	require.Equal(t, id, revoker.userID.String())
}

func TestAuthRevokeDeviceNotFound(t *testing.T) {
	revoker := &stubRefreshRevoker{}
	repo := &accountstest.MockAccountRepository{
		RevokeFunc: func(context.Context, string) (domain.Account, error) {
			return domain.Account{}, domain.ErrAccountNotFound
		},
	}
	auth := NewAuth(nil, repo, nil, revoker)

	_, err := auth.RevokeDevice(context.Background(), uuid.NewString())
	require.ErrorIs(t, err, domain.ErrAccountNotFound)
	require.False(t, revoker.called)
}

func TestAuthSetRoleRules(t *testing.T) {
	owner := domain.Account{ID: uuid.NewString(), Role: domain.RoleOwner, Status: domain.StatusActive}
	admin := domain.Account{ID: uuid.NewString(), Role: domain.RoleAdmin, Status: domain.StatusActive}
	deviceTarget := domain.Account{ID: uuid.NewString(), Role: domain.RoleDevice, Status: domain.StatusActive}
	ownerTarget := domain.Account{ID: uuid.NewString(), Role: domain.RoleOwner, Status: domain.StatusActive}

	newRepo := func(target domain.Account) *accountstest.MockAccountRepository {
		return &accountstest.MockAccountRepository{
			FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
				return target, nil
			},
			SetRoleFunc: func(_ context.Context, id string, role domain.Role) (domain.Account, error) {
				target.Role = role
				return target, nil
			},
		}
	}

	t.Run("owner may grant owner", func(t *testing.T) {
		auth := NewAuth(nil, newRepo(deviceTarget), nil, &stubRefreshRevoker{})
		updated, err := auth.SetRole(context.Background(), owner, deviceTarget.ID, domain.RoleOwner)
		require.NoError(t, err)
		require.Equal(t, domain.RoleOwner, updated.Role)
	})

	t.Run("admin may set admin or device", func(t *testing.T) {
		auth := NewAuth(nil, newRepo(deviceTarget), nil, &stubRefreshRevoker{})
		updated, err := auth.SetRole(context.Background(), admin, deviceTarget.ID, domain.RoleAdmin)
		require.NoError(t, err)
		require.Equal(t, domain.RoleAdmin, updated.Role)
	})

	t.Run("admin may not grant owner", func(t *testing.T) {
		auth := NewAuth(nil, newRepo(deviceTarget), nil, &stubRefreshRevoker{})
		_, err := auth.SetRole(context.Background(), admin, deviceTarget.ID, domain.RoleOwner)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("admin may not change an owner", func(t *testing.T) {
		auth := NewAuth(nil, newRepo(ownerTarget), nil, &stubRefreshRevoker{})
		_, err := auth.SetRole(context.Background(), admin, ownerTarget.ID, domain.RoleDevice)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("invalid role is rejected", func(t *testing.T) {
		auth := NewAuth(nil, newRepo(deviceTarget), nil, &stubRefreshRevoker{})
		_, err := auth.SetRole(context.Background(), owner, deviceTarget.ID, domain.Role("root"))
		require.ErrorIs(t, err, domain.ErrInvalidRole)
	})
}

func TestAuthSetLabel(t *testing.T) {
	target := domain.Account{ID: uuid.NewString(), Role: domain.RoleDevice, Status: domain.StatusActive, Label: "old"}
	repo := &accountstest.MockAccountRepository{
		UpdateLabelFunc: func(_ context.Context, id string, label string) error {
			if id != target.ID {
				return domain.ErrAccountNotFound
			}
			target.Label = label
			return nil
		},
	}
	auth := NewAuth(nil, repo, nil, &stubRefreshRevoker{})

	t.Run("trims and stores a valid label", func(t *testing.T) {
		require.NoError(t, auth.SetLabel(context.Background(), target.ID, "  Kitchen tablet  "))
		require.Equal(t, "Kitchen tablet", target.Label)
	})

	t.Run("rejects an empty label", func(t *testing.T) {
		require.ErrorIs(t, auth.SetLabel(context.Background(), target.ID, "   "), domain.ErrInvalidLabel)
	})

	t.Run("rejects a label longer than 64 characters", func(t *testing.T) {
		require.ErrorIs(t, auth.SetLabel(context.Background(), target.ID, strings.Repeat("a", 65)), domain.ErrInvalidLabel)
	})

	t.Run("propagates an unknown account", func(t *testing.T) {
		require.ErrorIs(t, auth.SetLabel(context.Background(), "missing", "ok"), domain.ErrAccountNotFound)
	})
}
