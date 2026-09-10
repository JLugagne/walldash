package homeassistant_test

import (
	"context"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/outbound/homeassistant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeAssistantClient_GetConfigFallback(t *testing.T) {
	client := homeassistant.NewClient("", "", nil)
	cfg, err := client.GetConfig(context.Background())
	require.NoError(t, err)
	assert.False(t, cfg.Configured)
	assert.Equal(t, "°C", cfg.TemperatureUnit)
	assert.Equal(t, "km", cfg.LengthUnit)
}
