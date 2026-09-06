package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/app"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	repohealthtest "github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/health/healthtest"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/uow/uowtest"
	svchealthtest "github.com/JLugagne/ha-dash/internal/dashboard/domain/service/health/healthtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthServiceContract(t *testing.T) {
	mockRepo := &repohealthtest.MockRepository{
		PingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	mockUow := &uowtest.MockUnitOfWork{
		DoFunc: func(ctx context.Context, fn func(repos uow.Repositories) error) error {
			return fn(uow.Repositories{Health: mockRepo})
		},
	}

	service := app.New(mockRepo, nil, nil, nil, nil, nil, nil, mockUow, "0.1.0")
	svchealthtest.HealthQueriesContractTesting(t, service)
}

func TestHealthService_GetHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("returns ok health status when database is reachable", func(t *testing.T) {
		mockRepo := &repohealthtest.MockRepository{
			PingFunc: func(ctx context.Context) error {
				return nil
			},
		}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockRepo, nil, nil, nil, nil, nil, nil, mockUow, "0.1.0")
		health, err := service.GetHealth(ctx)

		require.NoError(t, err)
		assert.Equal(t, domain.HealthStatusOK, health.Status)
		assert.Equal(t, "ok", health.Database)
		assert.Equal(t, "0.1.0", health.Version)
		assert.False(t, health.Timestamp.IsZero())
	})

	t.Run("returns domain error when repository ping fails", func(t *testing.T) {
		pingErr := errors.New("connection refused")
		mockRepo := &repohealthtest.MockRepository{
			PingFunc: func(ctx context.Context) error {
				return pingErr
			},
		}
		mockUow := &uowtest.MockUnitOfWork{}

		service := app.New(mockRepo, nil, nil, nil, nil, nil, nil, mockUow, "0.1.0")
		_, err := service.GetHealth(ctx)

		require.Error(t, err)
		assert.True(t, domain.IsDomainError(err))
		assert.ErrorIs(t, err, domain.ErrHealthCheckFailed)
		assert.ErrorIs(t, err, pingErr)
	})
}
