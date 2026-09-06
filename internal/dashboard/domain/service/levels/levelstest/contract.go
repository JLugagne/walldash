package levelstest

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLevelQueries is a function-based mock implementation of levels.LevelQueries.
type MockLevelQueries struct {
	ListLevelsFunc func(ctx context.Context) ([]domain.Level, error)
	GetLevelFunc   func(ctx context.Context, id string) (domain.Level, error)
	GetPlanFunc    func(ctx context.Context, levelID string) (domain.Plan, error)
}

func (m *MockLevelQueries) ListLevels(ctx context.Context) ([]domain.Level, error) {
	if m.ListLevelsFunc == nil {
		panic("called not defined ListLevelsFunc")
	}
	return m.ListLevelsFunc(ctx)
}

func (m *MockLevelQueries) GetLevel(ctx context.Context, id string) (domain.Level, error) {
	if m.GetLevelFunc == nil {
		panic("called not defined GetLevelFunc")
	}
	return m.GetLevelFunc(ctx, id)
}

func (m *MockLevelQueries) GetPlan(ctx context.Context, levelID string) (domain.Plan, error) {
	if m.GetPlanFunc == nil {
		panic("called not defined GetPlanFunc")
	}
	return m.GetPlanFunc(ctx, levelID)
}

// MockLevelCommands is a function-based mock implementation of levels.LevelCommands.
type MockLevelCommands struct {
	ListLevelsFunc    func(ctx context.Context) ([]domain.Level, error)
	GetLevelFunc      func(ctx context.Context, id string) (domain.Level, error)
	GetPlanFunc       func(ctx context.Context, levelID string) (domain.Plan, error)
	CreateLevelFunc   func(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error)
	UpdateLevelFunc   func(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error)
	DeleteLevelFunc   func(ctx context.Context, actor domain.Actor, id string) error
	ReorderLevelsFunc func(ctx context.Context, actor domain.Actor, orderedIDs []string) error
	SavePlanFunc      func(ctx context.Context, actor domain.Actor, plan domain.Plan) (domain.Plan, error)
}

func (m *MockLevelCommands) ListLevels(ctx context.Context) ([]domain.Level, error) {
	if m.ListLevelsFunc == nil {
		panic("called not defined ListLevelsFunc")
	}
	return m.ListLevelsFunc(ctx)
}

func (m *MockLevelCommands) GetLevel(ctx context.Context, id string) (domain.Level, error) {
	if m.GetLevelFunc == nil {
		panic("called not defined GetLevelFunc")
	}
	return m.GetLevelFunc(ctx, id)
}

func (m *MockLevelCommands) GetPlan(ctx context.Context, levelID string) (domain.Plan, error) {
	if m.GetPlanFunc == nil {
		panic("called not defined GetPlanFunc")
	}
	return m.GetPlanFunc(ctx, levelID)
}

func (m *MockLevelCommands) CreateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
	if m.CreateLevelFunc == nil {
		panic("called not defined CreateLevelFunc")
	}
	return m.CreateLevelFunc(ctx, actor, level)
}

func (m *MockLevelCommands) UpdateLevel(ctx context.Context, actor domain.Actor, level domain.Level) (domain.Level, error) {
	if m.UpdateLevelFunc == nil {
		panic("called not defined UpdateLevelFunc")
	}
	return m.UpdateLevelFunc(ctx, actor, level)
}

func (m *MockLevelCommands) DeleteLevel(ctx context.Context, actor domain.Actor, id string) error {
	if m.DeleteLevelFunc == nil {
		panic("called not defined DeleteLevelFunc")
	}
	return m.DeleteLevelFunc(ctx, actor, id)
}

func (m *MockLevelCommands) ReorderLevels(ctx context.Context, actor domain.Actor, orderedIDs []string) error {
	if m.ReorderLevelsFunc == nil {
		panic("called not defined ReorderLevelsFunc")
	}
	return m.ReorderLevelsFunc(ctx, actor, orderedIDs)
}

func (m *MockLevelCommands) SavePlan(ctx context.Context, actor domain.Actor, plan domain.Plan) (domain.Plan, error) {
	if m.SavePlanFunc == nil {
		panic("called not defined SavePlanFunc")
	}
	return m.SavePlanFunc(ctx, actor, plan)
}

// LevelQueriesContractTesting verifies that any levels.LevelQueries implementation adheres to query contracts.
func LevelQueriesContractTesting(t *testing.T, queries levels.LevelQueries) {
	ctx := context.Background()

	t.Run("Contract: ListLevels returns without error", func(t *testing.T) {
		lvls, err := queries.ListLevels(ctx)
		require.NoError(t, err)
		assert.NotNil(t, lvls)
	})

	t.Run("Contract: GetLevel returns ErrLevelNotFound for unknown level", func(t *testing.T) {
		_, err := queries.GetLevel(ctx, "unknown-level-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrLevelNotFound)
	})

	t.Run("Contract: GetPlan returns ErrPlanNotFound for unknown level plan", func(t *testing.T) {
		_, err := queries.GetPlan(ctx, "unknown-level-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPlanNotFound)
	})
}
