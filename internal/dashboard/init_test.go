package dashboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	dashboard "github.com/JLugagne/ha-dash/internal/dashboard"
	pkgdashboard "github.com/JLugagne/ha-dash/pkg/dashboard"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardNew(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	router := mux.NewRouter()
	conf := dashboard.Config{
		DBPath:  dbPath,
		Version: "0.1.0-test",
	}

	dash, err := dashboard.New(ctx, conf, router)
	require.NoError(t, err)
	require.NotNil(t, dash)
	defer func() {
		_ = dash.Close()
	}()

	// 1. Verify health endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. Create a Level via POST /api/levels
	createPayload := pkgdashboard.CreateLevelRequest{
		Name:      "Rez-de-chaussée",
		IsOutdoor: false,
	}
	body, _ := json.Marshal(createPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/levels", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var createResp struct {
		Status string                     `json:"status"`
		Data   pkgdashboard.LevelResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&createResp)
	require.NoError(t, err)
	assert.Equal(t, "Rez-de-chaussée", createResp.Data.Name)
	levelID := createResp.Data.ID
	assert.NotEmpty(t, levelID)

	// 3. List levels via GET /api/levels
	req = httptest.NewRequest(http.MethodGet, "/api/levels", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var listResp struct {
		Status string                       `json:"status"`
		Data   []pkgdashboard.LevelResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&listResp)
	require.NoError(t, err)
	assert.Len(t, listResp.Data, 1)

	// 4. Save Plan via PUT /api/levels/{id}/plan
	savePlanPayload := pkgdashboard.SavePlanRequest{
		Walls: []pkgdashboard.WallSegmentDTO{
			{ID: "w-test", X1: 0, Y1: 0, X2: 100, Y2: 0, Thickness: 10},
		},
		Zones: []pkgdashboard.ZoneDTO{
			{
				ID:    "z-test",
				Name:  "Séjour",
				Color: "#3b82f6",
				Points: []pkgdashboard.Point2DDTO{
					{X: 0, Y: 0},
					{X: 100, Y: 0},
					{X: 100, Y: 100},
				},
			},
		},
	}
	planBody, _ := json.Marshal(savePlanPayload)
	req = httptest.NewRequest(http.MethodPut, "/api/levels/"+levelID+"/plan", bytes.NewReader(planBody))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// 5. Get Plan via GET /api/levels/{id}/plan
	req = httptest.NewRequest(http.MethodGet, "/api/levels/"+levelID+"/plan", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var planResp struct {
		Status string                    `json:"status"`
		Data   pkgdashboard.PlanResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&planResp)
	require.NoError(t, err)
	assert.Equal(t, levelID, planResp.Data.LevelID)
	assert.Len(t, planResp.Data.Walls, 1)
	assert.Len(t, planResp.Data.Zones, 1)
}
