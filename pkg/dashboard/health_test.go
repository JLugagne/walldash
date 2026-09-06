package dashboard_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthResponseValidation(t *testing.T) {
	validate := validator.New()

	t.Run("valid HealthResponse passes validation", func(t *testing.T) {
		res := pkgdashboard.HealthResponse{
			Status:    "ok",
			Database:  "ok",
			Version:   "0.1.0",
			Timestamp: time.Now(),
		}
		err := validate.Struct(res)
		require.NoError(t, err)
	})

	t.Run("invalid HealthResponse with missing fields fails validation", func(t *testing.T) {
		res := pkgdashboard.HealthResponse{
			Status: "invalid_status",
		}
		err := validate.Struct(res)
		require.Error(t, err)
		assert.True(t, domain.IsDomainError(pkgdashboard.ErrInvalidHealthResponse))
	})
}
