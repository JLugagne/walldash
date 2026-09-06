package queries_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/ha-dash/internal/dashboard/inbound"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/middleware"
	"github.com/JLugagne/ha-dash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFHandler_GetToken(t *testing.T) {
	controller := inbound.NewController()
	tokenManager := middleware.NewCSRFTokenManager()
	handler := queries.NewCSRFHandler(controller, tokenManager)

	req := httptest.NewRequest(http.MethodGet, "/api/csrf-token", nil)
	rec := httptest.NewRecorder()

	handler.GetToken(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Status string                         `json:"status"`
		Data   pkgdashboard.CSRFTokenResponse `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "success", resp.Status)
	assert.NotEmpty(t, resp.Data.CSRFToken)
	assert.True(t, tokenManager.ValidateToken(resp.Data.CSRFToken))
}
