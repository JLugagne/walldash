package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainError(t *testing.T) {
	t.Run("Error returns Message", func(t *testing.T) {
		err := &domain.Error{
			Code:    "TEST_CODE",
			Message: "test message",
		}
		assert.Equal(t, "test message", err.Error())
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		cause := fmt.Errorf("cause error")
		err := &domain.Error{
			Code:    "TEST_CODE",
			Message: "wrapped error",
			Err:     cause,
		}
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("IsDomainError returns true for direct domain error", func(t *testing.T) {
		err := domain.ErrNotFound
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("IsDomainError returns true when joined with standard errors.Join", func(t *testing.T) {
		cause := fmt.Errorf("sql connection failed")
		joined := errors.Join(domain.ErrDatabaseUnavailable, cause)
		assert.True(t, domain.IsDomainError(joined))
	})

	t.Run("IsDomainError returns false for non-domain error", func(t *testing.T) {
		err := fmt.Errorf("regular error")
		assert.False(t, domain.IsDomainError(err))
	})

	t.Run("AsDomainError extracts domain error", func(t *testing.T) {
		cause := fmt.Errorf("disk full")
		joined := errors.Join(domain.ErrHealthCheckFailed, cause)
		extracted, ok := domain.AsDomainError(joined)
		require.True(t, ok)
		assert.Equal(t, "HEALTH_CHECK_FAILED", extracted.Code)
		assert.Equal(t, "health check failed", extracted.Message)
	})
}
