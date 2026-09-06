package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain/service/actions/actionstest"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/commands"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestActionsHandler_ExecuteAction(t *testing.T) {
	controller := inbound.NewController()

	t.Run("valid action returns HTTP 200 and success", func(t *testing.T) {
		var executedCmd domain.ActionCommand
		mockCommands := &actionstest.MockActionCommands{
			ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
				executedCmd = cmd
				return nil
			},
		}

		router := mux.NewRouter()
		commands.SetupActionRoutes(router, controller, mockCommands)

		reqBody, _ := json.Marshal(pkgdashboard.ActionRequest{
			EntityID: "light.salon_plafond",
			Action:   "toggle",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/actions", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "light.salon_plafond", executedCmd.EntityID)
		assert.Equal(t, "toggle", executedCmd.Action)
	})

	t.Run("action violating whitelist returns 400 fail", func(t *testing.T) {
		mockCommands := &actionstest.MockActionCommands{
			ExecuteActionFunc: func(ctx context.Context, actor domain.Actor, cmd domain.ActionCommand) error {
				return domain.ErrActionNotAllowed
			},
		}

		router := mux.NewRouter()
		commands.SetupActionRoutes(router, controller, mockCommands)

		reqBody, _ := json.Marshal(pkgdashboard.ActionRequest{
			EntityID: "sensor.temp",
			Action:   "toggle",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/actions", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
