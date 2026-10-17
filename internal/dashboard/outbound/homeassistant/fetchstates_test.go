package homeassistant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchStates_UnconfiguredReturnsError(t *testing.T) {
	client := NewClient("", "", nil)

	devices, err := client.FetchStates(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHomeAssistantUnavailable)
	assert.Nil(t, devices)

	// The REST-facing GetStates still falls back to mock devices for an unconfigured instance.
	fallback, err := client.GetStates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, fallback)
}

func TestFetchStates_ServerErrorReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", server.Client())
	devices, err := client.FetchStates(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHomeAssistantUnavailable)
	assert.Nil(t, devices)
}

func TestFetchStates_FiltersUnsupportedDomains(t *testing.T) {
	states := []map[string]any{
		{"entity_id": "light.living_room", "state": "on", "attributes": map[string]any{"friendly_name": "Living Room Light"}, "last_updated": "2026-01-01T12:00:00+00:00"},
		{"entity_id": "camera.front_door", "state": "idle", "attributes": map[string]any{"friendly_name": "Camera"}, "last_updated": "2026-01-01T12:00:00+00:00"},
		{"entity_id": "switch.kitchen_plug", "state": "unavailable", "attributes": map[string]any{"friendly_name": "Kitchen Plug"}, "last_updated": "2026-01-01T12:05:00+00:00"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(states)
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", server.Client())
	devices, err := client.FetchStates(context.Background())
	require.NoError(t, err)
	require.Len(t, devices, 2)
	assert.Equal(t, "light.living_room", devices[0].ID)
	assert.Equal(t, "switch.kitchen_plug", devices[1].ID)
	assert.Equal(t, "unavailable", devices[1].State)
}
