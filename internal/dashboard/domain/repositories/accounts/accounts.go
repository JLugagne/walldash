package accounts

import (
	"context"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
)

// AccountRepository defines repository operations for device auth accounts.
type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (domain.Account, error)
	FindByID(ctx context.Context, id string) (domain.Account, error)
	FindAll(ctx context.Context) ([]domain.Account, error)
	SetRole(ctx context.Context, id string, role domain.Role) (domain.Account, error)
	UpdateLabel(ctx context.Context, id string, label string) error
	Revoke(ctx context.Context, id string) (domain.Account, error)
	Touch(ctx context.Context, id string) error
}
