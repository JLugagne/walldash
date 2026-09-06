package dashboard_test

import (
	"testing"

	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestCSRFTokenResponse(t *testing.T) {
	resp := pkgdashboard.CSRFTokenResponse{
		CSRFToken: "test-token-12345",
	}
	assert.Equal(t, "test-token-12345", resp.CSRFToken)
}
