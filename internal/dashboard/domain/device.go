package domain

import (
	"errors"
	"strings"
	"time"
)

// Supported Home Assistant domain constants
const (
	DomainLight       = "light"
	DomainSwitch      = "switch"
	DomainMediaPlayer = "media_player"
	DomainSensor      = "sensor"
	DomainClimate     = "climate"
)

// SupportedDomains list of domain strings supported by the dashboard
var SupportedDomains = []string{
	DomainLight,
	DomainSwitch,
	DomainMediaPlayer,
	DomainSensor,
	DomainClimate,
}

// IsSupportedDomain checks if a given domain string is supported
func IsSupportedDomain(domain string) bool {
	for _, d := range SupportedDomains {
		if d == domain {
			return true
		}
	}
	return false
}

// Device represents an entity retrieved from Home Assistant.
// LastUpdated is the instant Home Assistant last wrote this state, zero when unknown.
// Home Assistant refreshes it when an entity becomes unavailable, so freshness alone
// never proves the value is trustworthy: the state must be checked too.
type Device struct {
	ID          string
	Name        string
	Domain      string
	State       string
	Attributes  map[string]any
	LastUpdated time.Time
}

// Validate ensures the Device has a valid ID and supported domain.
func (d Device) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return errors.Join(ErrInvalidDevice, errors.New("device id cannot be empty"))
	}
	if !IsSupportedDomain(d.Domain) {
		return errors.Join(ErrUnsupportedDomain, errors.New("unsupported device domain: "+d.Domain))
	}
	return nil
}

// DefaultPlacementLayer defines the default layer assigned to a placed device.
const DefaultPlacementLayer = "controls"

// DevicePlacement represents the 2D placement of a device on a specific level's plan.
type DevicePlacement struct {
	ID           string
	LevelID      string
	DeviceID     string
	X            float64
	Y            float64
	Icon         string
	RenderDomain string
	CustomName   string
	Layer        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate ensures the DevicePlacement is well-formed.
func (p DevicePlacement) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.Join(ErrInvalidPlacement, errors.New("placement id cannot be empty"))
	}
	if strings.TrimSpace(p.LevelID) == "" {
		return errors.Join(ErrInvalidPlacement, errors.New("placement level_id cannot be empty"))
	}
	if strings.TrimSpace(p.DeviceID) == "" {
		return errors.Join(ErrInvalidPlacement, errors.New("placement device_id cannot be empty"))
	}
	if p.X < 0 || p.Y < 0 {
		return errors.Join(ErrInvalidPlacement, errors.New("placement coordinates must be non-negative"))
	}
	if p.RenderDomain != "" && !IsSupportedDomain(p.RenderDomain) {
		return errors.Join(ErrInvalidPlacement, errors.New("placement render_domain must be a supported domain"))
	}
	return nil
}
