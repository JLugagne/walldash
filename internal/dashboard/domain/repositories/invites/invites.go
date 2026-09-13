package invites

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// Repository defines persistence operations for device invitations.
type Repository interface {
	// Create stores a new invitation and returns it.
	Create(ctx context.Context, invite domain.Invite) (domain.Invite, error)
	// FindBySelector loads an invitation by its selector, regardless of state.
	FindBySelector(ctx context.Context, selector string) (domain.Invite, error)
	// FindAll returns every invitation, newest first.
	FindAll(ctx context.Context) ([]domain.Invite, error)
	// Consume atomically marks an active invitation as used by consumedBy. It reports
	// ErrInviteConsumed, ErrInviteExpired or ErrInviteRevoked when the invitation is no
	// longer usable, and ErrInviteNotFound for an unknown selector.
	Consume(ctx context.Context, selector string, consumedBy string) (domain.Invite, error)
	// Revoke marks an active invitation as revoked and returns it.
	Revoke(ctx context.Context, selector string) (domain.Invite, error)
	// DeleteExpired removes invitations past their expiry and returns how many were deleted.
	DeleteExpired(ctx context.Context) (int64, error)
}
