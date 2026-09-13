package dashboard

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDashboardBackgroundImageIsValidated is the regression test for the audit finding
// "Dashboard background_image is stored unvalidated and interpolated into a CSS url() value".
// The value is rendered inside url(...) in the browser, so it must be one of the bundled
// backgrounds — a path the application ships — or empty. Remote URLs and CSS fragments are refused
// on every write path, including a restored backup.
func TestDashboardBackgroundImageIsValidated(t *testing.T) {
	ctx, dash, _, owner := setupAuthServer(t)
	_, role := enrollClient(t, ctx, dash, owner)
	require.Equal(t, "owner", role)

	dashboardID := defaultDashboardID(t, owner)

	hostile := []string{
		"x); background: red; position: fixed",
		"javascript:alert(1)",
		"https://evil.example/track.png",
		"/backgrounds/../../etc/passwd",
		"data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=",
	}

	for _, value := range hostile {
		resp := owner.do(http.MethodPut, "/api/dashboards/"+dashboardID, map[string]any{
			"name": "Home", "background_image": value,
		}, nil)
		body := readBody(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"background_image %q must be refused, got %d: %s", value, resp.StatusCode, body)
		require.Contains(t, body, "INVALID_DASHBOARD")
	}

	// The bundled sample still works, and so does clearing the background.
	for _, value := range []string{"/backgrounds/meadow.jpg", ""} {
		resp := owner.do(http.MethodPut, "/api/dashboards/"+dashboardID, map[string]any{
			"name": "Home", "background_image": value,
		}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode,
			"background_image %q must be accepted: %s", value, readBody(t, resp))
	}

	// A restored backup cannot smuggle the value back in either.
	snapshot := exportedSnapshot(t, owner)
	findDashboard(t, snapshot, "Home")["background_image"] = "https://evil.example/track.png"
	code, body := restoreSnapshot(t, owner, snapshot)
	require.Equal(t, http.StatusBadRequest, code,
		"a restore must apply the same background rule, got %d: %s", code, body)
}

// defaultDashboardID returns the id of the dashboard the application ships with.
func defaultDashboardID(t *testing.T, owner *authTestClient) string {
	t.Helper()
	resp := owner.do(http.MethodGet, "/api/dashboards", nil, nil)
	body := readBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &out))
	require.NotEmpty(t, out.Data)
	return out.Data[0].ID
}
