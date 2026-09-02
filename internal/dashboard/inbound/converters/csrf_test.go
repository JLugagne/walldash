package converters_test

import (
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	"github.com/stretchr/testify/assert"
)

func TestToPublicCSRFToken(t *testing.T) {
	rawToken := "sample-csrf-token-xyz"
	resp := converters.ToPublicCSRFToken(rawToken)

	assert.Equal(t, rawToken, resp.CSRFToken)
}
