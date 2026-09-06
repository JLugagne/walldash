package domain_test

import (
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowedActions(t *testing.T) {
	allowed := []string{
		domain.ActionToggle,
		domain.ActionTurnOn,
		domain.ActionTurnOff,
		domain.ActionTrigger,
	}

	for _, a := range allowed {
		t.Run("allowed action: "+a, func(t *testing.T) {
			assert.True(t, domain.IsAllowedAction(a))
		})
	}

	disallowed := []string{"", "dim", "set_color", "press", "delete"}
	for _, a := range disallowed {
		t.Run("disallowed action: "+a, func(t *testing.T) {
			assert.False(t, domain.IsAllowedAction(a))
		})
	}
}

func TestAllowedActionDomains(t *testing.T) {
	allowed := []string{
		domain.DomainLight,
		domain.DomainSwitch,
		domain.DomainMediaPlayer,
		domain.DomainAutomation,
	}

	for _, d := range allowed {
		t.Run("allowed action domain: "+d, func(t *testing.T) {
			assert.True(t, domain.IsAllowedActionDomain(d))
		})
	}

	disallowed := []string{"sensor", "climate", "camera", "vacuum"}
	for _, d := range disallowed {
		t.Run("disallowed action domain: "+d, func(t *testing.T) {
			assert.False(t, domain.IsAllowedActionDomain(d))
		})
	}
}

func TestActionCommandValidation(t *testing.T) {
	t.Run("valid action commands succeed", func(t *testing.T) {
		validCmds := []domain.ActionCommand{
			{EntityID: "light.salon_plafond", Action: "toggle"},
			{EntityID: "light.cuisine_spot", Action: "turn_on"},
			{EntityID: "switch.machine_a_cafe", Action: "turn_off"},
			{EntityID: "media_player.enceinte_salon", Action: "toggle"},
			{EntityID: "automation.eteindre_tout", Action: "trigger"},
		}

		for _, cmd := range validCmds {
			require.NoError(t, cmd.Validate())
		}
	})

	t.Run("disallowed actions return ErrActionNotAllowed", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "light.salon_plafond",
			Action:   "set_brightness",
		}
		err := cmd.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("trigger on device domain returns ErrActionNotAllowed", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "light.salon_plafond",
			Action:   "trigger",
		}
		err := cmd.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
	})

	t.Run("toggle on automation domain returns ErrActionNotAllowed", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "automation.eteindre_tout",
			Action:   "toggle",
		}
		err := cmd.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
	})

	t.Run("disallowed domains return ErrActionNotAllowed", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "sensor.temperature_salon",
			Action:   "toggle",
		}
		err := cmd.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		assert.True(t, domain.IsDomainError(err))
	})

	t.Run("climate domain not in action whitelist", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "climate.thermostat_salon",
			Action:   "turn_on",
		}
		err := cmd.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
	})

	t.Run("malformed entity ID returns ErrActionNotAllowed", func(t *testing.T) {
		badIDs := []string{"", "   ", "invalid_id_without_dot", ".light"}
		for _, id := range badIDs {
			cmd := domain.ActionCommand{
				EntityID: id,
				Action:   "toggle",
			}
			err := cmd.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrActionNotAllowed)
		}
	})
}
