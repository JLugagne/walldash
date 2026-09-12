package dashboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JLugagne/egauth/tokens/basic"
	dashboard "github.com/JLugagne/walldash/internal/dashboard"
	pkgdashboard "github.com/JLugagne/walldash/pkg/dashboard"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardNew(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	router := mux.NewRouter()
	conf := dashboard.Config{
		DBPath:         dbPath,
		Version:        "0.1.0-test",
		AllowedOrigins: []string{"http://localhost:8080"},
	}

	dash, err := dashboard.New(ctx, conf, router)
	require.NoError(t, err)
	require.NotNil(t, dash)

	pair, err := dash.Issuer.IssueTokenPair(ctx, basic.Claims{Subject: uuid.New(), Scopes: []string{"setup:manage"}})
	require.NoError(t, err)
	authCookie := &http.Cookie{Name: dash.Cookies.AccessName, Value: pair.AccessToken, Path: "/"}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.AddCookie(authCookie)
		router.ServeHTTP(w, r)
	})
	defer func() {
		_ = dash.Close()
	}()

	// 1. Verify health endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. Obtain CSRF token via GET /api/csrf-token
	req = httptest.NewRequest(http.MethodGet, "/api/csrf-token", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var csrfResp struct {
		Status string                         `json:"status"`
		Data   pkgdashboard.CSRFTokenResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&csrfResp)
	require.NoError(t, err)
	csrfToken := csrfResp.Data.CSRFToken
	require.NotEmpty(t, csrfToken)

	// 3. Verify CORS preflight on /api/levels
	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/levels", nil)
	preflightReq.Header.Set("Origin", "http://localhost:8080")
	preflightReq.Header.Set("Access-Control-Request-Method", "POST")
	preflightRec := httptest.NewRecorder()
	handler.ServeHTTP(preflightRec, preflightReq)
	assert.Equal(t, http.StatusNoContent, preflightRec.Code)
	assert.NotEmpty(t, preflightRec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, preflightRec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, preflightRec.Header().Get("Access-Control-Allow-Headers"), "X-CSRF-Token")

	// 4. Verify CSRF rejection on state-changing method without CSRF header
	createPayload := pkgdashboard.CreateLevelRequest{
		Name:      "Ground Floor",
		IsOutdoor: false,
	}
	body, _ := json.Marshal(createPayload)
	unauthReq := httptest.NewRequest(http.MethodPost, "/api/levels", bytes.NewReader(body))
	unauthReq.Header.Set("Origin", "http://evil.example")
	unauthRec := httptest.NewRecorder()
	handler.ServeHTTP(unauthRec, unauthReq)
	assert.Equal(t, http.StatusForbidden, unauthRec.Code, "cross-origin mutation must be rejected")

	// 5. Create a Level via POST /api/levels with valid X-CSRF-Token
	req = httptest.NewRequest(http.MethodPost, "/api/levels", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Origin", "http://localhost:8080")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var createResp struct {
		Status string                     `json:"status"`
		Data   pkgdashboard.LevelResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&createResp)
	require.NoError(t, err)
	assert.Equal(t, "Ground Floor", createResp.Data.Name)
	levelID := createResp.Data.ID
	assert.NotEmpty(t, levelID)

	// 6. List levels via GET /api/levels
	req = httptest.NewRequest(http.MethodGet, "/api/levels", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var listResp struct {
		Status string                       `json:"status"`
		Data   []pkgdashboard.LevelResponse `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&listResp)
	require.NoError(t, err)
	assert.Len(t, listResp.Data, 1)

	// 7. Save Plan via PUT /api/levels/{id}/plan with custom header X-Requested-With
	savePlanPayload := pkgdashboard.SavePlanRequest{
		Walls: []pkgdashboard.WallSegmentDTO{
			{ID: "w-test", X1: 0, Y1: 0, X2: 100, Y2: 0, Thickness: 10},
		},
		Zones: []pkgdashboard.ZoneDTO{
			{
				ID:    "z-test",
				Name:  "Living Room",
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
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "http://localhost:8080")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// 8. Get Plan via GET /api/levels/{id}/plan
	req = httptest.NewRequest(http.MethodGet, "/api/levels/"+levelID+"/plan", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
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

	// 9. Embedded Frontend Serving and SPA fallback
	t.Run("embedded frontend serves index.html at root", func(t *testing.T) {
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "<div id=\"root\"></div>")
	})

	t.Run("embedded frontend falls back to index.html for client route", func(t *testing.T) {
		req = httptest.NewRequest(http.MethodGet, "/levels/view/123", nil)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "<div id=\"root\"></div>")
	})

	t.Run("unknown api route returns 404 not SPA index.html", func(t *testing.T) {
		req = httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	// 10. Test WebSocket /api/ws integration
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws"
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer conn.Close()

	// Receive connected handshake
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	var connectMsg map[string]any
	err = json.Unmarshal(msg, &connectMsg)
	require.NoError(t, err)
	assert.Equal(t, "connected", connectMsg["type"])

	// Send toggle action
	actionPayload := map[string]string{
		"type":      "action",
		"entity_id": "light.salon_plafond",
		"action":    "toggle",
	}
	err = conn.WriteJSON(actionPayload)
	require.NoError(t, err)

	// Read response (action_success or state_changed)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg1, err := conn.ReadMessage()
	require.NoError(t, err)
	var event1 map[string]any
	err = json.Unmarshal(msg1, &event1)
	require.NoError(t, err)
	assert.NotEmpty(t, event1["type"])
}
