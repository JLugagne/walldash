package levelstest

import (
	"context"
	"testing"
	"time"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLevelRepository is a function-based mock implementation of levels.LevelRepository.
type MockLevelRepository struct {
	CreateFunc   func(ctx context.Context, level domain.Level) (domain.Level, error)
	FindByIDFunc func(ctx context.Context, id string) (domain.Level, error)
	FindAllFunc  func(ctx context.Context) ([]domain.Level, error)
	UpdateFunc   func(ctx context.Context, level domain.Level) (domain.Level, error)
	DeleteFunc   func(ctx context.Context, id string) error
	ReorderFunc  func(ctx context.Context, orderedIDs []string) error
}

func (m *MockLevelRepository) Create(ctx context.Context, level domain.Level) (domain.Level, error) {
	if m.CreateFunc == nil {
		panic("called not defined CreateFunc")
	}
	return m.CreateFunc(ctx, level)
}

func (m *MockLevelRepository) FindByID(ctx context.Context, id string) (domain.Level, error) {
	if m.FindByIDFunc == nil {
		panic("called not defined FindByIDFunc")
	}
	return m.FindByIDFunc(ctx, id)
}

func (m *MockLevelRepository) FindAll(ctx context.Context) ([]domain.Level, error) {
	if m.FindAllFunc == nil {
		panic("called not defined FindAllFunc")
	}
	return m.FindAllFunc(ctx)
}

func (m *MockLevelRepository) Update(ctx context.Context, level domain.Level) (domain.Level, error) {
	if m.UpdateFunc == nil {
		panic("called not defined UpdateFunc")
	}
	return m.UpdateFunc(ctx, level)
}

func (m *MockLevelRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc == nil {
		panic("called not defined DeleteFunc")
	}
	return m.DeleteFunc(ctx, id)
}

func (m *MockLevelRepository) Reorder(ctx context.Context, orderedIDs []string) error {
	if m.ReorderFunc == nil {
		panic("called not defined ReorderFunc")
	}
	return m.ReorderFunc(ctx, orderedIDs)
}

// LevelRepositoryContractTesting runs all contract tests for a LevelRepository implementation.
func LevelRepositoryContractTesting(t *testing.T, repo levels.LevelRepository) {
	ctx := context.Background()

	t.Run("Contract: Create and FindByID retrieves stored level", func(t *testing.T) {
		lvl := domain.Level{
			ID:        "lvl-contract-1",
			Name:      "Ground Floor",
			Order:     1,
			IsOutdoor: false,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}

		created, err := repo.Create(ctx, lvl)
		require.NoError(t, err)
		assert.Equal(t, lvl.ID, created.ID)
		assert.Equal(t, lvl.Name, created.Name)
		assert.Equal(t, lvl.Order, created.Order)
		assert.Equal(t, lvl.IsOutdoor, created.IsOutdoor)
		assert.Equal(t, []string{"controls", "sensors"}, created.Layers)

		found, err := repo.FindByID(ctx, lvl.ID)
		require.NoError(t, err)
		assert.Equal(t, lvl.ID, found.ID)
		assert.Equal(t, lvl.Name, found.Name)
		assert.Equal(t, lvl.Order, found.Order)
		assert.Equal(t, lvl.IsOutdoor, found.IsOutdoor)
		assert.Equal(t, []string{"controls", "sensors"}, found.Layers)
	})

	t.Run("Contract: FindByID returns ErrLevelNotFound for missing level", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: FindAll returns levels in ascending order", func(t *testing.T) {
		l1 := domain.Level{
			ID:        "lvl-sort-2",
			Name:      "1st Floor",
			Order:     2,
			IsOutdoor: false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		l2 := domain.Level{
			ID:        "lvl-sort-1",
			Name:      "Sous-sol",
			Order:     0,
			IsOutdoor: false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		_, err := repo.Create(ctx, l1)
		require.NoError(t, err)
		_, err = repo.Create(ctx, l2)
		require.NoError(t, err)

		all, err := repo.FindAll(ctx)
		require.NoError(t, err)
		require.True(t, len(all) >= 2)

		// Verify order is strictly non-decreasing
		for i := 0; i < len(all)-1; i++ {
			assert.True(t, all[i].Order <= all[i+1].Order, "Levels must be sorted by Order ascending")
		}
	})

	t.Run("Contract: Update modifies level details", func(t *testing.T) {
		lvl := domain.Level{
			ID:        "lvl-update-1",
			Name:      "Original Garden",
			Order:     3,
			IsOutdoor: true,
			CreatedAt: time.Now().UTC().Truncate(time.Second),
			UpdatedAt: time.Now().UTC().Truncate(time.Second),
		}
		_, err := repo.Create(ctx, lvl)
		require.NoError(t, err)

		lvl.Name = "Modified Garden"
		lvl.IsOutdoor = true
		lvl.Layers = []string{"controls", "plants", "lighting"}
		lvl.UpdatedAt = time.Now().UTC().Truncate(time.Second)

		updated, err := repo.Update(ctx, lvl)
		require.NoError(t, err)
		assert.Equal(t, "Modified Garden", updated.Name)
		assert.Equal(t, []string{"controls", "plants", "lighting"}, updated.Layers)

		found, err := repo.FindByID(ctx, lvl.ID)
		require.NoError(t, err)
		assert.Equal(t, "Modified Garden", found.Name)
		assert.Equal(t, []string{"controls", "plants", "lighting"}, found.Layers)
	})

	t.Run("Contract: Update returns ErrLevelNotFound for non-existent level", func(t *testing.T) {
		missing := domain.Level{
			ID:        "lvl-missing-update",
			Name:      "Inconnu",
			Order:     99,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.Update(ctx, missing)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("Contract: Reorder updates ordering of existing levels", func(t *testing.T) {
		lvlA := domain.Level{
			ID:        "lvl-reorder-a",
			Name:      "A",
			Order:     10,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		lvlB := domain.Level{
			ID:        "lvl-reorder-b",
			Name:      "B",
			Order:     20,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.Create(ctx, lvlA)
		require.NoError(t, err)
		_, err = repo.Create(ctx, lvlB)
		require.NoError(t, err)

		err = repo.Reorder(ctx, []string{"lvl-reorder-b", "lvl-reorder-a"})
		require.NoError(t, err)

		foundB, err := repo.FindByID(ctx, "lvl-reorder-b")
		require.NoError(t, err)
		foundA, err := repo.FindByID(ctx, "lvl-reorder-a")
		require.NoError(t, err)

		assert.Equal(t, 0, foundB.Order)
		assert.Equal(t, 1, foundA.Order)
	})

	t.Run("Contract: Delete removes an existing level", func(t *testing.T) {
		lvl := domain.Level{
			ID:        "lvl-del-1",
			Name:      "To Delete",
			Order:     50,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_, err := repo.Create(ctx, lvl)
		require.NoError(t, err)

		err = repo.Delete(ctx, lvl.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, lvl.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("Contract: Delete returns ErrLevelNotFound for missing level", func(t *testing.T) {
		err := repo.Delete(ctx, "lvl-missing-delete")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
		assert.True(t, domain.IsDomainError(err))
	})
}
