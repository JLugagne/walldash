package converters_test

import (
	"testing"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestHealthConverters(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("ToPublicHealth converts domain.Health to pkgdashboard.HealthResponse", func(t *testing.T) {
		domainHealth := domain.Health{
			Status:    domain.HealthStatusOK,
			Database:  "ok",
			Version:   "0.1.0",
			Timestamp: now,
		}

		res := converters.ToPublicHealth(domainHealth)

		assert.Equal(t, "ok", res.Status)
		assert.Equal(t, "ok", res.Database)
		assert.Equal(t, "0.1.0", res.Version)
		assert.Equal(t, now, res.Timestamp)
	})

	t.Run("ToDomainHealth converts pkgdashboard.HealthResponse to domain.Health", func(t *testing.T) {
		res := pkgdashboard.HealthResponse{
			Status:    "ok",
			Database:  "ok",
			Version:   "0.1.0",
			Timestamp: now,
		}

		domainHealth := converters.ToDomainHealth(res)

		assert.Equal(t, domain.HealthStatusOK, domainHealth.Status)
		assert.Equal(t, "ok", domainHealth.Database)
		assert.Equal(t, "0.1.0", domainHealth.Version)
		assert.Equal(t, now, domainHealth.Timestamp)
	})
}
