package app

import (
	"context"
	"testing"

	"github.com/JLugagne/egauth/revocation"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts/accountstest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSessionCloser struct {
	closed []string
}

func (f *fakeSessionCloser) CloseUser(subject string) {
	f.closed = append(f.closed, subject)
}

func TestSetRoleInvalidatesTargetSessions(t *testing.T) {
	targetID := uuid.NewString()
	target := domain.Account{ID: targetID, Role: domain.RoleAdmin, Status: domain.StatusActive}
	actor := domain.Account{ID: uuid.NewString(), Role: domain.RoleOwner, Status: domain.StatusActive}
	revoker := &stubRefreshRevoker{}
	bus := revocation.NewMemBus()
	var published []revocation.Revocation
	bus.Subscribe(revocation.TargetUser, revocation.HandlerFunc(func(_ context.Context, rev revocation.Revocation) error {
		published = append(published, rev)
		return nil
	}))
	repo := &accountstest.MockAccountRepository{
		FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
			require.Equal(t, targetID, id)
			return target, nil
		},
		SetRoleFunc: func(_ context.Context, id string, role domain.Role) (domain.Account, error) {
			require.Equal(t, targetID, id)
			target.Role = role
			return target, nil
		},
	}
	auth := NewAuth(repo, nil, nil, revoker, bus, false)
	closer := &fakeSessionCloser{}
	auth.SetSessionCloser(closer)

	updated, err := auth.SetRole(context.Background(), actor, targetID, domain.RoleDevice)
	require.NoError(t, err)
	require.Equal(t, domain.RoleDevice, updated.Role)
	require.True(t, revoker.called)
	require.Equal(t, targetID, revoker.userID.String())
	require.Equal(t, []string{targetID}, closer.closed)
	require.Len(t, published, 1)
	assert.Equal(t, targetID, published[0].TargetID)
	assert.Equal(t, revocation.ReasonLogoutEverywhere, published[0].Reason)
}

func TestRevokeDeviceClosesSessions(t *testing.T) {
	targetID := uuid.NewString()
	revoker := &stubRefreshRevoker{}
	repo := &accountstest.MockAccountRepository{
		FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
			return domain.Account{ID: id, Role: domain.RoleDevice, Status: domain.StatusActive}, nil
		},
		RevokeFunc: func(_ context.Context, id string) (domain.Account, error) {
			return domain.Account{ID: id, Status: domain.StatusRevoked}, nil
		},
	}
	auth := NewAuth(repo, nil, nil, revoker, nil, false)
	closer := &fakeSessionCloser{}
	auth.SetSessionCloser(closer)
	actor := domain.Account{ID: uuid.NewString(), Role: domain.RoleOwner, Status: domain.StatusActive}

	acct, err := auth.RevokeDevice(context.Background(), actor, targetID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusRevoked, acct.Status)
	require.True(t, revoker.called)
	require.Equal(t, []string{targetID}, closer.closed)
}

func TestSetRoleRejectsNonPrivilegedActor(t *testing.T) {
	lookedUp := false
	repo := &accountstest.MockAccountRepository{
		FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
			lookedUp = true
			return domain.Account{ID: id, Role: domain.RoleDevice, Status: domain.StatusActive}, nil
		},
	}
	auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)
	actor := domain.Account{ID: uuid.NewString(), Role: domain.RoleDevice, Status: domain.StatusActive}

	_, err := auth.SetRole(context.Background(), actor, uuid.NewString(), domain.RoleAdmin)
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.False(t, lookedUp)
}

func TestCanManageAccount(t *testing.T) {
	owner := domain.Account{Role: domain.RoleOwner}
	admin := domain.Account{Role: domain.RoleAdmin}
	device := domain.Account{Role: domain.RoleDevice}

	require.NoError(t, canManageAccount(owner, owner))
	require.NoError(t, canManageAccount(owner, admin))
	require.NoError(t, canManageAccount(owner, device))
	require.NoError(t, canManageAccount(admin, admin))
	require.NoError(t, canManageAccount(admin, device))
	require.ErrorIs(t, canManageAccount(admin, owner), domain.ErrForbidden)
	require.ErrorIs(t, canManageAccount(device, owner), domain.ErrForbidden)
	require.ErrorIs(t, canManageAccount(device, device), domain.ErrForbidden)
}

// TestSetRolePromotionKeepsSession pins the rule that promoting a device does not sign it out:
// only privilege reduction ends the session. Nothing is revoked either — not the refresh family,
// not the live socket, and not the issued access token. A promotion grants privileges, so the
// target's older token carries only a smaller scope set; the client re-issues it by refreshing
// once when a setup-scoped route answers 403. Revoking it instead (as an "access-token-only"
// revocation cutoff) is what used to drop promoted devices: the tracker rejects any token whose
// second-truncated `iat` is at or before the sub-second cutoff, so the freshly re-issued token was
// rejected too, and the client's single 401 retry landed on the login screen.
func TestSetRolePromotionKeepsSession(t *testing.T) {
	targetID := uuid.NewString()
	target := domain.Account{ID: targetID, Role: domain.RoleDevice, Status: domain.StatusActive}
	actor := domain.Account{ID: uuid.NewString(), Role: domain.RoleOwner, Status: domain.StatusActive}
	revoker := &stubRefreshRevoker{}
	bus := revocation.NewMemBus()
	var published []revocation.Revocation
	bus.Subscribe(revocation.TargetUser, revocation.HandlerFunc(func(_ context.Context, rev revocation.Revocation) error {
		published = append(published, rev)
		return nil
	}))
	repo := &accountstest.MockAccountRepository{
		FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
			require.Equal(t, targetID, id)
			return target, nil
		},
		SetRoleFunc: func(_ context.Context, id string, role domain.Role) (domain.Account, error) {
			target.Role = role
			return target, nil
		},
	}
	auth := NewAuth(repo, nil, nil, revoker, bus, false)
	closer := &fakeSessionCloser{}
	auth.SetSessionCloser(closer)

	updated, err := auth.SetRole(context.Background(), actor, targetID, domain.RoleAdmin)
	require.NoError(t, err)
	require.Equal(t, domain.RoleAdmin, updated.Role)
	require.False(t, revoker.called, "promotion must not revoke the refresh family")
	require.Empty(t, closer.closed, "promotion must not close live sockets")
	require.Empty(t, published, "promotion must not publish any access-token revocation")
}
