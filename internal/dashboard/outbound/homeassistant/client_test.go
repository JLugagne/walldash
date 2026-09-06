package homeassistant_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha/hatest"
	"github.com/JLugagne/ha-dash/internal/dashboard/outbound/homeassistant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeAssistantClient_FallbackContract(t *testing.T) {
	// When empty URL or token, client should use fallback/mock devices and satisfy contract test
	client := homeassistant.NewClient("", "", nil)
	hatest.HomeAssistantRepositoryContractTesting(t, client)
}

func TestHomeAssistantClient_RealServer(t *testing.T) {
	ctx := context.Background()

	mockStates := []map[string]any{
		{
			"entity_id": "light.living_room",
			"state":     "on",
			"attributes": map[string]any{
				"friendly_name": "Salon Plafonnier",
				"brightness":    255,
			},
		},
		{
			"entity_id": "switch.kitchen_plug",
			"state":     "off",
			"attributes": map[string]any{
				"friendly_name": "Prise Cuisine",
			},
		},
		{
			"entity_id": "sensor.living_temp",
			"state":     "21.5",
			"attributes": map[string]any{
				"friendly_name":       "Température Salon",
				"unit_of_measurement": "°C",
			},
		},
		{
			// Unsupported domain: should be filtered out
			"entity_id": "camera.front_door",
			"state":     "idle",
			"attributes": map[string]any{
				"friendly_name": "Caméra Entrée",
			},
		},
		{
			// Unsupported domain: should be filtered out
			"entity_id": "automation.morning_routine",
			"state":     "on",
			"attributes": map[string]any{
				"friendly_name": "Routine Matin",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-secret-token", r.Header.Get("Authorization"))

		switch r.URL.Path {
		case "/api/states":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockStates)
		case "/api/states/light.living_room":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockStates[0])
		case "/api/states/non_existent.id":
			w.WriteHeader(http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := homeassistant.NewClient(server.URL, "test-secret-token", server.Client())

	t.Run("GetStates filters unsupported domains", func(t *testing.T) {
		devices, err := client.GetStates(ctx)
		require.NoError(t, err)
		require.Len(t, devices, 3)

		ids := make([]string, len(devices))
		for i, d := range devices {
			ids[i] = d.ID
			assert.True(t, domain.IsSupportedDomain(d.Domain))
		}
		assert.Contains(t, ids, "light.living_room")
		assert.Contains(t, ids, "switch.kitchen_plug")
		assert.Contains(t, ids, "sensor.living_temp")
		assert.NotContains(t, ids, "camera.front_door")
		assert.NotContains(t, ids, "automation.morning_routine")
	})

	t.Run("GetState returns device entity", func(t *testing.T) {
		dev, err := client.GetState(ctx, "light.living_room")
		require.NoError(t, err)
		assert.Equal(t, "light.living_room", dev.ID)
		assert.Equal(t, "Salon Plafonnier", dev.Name)
		assert.Equal(t, "on", dev.State)
	})

	t.Run("GetState returns ErrDeviceNotFound for 404", func(t *testing.T) {
		_, err := client.GetState(ctx, "non_existent.id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDeviceNotFound)
	})
}
