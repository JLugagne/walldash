package healthtest

import (
	"context"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health"
	"github.com/stretchr/testify/require"
)

// MockRepository is a function-based mock implementation of the health.Repository interface.
//
// Example usage:
//
//	mock := &healthtest.MockRepository{
//		PingFunc: func(ctx context.Context) error {
//			return nil
//		},
//	}
type MockRepository struct {
	PingFunc func(ctx context.Context) error
}

func (m *MockRepository) Ping(ctx context.Context) error {
	if m.PingFunc == nil {
		panic("called not defined PingFunc")
	}
	return m.PingFunc(ctx)
}

// RepositoryContractTesting runs all contract tests for a health.Repository implementation.
func RepositoryContractTesting(t *testing.T, repo health.Repository) {
	ctx := context.Background()

	t.Run("Contract: Ping succeeds on active repository", func(t *testing.T) {
		err := repo.Ping(ctx)
		require.NoError(t, err, "Ping should succeed on an active repository")
	})
}
