package healthtest

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHealthQueries is a function-based mock implementation of health.HealthQueries.
//
// Example usage:
//
//	mock := &healthtest.MockHealthQueries{
//		GetHealthFunc: func(ctx context.Context) (domain.Health, error) {
//			return domain.Health{Status: domain.HealthStatusOK}, nil
//		},
//	}
type MockHealthQueries struct {
	GetHealthFunc func(ctx context.Context) (domain.Health, error)
}

func (m *MockHealthQueries) GetHealth(ctx context.Context) (domain.Health, error) {
	if m.GetHealthFunc == nil {
		panic("called not defined GetHealthFunc")
	}
	return m.GetHealthFunc(ctx)
}

// MockHealthCommands is a function-based mock implementation of health.HealthCommands.
type MockHealthCommands struct {
	GetHealthFunc func(ctx context.Context) (domain.Health, error)
}

func (m *MockHealthCommands) GetHealth(ctx context.Context) (domain.Health, error) {
	if m.GetHealthFunc == nil {
		panic("called not defined GetHealthFunc")
	}
	return m.GetHealthFunc(ctx)
}

// HealthQueriesContractTesting verifies that any health.HealthQueries implementation behaves according to contract.
func HealthQueriesContractTesting(t *testing.T, queries health.HealthQueries) {
	ctx := context.Background()

	t.Run("Contract: GetHealth returns health status successfully", func(t *testing.T) {
		h, err := queries.GetHealth(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, h.Status)
	})
}
