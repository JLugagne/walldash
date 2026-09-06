package homeassistant_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/repositories/ha/hatest"
	"github.com/JLugagne/walldash/internal/dashboard/outbound/homeassistant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeAssistantClient_FallbackContract(t *testing.T) {
	// When empty URL or token, client should use fallback/mock devices and automations and satisfy contract test
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
				"friendly_name": "Living Room Ceiling Light",
				"brightness":    255,
			},
		},
		{
			"entity_id": "switch.kitchen_plug",
			"state":     "off",
			"attributes": map[string]any{
				"friendly_name": "Kitchen Plug",
			},
		},
		{
			"entity_id": "sensor.living_temp",
			"state":     "21.5",
			"attributes": map[string]any{
				"friendly_name":       "Living Room Temperature",
				"unit_of_measurement": "°C",
			},
		},
		{
			// Unsupported domain: should be filtered out from devices
			"entity_id": "camera.front_door",
			"state":     "idle",
			"attributes": map[string]any{
				"friendly_name": "Entrance Camera",
			},
		},
		{
			"entity_id": "automation.morning_routine",
			"state":     "on",
			"attributes": map[string]any{
				"friendly_name":  "Routine Matin",
				"current":        0,
				"last_triggered": "2026-09-02T07:30:00+00:00",
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
		case "/api/states/automation.morning_routine":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockStates[4])
		case "/api/states/non_existent.id":
			w.WriteHeader(http.StatusNotFound)
		case "/api/services/light/toggle":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]any{mockStates[0]})
		case "/api/services/automation/trigger":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]any{mockStates[4]})
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
		assert.Equal(t, "Living Room Ceiling Light", dev.Name)
		assert.Equal(t, "on", dev.State)
	})

	t.Run("GetState returns ErrDeviceNotFound for 404", func(t *testing.T) {
		_, err := client.GetState(ctx, "non_existent.id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDeviceNotFound)
	})

	t.Run("CallService executes POST against HA services API", func(t *testing.T) {
		err := client.CallService(ctx, "light", "toggle", "light.living_room")
		require.NoError(t, err)
	})

	t.Run("GetAutomations returns automations from states", func(t *testing.T) {
		automations, err := client.GetAutomations(ctx)
		require.NoError(t, err)
		require.Len(t, automations, 1)
		assert.Equal(t, "automation.morning_routine", automations[0].ID)
		assert.Equal(t, "Routine Matin", automations[0].Name)
		assert.Equal(t, "on", automations[0].State)
		assert.Equal(t, 0, automations[0].Current)
		require.NotNil(t, automations[0].LastTriggered)
	})

	t.Run("GetAutomation returns single automation", func(t *testing.T) {
		auto, err := client.GetAutomation(ctx, "automation.morning_routine")
		require.NoError(t, err)
		assert.Equal(t, "automation.morning_routine", auto.ID)
		assert.Equal(t, "Routine Matin", auto.Name)
	})

	t.Run("TriggerAutomation posts to /api/services/automation/trigger", func(t *testing.T) {
		err := client.TriggerAutomation(ctx, "automation.morning_routine")
		require.NoError(t, err)
	})
}

func TestHomeAssistantClient_CallServiceFallback(t *testing.T) {
	ctx := context.Background()
	client := homeassistant.NewClient("", "", nil)

	t.Run("toggle toggles state between on and off", func(t *testing.T) {
		dev, err := client.GetState(ctx, "light.salon_plafond")
		require.NoError(t, err)
		initialState := dev.State

		err = client.CallService(ctx, "light", "toggle", "light.salon_plafond")
		require.NoError(t, err)

		updated, err := client.GetState(ctx, "light.salon_plafond")
		require.NoError(t, err)
		if initialState == "on" {
			assert.Equal(t, "off", updated.State)
		} else {
			assert.Equal(t, "on", updated.State)
		}

		// Toggle back
		err = client.CallService(ctx, "light", "toggle", "light.salon_plafond")
		require.NoError(t, err)

		restored, err := client.GetState(ctx, "light.salon_plafond")
		require.NoError(t, err)
		assert.Equal(t, initialState, restored.State)
	})

	t.Run("turn_on and turn_off set state explicitly", func(t *testing.T) {
		err := client.CallService(ctx, "switch", "turn_off", "switch.machine_a_cafe")
		require.NoError(t, err)

		dev, err := client.GetState(ctx, "switch.machine_a_cafe")
		require.NoError(t, err)
		assert.Equal(t, "off", dev.State)

		err = client.CallService(ctx, "switch", "turn_on", "switch.machine_a_cafe")
		require.NoError(t, err)

		dev, err = client.GetState(ctx, "switch.machine_a_cafe")
		require.NoError(t, err)
		assert.Equal(t, "on", dev.State)
	})

	t.Run("CallService on non-existent device returns ErrDeviceNotFound", func(t *testing.T) {
		err := client.CallService(ctx, "light", "toggle", "light.non_existent")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDeviceNotFound)
	})

	t.Run("TriggerAutomation updates last triggered in fallback mode", func(t *testing.T) {
		err := client.TriggerAutomation(ctx, "automation.eteindre_toutes_les_lumieres")
		require.NoError(t, err)

		auto, err := client.GetAutomation(ctx, "automation.eteindre_toutes_les_lumieres")
		require.NoError(t, err)
		require.NotNil(t, auto.LastTriggered)
	})
}
