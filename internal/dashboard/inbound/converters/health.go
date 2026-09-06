package converters

import (
	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
)

// ToPublicHealth converts a domain.Health entity into a public pkgdashboard.HealthResponse.
func ToPublicHealth(h domain.Health) pkgdashboard.HealthResponse {
	return pkgdashboard.HealthResponse{
		Status:    string(h.Status),
		Database:  h.Database,
		Version:   h.Version,
		Timestamp: h.Timestamp,
	}
}

// ToDomainHealth converts a public pkgdashboard.HealthResponse into a domain.Health entity.
func ToDomainHealth(r pkgdashboard.HealthResponse) domain.Health {
	return domain.Health{
		Status:    domain.HealthStatus(r.Status),
		Database:  r.Database,
		Version:   r.Version,
		Timestamp: r.Timestamp,
	}
}
