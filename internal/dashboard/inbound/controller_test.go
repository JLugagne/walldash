package inbound_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestController(t *testing.T) {
	c := inbound.NewController()

	t.Run("SendSuccess writes 200 with success status and data", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		c.SendSuccess(rec, req, map[string]string{"foo": "bar"})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &body)
		require.NoError(t, err)
		assert.Equal(t, "success", body["status"])
		data := body["data"].(map[string]any)
		assert.Equal(t, "bar", data["foo"])
	})

	t.Run("SendFail writes 400 with fail status and error details", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)

		c.SendFail(rec, req, map[string]string{"field": "name"}, domain.ErrInvalidRequest)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &body)
		require.NoError(t, err)
		assert.Equal(t, "fail", body["status"])
	})

	t.Run("SendError writes 500 or error status with error message", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		c.SendError(rec, req, errors.New("unexpected crash"))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &body)
		require.NoError(t, err)
		assert.Equal(t, "error", body["status"])
		assert.Equal(t, "internal error", body["message"])
		assert.NotContains(t, body["message"], "unexpected crash")
	})
}
