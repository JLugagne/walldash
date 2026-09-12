package domain

import (
	"errors"
	"strings"
)

// Allowed action constants according to ADR 0002 Action Whitelist
const (
	ActionToggle  = "toggle"
	ActionTurnOn  = "turn_on"
	ActionTurnOff = "turn_off"
	ActionTrigger = "trigger"
)

// DomainAutomation defines the Home Assistant automation domain.
const DomainAutomation = "automation"

// AllowedActions defines the strict whitelist of permissible user actions.
var AllowedActions = []string{
	ActionToggle,
	ActionTurnOn,
	ActionTurnOff,
	ActionTrigger,
}

// AllowedActionDomains defines the domains that accept actions.
var AllowedActionDomains = []string{
	DomainLight,
	DomainSwitch,
	DomainMediaPlayer,
	DomainAutomation,
}

// IsAllowedAction returns true if the action string is strictly allowed by the whitelist.
func IsAllowedAction(action string) bool {
	for _, a := range AllowedActions {
		if a == action {
			return true
		}
	}
	return false
}

// IsAllowedActionDomain returns true if the domain is permitted to receive user actions.
func IsAllowedActionDomain(domain string) bool {
	for _, d := range AllowedActionDomains {
		if d == domain {
			return true
		}
	}
	return false
}

// ActionCommand specifies an action to execute against a specific device or automation entity.
type ActionCommand struct {
	EntityID string
	Action   string
}

// Validate ensures the action and target entity conform to the strict Action Whitelist.
func (c ActionCommand) Validate() error {
	entityID := strings.TrimSpace(c.EntityID)
	if entityID == "" {
		return errors.Join(ErrActionNotAllowed, errors.New("entity_id cannot be empty"))
	}

	if !isSafeEntityID(entityID) {
		return errors.Join(ErrActionNotAllowed, errors.New("malformed entity_id: "+entityID))
	}
	parts := strings.Split(entityID, ".")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return errors.Join(ErrActionNotAllowed, errors.New("malformed entity_id: "+entityID))
	}

	domainStr := parts[0]
	if !IsAllowedActionDomain(domainStr) {
		return errors.Join(ErrActionNotAllowed, errors.New("domain "+domainStr+" is not permitted by action whitelist"))
	}

	if !IsAllowedAction(c.Action) {
		return errors.Join(ErrActionNotAllowed, errors.New("action "+c.Action+" is not permitted by whitelist"))
	}

	// Domain-specific action restrictions per ADR 0002
	if domainStr == DomainAutomation {
		if c.Action != ActionTrigger {
			return errors.Join(ErrActionNotAllowed, errors.New("automation domain only permits trigger action"))
		}
	} else {
		if c.Action == ActionTrigger {
			return errors.Join(ErrActionNotAllowed, errors.New("device domains do not permit trigger action"))
		}
	}

	return nil
}

// DeviceBroadcaster represents an outbound broadcaster capable of publishing device updates.
type DeviceBroadcaster interface {
	BroadcastDevice(device Device)
}

func isSafeEntityID(s string) bool {
	if strings.Count(s, ".") != 1 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_':
		case r == '.':
		default:
			return false
		}
	}
	return true
}
