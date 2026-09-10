package queries_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JLugagne/walldash/internal/dashboard/domain"
	weathertest "github.com/JLugagne/walldash/internal/dashboard/domain/service/weather/weathertest"
	"github.com/JLugagne/walldash/internal/dashboard/inbound"
	"github.com/JLugagne/walldash/internal/dashboard/inbound/queries"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeatherConfigHandler(t *testing.T) {
	controller := inbound.NewController()

	t.Run("returns configured location and units", func(t *testing.T) {
		mockQueries := &weathertest.MockWeatherQueries{
			GetWeatherConfigFunc: func(ctx context.Context) (domain.HomeConfig, error) {
				return domain.HomeConfig{
					Configured:      true,
					Latitude:        48.8566,
					Longitude:       2.3522,
					LocationName:    "Paris",
					TemperatureUnit: "°C",
					LengthUnit:      "km",
				}, nil
			},
		}
		router := mux.NewRouter()
		queries.SetupWeatherRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/weather/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Status string                             `json:"status"`
			Data   pkgdashboard.WeatherConfigResponse `json:"data"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "success", resp.Status)
		assert.True(t, resp.Data.Configured)
		assert.InDelta(t, 48.8566, resp.Data.Latitude, 0.0001)
		assert.Equal(t, "Paris", resp.Data.Name)
		assert.Equal(t, "celsius", resp.Data.TemperatureUnit)
		assert.Equal(t, "kmh", resp.Data.WindUnit)
	})

	t.Run("surfaces repository errors", func(t *testing.T) {
		mockQueries := &weathertest.MockWeatherQueries{
			GetWeatherConfigFunc: func(ctx context.Context) (domain.HomeConfig, error) {
				return domain.HomeConfig{}, errors.Join(domain.ErrDatabaseUnavailable, errors.New("boom"))
			},
		}
		router := mux.NewRouter()
		queries.SetupWeatherRoutes(router, controller, mockQueries)

		req := httptest.NewRequest(http.MethodGet, "/api/weather/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.GreaterOrEqual(t, rec.Code, http.StatusBadRequest)
	})
}
