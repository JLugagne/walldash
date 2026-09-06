package homeassistant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/domain/repositories/ha"
	"github.com/JLugagne/ha-dash/internal/pkg/logger"
)

var _ ha.HomeAssistantRepository = (*Client)(nil)

// Client interacts with the Home Assistant HTTP REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient initializes a Home Assistant HTTP client.
func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}
}

type haEntityState struct {
	EntityID   string         `json:"entity_id"`
	State      string         `json:"state"`
	Attributes map[string]any `json:"attributes"`
}

var (
	fallbackMu      sync.RWMutex
	fallbackDevices = []domain.Device{
		{
			ID:     "light.salon_plafond",
			Name:   "Plafonnier Salon",
			Domain: domain.DomainLight,
			State:  "on",
			Attributes: map[string]any{
				"friendly_name": "Plafonnier Salon",
				"brightness":    255,
			},
		},
		{
			ID:     "light.cuisine_spot",
			Name:   "Spots Cuisine",
			Domain: domain.DomainLight,
			State:  "off",
			Attributes: map[string]any{
				"friendly_name": "Spots Cuisine",
			},
		},
		{
			ID:     "light.chambre_chevet",
			Name:   "Lampe Chevet Chambre",
			Domain: domain.DomainLight,
			State:  "on",
			Attributes: map[string]any{
				"friendly_name": "Lampe Chevet Chambre",
				"brightness":    128,
			},
		},
		{
			ID:     "switch.machine_a_cafe",
			Name:   "Machine à Café",
			Domain: domain.DomainSwitch,
			State:  "on",
			Attributes: map[string]any{
				"friendly_name": "Machine à Café",
			},
		},
		{
			ID:     "switch.prise_tv",
			Name:   "Prise TV Salon",
			Domain: domain.DomainSwitch,
			State:  "on",
			Attributes: map[string]any{
				"friendly_name": "Prise TV Salon",
			},
		},
		{
			ID:     "sensor.temperature_salon",
			Name:   "Température Salon",
			Domain: domain.DomainSensor,
			State:  "21.4",
			Attributes: map[string]any{
				"friendly_name":       "Température Salon",
				"unit_of_measurement": "°C",
			},
		},
		{
			ID:     "sensor.humidite_sdb",
			Name:   "Humidité Salle de Bain",
			Domain: domain.DomainSensor,
			State:  "62",
			Attributes: map[string]any{
				"friendly_name":       "Humidité Salle de Bain",
				"unit_of_measurement": "%",
			},
		},
		{
			ID:     "climate.thermostat_salon",
			Name:   "Thermostat Salon",
			Domain: domain.DomainClimate,
			State:  "heat",
			Attributes: map[string]any{
				"friendly_name":       "Thermostat Salon",
				"current_temperature": 20.8,
				"temperature":         21.5,
			},
		},
		{
			ID:     "media_player.enceinte_salon",
			Name:   "Sonos Salon",
			Domain: domain.DomainMediaPlayer,
			State:  "playing",
			Attributes: map[string]any{
				"friendly_name": "Sonos Salon",
				"media_title":   "Get Lucky",
				"media_artist":  "Daft Punk",
				"volume_level":  0.45,
			},
		},
	}
)

