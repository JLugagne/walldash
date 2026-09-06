package commands_test

import (
	"testing"

	svchealthtest "github.com/JLugagne/walldash/internal/dashboard/domain/service/health/healthtest"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/commands"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	router := mux.NewRouter()
	c := inbound.NewController()
	mockCommands := &svchealthtest.MockHealthCommands{}

	assert.NotPanics(t, func() {
		commands.SetupRoutes(router, c, mockCommands)
	})
}
