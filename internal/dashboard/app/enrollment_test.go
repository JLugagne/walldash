package app

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/accounts/accountstest"
	"github.com/stretchr/testify/require"
)

func TestAuthConnectPendingAndApprove(t *testing.T) {
	var created domain.Account
	repo := &accountstest.MockAccountRepository{
		CountFunc: func(_ context.Context) (int, error) { return 1, nil },
		CreateFunc: func(_ context.Context, account domain.Account) (domain.Account, error) {
			created = account
			account.Status = domain.StatusActive
			return account, nil
		},
	}
	auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)

	outcome, err := auth.Connect(context.Background(), "Kitchen tablet", "10.0.0.1")
	require.NoError(t, err)
	require.Equal(t, EnrollmentPending, outcome.Status)
	require.NotEmpty(t, outcome.Pending.PendingID)

	acct, err := auth.ApprovePending(context.Background(), outcome.Pending.PendingID)
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
	repo := &accountstest.MockAccountRepository{
		CountFunc: func(_ context.Context) (int, error) { return 1, nil },
	}
	auth := NewAuth(repo, nil, nil, &stubRefreshRevoker{}, nil, false)

	outcome, err := auth.Connect(context.Background(), "Phone", "")
	require.NoError(t, err)
	require.NoError(t, auth.DenyPending(context.Background(), outcome.Pending.PendingID))

	_, _, err = auth.Redeem(context.Background(), outcome.Pending.PendingID)
	require.ErrorIs(t, err, ErrInvalidEnrollment)
}

func TestAuthCreateInviteReservesOwnerRole(t *testing.T) {
	auth := NewAuth(&accountstest.MockAccountRepository{}, nil, nil, &stubRefreshRevoker{}, nil, false)
	_, _, err := auth.CreateInvite(context.Background(), "owner-1", domain.RoleOwner)
	require.ErrorIs(t, err, domain.ErrInvalidInviteRole)
}

func TestAuthRedeemInviteRejectsMalformedToken(t *testing.T) {
	auth := NewAuth(&accountstest.MockAccountRepository{}, nil, nil, &stubRefreshRevoker{}, nil, false)
	_, _, err := auth.RedeemInvite(context.Background(), "not-a-token", "tablet")
	require.ErrorIs(t, err, ErrInvalidInvite)
}
