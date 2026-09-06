package uowtest

import (
	"context"
	"errors"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUnitOfWork is a function-based mock implementation of uow.UnitOfWork.
//
// Example usage:
//
//	mock := &uowtest.MockUnitOfWork{
//		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
//			return fn(uow.Repositories{})
//		},
//	}
type MockUnitOfWork struct {
	DoFunc func(ctx context.Context, fn func(repos uow.Repositories) error) error
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(repos uow.Repositories) error) error {
	if m.DoFunc == nil {
		panic("called not defined DoFunc")
	}
	return m.DoFunc(ctx, fn)
}

// UnitOfWorkContractTesting verifies that any UnitOfWork implementation executes transactions properly.
func UnitOfWorkContractTesting(t *testing.T, unit uow.UnitOfWork) {
	ctx := context.Background()

	t.Run("Contract: Do executes callback and commits transaction on nil error", func(t *testing.T) {
		executed := false
		err := unit.Do(ctx, func(repos uow.Repositories) error {
			executed = true
			if repos.Health != nil {
				return repos.Health.Ping(ctx)
			}
			return nil
		})
		require.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("Contract: Do returns error and aborts transaction on callback error", func(t *testing.T) {
		expectedErr := errors.New("something went wrong inside transaction")
		err := unit.Do(ctx, func(repos uow.Repositories) error {
			return expectedErr
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}