// GetStates fetches all states from Home Assistant, filtering to supported domains.
// If the remote server is unreachable or unconfigured, it gracefully falls back to mock devices.
func (c *Client) GetStates(ctx context.Context) ([]domain.Device, error) {
	log := logger.LoggerFromContext(ctx)

	if c.baseURL == "" || c.token == "" {
		log.Info("HA URL or token unconfigured, using fallback devices")
		return copyFallbackDevices(), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/states", nil)
	if err != nil {
		log.WithError(err).Error("failed to create HA states request")
		return copyFallbackDevices(), nil
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("HA API unreachable, falling back to mock devices")
		return copyFallbackDevices(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status_code", resp.StatusCode).Warn("HA API returned non-200, falling back to mock devices")
		return copyFallbackDevices(), nil
	}

	var rawStates []haEntityState
	if err := json.NewDecoder(resp.Body).Decode(&rawStates); err != nil {
		log.WithError(err).Error("failed to decode HA states JSON")
		return copyFallbackDevices(), nil
	}

	var devices []domain.Device
	for _, raw := range rawStates {
		parts := strings.Split(raw.EntityID, ".")
		if len(parts) < 2 {
			continue
		}
		domainStr := parts[0]
		if !domain.IsSupportedDomain(domainStr) {
			continue
		}

		name := raw.EntityID
		if fn, ok := raw.Attributes["friendly_name"].(string); ok && strings.TrimSpace(fn) != "" {
			name = fn
		}

		devices = append(devices, domain.Device{
			ID:         raw.EntityID,
			Name:       name,
			Domain:     domainStr,
			State:      raw.State,
			Attributes: raw.Attributes,
		})
	}

	if devices == nil {
		devices = []domain.Device{}
	}
	return devices, nil
}

// GetState fetches a single entity's state from Home Assistant.
func (c *Client) GetState(ctx context.Context, entityID string) (domain.Device, error) {
	log := logger.LoggerFromContext(ctx)

	if c.baseURL == "" || c.token == "" {
		fallbackMu.RLock()
		defer fallbackMu.RUnlock()
		for _, d := range fallbackDevices {
			if d.ID == entityID {
				return d, nil
			}
		}
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("entity %s not found in fallback devices", entityID))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/states/"+entityID, nil)
	if err != nil {
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("HA API entity fetch failed, trying fallback")
		fallbackMu.RLock()
		defer fallbackMu.RUnlock()
		for _, d := range fallbackDevices {
			if d.ID == entityID {
				return d, nil
			}
		}
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("entity %s not found on HA", entityID))
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("HA returned status %d for entity %s", resp.StatusCode, entityID))
	}

	var raw haEntityState
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return domain.Device{}, errors.Join(domain.ErrDeviceNotFound, err)
	}

	parts := strings.Split(raw.EntityID, ".")
	if len(parts) < 2 || !domain.IsSupportedDomain(parts[0]) {
		return domain.Device{}, errors.Join(domain.ErrUnsupportedDomain, fmt.Errorf("entity %s belongs to unsupported domain", entityID))
	}

	name := raw.EntityID
	if fn, ok := raw.Attributes["friendly_name"].(string); ok && strings.TrimSpace(fn) != "" {
		name = fn
	}

	return domain.Device{
		ID:         raw.EntityID,
		Name:       name,
		Domain:     parts[0],
		State:      raw.State,
		Attributes: raw.Attributes,
	}, nil
}

// CallService calls a Home Assistant service for the specified entity.
// In dev/mock mode or when HA is unreachable, it mutates state in fallback devices.
func (c *Client) CallService(ctx context.Context, domainStr string, service string, entityID string) error {
	log := logger.LoggerFromContext(ctx)

	if c.baseURL == "" || c.token == "" {
		return c.mutateFallbackDevice(ctx, service, entityID)
	}

	payload, err := json.Marshal(map[string]string{
		"entity_id": entityID,
	})
	if err != nil {
		return errors.Join(domain.ErrInvalidRequest, err)
	}

	url := fmt.Sprintf("%s/api/services/%s/%s", c.baseURL, domainStr, service)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return errors.Join(domain.ErrHealthCheckFailed, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("HA API unreachable, falling back to mock device mutation")
		return c.mutateFallbackDevice(ctx, service, entityID)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("entity %s or service %s not found on HA", entityID, service))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Join(domain.ErrHealthCheckFailed, fmt.Errorf("HA returned HTTP %d for service %s/%s", resp.StatusCode, domainStr, service))
	}

	return nil
}

func (c *Client) mutateFallbackDevice(ctx context.Context, service, entityID string) error {
	log := logger.LoggerFromContext(ctx)

	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	found := false
	for i := range fallbackDevices {
		if fallbackDevices[i].ID == entityID {
			found = true
			switch service {
			case "toggle":
				if fallbackDevices[i].State == "on" {
					fallbackDevices[i].State = "off"
				} else {
					fallbackDevices[i].State = "on"
				}
			case "turn_on":
				fallbackDevices[i].State = "on"
			case "turn_off":
				fallbackDevices[i].State = "off"
			}
			log.WithField("entity_id", entityID).WithField("state", fallbackDevices[i].State).Info("fallback device state updated")
			break
		}
	}

	if !found {
		return errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("entity %s not found in fallback devices", entityID))
	}
	return nil
}

func copyFallbackDevices() []domain.Device {
	fallbackMu.RLock()
	defer fallbackMu.RUnlock()
	res := make([]domain.Device, len(fallbackDevices))
	copy(res, fallbackDevices)
	return res
}
