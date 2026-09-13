package app

import (
	"context"
	"errors"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts/accountstest"
	"github.com/stretchr/testify/require"
)

// testManager is the active owner account used as the caller for privileged use-cases: these
// use-cases re-load the actor from the store, so the mock has to know it.
var testManager = domain.Account{ID: "owner-1", Role: domain.RoleOwner, Status: domain.StatusActive}

// withManager teaches a mock repository about testManager.
func withManager(repo *accountstest.MockAccountRepository) *accountstest.MockAccountRepository {
	repo.FindByIDFunc = func(_ context.Context, id string) (domain.Account, error) {
		if id == testManager.ID {
			return testManager, nil
		}
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return repo
}

func TestAuthConnectPendingAndApprove(t *testing.T) {
	var created domain.Account
	repo := withManager(&accountstest.MockAccountRepository{
		CountFunc: func(_ context.Context) (int, error) { return 1, nil },
		CreateFunc: func(_ context.Context, account domain.Account) (domain.Account, error) {
			created = account
			account.Status = domain.StatusActive
			return account, nil
		},
	})
	auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)

	outcome, err := auth.Connect(context.Background(), "Kitchen tablet", "10.0.0.1")
	require.NoError(t, err)
	require.Equal(t, EnrollmentPending, outcome.Status)
	require.NotEmpty(t, outcome.Pending.PendingID)

	acct, err := auth.ApprovePending(context.Background(), testManager, outcome.Pending.PendingID)
	require.NoError(t, err)
	require.Equal(t, domain.RoleDevice, acct.Role)
	require.Equal(t, domain.RoleDevice, created.Role)
	require.Equal(t, "Kitchen tablet", created.Label)
}

func TestAuthRedeemUnknownPendingIsInvalid(t *testing.T) {
	auth := NewAuth(&accountstest.MockAccountRepository{}, nil, nil, &stubRefreshRevoker{}, nil, false)
	_, _, err := auth.Redeem(context.Background(), "does-not-exist")
	require.ErrorIs(t, err, ErrInvalidEnrollment)
}

func TestAuthDenyPendingDropsEnrollment(t *testing.T) {
	repo := withManager(&accountstest.MockAccountRepository{
		CountFunc: func(_ context.Context) (int, error) { return 1, nil },
	})
	auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)

	outcome, err := auth.Connect(context.Background(), "Phone", "")
	require.NoError(t, err)
	require.NoError(t, auth.DenyPending(context.Background(), testManager, outcome.Pending.PendingID))

	_, _, err = auth.Redeem(context.Background(), outcome.Pending.PendingID)
	require.ErrorIs(t, err, ErrInvalidEnrollment)
}

func TestAuthCreateInviteReservesOwnerRole(t *testing.T) {
	auth := NewAuth(withManager(&accountstest.MockAccountRepository{}), nil, nil, &stubRefreshRevoker{}, nil, false)
	_, _, err := auth.CreateInvite(context.Background(), testManager, domain.RoleOwner)
	require.ErrorIs(t, err, domain.ErrInvalidInviteRole)
}

func TestAuthRedeemInviteRejectsMalformedToken(t *testing.T) {
	auth := NewAuth(&accountstest.MockAccountRepository{}, nil, nil, &stubRefreshRevoker{}, nil, false)
	_, _, err := auth.RedeemInvite(context.Background(), "not-a-token", "tablet")
	require.ErrorIs(t, err, ErrInvalidInvite)
}

// TestPrivilegedUseCasesRejectStaleActors covers the authorization gate added after the audit
// finding "Access-token revocation does not survive a restart": every privileged use-case
// re-loads the caller from the store, so a token whose account was demoted or revoked since it
// was issued can no longer mint invitations, approve enrollments or revoke invitations.
func TestPrivilegedUseCasesRejectStaleActors(t *testing.T) {
	cases := map[string]struct {
		stored  *domain.Account
		findErr error
	}{
		"demoted to device": {stored: &domain.Account{ID: "ex-admin", Role: domain.RoleDevice, Status: domain.StatusActive}},
		"revoked":           {stored: &domain.Account{ID: "ex-admin", Role: domain.RoleAdmin, Status: domain.StatusRevoked}},
		"deleted":           {findErr: domain.ErrAccountNotFound},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &accountstest.MockAccountRepository{
				FindByIDFunc: func(_ context.Context, id string) (domain.Account, error) {
					if tc.findErr != nil || tc.stored == nil || id != tc.stored.ID {
						return domain.Account{}, errors.Join(domain.ErrAccountNotFound, tc.findErr)
					}
					return *tc.stored, nil
				},
			}
			auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)
			stale := domain.Account{ID: "ex-admin", Role: domain.RoleAdmin, Status: domain.StatusActive}

			_, _, err := auth.CreateInvite(context.Background(), stale, domain.RoleAdmin)
			require.ErrorIs(t, err, domain.ErrForbidden)

			_, err = auth.ApprovePending(context.Background(), stale, "pending-1")
			require.ErrorIs(t, err, domain.ErrForbidden)

			require.ErrorIs(t, auth.DenyPending(context.Background(), stale, "pending-1"), domain.ErrForbidden)

			_, err = auth.RevokeInvite(context.Background(), stale, "selector-1")
			require.ErrorIs(t, err, domain.ErrForbidden)
		})
	}
}
