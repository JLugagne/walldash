package converters

import (
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
)

// ToPublicCSRFToken converts a raw CSRF token string into the public response DTO.
func ToPublicCSRFToken(token string) pkgdashboard.CSRFTokenResponse {
	return pkgdashboard.CSRFTokenResponse{
		CSRFToken: token,
	}
}
