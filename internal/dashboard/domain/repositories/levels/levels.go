package levels

import (
	"context"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
)

// LevelRepository defines repository operations for Level domain entities.
type LevelRepository interface {
	Create(ctx context.Context, level domain.Level) (domain.Level, error)
	FindByID(ctx context.Context, id string) (domain.Level, error)
	FindAll(ctx context.Context) ([]domain.Level, error)
	Update(ctx context.Context, level domain.Level) (domain.Level, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, orderedIDs []string) error
}
