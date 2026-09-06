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
			Timeout: 25 * time.Second,
		}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}
}

type haEntityState struct {
	EntityID    string         `json:"entity_id"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
	LastUpdated time.Time      `json:"last_updated"`
}

var (
	defaultLastTriggered1 = time.Now().UTC().Add(-45 * time.Minute)
	defaultLastTriggered2 = time.Now().UTC().Add(-3 * time.Hour)
	defaultLastTriggered3 = time.Now().UTC().Add(-12 * time.Minute)
	fallbackLastUpdated   = time.Now().UTC().Add(-2 * time.Minute)

	fallbackMu      sync.RWMutex
	fallbackDevices = []domain.Device{
		{
			ID:          "light.salon_plafond",
			Name:        "Plafonnier Salon",
			Domain:      domain.DomainLight,
			State:       "on",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Plafonnier Salon",
				"brightness":    255,
			},
		},
		{
			ID:          "light.cuisine_spot",
			Name:        "Spots Cuisine",
			Domain:      domain.DomainLight,
			State:       "off",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Spots Cuisine",
			},
		},
		{
			ID:          "light.chambre_chevet",
			Name:        "Lampe Chevet Chambre",
			Domain:      domain.DomainLight,
			State:       "on",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Lampe Chevet Chambre",
				"brightness":    128,
			},
		},
		{
			ID:          "switch.machine_a_cafe",
			Name:        "Machine à Café",
			Domain:      domain.DomainSwitch,
			State:       "on",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Machine à Café",
			},
		},
		{
			ID:          "switch.prise_tv",
			Name:        "Prise TV Salon",
			Domain:      domain.DomainSwitch,
			State:       "on",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Prise TV Salon",
			},
		},
		{
			ID:          "sensor.temperature_salon",
			Name:        "Température Salon",
			Domain:      domain.DomainSensor,
			State:       "21.4",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name":       "Température Salon",
				"unit_of_measurement": "°C",
			},
		},
		{
			ID:          "sensor.humidite_sdb",
			Name:        "Humidité Salle de Bain",
			Domain:      domain.DomainSensor,
			State:       "62",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name":       "Humidité Salle de Bain",
				"unit_of_measurement": "%",
			},
		},
		{
			ID:          "climate.thermostat_salon",
			Name:        "Thermostat Salon",
			Domain:      domain.DomainClimate,
			State:       "heat",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name":       "Thermostat Salon",
				"current_temperature": 20.8,
				"temperature":         21.5,
			},
		},
		{
			ID:          "media_player.enceinte_salon",
			Name:        "Sonos Salon",
			Domain:      domain.DomainMediaPlayer,
			State:       "playing",
			LastUpdated: fallbackLastUpdated,
			Attributes: map[string]any{
				"friendly_name": "Sonos Salon",
				"media_title":   "Get Lucky",
				"media_artist":  "Daft Punk",
				"volume_level":  0.45,
			},
		},
	}

	fallbackAutomations = []domain.Automation{
		{
			ID:            "automation.eteindre_toutes_les_lumieres",
			Name:          "Éteindre toutes les lumières",
			State:         "on",
			Current:       0,
			LastTriggered: &defaultLastTriggered1,
		},
		{
			ID:            "automation.scenario_depart_maison",
			Name:          "Scénario Départ Maison",
			State:         "on",
			Current:       0,
			LastTriggered: &defaultLastTriggered2,
		},
		{
			ID:            "automation.arrosage_automatique_jardin",
			Name:          "Arrosage Automatique Jardin",
			State:         "on",
			Current:       1,
			LastTriggered: &defaultLastTriggered3,
		},
		{
			ID:            "automation.simulation_presence",
			Name:          "Simulation de Présence Soirée",
			State:         "off",
			Current:       0,
			LastTriggered: nil,
		},
	}
)

// GetStates fetches all states from Home Assistant, filtering to supported domains.
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
			ID:          raw.EntityID,
			Name:        name,
			Domain:      domainStr,
			State:       raw.State,
			Attributes:  raw.Attributes,
			LastUpdated: raw.LastUpdated.UTC(),
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
		ID:          raw.EntityID,
		Name:        name,
		Domain:      parts[0],
		State:       raw.State,
		Attributes:  raw.Attributes,
		LastUpdated: raw.LastUpdated.UTC(),
	}, nil
}

// CallService calls a Home Assistant service for the specified entity.
func (c *Client) CallService(ctx context.Context, domainStr string, service string, entityID string) error {
	log := logger.LoggerFromContext(ctx)

	if c.baseURL == "" || c.token == "" {
		if domainStr == "automation" && service == "trigger" {
			return c.mutateFallbackAutomation(ctx, entityID)
		}
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
		log.WithError(err).Error("failed to call Home Assistant service")
		return errors.Join(domain.ErrHealthCheckFailed, err)
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

// GetAutomations fetches all automations (domain == "automation") from Home Assistant states.
func (c *Client) GetAutomations(ctx context.Context) ([]domain.Automation, error) {
	log := logger.LoggerFromContext(ctx)

	if c.baseURL == "" || c.token == "" {
		log.Info("HA URL or token unconfigured, using fallback automations")
		return copyFallbackAutomations(), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/states", nil)
	if err != nil {
		log.WithError(err).Error("failed to create HA states request for automations")
		return copyFallbackAutomations(), nil
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("HA API unreachable, falling back to mock automations")
		return copyFallbackAutomations(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status_code", resp.StatusCode).Warn("HA API returned non-200 for automations, falling back to mock automations")
		return copyFallbackAutomations(), nil
	}

	var rawStates []haEntityState
	if err := json.NewDecoder(resp.Body).Decode(&rawStates); err != nil {
		log.WithError(err).Error("failed to decode HA states JSON")
		return copyFallbackAutomations(), nil
	}

	var automations []domain.Automation
	for _, raw := range rawStates {
		if !strings.HasPrefix(raw.EntityID, "automation.") {
			continue
		}
		automations = append(automations, parseAutomation(raw))
	}

	if automations == nil {
		automations = []domain.Automation{}
	}
	return automations, nil
}

// GetAutomation fetches a single automation from Home Assistant.
func (c *Client) GetAutomation(ctx context.Context, entityID string) (domain.Automation, error) {
	log := logger.LoggerFromContext(ctx)

	if !strings.HasPrefix(entityID, "automation.") {
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("invalid automation id: %s", entityID))
	}

	if c.baseURL == "" || c.token == "" {
		fallbackMu.RLock()
		defer fallbackMu.RUnlock()
		for _, a := range fallbackAutomations {
			if a.ID == entityID {
				return a, nil
			}
		}
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("automation %s not found in fallback automations", entityID))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/states/"+entityID, nil)
	if err != nil {
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Warn("HA API automation fetch failed, trying fallback")
		fallbackMu.RLock()
		defer fallbackMu.RUnlock()
		for _, a := range fallbackAutomations {
			if a.ID == entityID {
				return a, nil
			}
		}
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("automation %s not found on HA", entityID))
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("HA returned status %d for automation %s", resp.StatusCode, entityID))
	}

	var raw haEntityState
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return domain.Automation{}, errors.Join(domain.ErrAutomationNotFound, err)
	}

	return parseAutomation(raw), nil
}

// TriggerAutomation triggers an automation via POST /api/services/automation/trigger.
func (c *Client) TriggerAutomation(ctx context.Context, entityID string) error {
	if !strings.HasPrefix(entityID, "automation.") {
		return errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("invalid automation entity id: %s", entityID))
	}
	return c.CallService(ctx, "automation", "trigger", entityID)
}

func parseAutomation(raw haEntityState) domain.Automation {
	name := raw.EntityID
	if fn, ok := raw.Attributes["friendly_name"].(string); ok && strings.TrimSpace(fn) != "" {
		name = fn
	}

	current := 0
	if currVal, ok := raw.Attributes["current"].(float64); ok {
		current = int(currVal)
	} else if currInt, ok := raw.Attributes["current"].(int); ok {
		current = currInt
	}

	var lastTriggered *time.Time
	if ltStr, ok := raw.Attributes["last_triggered"].(string); ok && ltStr != "" && ltStr != "null" {
		if t, err := time.Parse(time.RFC3339, ltStr); err == nil {
			utc := t.UTC()
			lastTriggered = &utc
		} else if t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", ltStr); err == nil {
			utc := t.UTC()
			lastTriggered = &utc
		}
	}

	return domain.Automation{
		ID:            raw.EntityID,
		Name:          name,
		State:         raw.State,
		Current:       current,
		LastTriggered: lastTriggered,
	}
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
			fallbackDevices[i].LastUpdated = time.Now().UTC()
			log.WithField("entity_id", entityID).WithField("state", fallbackDevices[i].State).Info("fallback device state updated")
			break
		}
	}

	if !found {
		return errors.Join(domain.ErrDeviceNotFound, fmt.Errorf("entity %s not found in fallback devices", entityID))
	}
	return nil
}

func (c *Client) mutateFallbackAutomation(ctx context.Context, entityID string) error {
	log := logger.LoggerFromContext(ctx)

	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	found := false
	now := time.Now().UTC()
	for i := range fallbackAutomations {
		if fallbackAutomations[i].ID == entityID {
			found = true
			fallbackAutomations[i].LastTriggered = &now
			log.WithField("automation_id", entityID).Info("fallback automation triggered")
			break
		}
	}

	if !found {
		return errors.Join(domain.ErrAutomationNotFound, fmt.Errorf("automation %s not found in fallback automations", entityID))
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

func copyFallbackAutomations() []domain.Automation {
	fallbackMu.RLock()
	defer fallbackMu.RUnlock()
	res := make([]domain.Automation, len(fallbackAutomations))
	copy(res, fallbackAutomations)
	return res
}
