package converters_test

import (
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/domain"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/converters"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestActionConverters(t *testing.T) {
	t.Run("ToDomainAction converts public request to domain action command", func(t *testing.T) {
		req := pkgdashboard.ActionRequest{
			EntityID: "light.salon_plafond",
			Action:   "toggle",
		}

		cmd := converters.ToDomainAction(req)
		assert.Equal(t, "light.salon_plafond", cmd.EntityID)
		assert.Equal(t, "toggle", cmd.Action)
	})

	t.Run("ToPublicAction converts domain command to public response", func(t *testing.T) {
		cmd := domain.ActionCommand{
			EntityID: "switch.coffee_maker",
			Action:   "turn_on",
		}

		resp := converters.ToPublicAction(cmd)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "switch.coffee_maker", resp.EntityID)
		assert.Equal(t, "turn_on", resp.Action)
	})
}
